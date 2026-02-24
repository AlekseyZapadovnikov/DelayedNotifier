package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/config"
	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/app"
)

func main() {
	if err := run(context.Background()); err != nil {
		slog.Error("application stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}

	if err := application.Run(ctx); err != nil {
		return fmt.Errorf("run app: %w", err)
	}

	return nil
}
