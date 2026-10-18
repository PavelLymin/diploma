package app

import (
	"context"
	"log/slog"

	"github.com/PavelLymin/diploma/pkg/rdb"
	httpserver "github.com/PavelLymin/diploma/pkg/server/http"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type App struct {
	server      *httpserver.Server
	redisClient *redis.Client
	logger      *slog.Logger
}

func New() (*App, error) {
	ctx := context.Background()

	logger := slog.Default()

	httpServer := initServer(logger)

	redisClient, err := initRedis(ctx)
	if err != nil {
		return nil, err
	}

	return &App{
		server:      httpServer,
		redisClient: redisClient,
		logger:      logger,
	}, nil
}

func (a *App) Run() error {
	if err := a.server.Run(); err != nil {
		return err
	}

	a.Stop()

	return nil
}

func initServer(
	logger *slog.Logger,
) *httpserver.Server {
	handler := gin.New()
	// Initialize your routes and middleware here

	httpServer := httpserver.New(httpserver.NewConfigMust(), handler, logger)

	return httpServer
}

func initRedis(ctx context.Context) (*redis.Client, error) {
	client, err := rdb.NewClient(ctx, rdb.NewConfigMust())
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (a *App) Stop() {
	a.logger.Info("Closing application resources...")

	if err := a.redisClient.Close(); err != nil {
		a.logger.Error("Failed to close redis connection", slog.String("error", err.Error()))
	} else {
		a.logger.Info("Redis connection successfully closed")
	}
}
