package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const refreshTokenBytes = 32

var (
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrRefreshTokenRevoked = errors.New("refresh token revoked")
)

type RefreshToken struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	TokenHash         string
	ExpiresAt         time.Time
	RevokedAt         *time.Time
	ReplacedByTokenID *uuid.UUID
	CreatedAt         time.Time
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, refreshToken RefreshToken) (RefreshToken, error)
	Rotate(ctx context.Context, tokenHash string, nextRefreshToken RefreshToken, revokedAt time.Time) (RefreshToken, error)
}

func newRefreshToken(userID uuid.UUID, duration time.Duration) (string, RefreshToken, error) {
	if duration <= 0 {
		return "", RefreshToken{}, fmt.Errorf("invalid refresh token duration: %s", duration)
	}

	tokenBytes := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", RefreshToken{}, fmt.Errorf("generate refresh token: %w", err)
	}

	plainToken := base64.RawURLEncoding.EncodeToString(tokenBytes)

	return plainToken, RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hashRefreshToken(plainToken),
		ExpiresAt: time.Now().UTC().Add(duration),
	}, nil
}

func hashRefreshToken(refreshToken string) string {
	tokenHash := sha256.Sum256([]byte(refreshToken))
	return hex.EncodeToString(tokenHash[:])
}

func normalizeRefreshToken(refreshToken string) (string, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return "", ErrInvalidRefreshToken
	}

	return refreshToken, nil
}
