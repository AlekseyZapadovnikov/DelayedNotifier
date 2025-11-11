package web

import (
	"log/slog"
	"net/http"
	"path/filepath"

	// Заменяем импорты chi на импорт вашей обертки ginext
	"github.com/wb-go/wbf/ginext" // <-- УКАЖИТЕ ПРАВИЛЬНЫЙ ПУТЬ К ПАКЕТУ ginext
)

type Configer interface {
	GetServerAddress() string
	GetStaticFilesPath() string
}

type RecordService interface{}

type Server struct {
	adress     string
	staticPath string
	httpServer *http.Server
	router     *ginext.Engine
	service    RecordService
}

func NewServer(c Configer, serv RecordService) *Server {
	adr := c.GetServerAddress()
	sp := c.GetStaticFilesPath()

	r := ginext.New("") // CHECK

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
	slog.Info("server was started by adress", "addres", s.adress)
	// Метод ListenAndServe остается стандартным, так как он часть http.Server
	return s.httpServer.ListenAndServe()
}

func (s *Server) routs() {
	router := s.router

	router.Use(ginext.Logger(), ginext.Recovery()) // Используем middleware
	router.Engine.StaticFS("/", http.Dir(s.staticPath)) // отдаём статические файлы

	// на непонятные запросы отдаём index.html
	router.Engine.NoRoute(func(c *ginext.Context) {
		slog.Info("Fallback to index.html", "requested", c.Request.URL.Path)
		c.File(filepath.Join(s.staticPath, "index.html"))
	})
}
