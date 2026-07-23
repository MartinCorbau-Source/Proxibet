package user

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidUserStatus  = errors.New("invalid user status")
	ErrInvalidUserEmail   = errors.New("invalid user email")
	ErrEmailAlreadyInUse  = errors.New("email already in use")
	ErrMissingPassword    = errors.New("password cannot be empty")
	ErrInvalidDisplayName = errors.New("invalid display name")

	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrAccountInactive           = errors.New("account is inactive")
	ErrInvalidToken              = errors.New("invalid token")
	ErrInvalidPasswordResetToken = errors.New("invalid password reset token")
	ErrTokenExpired              = errors.New("token has expired")
)
