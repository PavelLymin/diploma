package phone

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nyaruka/phonenumbers/v2"
)

var (
	ErrEmptyPhone    = errors.New("phone must not be empty")
	ErrInvalidPhone  = errors.New("phone has an invalid format")
	ErrNotE164Format = errors.New("phone must be in E.164 format")
)

func NormalizePhone(phoneNumber string, defaultRegion string) (string, error) {
	phone := strings.TrimSpace(phoneNumber)
	if phone == "" {
		return "", ErrEmptyPhone
	}
	number, err := phonenumbers.Parse(phone, defaultRegion)
	if err != nil || !phonenumbers.IsValidNumber(number) {
		return "", ErrInvalidPhone
	}
	return phonenumbers.Format(number, phonenumbers.E164), nil
}

func ValidatePhone(phoneNumber string, defaultRegion string) error {
	if phoneNumber == "" {
		return ErrEmptyPhone
	}
	normalizedPhone, err := NormalizePhone(phoneNumber, defaultRegion)
	if err != nil {
		return fmt.Errorf("phone: %w", err)
	}
	if normalizedPhone != phoneNumber {
		return ErrNotE164Format
	}
	return nil
}
