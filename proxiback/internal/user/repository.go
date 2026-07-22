package user

import (
	"context"

	"github.com/google/uuid"
)

type UpdateProfile struct {
	DisplayName *string
	Email       *string
}

type Repository interface {
	Create(ctx context.Context, user User) (User, error)
	FindByID(ctx context.Context, id uuid.UUID) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, update UpdateProfile) (User, error)
}
