package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	server *http.Server
	logger *slog.Logger
	config Config
}

func New(
	cfg Config,
	handler *gin.Engine,
	logger *slog.Logger,
) (*Server, error) {
	if handler == nil {
		return nil, errors.New("HTTP handler is required")
	}

	if logger == nil {
		logger = slog.Default()
	}

	httpServer := &http.Server{
		Addr:              cfg.Address,
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
	}

	return &Server{
		server: httpServer,
		logger: logger,
		config: cfg,
	}, nil
}

func (s *Server) Run() error {
	mainCtx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	s.logger.Info("HTTP Server is starting", slog.String("address", s.server.Addr))

	g, gCtx := errgroup.WithContext(mainCtx)
	g.Go(func() error {
		err := s.server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("listen and serve: %w", err)
		}

		return nil
	})
	g.Go(func() error {
		<-gCtx.Done()

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			s.config.ShutdownTimeout,
		)
		defer cancel()

		if err := s.server.Shutdown(shutdownCtx); err != nil {
			_ = s.server.Close()

			return fmt.Errorf("graceful shutdown: %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("HTTP server failed: %w", err)
	}

	s.logger.Info("HTTP Server gracefully stopped")

	return nil
}
