package security

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	OTPHMACKey string `env:"OTP_HMAC_KEY"`
}

func NewConfig() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse security config: %w", err)
	}
	return cfg, nil
}

func NewConfigMust() Config {
	cfg, err := NewConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to create security config: %v", err))
	}
	return cfg
}
