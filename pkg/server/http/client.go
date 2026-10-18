package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
	server *http.Server
	logger *slog.Logger
	config Config
	notify chan error
}

func New(cfg Config, handler *gin.Engine, logger *slog.Logger) *Server {
	httpServer := &http.Server{
		Addr:           cfg.Address,
		Handler:        handler,
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
	}

	return &Server{
		router: handler,
		server: httpServer,
		logger: logger,
		config: cfg,
		notify: make(chan error, 1),
	}
}

func (s *Server) start() {
	s.logger.Info("HTTP Server is starting", slog.String("address", s.server.Addr))

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.notify <- err
		}
		close(s.notify)
	}()
}

func (s *Server) Run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	s.start()

	select {
	case err := <-s.notify:
		return err

	case sig := <-quit:
		s.logger.Info("Shutting down HTTP server...", slog.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
		defer cancel()

		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.Error("Server forced to shutdown", slog.String("error", err.Error()))
			_ = s.server.Close()
			return err
		}

		s.logger.Info("HTTP Server gracefully stopped")
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
