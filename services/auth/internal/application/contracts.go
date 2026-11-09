package application

import (
	"context"

	"github.com/PavelLymin/diploma/services/auth/internal/domain"
	"github.com/google/uuid"
)

type IOTPChallengeUseCase interface {
	CreateChallenges(ctx context.Context, phone string, personalDataConsent bool) (*domain.OTPChallenge, error)
	VerifyChallenges(ctx context.Context, challengeId uuid.UUID, code string) error
}
