package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/OlegRozh/subscriptions-service/docs"
	"github.com/OlegRozh/subscriptions-service/internal/handlers"
	"github.com/OlegRozh/subscriptions-service/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	handler    *handlers.Handler
	logger     *slog.Logger
	port       string
	router     *chi.Mux
	httpServer *http.Server
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func New(storage *storage.Storage, logger *slog.Logger, port string) *Server {
	handler := handlers.NewHandler(storage, logger)
	router := chi.NewRouter()

	return &Server{
		handler: handler,
		logger:  logger,
		port:    port,
		router:  router,
		httpServer: &http.Server{
			Addr: ":" + port,
		},
	}
}

func (s *Server) RegisterRoutes() {
	// Swagger UI
	s.router.Get("/swagger/*", httpSwagger.WrapHandler)
	s.router.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", s.handler.Create)
		r.Get("/sum", s.handler.GetSum)
		r.Get("/{id}", s.handler.Get)
		r.Get("/", s.handler.GetList)
		r.Put("/{id}", s.handler.Update)
		r.Delete("/{id}", s.handler.Delete)
	})
}

func (s *Server) RegisterMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)
			s.logger.Info("HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", middleware.GetReqID(r.Context()),
			)
		})
	})
}

func (s *Server) Run() error {
	s.RegisterMiddleware()
	s.RegisterRoutes()
	s.httpServer.Handler = s.router

	go func() {
		s.logger.Info("starting server", "port", s.port)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("server forced to shutdown", "error", err)
		return err
	}

	s.logger.Info("server exited")
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
