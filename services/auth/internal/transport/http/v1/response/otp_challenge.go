package response

import (
	"time"

	"github.com/google/uuid"
)

type ChallengeResponse struct {
	ChallengeID       uuid.UUID `json:"challenge_id"`
	SentAt            time.Time `json:"sent_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	ResendAvailableAt time.Time `json:"resend_available_at"`
	ServerTime        time.Time `json:"server_time"`
}
