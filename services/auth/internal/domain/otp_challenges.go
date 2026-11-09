package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/PavelLymin/diploma/pkg/phone"
	"github.com/google/uuid"
)

type OTPChallenge struct {
	ID                uuid.UUID  // Уникальный ID challenge
	Phone             string     // Нормализованный номер телефона
	CodeHash          string     // HMAC-хеш OTP
	Attempts          int        // Число неправильных попыток
	CreatedAt         time.Time  // Время создания
	ExpiresAt         time.Time  // Время истечения
	ResendAvailableAt time.Time  // Когда разрешена повторная отправка
	BlockedUntil      *time.Time // Когда заканчивается блокировка
}

func NewOTPChallenge(
	number,
	codeHash string,
	ttl,
	resendCooldown time.Duration,
	now time.Time,
) (*OTPChallenge, error) {
	if strings.TrimSpace(codeHash) == "" {
		return nil, fmt.Errorf("code hash: %w", ErrInvalidChallenge)
	}
	if ttl <= 0 || resendCooldown < 0 || resendCooldown >= ttl {
		return nil, fmt.Errorf("invalid challenge TTL or resend cooldown: %w", ErrInvalidChallenge)
	}
	normalizedPhone, err := phone.NormalizePhone(number, "ZZ")
	if err != nil {
		return nil, fmt.Errorf("normalize phone: %w", err)
	}

	now = now.UTC()
	challenge := &OTPChallenge{
		ID:                uuid.New(),
		Phone:             normalizedPhone,
		CodeHash:          codeHash,
		Attempts:          0,
		CreatedAt:         now,
		ExpiresAt:         now.Add(ttl),
		ResendAvailableAt: now.Add(resendCooldown),
	}

	if err := challenge.Validate(); err != nil {
		return nil, err
	}

	return challenge, nil
}

func (c *OTPChallenge) Validate() error {
	if c == nil || c.ID == uuid.Nil || strings.TrimSpace(c.CodeHash) == "" || c.Attempts < 0 {
		return ErrInvalidChallenge
	}
	if c.CreatedAt.IsZero() || c.ExpiresAt.IsZero() || !c.ExpiresAt.After(c.CreatedAt) {
		return fmt.Errorf("invalid challenge timestamps: %w", ErrInvalidChallenge)
	}
	if c.ResendAvailableAt.IsZero() || c.ResendAvailableAt.Before(c.CreatedAt) || !c.ResendAvailableAt.Before(c.ExpiresAt) {
		return fmt.Errorf("invalid resend timestamp: %w", ErrInvalidChallenge)
	}
	if err := phone.ValidatePhone(c.Phone, "ZZ"); err != nil {
		return fmt.Errorf("invalid challenge phone: %w", err)
	}
	return nil
}

// Проверяет, истёк ли challenge
func (c *OTPChallenge) IsExpired(now time.Time) bool {
	return c == nil || !now.UTC().Before(c.ExpiresAt)
}

// Проверяет, разрешена ли повторная отправка, когда наступило время ResendAvailableAt
func (c *OTPChallenge) CanResend(now time.Time) bool {
	return c != nil &&
		!c.IsExpired(now) &&
		!now.UTC().Before(c.ResendAvailableAt)
}

// Увеличивает число неправильных попыток
func (o *OTPChallenge) IncrementAttempt(maxAttempts int) error {
	if o.Attempts >= maxAttempts {
		return ErrMaxAttemptsExceeded
	}
	o.Attempts++
	return nil
}

// Проверяет, действует ли блокировка challenge
func (c *OTPChallenge) IsBlocked(now time.Time) bool {
	return c != nil && c.BlockedUntil != nil && now.UTC().Before(*c.BlockedUntil)
}
