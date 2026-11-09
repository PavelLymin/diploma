package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
)

const otpDigits = 6

type Crypto struct {
	otpHMACKey []byte
}

func NewCrypto(otpHMACKey []byte) (*Crypto, error) {
	if len(otpHMACKey) < 32 {
		return nil, errors.New("OTP HMAC key must contain at least 32 bytes")
	}
	return &Crypto{otpHMACKey: append([]byte(nil), otpHMACKey...)}, nil
}

func (c *Crypto) GenerateOTP() (string, error) {
	if c == nil || len(c.otpHMACKey) < 32 {
		return "", errors.New("OTP crypto is not configured")
	}
	limit := big.NewInt(1_000_000)
	n, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return "", fmt.Errorf("generate OTP: %w", err)
	}
	return fmt.Sprintf("%0*d", otpDigits, n.Int64()), nil
}

func (c *Crypto) HashToken(value string) string {
	mac := hmac.New(sha256.New, c.otpHMACKey)
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func (c *Crypto) CompareToken(value, expectedHash string) bool {
	actual := c.HashToken(value)
	if len(actual) != len(expectedHash) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expectedHash)) == 1
}

func HashRefreshToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return append([]byte(nil), sum[:]...)
}
