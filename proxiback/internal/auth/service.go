package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/user"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 72
)

var (
	ErrInvalidEmailFormat       = errors.New("invalid email format")
	ErrInvalidPasswordFormat    = errors.New("invalid password format")
	ErrInvalidDisplayNameFormat = errors.New("invalid display name format")
	ErrInvalidDisplayName       = errors.New("Invalid display name")
	ErrPasswordTooShort         = fmt.Errorf("password must be at least %d characters long", MinPasswordLength)
	ErrPasswordTooLong          = fmt.Errorf("password must be at most %d characters long", MaxPasswordLength)
)

type Service struct {
	userRepo             user.Repository
	refreshTokenRepo     RefreshTokenRepository
	tokenGenerator       *TokenGenerator
	refreshTokenDuration time.Duration
}

func NewService(
	userRepo user.Repository,
	refreshTokenRepo RefreshTokenRepository,
	tokenGenerator *TokenGenerator,
	refreshTokenDuration time.Duration,
) *Service {
	return &Service{
		userRepo:             userRepo,
		refreshTokenRepo:     refreshTokenRepo,
		tokenGenerator:       tokenGenerator,
		refreshTokenDuration: refreshTokenDuration,
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
		ID:          uuid.New(),
		DisplayName: displayName,
		Email:       email,
		Password:    string(passwordHash),
		Status:      user.StatusActive,
	}

	createdUser, err := service.userRepo.Create(ctx, newUser)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("failed to create user: %w", err)
	}

	return RegisterResponse{ID: createdUser.ID.String(), DisplayName: createdUser.DisplayName, Email: createdUser.Email}, nil
}

func (service *Service) Login(ctx context.Context, email, password string) (LoginResponse, error) {
	email = user.NormalizeEmail(email)

	authenticatedUser, err := service.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("failed to get user by email: %w", err)
	}

	if authenticatedUser.Status != user.StatusActive {
		return LoginResponse{}, user.ErrAccountInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(authenticatedUser.Password), []byte(password)); err != nil {
		return LoginResponse{}, user.ErrInvalidCredentials
	}

	accessToken, accessTokenExpiresAt, err := service.tokenGenerator.Generate(authenticatedUser)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("failed to generate token: %w", err)
	}

	plainRefreshToken, refreshToken, err := newRefreshToken(authenticatedUser.ID, service.refreshTokenDuration)
	if err != nil {
		return LoginResponse{}, err
	}

	if _, err := service.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return LoginResponse{}, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: plainRefreshToken,
		ExpiresIn:    accessTokenExpiresAt.Unix(),
		User:         newUserResponse(authenticatedUser),
	}, nil
}

func (service *Service) Refresh(ctx context.Context, refreshToken string) (RefreshResponse, error) {
	refreshToken, err := normalizeRefreshToken(refreshToken)
	if err != nil {
		return RefreshResponse{}, err
	}

	plainNextRefreshToken, nextRefreshToken, err := newRefreshToken(uuid.Nil, service.refreshTokenDuration)
	if err != nil {
		return RefreshResponse{}, err
	}

	rotatedRefreshToken, err := service.refreshTokenRepo.Rotate(
		ctx,
		hashRefreshToken(refreshToken),
		nextRefreshToken,
		time.Now().UTC(),
	)
	if err != nil {
		return RefreshResponse{}, err
	}

	authenticatedUser, err := service.userRepo.FindByID(ctx, rotatedRefreshToken.UserID)
	if err != nil {
		return RefreshResponse{}, fmt.Errorf("failed to get user by id: %w", err)
	}

	if authenticatedUser.Status != user.StatusActive {
		return RefreshResponse{}, user.ErrAccountInactive
	}

	accessToken, accessTokenExpiresAt, err := service.tokenGenerator.Generate(authenticatedUser)
	if err != nil {
		return RefreshResponse{}, fmt.Errorf("failed to generate token: %w", err)
	}

	return RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: plainNextRefreshToken,
		ExpiresIn:    accessTokenExpiresAt.Unix(),
		User:         newUserResponse(authenticatedUser),
	}, nil
}

func newUserResponse(authenticatedUser user.User) UserResponse {
	return UserResponse{
		ID:          authenticatedUser.ID.String(),
		DisplayName: authenticatedUser.DisplayName,
		Email:       authenticatedUser.Email,
	}
}

func validateRegisterRequest(email, password, displayName string) error {
	if err := validateEmail(email); err != nil {
		return err
	}

	displayNameLength := utf8.RuneCountInString(displayName)

	if displayNameLength < 2 || displayNameLength > 100 {
		return ErrInvalidDisplayName
	}

	passwordLength := len([]byte(password))
	if passwordLength < MinPasswordLength {
		return ErrPasswordTooShort
	}

	if passwordLength > MaxPasswordLength {
		return ErrPasswordTooLong
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
