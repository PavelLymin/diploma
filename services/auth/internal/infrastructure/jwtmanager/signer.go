package jwtmanager

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Signer struct {
	privateKey *rsa.PrivateKey
	keyID      string
	issuer     string
	audience   string
	ttl        time.Duration
}

func New(
	key *rsa.PrivateKey,
	keyID,
	issuer,
	audience string,
	ttl time.Duration,
) (*Signer, error) {
	switch {
	case key == nil:
		return nil, errors.New("JWT private key is required")
	case key.N == nil || key.N.BitLen() < 2048:
		return nil, errors.New("JWT RSA key must be at least 2048 bits")
	case strings.TrimSpace(keyID) == "":
		return nil, errors.New("JWT key ID is required")
	case strings.TrimSpace(issuer) == "":
		return nil, errors.New("JWT issuer is required")
	case strings.TrimSpace(audience) == "":
		return nil, errors.New("JWT audience is required")
	case ttl <= 0:
		return nil, errors.New("JWT access TTL must be positive")
	}
	if err := key.Validate(); err != nil {
		return nil, fmt.Errorf("validate JWT RSA key: %w", err)
	}

	return &Signer{
		privateKey: key,
		keyID:      keyID,
		issuer:     issuer,
		audience:   audience,
		ttl:        ttl,
	}, nil
}

func (s *Signer) AccessTTL() time.Duration { return s.ttl }

func (s *Signer) IssueAccessToken(
	userID,
	familyID uuid.UUID,
	now time.Time,
) (string, error) {
	if s == nil || s.privateKey == nil {
		return "", errors.New("JWT signer is not configured")
	}
	if userID == uuid.Nil || familyID == uuid.Nil {
		return "", errors.New("JWT user ID and session family ID are required")
	}
	if now.IsZero() {
		return "", errors.New("JWT issue time is required")
	}
	now = now.UTC()

	c := claims{
		TokenUse: "access",
		FamilyID: familyID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID.String(),
			Audience:  jwt.ClaimStrings{s.audience},
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, c)
	token.Header["kid"] = s.keyID
	return token.SignedString(s.privateKey)
}

func GenerateRefreshToken() (string, []byte, error) {
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", nil, err
	}
	text := base64.RawURLEncoding.EncodeToString(raw[:])
	hash := sha256.Sum256([]byte(text))
	return text, hash[:], nil
}
