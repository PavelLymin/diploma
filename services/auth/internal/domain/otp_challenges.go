package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/PavelLymin/diploma/pkg/phone"
	"github.com/google/uuid"
)

var (
	ErrEmptyChallengeID    = errors.New("otp challenge id must not be empty")
	ErrInvalidPhone        = errors.New("invalid phone number")
	ErrEmptyCodeHash       = errors.New("code hash cannot be empty")
	ErrChallengeExpired    = errors.New("otp challenge is expired")
	ErrAlreadyConsumed     = errors.New("otp challenge already consumed")
	ErrMaxAttemptsExceeded = errors.New("max attempts exceeded")
)

type OTPChallenge struct {
	ID         uuid.UUID
	Phone      string
	CodeHash   []byte
	ExpiresAt  time.Time
	Attempts   int
	ConsumedAt *time.Time
	CreatedAt  time.Time
}

func NewOTPChallenge(number string, codeHash []byte, ttl time.Duration) (*OTPChallenge, error) {
	normalizedPhone, err := phone.NormalizePhone(number, "ZZ")
	if err != nil {
		return nil, fmt.Errorf("phone: %w", err)
	}

	now := time.Now()
	otpChallenge := &OTPChallenge{
		ID:        uuid.New(),
		Phone:     normalizedPhone,
		CodeHash:  codeHash,
		ExpiresAt: now.Add(ttl),
		Attempts:  0,
		CreatedAt: now,
	}

	if err := otpChallenge.validateOtherFields(); err != nil {
		return nil, err
	}
	return otpChallenge, nil
}

func (o *OTPChallenge) Validate() error {
	if err := phone.ValidatePhone(o.Phone, "ZZ"); err != nil {
		return err
	}
	if err := o.validateOtherFields(); err != nil {
		return err
	}
	return nil
}

func (o *OTPChallenge) validateOtherFields() error {
	if o.ID == uuid.Nil {
		return ErrEmptyChallengeID
	}
	if len(o.CodeHash) == 0 {
		return ErrEmptyCodeHash
	}
	return nil
}

func (o *OTPChallenge) IsExpired(now time.Time) bool {
	return now.After(o.ExpiresAt)
}

func (o *OTPChallenge) IsConsumed() bool {
	return o.ConsumedAt != nil
}

func (o *OTPChallenge) Consume(now time.Time) error {
	if o.IsConsumed() {
		return ErrAlreadyConsumed
	}
	if o.IsExpired(now) {
		return ErrChallengeExpired
	}
	o.ConsumedAt = &now
	return nil
}

func (o *OTPChallenge) IncrementAttempt(maxAttempts int) error {
	if o.Attempts >= maxAttempts {
		return ErrMaxAttemptsExceeded
	}
	o.Attempts++
	return nil
}
