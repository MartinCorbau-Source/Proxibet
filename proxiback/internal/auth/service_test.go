package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/user"
)

func TestChangePasswordUpdatesPasswordHash(t *testing.T) {
	t.Parallel()

	storedUser := testUser(t, "current-password")
	repository := &fakeUserRepository{storedUser: storedUser}
	tokenGenerator := testTokenGenerator(t)
	service := NewService(repository, tokenGenerator)
	accessToken := testAccessToken(t, tokenGenerator, storedUser)

	err := service.ChangePassword(context.Background(), accessToken, "current-password", "new-password")
	if err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}

	if !repository.updatePasswordCalled {
		t.Fatal("expected password update to be called")
	}

	if repository.updatedUserID != storedUser.ID {
		t.Fatalf("updated user ID = %s, want %s", repository.updatedUserID, storedUser.ID)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(repository.updatedPasswordHash), []byte("new-password")); err != nil {
		t.Fatalf("updated password hash does not match new password: %v", err)
	}

	if repository.updatedPasswordHash == storedUser.Password {
		t.Fatal("updated password hash should not reuse the old hash")
	}
}

func TestChangePasswordRejectsInvalidCurrentPassword(t *testing.T) {
	t.Parallel()

	storedUser := testUser(t, "current-password")
	repository := &fakeUserRepository{storedUser: storedUser}
	tokenGenerator := testTokenGenerator(t)
	service := NewService(repository, tokenGenerator)
	accessToken := testAccessToken(t, tokenGenerator, storedUser)

	err := service.ChangePassword(context.Background(), accessToken, "wrong-password", "new-password")
	if !errors.Is(err, user.ErrInvalidCredentials) {
		t.Fatalf("ChangePassword() error = %v, want %v", err, user.ErrInvalidCredentials)
	}

	if repository.updatePasswordCalled {
		t.Fatal("password should not be updated when current password is invalid")
	}
}

func TestChangePasswordRejectsInvalidNewPassword(t *testing.T) {
	t.Parallel()

	storedUser := testUser(t, "current-password")
	repository := &fakeUserRepository{storedUser: storedUser}
	tokenGenerator := testTokenGenerator(t)
	service := NewService(repository, tokenGenerator)
	accessToken := testAccessToken(t, tokenGenerator, storedUser)

	err := service.ChangePassword(context.Background(), accessToken, "current-password", "short")
	if !errors.Is(err, ErrPasswordTooShort) {
		t.Fatalf("ChangePassword() error = %v, want %v", err, ErrPasswordTooShort)
	}

	if repository.updatePasswordCalled {
		t.Fatal("password should not be updated when new password is invalid")
	}
}

func TestChangePasswordRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	repository := &fakeUserRepository{storedUser: testUser(t, "current-password")}
	tokenGenerator := testTokenGenerator(t)
	service := NewService(repository, tokenGenerator)

	err := service.ChangePassword(context.Background(), "not-a-token", "current-password", "new-password")
	if !errors.Is(err, user.ErrInvalidToken) {
		t.Fatalf("ChangePassword() error = %v, want %v", err, user.ErrInvalidToken)
	}

	if repository.findByIDCalled {
		t.Fatal("user repository should not be called when token is invalid")
	}
}

func TestRequestPasswordResetCreatesHashedToken(t *testing.T) {
	t.Parallel()

	storedUser := testUser(t, "current-password")
	repository := &fakeUserRepository{storedUser: storedUser}
	service := NewService(repository, testTokenGenerator(t))

	ticket, err := service.RequestPasswordReset(context.Background(), " MARTIN@EXAMPLE.COM ")
	if err != nil {
		t.Fatalf("RequestPasswordReset() error = %v", err)
	}

	if ticket.Token == "" {
		t.Fatal("expected reset token to be returned to the delivery layer")
	}

	if !repository.createPasswordResetTokenCalled {
		t.Fatal("expected password reset token to be created")
	}

	createdToken := repository.createdPasswordResetToken
	if createdToken.UserID != storedUser.ID {
		t.Fatalf("reset token user ID = %s, want %s", createdToken.UserID, storedUser.ID)
	}

	if createdToken.TokenHash != hashPasswordResetToken(ticket.Token) {
		t.Fatal("stored reset token hash does not match returned token")
	}

	if createdToken.TokenHash == ticket.Token {
		t.Fatal("raw reset token should not be stored")
	}

	if !createdToken.ExpiresAt.Equal(ticket.ExpiresAt) {
		t.Fatalf("created token expiry = %s, want %s", createdToken.ExpiresAt, ticket.ExpiresAt)
	}

	if createdToken.ExpiresAt.Sub(createdToken.CreatedAt) != PasswordResetTokenDuration {
		t.Fatalf("reset token duration = %s, want %s", createdToken.ExpiresAt.Sub(createdToken.CreatedAt), PasswordResetTokenDuration)
	}
}

func TestRequestPasswordResetDoesNotRevealUnknownEmail(t *testing.T) {
	t.Parallel()

	repository := &fakeUserRepository{storedUser: testUser(t, "current-password")}
	service := NewService(repository, testTokenGenerator(t))

	ticket, err := service.RequestPasswordReset(context.Background(), "unknown@example.com")
	if err != nil {
		t.Fatalf("RequestPasswordReset() error = %v", err)
	}

	if ticket.Token != "" {
		t.Fatal("unknown email should not produce a reset token")
	}

	if repository.createPasswordResetTokenCalled {
		t.Fatal("password reset token should not be created for an unknown email")
	}
}

func TestRequestPasswordResetDoesNotCreateTokenForInactiveUser(t *testing.T) {
	t.Parallel()

	storedUser := testUser(t, "current-password")
	storedUser.Status = user.StatusInactive
	repository := &fakeUserRepository{storedUser: storedUser}
	service := NewService(repository, testTokenGenerator(t))

	ticket, err := service.RequestPasswordReset(context.Background(), storedUser.Email)
	if err != nil {
		t.Fatalf("RequestPasswordReset() error = %v", err)
	}

	if ticket.Token != "" {
		t.Fatal("inactive user should not produce a reset token")
	}

	if repository.createPasswordResetTokenCalled {
		t.Fatal("password reset token should not be created for an inactive user")
	}
}

func TestResetPasswordUsesTokenHashAndUpdatesPassword(t *testing.T) {
	t.Parallel()

	repository := &fakeUserRepository{storedUser: testUser(t, "current-password")}
	service := NewService(repository, testTokenGenerator(t))

	err := service.ResetPassword(context.Background(), "reset-token", "new-password")
	if err != nil {
		t.Fatalf("ResetPassword() error = %v", err)
	}

	if !repository.resetPasswordWithTokenCalled {
		t.Fatal("expected password reset to be called")
	}

	if repository.resetPasswordTokenHash != hashPasswordResetToken("reset-token") {
		t.Fatal("password reset should use the token hash")
	}

	if repository.resetPasswordTokenHash == "reset-token" {
		t.Fatal("raw reset token should not be sent to the repository")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(repository.resetPasswordHash), []byte("new-password")); err != nil {
		t.Fatalf("reset password hash does not match new password: %v", err)
	}

	if repository.resetPasswordNow.IsZero() {
		t.Fatal("expected reset timestamp to be passed to repository")
	}
}

func TestResetPasswordRejectsMissingToken(t *testing.T) {
	t.Parallel()

	repository := &fakeUserRepository{storedUser: testUser(t, "current-password")}
	service := NewService(repository, testTokenGenerator(t))

	err := service.ResetPassword(context.Background(), "   ", "new-password")
	if !errors.Is(err, user.ErrInvalidPasswordResetToken) {
		t.Fatalf("ResetPassword() error = %v, want %v", err, user.ErrInvalidPasswordResetToken)
	}

	if repository.resetPasswordWithTokenCalled {
		t.Fatal("password reset should not be called when token is missing")
	}
}

func TestResetPasswordRejectsInvalidNewPassword(t *testing.T) {
	t.Parallel()

	repository := &fakeUserRepository{storedUser: testUser(t, "current-password")}
	service := NewService(repository, testTokenGenerator(t))

	err := service.ResetPassword(context.Background(), "reset-token", "short")
	if !errors.Is(err, ErrPasswordTooShort) {
		t.Fatalf("ResetPassword() error = %v, want %v", err, ErrPasswordTooShort)
	}

	if repository.resetPasswordWithTokenCalled {
		t.Fatal("password reset should not be called when new password is invalid")
	}
}

type fakeUserRepository struct {
	storedUser user.User

	findByIDCalled                 bool
	updatePasswordCalled           bool
	updatedUserID                  uuid.UUID
	updatedPasswordHash            string
	createPasswordResetTokenCalled bool
	createdPasswordResetToken      user.PasswordResetToken
	resetPasswordWithTokenCalled   bool
	resetPasswordTokenHash         string
	resetPasswordHash              string
	resetPasswordNow               time.Time
}

func (repository *fakeUserRepository) Create(ctx context.Context, newUser user.User) (user.User, error) {
	return newUser, nil
}

func (repository *fakeUserRepository) FindByID(ctx context.Context, id uuid.UUID) (user.User, error) {
	repository.findByIDCalled = true

	if id != repository.storedUser.ID {
		return user.User{}, user.ErrUserNotFound
	}

	return repository.storedUser, nil
}

func (repository *fakeUserRepository) FindByEmail(ctx context.Context, email string) (user.User, error) {
	if email != repository.storedUser.Email {
		return user.User{}, user.ErrUserNotFound
	}

	return repository.storedUser, nil
}

func (repository *fakeUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	repository.updatePasswordCalled = true
	repository.updatedUserID = id
	repository.updatedPasswordHash = passwordHash

	return nil
}

func (repository *fakeUserRepository) CreatePasswordResetToken(ctx context.Context, resetToken user.PasswordResetToken) error {
	repository.createPasswordResetTokenCalled = true
	repository.createdPasswordResetToken = resetToken

	return nil
}

func (repository *fakeUserRepository) ResetPasswordWithToken(ctx context.Context, tokenHash string, passwordHash string, now time.Time) error {
	repository.resetPasswordWithTokenCalled = true
	repository.resetPasswordTokenHash = tokenHash
	repository.resetPasswordHash = passwordHash
	repository.resetPasswordNow = now

	return nil
}

func testUser(t *testing.T, password string) user.User {
	t.Helper()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	return user.User{
		ID:          uuid.New(),
		DisplayName: "Martin",
		Email:       "martin@example.com",
		Password:    string(passwordHash),
		Status:      user.StatusActive,
	}
}

func testTokenGenerator(t *testing.T) *TokenGenerator {
	t.Helper()

	tokenGenerator, err := NewTokenGenerator("test-secret", "proxibet-test", 15*time.Minute)
	if err != nil {
		t.Fatalf("NewTokenGenerator() error = %v", err)
	}

	return tokenGenerator
}

func testAccessToken(t *testing.T, tokenGenerator *TokenGenerator, authenticatedUser user.User) string {
	t.Helper()

	accessToken, _, err := tokenGenerator.Generate(authenticatedUser)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	return accessToken
}
