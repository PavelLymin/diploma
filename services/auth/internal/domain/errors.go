package domain

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var (
	ErrInvalidChallenge       = errors.New("invalid OTP challenge")
	ErrChallengeNotFound      = errors.New("OTP challenge not found")
	ErrChallengeExpired       = errors.New("OTP challenge expired")
	ErrInvalidOTP             = errors.New("invalid OTP code")
	ErrMaxAttemptsExceeded    = errors.New("maximum OTP attempts exceeded")
	ErrRateLimited            = errors.New("OTP operation is rate limited")
	ErrChallengeAlreadyActive = errors.New("an active OTP challenge already exists")
	ErrStaleChallenge         = errors.New("OTP challenge is no longer active")
	ErrIdempotencyConflict    = errors.New("request_id was already used with different parameters")
	ErrConsentRequired        = errors.New("personal data consent is required")
	ErrSessionExpired         = errors.New("refresh session is expired")
	ErrAlreadyRevoked         = errors.New("refresh session already revoked")
	ErrEmptyRequestID         = errors.New("request id must not be empty")
	ErrEmptyChallengeID       = errors.New("OTP challenge id must not be empty")
	ErrEmptyCodeHash          = errors.New("code hash cannot be empty")
	ErrEmptyRefreshSessionID  = errors.New("refresh session id must not be empty")
	ErrEmptyUserID            = errors.New("user id must not be empty")
	ErrEmptyFamilyID          = errors.New("family id must not be empty")
	ErrEmptyTokenHash         = errors.New("token hash cannot be empty")
	ErrInvalidOTPFormat       = errors.New("invalid OTP format")
)

const (
	RateLimitCodeResendTooEarly = "otp_resend_too_early"
	RateLimitCodePhoneBlocked   = "otp_phone_blocked"
)

type RateLimitError struct {
	Code              string
	RetryAfterSeconds int
	ResendAvailableAt *time.Time
	BlockedUntil      *time.Time
	ServerTime        time.Time
}

func (e *RateLimitError) Error() string {
	if e == nil {
		return ErrRateLimited.Error()
	}
	return fmt.Sprintf("%s: %s", ErrRateLimited, e.Code)
}

func (e *RateLimitError) Unwrap() error { return ErrRateLimited }

func NewRateLimitError(
	code string,
	serverTime time.Time,
	resendAt,
	blockedUntil *time.Time,
) *RateLimitError {
	serverTime = serverTime.UTC()
	var normalizedBlockedUntil *time.Time
	if blockedUntil != nil {
		t := blockedUntil.UTC()
		normalizedBlockedUntil = &t
	}
	var normalizedResendAt *time.Time
	if resendAt != nil {
		t := resendAt.UTC()
		normalizedResendAt = &t
	}

	var retryUntil *time.Time
	for _, candidate := range []*time.Time{normalizedBlockedUntil, normalizedResendAt} {
		if candidate == nil || !candidate.After(serverTime) {
			continue
		}

		if retryUntil == nil || candidate.After(*retryUntil) {
			retryUntil = candidate
		}
	}

	retryAfter := 0
	if retryUntil != nil {
		retryAfter = int(math.Ceil(retryUntil.Sub(serverTime).Seconds()))
	}

	return &RateLimitError{
		Code:              code,
		RetryAfterSeconds: retryAfter,
		ResendAvailableAt: normalizedResendAt,
		BlockedUntil:      normalizedBlockedUntil,
		ServerTime:        serverTime,
	}
}
