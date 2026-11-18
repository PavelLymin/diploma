package exolve

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	APIKey  string `env:"DELIVERY_API_KEY"`
	Sender  string `env:"DELIVERY_SENDER"`
	BaseURL string `env:"DELIVERY_BASE_URL"`
}

func NewConfig() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse delivery config: %w", err)
	}
	return cfg, nil
}

func NewConfigMust() Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to create delivery config: %v", err))
	}
	return cfg
}
