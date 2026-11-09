package autherrors

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	httperrors "github.com/PavelLymin/diploma/pkg/transport/http"
	"github.com/PavelLymin/diploma/services/auth/internal/domain"
	"github.com/gin-gonic/gin"
)

type RateLimitDetails struct {
	Code              string     `json:"code"`
	RetryAfterSeconds int        `json:"retry_after_seconds"`
	ResendAvailableAt *time.Time `json:"resend_available_at,omitempty"`
	BlockedUntil      *time.Time `json:"blocked_until,omitempty"`
	ServerTime        time.Time  `json:"server_time"`
}

func MapAuthError(
	err error,
) (int, httperrors.ErrorResponse, http.Header) {
	headers := make(http.Header)

	if err == nil {
		err = errors.New("nil error passed to MapAuthError")
	}

	var rateLimitErr *domain.RateLimitError

	if errors.As(err, &rateLimitErr) && rateLimitErr != nil {
		retryAfter := rateLimitErr.RetryAfterSeconds
		if retryAfter < 0 {
			retryAfter = 0
		}

		headers.Set("Retry-After", strconv.Itoa(retryAfter))

		return http.StatusTooManyRequests, httperrors.ErrorResponse{
			Code:    "too_many_requests",
			Message: "too many requests",
			Details: RateLimitDetails{
				Code:              rateLimitErr.Code,
				RetryAfterSeconds: retryAfter,
				ResendAvailableAt: rateLimitErr.ResendAvailableAt,
				BlockedUntil:      rateLimitErr.BlockedUntil,
				ServerTime:        rateLimitErr.ServerTime,
			},
		}, headers
	}

	switch {
	case errors.Is(err, domain.ErrInvalidChallenge):
		return http.StatusBadRequest, httperrors.ErrorResponse{
			Code:    "invalid_challenge",
			Message: "invalid OTP challenge",
		}, headers

	case errors.Is(err, domain.ErrChallengeNotFound):
		return http.StatusNotFound, httperrors.ErrorResponse{
			Code:    "challenge_not_found",
			Message: "OTP challenge not found",
		}, headers

	case errors.Is(err, domain.ErrChallengeExpired):
		return http.StatusGone, httperrors.ErrorResponse{
			Code:    "challenge_expired",
			Message: "OTP challenge expired",
		}, headers

	case errors.Is(err, domain.ErrInvalidOTP):
		return http.StatusBadRequest, httperrors.ErrorResponse{
			Code:    "invalid_otp",
			Message: "invalid OTP code",
		}, headers

	case errors.Is(err, domain.ErrInvalidOTPFormat):
		return http.StatusUnprocessableEntity, httperrors.ErrorResponse{
			Code:    "validation_error",
			Message: "invalid OTP code format",
			Details: &httperrors.Details{
				Field:   "code",
				Message: "must contain exactly six digits",
			},
		}, headers

	case errors.Is(err, domain.ErrMaxAttemptsExceeded):
		return http.StatusConflict, httperrors.ErrorResponse{
			Code:    "otp_max_attempts_exceeded",
			Message: "maximum OTP attempts exceeded; request a new challenge",
		}, headers

	case errors.Is(err, domain.ErrRateLimited):
		return http.StatusTooManyRequests, httperrors.ErrorResponse{
			Code:    "too_many_requests",
			Message: "too many requests",
		}, headers

	case errors.Is(err, domain.ErrChallengeAlreadyActive):
		return http.StatusConflict, httperrors.ErrorResponse{
			Code:    "challenge_already_active",
			Message: "an active OTP challenge already exists",
		}, headers

	case errors.Is(err, domain.ErrStaleChallenge):
		return http.StatusConflict, httperrors.ErrorResponse{
			Code:    "stale_challenge",
			Message: "OTP challenge is no longer active",
		}, headers

	case errors.Is(err, domain.ErrIdempotencyConflict):
		return http.StatusConflict, httperrors.ErrorResponse{
			Code:    "idempotency_conflict",
			Message: "request_id was already used with different parameters",
		}, headers

	case errors.Is(err, domain.ErrConsentRequired):
		return http.StatusForbidden, httperrors.ErrorResponse{
			Code:    "consent_required",
			Message: "personal data consent is required",
		}, headers

	case errors.Is(err, domain.ErrSessionExpired):
		return http.StatusUnauthorized, httperrors.ErrorResponse{
			Code:    "session_expired",
			Message: "refresh session expired",
		}, headers

	case errors.Is(err, domain.ErrAlreadyRevoked):
		return http.StatusUnauthorized, httperrors.ErrorResponse{
			Code:    "session_revoked",
			Message: "refresh session is no longer active",
		}, headers

	case errors.Is(err, domain.ErrEmptyChallengeID):
		return http.StatusBadRequest, httperrors.ErrorResponse{
			Code:    "bad_request",
			Message: "invalid request",
			Details: &httperrors.Details{
				Field:   "challenge_id",
				Message: "must not be empty",
			},
		}, headers

	case errors.Is(err, domain.ErrEmptyRequestID):
		return http.StatusUnprocessableEntity, httperrors.ErrorResponse{
			Code:    "validation_error",
			Message: "invalid request",
			Details: &httperrors.Details{
				Field:   "request_id",
				Message: "must not be empty",
			},
		}, headers

	default:
		status, response := httperrors.MapError(err)
		return status, response, headers
	}
}

func WriteError(c *gin.Context, err error) {
	status, response, headers := MapAuthError(err)
	for k, v := range headers {
		c.Header(k, v[0])
	}

	c.AbortWithStatusJSON(status, response)
}
