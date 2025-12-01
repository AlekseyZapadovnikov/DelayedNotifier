package web

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/AlekseyZapadovnikov/DelayedNotifier/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Configer interface {
	GetServerAddress() string
	GetStaticFilesPath() string
}

type NotifydService interface {
	CreateNotify(ctx context.Context, record models.Record) error
	GetNotifyStatByID(ctx context.Context, id int64) error
	DeleteNotifyByID(ctx context.Context, id int64) error
}

type Server struct {
	adress     string
	staticPath string
	httpServer *http.Server
	router     *chi.Mux
	service    NotifydService
}

func NewServer(c Configer, serv NotifydService) *Server {
	adr := c.GetServerAddress()
	sp := c.GetStaticFilesPath()

	// Создаём роутер с помощью chi
	r := chi.NewRouter()

	httpServ := &http.Server{
		Addr:    adr,
		Handler: r,
	}

	srv := &Server{
		adress:     adr,
		router:     r,
		service:    serv,
		httpServer: httpServ,
		staticPath: sp,
	}

	return srv
}

func (s *Server) Start() error {
	s.routs()
	slog.Info("server was started by address", "address", s.adress)
	return s.httpServer.ListenAndServe()
}

func (s *Server) routs() {
	// Используем встроенные middleware chi
	s.router.Use(middleware.Logger)    // логирование запросов
	s.router.Use(middleware.Recoverer) // восстановление после паник

	// Отдаём статические файлы
	fileServer := http.StripPrefix("/", http.FileServer(http.Dir(s.staticPath)))
	s.router.Handle("/*", fileServer)

	s.router.Post("/notify", s.createNotify)
	s.router.Get("/notify/{id}/", s.getNotifyStatByID)
	s.router.Delete("/notify/{id}/", s.deleteNotifyByID)
}