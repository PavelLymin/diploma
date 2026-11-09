package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PavelLymin/diploma/services/auth/internal/domain"
	"github.com/PavelLymin/diploma/services/auth/internal/infrastructure/repository"
	"github.com/PavelLymin/diploma/services/auth/internal/infrastructure/security"
	"github.com/PavelLymin/diploma/services/auth/internal/transport/http/v1/request"
	"github.com/PavelLymin/diploma/services/auth/internal/transport/http/v1/response"

	"github.com/google/uuid"
)

type UseCase struct {
	repository repository.IOTPRepository // Хранение challenge
	crypto     *security.Crypto          // Генерация OTP и HMAC
	delivery   repository.IOTPDelivery   // Передача OTP пользователю
	clock      Clock                     // Получение времени
	config     Config                    // Настройки ограничений
}

func New(
	r repository.IOTPRepository,
	crypt *security.Crypto,
	d repository.IOTPDelivery,
	c Clock,
	cfg Config,
) (*UseCase, error) {
	if r == nil {
		return nil, errors.New("OTP repository is required")
	}
	if crypt == nil {
		return nil, errors.New("OTP crypto is required")
	}
	if d == nil {
		return nil, errors.New("OTP delivery adapter is required")
	}
	if c == nil {
		c = systemClock{}
	}
	if cfg.ChallengeTTL == 0 && cfg.ResendCooldown == 0 && cfg.BlockDuration == 0 && cfg.IdempotencyTTL == 0 && cfg.MaxAttempts == 0 {
		cfg = DefaultConfig()
	}
	if cfg.ChallengeTTL <= 0 || cfg.ResendCooldown < 0 || cfg.BlockDuration <= 0 || cfg.IdempotencyTTL <= 0 || cfg.MaxAttempts < 1 {
		return nil, errors.New("invalid OTP configuration")
	}
	if cfg.ResendCooldown >= cfg.ChallengeTTL {
		return nil, errors.New("resend cooldown must be shorter than challenge TTL")
	}
	return &UseCase{
		repository: r,
		crypto:     crypt,
		delivery:   d,
		clock:      c,
		config:     cfg,
	}, nil
}

func (uc *UseCase) CreateChallenges(
	ctx context.Context,
	req request.CreateChallengeRequest,
) (*response.ChallengeResponse, error) {
	if req.RequestID == uuid.Nil {
		return nil, domain.ErrEmptyRequestID
	}
	if !req.PersonalDataConsent {
		return nil, domain.ErrConsentRequired
	}

	code, err := uc.crypto.GenerateOTP()
	if err != nil {
		return nil, fmt.Errorf("generate OTP: %w", err)
	}
	challenge, err := domain.NewOTPChallenge(
		req.Phone,
		uc.crypto.HashToken(code),
		uc.config.ChallengeTTL,
		uc.config.ResendCooldown,
		uc.clock.Now(),
	)
	if err != nil {
		return nil, err
	}

	fingerprint := fingerprint("create", challenge.Phone, req.PersonalDataConsent)
	stored, created, err := uc.repository.CreateChallenge(
		ctx,
		req.RequestID,
		fingerprint,
		challenge,
		uc.config.ChallengeTTL,
		uc.config.IdempotencyTTL,
	)
	if err != nil {
		return nil, err
	}

	if created {
		if err := uc.delivery.SendOTP(ctx, stored.Phone, code); err != nil {
			cleanupErr := uc.repository.DeletePending(ctx, "create", req.RequestID, stored.ID, stored.Phone)
			if cleanupErr != nil {
				return nil, fmt.Errorf("OTP delivery failed (%v); challenge cleanup failed: %w", err, cleanupErr)
			}
			return nil, fmt.Errorf("OTP delivery failed: %w", err)
		}
	}
	serverTime, err := uc.repository.ServerTime(ctx)
	if err != nil {
		return nil, err
	}
	return responseFromChallenge(stored, serverTime), nil
}

func (uc *UseCase) ResendChallenge(
	ctx context.Context,
	req request.ResendChallengeRequest,
) (*response.ChallengeResponse, error) {
	if req.RequestID == uuid.Nil {
		return nil, domain.ErrEmptyRequestID
	}
	if req.ChallengeID == uuid.Nil {
		return nil, domain.ErrEmptyChallengeID
	}
	fingerprint := fingerprint("resend", req.ChallengeID.String())
	if previous, found, err := uc.repository.GetIdempotentChallenge(ctx, "resend", req.RequestID, fingerprint); err != nil {
		return nil, err
	} else if found {
		serverTime, err := uc.repository.ServerTime(ctx)
		if err != nil {
			return nil, err
		}
		return responseFromChallenge(previous, serverTime), nil
	}

	old, err := uc.repository.GetByID(ctx, req.ChallengeID)
	if err != nil {
		return nil, err
	}
	code, err := uc.crypto.GenerateOTP()
	if err != nil {
		return nil, fmt.Errorf("generate replacement OTP: %w", err)
	}
	challenge, err := domain.NewOTPChallenge(
		old.Phone,
		uc.crypto.HashToken(code),
		uc.config.ChallengeTTL,
		uc.config.ResendCooldown,
		uc.clock.Now(),
	)
	if err != nil {
		return nil, err
	}
	stored, created, err := uc.repository.ResendChallenge(
		ctx,
		req.RequestID,
		old.ID,
		fingerprint,
		challenge,
		uc.config.ChallengeTTL,
		uc.config.IdempotencyTTL,
	)
	if err != nil {
		return nil, err
	}

	if created {
		if err := uc.delivery.SendOTP(ctx, stored.Phone, code); err != nil {
			cleanupErr := uc.repository.DeletePending(ctx, "resend", req.RequestID, stored.ID, stored.Phone)
			if cleanupErr != nil {
				return nil, fmt.Errorf("OTP delivery failed (%v); replacement challenge cleanup failed: %w", err, cleanupErr)
			}
			return nil, fmt.Errorf("OTP delivery failed: %w", err)
		}
	}
	serverTime, err := uc.repository.ServerTime(ctx)
	if err != nil {
		return nil, err
	}
	return responseFromChallenge(stored, serverTime), nil
}

func (uc *UseCase) VerifyAndConsume(
	ctx context.Context,
	challengeID uuid.UUID,
	req request.VerifyRequest,
) (string, error) {
	if challengeID == uuid.Nil {
		return "", domain.ErrEmptyChallengeID
	}
	if len(req.Code) != 6 {
		return "", domain.ErrInvalidOTPFormat
	}
	for _, r := range req.Code {
		if r < '0' || r > '9' {
			return "", domain.ErrInvalidOTPFormat
		}
	}

	challenge, err := uc.repository.VerifyAndConsume(
		ctx,
		challengeID,
		uc.crypto.HashToken(req.Code),
		uc.config.MaxAttempts,
		uc.config.BlockDuration,
	)
	if err != nil {
		return "", err
	}
	return challenge.Phone, nil
}

func responseFromChallenge(
	challenge *domain.OTPChallenge,
	serverTime time.Time,
) *response.ChallengeResponse {
	return &response.ChallengeResponse{
		ChallengeID:       challenge.ID,
		SentAt:            challenge.CreatedAt.UTC(),
		ExpiresAt:         challenge.ExpiresAt.UTC(),
		ResendAvailableAt: challenge.ResendAvailableAt.UTC(),
		ServerTime:        serverTime.UTC(),
	}
}

func fingerprint(parts ...any) string {
	var value strings.Builder
	for _, part := range parts {
		value.WriteString(fmt.Sprint(part))
		value.WriteByte(0)
	}
	sum := sha256.Sum256([]byte(value.String()))
	return hex.EncodeToString(sum[:])
}
