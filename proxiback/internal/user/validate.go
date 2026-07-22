package user

import (
	"net/mail"
	"strings"
)

func Validate(user User) error {

	if err := ValidateEmail(user.Email); err != nil {
		return err
	}

	displayNameLength := len([]rune(strings.TrimSpace(user.DisplayName)))
	if displayNameLength < 2 || displayNameLength > 100 {
		return ErrInvalidDisplayName
	}

	if strings.TrimSpace(user.Password) == "" {
		return ErrMissingPassword
	}

	switch user.Status {
	case StatusActive, StatusInactive:
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
