package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/user"
)

func TestTokenGeneratorValidate(t *testing.T) {
	t.Parallel()

	generator, err := NewTokenGenerator("secret", "proxibet-api", time.Hour)
	if err != nil {
		t.Fatalf("NewTokenGenerator() error = %v", err)
	}

	authenticatedUser := user.User{
		ID:          uuid.New(),
		DisplayName: "Martin",
		Email:       "martin@example.com",
	}

	token, _, err := generator.Generate(authenticatedUser)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	userID, err := generator.Validate(token)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if userID != authenticatedUser.ID {
		t.Fatalf("Validate() userID = %s, want %s", userID, authenticatedUser.ID)
	}
}

func TestTokenGeneratorValidateRejectsExpiredToken(t *testing.T) {
	t.Parallel()

	generator, err := NewTokenGenerator("secret", "proxibet-api", time.Hour)
	if err != nil {
		t.Fatalf("NewTokenGenerator() error = %v", err)
	}

	now := time.Now().UTC()
	expiredClaims := AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "proxibet-api",
			Subject:   uuid.New().String(),
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	_, err = generator.Validate(token)
	if !errors.Is(err, ErrInvalidAccessToken) {
		t.Fatalf("Validate() error = %v, want %v", err, ErrInvalidAccessToken)
	}
}
