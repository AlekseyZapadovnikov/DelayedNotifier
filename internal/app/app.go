package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/config"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/cache"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/repository/postgres"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/service/notify"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/service/rabbit"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/service/sender"
	mailsend "github.com/AlekseyZapadovnikov/DelayedNotifier/internal/service/sender/mailsender"
	tgsend "github.com/AlekseyZapadovnikov/DelayedNotifier/internal/service/sender/tgsender"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/web"
)

type Closer interface {
	Close() error
}

type App struct {
	cfg        *config.Config
	httpServer *web.Server
	notifier   *notify.Notifier
	closers    []Closer
}

func New(cfg *config.Config) (*App, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	repo := postgres.NewRecordRepo(&cfg.PostgresConfig)
	cache, err := cache.New(context.Background(), cfg.RedisConfig, repo)
	if err != nil {
		return nil, fmt.Errorf("init cache: %w", err)
	}

	mailSender, err := mailsend.NewEmailSender(cfg.EmailSenderConfig)
	if err != nil {
		return nil, fmt.Errorf("init email sender: %w", err)
	}

	var telegramSender sender.Sender
	if strings.TrimSpace(cfg.TelegramSenderConfig.BotToken) != "" {
		telegramSender, err = tgsend.NewTgSender(cfg.TelegramSenderConfig)
		if err != nil {
			return nil, fmt.Errorf("init telegram sender: %w", err)
		}
	} else {
		slog.Warn("telegram sender disabled: TG_SENDER_BOT_TOKEN is empty")
	}
	rabbitClient := rabbit.NewRabbitClient(&cfg.Rabbit)
	sendWrap := sender.NewSenderWrap(telegramSender, mailSender)

	notifyService := notify.NewNotifier(sendWrap, cache, rabbitClient, cfg.Notifier)
	httpServer := web.NewServer(&cfg.HTTP, notifyService)
	httpServer.SetDefaultFrom(cfg.EmailSenderConfig.SMTPUser)

	return &App{
		cfg:        cfg,
		httpServer: httpServer,
		notifier:   notifyService,
		closers:    []Closer{cache, rabbitClient},
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	if a == nil {
		return errors.New("app is nil")
	}
	if a.httpServer == nil {
		return errors.New("http server is nil")
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	serverErrCh := make(chan error, 1)

	go func() {
		err := a.httpServer.Start()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			return
		}
		serverErrCh <- nil
	}()

	go func() {
		if a.notifier == nil {
			slog.Warn("notification worker disabled: notifier is nil")
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			err := a.notifier.SendMessageWorkerContext(ctx)
			if err == nil || ctx.Err() != nil {
				return
			}

			slog.Error("notification worker failed; retrying",
				"error", err,
			)

			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}()

	slog.Info("app started")

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-serverErrCh:
		if err != nil {
			return fmt.Errorf("http server failed: %w", err)
		}
		return nil
	}

	if err := a.gracefulShutdown(); err != nil {
		return err
	}

	if err := <-serverErrCh; err != nil {
		return fmt.Errorf("http server failed during shutdown: %w", err)
	}

	slog.Info("app stopped")
	return nil
}

func (a *App) gracefulShutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var shutdownErr error

	if err := a.httpServer.Shutdown(ctx); err != nil {
		shutdownErr = errors.Join(shutdownErr, fmt.Errorf("shutdown http server: %w", err))
	}

	for i := len(a.closers) - 1; i >= 0; i-- {
		if err := a.closers[i].Close(); err != nil {
			shutdownErr = errors.Join(shutdownErr, err)
		}
	}

	return shutdownErr
}
