package usecase

import "time"

type Config struct {
	ChallengeTTL   time.Duration
	ResendCooldown time.Duration
	BlockDuration  time.Duration
	IdempotencyTTL time.Duration
	MaxAttempts    int
}

func DefaultConfig() Config {
	return Config{
		ChallengeTTL:   5 * time.Minute,
		ResendCooldown: 60 * time.Second,
		BlockDuration:  15 * time.Minute,
		IdempotencyTTL: 24 * time.Hour,
		MaxAttempts:    5,
	}
}
