package request

import (
	"github.com/google/uuid"
)

type CreateChallengeRequest struct {
	RequestID           uuid.UUID `json:"request_id" binding:"required"`
	Phone               string    `json:"phone" binding:"required,phone"`
	PersonalDataConsent bool      `json:"personal_data_consent" binding:"required"`
}

type ResendChallengeRequest struct {
	RequestID   uuid.UUID `json:"request_id" binding:"required"`
	ChallengeID uuid.UUID `json:"challenge_id" binding:"required"`
}

type VerifyRequest struct {
	Code string `json:"code" binding:"required"`
}
