package rdb

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Addr         string        `env:"REDIS_ADDR,required,notEmpty"`
	Password     string        `env:"REDIS_PASSWORD" envDefault:""`
	DB           int           `env:"REDIS_DB" envDefault:"0"`
	PoolSize     int           `env:"REDIS_POOL_SIZE" envDefault:"100"`
	MinIdleConns int           `env:"REDIS_MIN_IDLE_CONNS" envDefault:"10"`
	PoolTimeout  time.Duration `env:"REDIS_POOL_TIMEOUT" envDefault:"5s"`
	DialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT" envDefault:"5s"`
	ReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT" envDefault:"3s"`
	WriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT" envDefault:"3s"`
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
