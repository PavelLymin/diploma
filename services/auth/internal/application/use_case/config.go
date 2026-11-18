package usecase

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ChallengeTTL   time.Duration `env:"OTP_CHALLENGE_TTL" envDefault:"5m"`
	ResendCooldown time.Duration `env:"OTP_RESEND_COOLDOWN" envDefault:"60s"`
	BlockDuration  time.Duration `env:"OTP_BLOCK_DURATION" envDefault:"15m"`
	IdempotencyTTL time.Duration `env:"OTP_IDEMPOTENCY_TTL" envDefault:"24h"`
	MaxAttempts    int           `env:"OTP_MAX_ATTEMPTS" envDefault:"5"`
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
