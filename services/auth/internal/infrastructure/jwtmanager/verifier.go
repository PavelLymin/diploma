package jwtmanager

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PavelLymin/diploma/services/auth/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Verifier struct {
	publicKeys map[string]*rsa.PublicKey
	issuer     string
	audience   string
	leeway     time.Duration
}

func NewVerifier(
	publicKeys map[string]*rsa.PublicKey,
	issuer,
	audience string,
) (*Verifier, error) {
	if strings.TrimSpace(issuer) == "" {
		return nil, errors.New("JWT issuer is required")
	}
	if strings.TrimSpace(audience) == "" {
		return nil, errors.New("JWT audience is required")
	}
	if len(publicKeys) == 0 {
		return nil, errors.New("JWT public keys are required")
	}

	keys := make(map[string]*rsa.PublicKey, len(publicKeys))
	for kid, key := range publicKeys {
		if strings.TrimSpace(kid) == "" {
			return nil, errors.New("JWT key ID is required")
		}
		if key == nil || key.N == nil || key.N.BitLen() < 2048 {
			return nil, fmt.Errorf("invalid RSA public key for kid %q", kid)
		}

		keys[kid] = key
	}

	return &Verifier{
		publicKeys: keys,
		issuer:     issuer,
		audience:   audience,
		leeway:     2 * time.Second,
	}, nil
}

func (v *Verifier) ParseAccessToken(raw string) (*domain.AccessClaims, error) {
	if v == nil || len(raw) == 0 || len(raw) > 8192 {
		return nil, errors.New("invalid access token size")
	}

	c := &claims{}
	token, err := jwt.ParseWithClaims(
		raw,
		c,
		v.keyFunc,
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(v.leeway),
	)
	if err != nil || token == nil || !token.Valid {
		return nil, errors.New("invalid access token")
	}
	if c.TokenUse != "access" || c.IssuedAt == nil || c.NotBefore == nil || c.ExpiresAt == nil || c.ID == "" {
		return nil, errors.New("required access token claims are missing")
	}
	if _, err := uuid.Parse(c.Subject); err != nil {
		return nil, errors.New("invalid access token subject")
	}
	if _, err := uuid.Parse(c.FamilyID); err != nil {
		return nil, errors.New("invalid access token session family")
	}
	if _, err := uuid.Parse(c.ID); err != nil {
		return nil, errors.New("invalid access token ID")
	}
	return &domain.AccessClaims{
		TokenUse: c.TokenUse,
		UserID:   c.Subject,
		FamilyID: c.FamilyID,
	}, nil
}

func (v *Verifier) keyFunc(token *jwt.Token) (any, error) {
	if token.Method != jwt.SigningMethodRS256 {
		return nil, errors.New("unexpected JWT signing method")
	}

	kid, ok := token.Header["kid"].(string)
	if !ok || strings.TrimSpace(kid) == "" {
		return nil, errors.New("missing JWT key ID")
	}

	key, ok := v.publicKeys[kid]
	if !ok || key == nil {
		return nil, errors.New("unknown JWT key ID")
	}

	return key, nil
}
