package rdb

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Addr         string        `env:"REDIS_ADDR" required:"true"`
	Password     string        `env:"REDIS_PASSWORD" default:""`
	DB           int           `env:"REDIS_DB" default:"0"`
	PoolSize     int           `env:"REDIS_POOL_SIZE" default:"100"`
	MinIdleConns int           `env:"REDIS_MIN_IDLE_CONNS" default:"10"`
	DialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT" default:"5s"`
	ReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT" default:"3s"`
	WriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT" default:"3s"`
}

func NewConfig() (Config, error) {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse redis config: %w", err)
	}

	return cfg, nil
}

func NewConfigMust() Config {
	cfg, err := NewConfig()

	if err != nil {
		panic(fmt.Errorf("failed to create redis config: %v", err))
	}

	return cfg
}
