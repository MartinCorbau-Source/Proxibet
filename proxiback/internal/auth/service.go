package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/user"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 64
)

var (
	ErrInvalidEmailFormat = errors.New("invalid email format")
	ErrInvalidPasswordFormat = errors.New("invalid password format")
	ErrInvalidDisplayNameFormat = errors.New("invalid display name format")
	ErrInvalidDisplayName = errors.New("Invalid display name")
	ErrPasswordTooShort   = fmt.Errorf("password must be at least %d characters long", MinPasswordLength)
	ErrPasswordTooLong    = fmt.Errorf("password must be at most %d characters long", MaxPasswordLength)
)

type Service struct {
	userRepo user.Repository
}

func NewService(userRepo user.Repository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}

func (service *Service) Register(ctx context.Context, email, password, displayName string) (RegisterResponse, error) {
	email = user.NormalizeEmail(email)
	displayName = user.NormalizeDisplayName(displayName)

	if err := validateRegisterRequest(email, password, displayName); err != nil {
		return RegisterResponse{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := user.User{
		ID:           uuid.New(),
		DisplayName:  displayName,
		Email:        email,
		Password: string(passwordHash),
		Status:       user.StatusActive,
	}

	createdUser, err := service.userRepo.Create(ctx, newUser)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("failed to create user: %w", err)
	}

	return RegisterResponse{ID: createdUser.ID.String(), DisplayName: createdUser.DisplayName, Email: createdUser.Email}, nil
}

func validateRegisterRequest(email, password, displayName string, ) error {
	if err := validateEmail(email); err != nil {
		return err
	}

	displayNameLength := utf8.RuneCountInString(displayName)

	if displayNameLength < 2 || displayNameLength > 100 {
		return ErrInvalidDisplayName
	}

	passwordLength := len([]byte(password))

	if passwordLength < 2 ||
		passwordLength > 100 {
		return ErrInvalidPasswordFormat
	}

	return nil
}

func validateEmail(email string) error {
	if len(email) == 0 {
		return ErrInvalidEmailFormat
	}

	if len(email) > 254 {
		return ErrInvalidEmailFormat
	}

	address, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmailFormat
	}

	if strings.ToLower(address.Address) != email {
		return ErrInvalidEmailFormat
	}

	return nil
}