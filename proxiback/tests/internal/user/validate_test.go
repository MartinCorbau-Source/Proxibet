package user_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/MartinCorbau-Source/proxibet/proxiback/internal/user"
)

func TestValidateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		user    user.User
		wantErr error
	}{
		{
			name: "valid user",
			user: user.User{
				ID:          uuid.New(),
				Email:       "martin@example.com",
				DisplayName: "Martin",
				Password:    "hashed-password",
				Status:      user.StatusActive,
			},
		},
		{
			name: "invalid email",
			user: user.User{
				ID:          uuid.New(),
				Email:       "invalid-email",
				DisplayName: "Martin",
				Password:    "hashed-password",
				Status:      user.StatusActive,
			},
			wantErr: user.ErrInvalidUserEmail,
		},
		{
			name: "display name too short",
			user: user.User{
				ID:          uuid.New(),
				Email:       "martin@example.com",
				DisplayName: "M",
				Password:    "hashed-password",
				Status:      user.StatusActive,
			},
			wantErr: user.ErrInvalidDisplayName,
		},
		{
			name: "missing password hash",
			user: user.User{
				ID:          uuid.New(),
				Email:       "martin@example.com",
				DisplayName: "Martin",
				Status:      user.StatusActive,
			},
			wantErr: user.ErrMissingPassword,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := user.Validate(test.user)

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
