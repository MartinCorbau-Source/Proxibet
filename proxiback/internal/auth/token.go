package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/user"
)

var (
	ErrMissingJWTSecret   = errors.New("missing JWT secret")
	ErrInvalidAccessToken = errors.New("invalid access token")
)

type TokenGenerator struct {
	secret   []byte
	issuer   string
	duration time.Duration
}

type AccessTokenClaims struct {
	jwt.RegisteredClaims

	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

func NewTokenGenerator(secret, issuer string, duration time.Duration) (*TokenGenerator, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, ErrMissingJWTSecret
	}

	if duration <= 0 {
		return nil, fmt.Errorf("invalid token duration: %s", duration)
	}

	if strings.TrimSpace(issuer) == "" {
		return nil, fmt.Errorf("invalid issuer: %s", issuer)
	}

	return &TokenGenerator{
		secret:   []byte(secret),
		issuer:   issuer,
		duration: duration,
	}, nil
}

func (generator *TokenGenerator) Generate(authenticatedUser user.User) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(generator.duration)

	claims := AccessTokenClaims{
		Email:       authenticatedUser.Email,
		DisplayName: authenticatedUser.DisplayName,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    generator.issuer,
			Subject:   authenticatedUser.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(generator.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return signedToken, expiresAt, nil
}

func (generator *TokenGenerator) Validate(signedToken string) (uuid.UUID, error) {
	signedToken = strings.TrimSpace(signedToken)
	if signedToken == "" {
		return uuid.Nil, ErrInvalidAccessToken
	}

	claims := &AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(
		signedToken,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, ErrInvalidAccessToken
			}

			return generator.secret, nil
		},
		jwt.WithIssuer(generator.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil || token == nil || !token.Valid {
		return uuid.Nil, ErrInvalidAccessToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidAccessToken
	}

	return userID, nil
}
