package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyRefreshSessionID = errors.New("refresh session id must not be empty")
	ErrEmptyUserID           = errors.New("user id must not be empty")
	ErrEmptyTokenHash        = errors.New("token hash cannot be empty")
	ErrSessionExpired        = errors.New("refresh session is expired")
	ErrAlreadyRevoked        = errors.New("refresh session already revoked")
)

type RefreshSession struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  []byte
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
	LastUsedAt *time.Time
}

func NewRefreshSession(userID uuid.UUID, tokenHash []byte, ttl time.Duration) (*RefreshSession, error) {
	now := time.Now()
	session := &RefreshSession{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
	}

	if err := session.Validate(); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *RefreshSession) Validate() error {
	if s.ID == uuid.Nil {
		return ErrEmptyRefreshSessionID
	}

	if s.UserID == uuid.Nil {
		return ErrEmptyUserID
	}

	if len(s.TokenHash) == 0 {
		return ErrEmptyTokenHash
	}

	return nil
}

func (s *RefreshSession) IsExpired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}

func (s *RefreshSession) IsRevoked() bool {
	return s.RevokedAt != nil
}

func (s *RefreshSession) Revoke(now time.Time) error {
	if s.IsRevoked() {
		return ErrAlreadyRevoked
	}
	s.RevokedAt = &now
	return nil
}

func (s *RefreshSession) Use(now time.Time) error {
	if s.IsRevoked() {
		return ErrAlreadyRevoked
	}
	if s.IsExpired(now) {
		return ErrSessionExpired
	}
	s.LastUsedAt = &now
	return nil
}
