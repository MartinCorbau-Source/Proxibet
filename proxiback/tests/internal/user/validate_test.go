package user

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestValidateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		user    User
		wantErr error
	}{
		{
			name: "valid user",
			user: User{
				ID:           uuid.New(),
				Email:        "martin@example.com",
				DisplayName:  "Martin",
				PasswordHash: "hashed-password",
				Status:       StatusActive,
			},
		},
		{
			name: "invalid email",
			user: User{
				ID:           uuid.New(),
				Email:        "invalid-email",
				DisplayName:  "Martin",
				PasswordHash: "hashed-password",
				Status:       StatusActive,
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "display name too short",
			user: User{
				ID:           uuid.New(),
				Email:        "martin@example.com",
				DisplayName:  "M",
				PasswordHash: "hashed-password",
				Status:       StatusActive,
			},
			wantErr: ErrInvalidDisplayName,
		},
		{
			name: "missing password hash",
			user: User{
				ID:          uuid.New(),
				Email:       "martin@example.com",
				DisplayName: "Martin",
				Status:      StatusActive,
			},
			wantErr: ErrMissingPasswordHash,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := Validate(test.user)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf(
					"Validate() error = %v, want %v",
					err,
					test.wantErr,
				)
			}
		})
	}
}