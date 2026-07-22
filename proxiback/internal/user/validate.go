package user

import (
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrInvalidUserEmail = errors.New("invalid user email")
	ErrEmailAlreadyInUse = errors.New("email already in use")
	ErrInvalidUserStatus = errors.New("invalid user status")
	ErrUserNotFound = errors.New("user not found")
	ErrMissingPassword = errors.New("password cannot be empty")
	ErrUserAlreadyExists = errors.New("user already exists")
)

func Validate(user User) error {

	if err := ValidateEmail(user.Email); err != nil {
		return err
	}

	displayNameLength := len([]rune(strings.TrimSpace(user.DisplayName)))
	if displayNameLength < 2 || displayNameLength > 100 {
		return errors.New("invalid display name")
	}

	if strings.TrimSpace(user.Password) == "" {
		return ErrMissingPassword
	}

	switch user.Status {
	case active, inactive:
		return nil
	default:
		return ErrInvalidUserStatus
	}
}

func ValidateEmail(email string) error {
	normalizedEmail := NormalizeEmail(email)

	if normalizedEmail == "" {
		return ErrInvalidUserEmail
	}

	if len(normalizedEmail) > 254 {
		return ErrInvalidUserEmail
	}

	adress, err := mail.ParseAddress(normalizedEmail)
	if err != nil || adress.Address != normalizedEmail {
		return ErrInvalidUserEmail
	}

	if adress.Address != normalizedEmail {
		return ErrInvalidUserEmail
	}

	return nil
}