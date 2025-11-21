package domain

import "errors"

var (
	ErrUserExists         = errors.New("user exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEventNotFound      = errors.New("event not found")
)
