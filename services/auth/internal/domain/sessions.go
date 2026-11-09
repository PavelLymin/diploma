package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID         uuid.UUID  `json:"id"`
	FamilyID   uuid.UUID  `json:"family_id"`
	UserID     uuid.UUID  `json:"user_id"`
	TokenHash  []byte     `json:"token_hash"`
	ExpiresAt  time.Time  `json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

func NewSession(
	userID,
	familyID uuid.UUID,
	tokenHash []byte,
	ttl time.Duration,
	now time.Time,
) (*Session, error) {
	if ttl <= 0 {
		return nil, errors.New("refresh session TTL must be positive")
	}
	now = now.UTC()
	session := &Session{
		ID:        uuid.New(),
		FamilyID:  familyID,
		UserID:    userID,
		TokenHash: append([]byte(nil), tokenHash...),
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
	}
	if err := session.Validate(); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Session) Validate() error {
	if s == nil || s.ID == uuid.Nil {
		return ErrEmptyRefreshSessionID
	}
	if s.UserID == uuid.Nil {
		return ErrEmptyUserID
	}
	if s.FamilyID == uuid.Nil {
		return ErrEmptyFamilyID
	}
	if len(s.TokenHash) == 0 {
		return ErrEmptyTokenHash
	}
	if s.CreatedAt.IsZero() || !s.ExpiresAt.After(s.CreatedAt) {
		return errors.New("invalid refresh session timestamps")
	}
	return nil
}

func (s *Session) IsExpired(now time.Time) bool {
	return s == nil || !now.UTC().Before(s.ExpiresAt)
}

func (s *Session) IsRevoked() bool {
	return s == nil || s.RevokedAt != nil
}

// Отмечает сессию как отозванную
func (s *Session) Revoke(now time.Time) error {
	if s == nil || s.RevokedAt != nil {
		return ErrAlreadyRevoked
	}
	now = now.UTC()
	s.RevokedAt = &now
	return nil
}

// Проверяет, что сессия:
// - не отозвана;
// - не истекла
func (s *Session) Use(now time.Time) error {
	if s == nil || s.IsRevoked() {
		return ErrAlreadyRevoked
	}
	if s.IsExpired(now) {
		return ErrSessionExpired
	}
	now = now.UTC()
	s.LastUsedAt = &now
	return nil
}
