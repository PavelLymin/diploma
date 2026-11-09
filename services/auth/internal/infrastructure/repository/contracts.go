package repository

import (
	"context"
	"time"

	"github.com/PavelLymin/diploma/services/auth/internal/domain"
	"github.com/google/uuid"
)

type IOTPRepository interface {
	CreateChallenge(
		ctx context.Context,
		requestID uuid.UUID,
		fingerprint string,
		challenge *domain.OTPChallenge,
		challengeTTL, idempotencyTTL time.Duration,
	) (*domain.OTPChallenge, bool, error)

	ResendChallenge(
		ctx context.Context,
		requestID, oldChallengeID uuid.UUID,
		fingerprint string,
		challenge *domain.OTPChallenge,
		challengeTTL, idempotencyTTL time.Duration,
	) (*domain.OTPChallenge, bool, error)

	VerifyAndConsume(
		ctx context.Context,
		challengeID uuid.UUID,
		codeHash string,
		maxAttempts int,
		blockDuration time.Duration,
	) (*domain.OTPChallenge, error)

	GetByID(
		ctx context.Context,
		challengeID uuid.UUID,
	) (*domain.OTPChallenge, error)

	GetIdempotentChallenge(
		ctx context.Context,
		operation string,
		requestID uuid.UUID,
		fingerprint string,
	) (*domain.OTPChallenge, bool, error)

	IsPhoneBlocked(
		ctx context.Context,
		normalizedPhone string,
	) (*time.Time, error)

	DeletePending(
		ctx context.Context,
		operation string,
		requestID, challengeID uuid.UUID,
		normalizedPhone string,
	) error

	ServerTime(ctx context.Context) (time.Time, error)
}

type IOTPDelivery interface {
	SendOTP(ctx context.Context, normalizedPhone, code string) error
}
