package domain

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PavelLymin/diploma/pkg/phone"
	"github.com/google/uuid"
)

var (
	ErrEmptyUserID       = errors.New("user id must not be empty")
	ErrEmptyEmail        = errors.New("email must not be empty")
	ErrEmailWhitespace   = errors.New("email must not contain leading or trailing whitespace")
	ErrEmailTooLong      = errors.New("email must not exceed 255 characters")
	ErrInvalidEmail      = errors.New("email has an invalid format")
	ErrFutureBirthDate   = errors.New("birth date must not be in the future")
	ErrInvalidUserStatus = errors.New("invalid user status")
)

type UserStatus string

const (
	UserStatusActive  UserStatus = "ACTIVE"
	UserStatusBlocked UserStatus = "BLOCKED"
)

type User struct {
	ID           uuid.UUID
	Email        *string
	Phone        string
	FirstName    *string
	LastName     *string
	MiddleName   *string
	FullName     *string
	BirthDate    *time.Time
	AvatarFileID *uuid.UUID
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(
	id uuid.UUID,
	email *string,
	number string,
	firstName *string,
	lastName *string,
	middleName *string,
	birthDate *time.Time,
) (*User, error) {
	normalizedPhone, err := phone.NormalizePhone(number, "ZZ")
	if err != nil {
		return nil, fmt.Errorf("phone: %w", err)
	}

	fName := normalizeOptionalText(firstName)
	lName := normalizeOptionalText(lastName)
	mName := normalizeOptionalText(middleName)

	user := &User{
		ID:         id,
		Email:      normalizeOptionalText(email),
		Phone:      normalizedPhone,
		FirstName:  fName,
		LastName:   lName,
		MiddleName: mName,
		FullName:   createFullName(lName, fName, mName),
		BirthDate:  birthDate,
		Status:     UserStatusActive,
	}

	if err := user.validateOtherFields(); err != nil {
		return nil, err
	}
	return user, nil
}

func (u User) Validate() error {
	if err := phone.ValidatePhone(u.Phone, "ZZ"); err != nil {
		return err
	}
	if err := u.validateOtherFields(); err != nil {
		return err
	}
	return nil
}

func (u User) validateOtherFields() error {
	if u.ID == uuid.Nil {
		return ErrEmptyUserID
	}
	if err := validateEmail(u.Email); err != nil {
		return err
	}
	if err := validateOptionalText("first_name", u.FirstName, 100); err != nil {
		return err
	}
	if err := validateOptionalText("last_name", u.LastName, 100); err != nil {
		return err
	}
	if err := validateOptionalText("middle_name", u.MiddleName, 100); err != nil {
		return err
	}
	if err := validateOptionalText("full_name", u.FullName, 300); err != nil {
		return err
	}
	if err := validateBirthDate(u.BirthDate); err != nil {
		return err
	}
	switch u.Status {
	case UserStatusActive, UserStatusBlocked:
	default:
		return fmt.Errorf("%w: %q", ErrInvalidUserStatus, u.Status)
	}
	return nil
}

func createFullName(lName *string, fName *string, mName *string) *string {
	parts := make([]string, 0, 3)
	if lName != nil {
		parts = append(parts, *lName)
	}
	if fName != nil {
		parts = append(parts, *fName)
	}
	if mName != nil {
		parts = append(parts, *mName)
	}
	var fullName *string
	if len(parts) > 0 {
		joined := strings.Join(parts, " ")
		fullName = &joined
	}
	return fullName
}

func normalizeOptionalText(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	return &normalized
}

func validateEmail(email *string) error {
	if email == nil {
		return nil
	}
	value := *email
	if value == "" {
		return ErrEmptyEmail
	}
	if strings.TrimSpace(value) != value {
		return ErrEmailWhitespace
	}
	if utf8.RuneCountInString(value) > 255 {
		return ErrEmailTooLong
	}
	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Address != value {
		return ErrInvalidEmail
	}
	return nil
}

func validateOptionalText(field string, value *string, maxLength int) error {
	if value == nil {
		return nil
	}
	text := *value
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("%s must not be empty", field)
	}
	if strings.TrimSpace(text) != text {
		return fmt.Errorf("%s must not contain leading or trailing whitespace", field)
	}
	if utf8.RuneCountInString(text) > maxLength {
		return fmt.Errorf("%s must not exceed %d characters", field, maxLength)
	}
	return nil
}

func validateBirthDate(birthDate *time.Time) error {
	if birthDate == nil {
		return nil
	}

	now := time.Now().In(birthDate.Location())

	year, month, day := birthDate.Date()
	currentYear, currentMonth, currentDay := now.Date()

	isFuture := year > currentYear ||
		(year == currentYear && month > currentMonth) ||
		(year == currentYear &&
			month == currentMonth &&
			day > currentDay)
	if isFuture {
		return ErrFutureBirthDate
	}
	return nil
}
