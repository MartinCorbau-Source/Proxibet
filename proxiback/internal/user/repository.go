package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, user User) (User, error)
	FindByID(ctx context.Context, id uuid.UUID) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	CreatePasswordResetToken(ctx context.Context, resetToken PasswordResetToken) error
	ResetPasswordWithToken(ctx context.Context, tokenHash string, passwordHash string, now time.Time) error
}
