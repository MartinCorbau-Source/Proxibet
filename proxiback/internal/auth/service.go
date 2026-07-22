package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
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

	PasswordResetTokenByteLength = 32
	PasswordResetTokenDuration   = time.Hour
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

type PasswordResetTicket struct {
	Token     string
	ExpiresAt time.Time
}

func NewService(userRepo user.Repository, tokenGenerator *TokenGenerator) *Service {
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

func (service *Service) ChangePassword(ctx context.Context, accessToken, currentPassword, newPassword string) error {
	claims, err := service.tokenGenerator.ParseAccessToken(accessToken)
	if err != nil {
		return err
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return user.ErrInvalidToken
	}

	if err := validatePassword(newPassword); err != nil {
		return err
	}

	authenticatedUser, err := service.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user by id: %w", err)
	}

	if authenticatedUser.Status != user.StatusActive {
		return user.ErrAccountInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(authenticatedUser.Password), []byte(currentPassword)); err != nil {
		return user.ErrInvalidCredentials
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := service.userRepo.UpdatePassword(ctx, userID, string(passwordHash)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (service *Service) RequestPasswordReset(ctx context.Context, email string) (PasswordResetTicket, error) {
	email = user.NormalizeEmail(email)
	if err := validateEmail(email); err != nil {
		return PasswordResetTicket{}, err
	}

	foundUser, err := service.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return PasswordResetTicket{}, nil
		}

		return PasswordResetTicket{}, fmt.Errorf("failed to get user by email: %w", err)
	}

	if foundUser.Status != user.StatusActive {
		return PasswordResetTicket{}, nil
	}

	resetToken, err := generatePasswordResetToken()
	if err != nil {
		return PasswordResetTicket{}, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(PasswordResetTokenDuration)

	err = service.userRepo.CreatePasswordResetToken(ctx, user.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    foundUser.ID,
		TokenHash: hashPasswordResetToken(resetToken),
		ExpiresAt: expiresAt,
		CreatedAt: now,
	})
	if err != nil {
		return PasswordResetTicket{}, fmt.Errorf("failed to create password reset token: %w", err)
	}

	return PasswordResetTicket{
		Token:     resetToken,
		ExpiresAt: expiresAt,
	}, nil
}

func (service *Service) ResetPassword(ctx context.Context, resetToken, newPassword string) error {
	resetToken = strings.TrimSpace(resetToken)
	if resetToken == "" {
		return user.ErrInvalidPasswordResetToken
	}

	if err := validatePassword(newPassword); err != nil {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = service.userRepo.ResetPasswordWithToken(
		ctx,
		hashPasswordResetToken(resetToken),
		string(passwordHash),
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	return nil
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

	return validatePassword(password)
}

func validatePassword(password string) error {
	passwordLength := len([]byte(password))
	if passwordLength < MinPasswordLength {
		return ErrPasswordTooShort
	}

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

func generatePasswordResetToken() (string, error) {
	tokenBytes := make([]byte, PasswordResetTokenByteLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate password reset token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(tokenBytes), nil
}

func hashPasswordResetToken(resetToken string) string {
	tokenHash := sha256.Sum256([]byte(resetToken))
	return hex.EncodeToString(tokenHash[:])
}
