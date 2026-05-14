package server

import (
	"log/slog"
	"net/http"

	_ "github.com/OlegRozh/subscriptions-service/docs"
	"github.com/OlegRozh/subscriptions-service/internal/handlers"
	"github.com/OlegRozh/subscriptions-service/internal/storage"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	handler *handlers.Handler
	logger  *slog.Logger
	port    string
	router  *chi.Mux
}

func New(storage *storage.Storage, logger *slog.Logger, port string) *Server {
	handler := handlers.NewHandler(storage, logger)
	router := chi.NewRouter()

	return &Server{
		handler: handler,
		logger:  logger,
		port:    port,
		router:  router,
	}
}

func (s *Server) RegisterRoutes() {
	// Swagger UI
	s.router.Get("/swagger/*", httpSwagger.WrapHandler)
	s.router.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", s.handler.Create)
		r.Get("/sum", s.handler.GetSum)
		r.Get("/{id}", s.handler.Get)
		r.Put("/{id}", s.handler.Update)
		r.Delete("/{id}", s.handler.Delete)
	})
}

func (s *Server) Run() error {
	s.RegisterRoutes()
	s.logger.Info("Server started", "port", s.port)
	return http.ListenAndServe(":"+s.port, s.router)
}
