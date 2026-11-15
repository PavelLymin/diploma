package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/PavelLymin/diploma/pkg/rdb"
	httpserver "github.com/PavelLymin/diploma/pkg/server/http"
	"github.com/PavelLymin/diploma/pkg/transport/http/middleware"
	usecase "github.com/PavelLymin/diploma/services/auth/internal/application/use_case"
	"github.com/PavelLymin/diploma/services/auth/internal/infrastructure/repository"
	redisotp "github.com/PavelLymin/diploma/services/auth/internal/infrastructure/repository/redis"
	"github.com/PavelLymin/diploma/services/auth/internal/infrastructure/security"
	v1 "github.com/PavelLymin/diploma/services/auth/internal/transport/http/v1"
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

	redisClient, err := initRedis(ctx)
	if err != nil {
		return nil, err
	}
	initialized := false

	defer func() {
		if !initialized {
			_ = redisClient.Close()
		}
	}()

	otpRepository, err := redisotp.New(redisClient)
	if err != nil {
		return nil, fmt.Errorf("initialize OTP repository: %w", err)
	}

	otpDelivery := repository.New()

	otpCrypto, err := initSecurity()
	if err != nil {
		return nil, err
	}

	otpUseCase, err := usecase.New(
		otpRepository,
		otpCrypto,
		otpDelivery,
		nil,
		usecase.DefaultConfig(),
	)
	if err != nil {
		return nil, fmt.Errorf("initialize OTP use case: %w", err)
	}

	otpHandler := v1.New(otpUseCase)

	httpServer, err := initServer(otpHandler, logger)
	if err != nil {
		return nil, err
	}

	initialized = true

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

	return nil
}

func initServer(
	otpHadler *v1.OTPHandler,
	logger *slog.Logger,
) (*httpserver.Server, error) {
	cfg, err := httpserver.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("load HTTP server config: %w", err)
	}

	handler := gin.New()
	handler.Use(
		gin.Recovery(),
		middleware.RequestID(),
		middleware.SecurityHeaders(),
	)

	v1.NewRouter(handler, otpHadler)

	httpServer := httpserver.New(cfg, handler, logger)

	return httpServer, nil
}

func initRedis(ctx context.Context) (*redis.Client, error) {
	cfg, err := rdb.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("load Redis config: %w", err)
	}

	client, err := rdb.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create Redis client: %w", err)
	}

	return client, nil
}

func initSecurity() (*security.Crypto, error) {
	securityConfig, err := security.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("initialize security config: %w", err)
	}

	crypto, err := security.NewCrypto(securityConfig)
	if err != nil {
		return nil, fmt.Errorf("initialize crypto: %w", err)
	}

	return crypto, nil
}

func (a *App) Stop() {
	a.logger.Info("Closing application resources...")

	if err := a.redisClient.Close(); err != nil {
		a.logger.Error("Failed to close redis connection", slog.String("error", err.Error()))
	} else {
		a.logger.Info("Redis connection successfully closed")
	}
}
