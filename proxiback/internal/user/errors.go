package user 

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUserStatus = errors.New("invalid user status")
	ErrInvalidUserEmail = errors.New("invalid user email")
	ErrEmailAlreadyInUse = errors.New("email already in use")
	ErrMissingPassword = errors.New("password cannot be empty")
)