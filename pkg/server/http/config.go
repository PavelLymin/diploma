package httpserver

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Address           string        `env:"HTTP_ADDRESS" envDefault:":8080"`
	ReadTimeout       time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"5s"`
	ReadHeaderTimeout time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" envDefault:"5s"`
	WriteTimeout      time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"10s"`
	ShutdownTimeout   time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"15s"`
	IdleTimeout       time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
	MaxHeaderBytes    int           `env:"HTTP_MAX_HEADER_BYTES" envDefault:"1048576"`
}

func NewConfig() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse server config: %w", err)
	}
	return cfg, nil
}

func NewConfigMust() Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to create server config: %v", err))
	}
	return cfg
}
