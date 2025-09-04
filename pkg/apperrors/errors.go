package apperrors

import "errors"

var (
	// common errors
	ErrNotFound     = errors.New("resource not found")
	ErrValidation   = errors.New("validation error")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")

	// specific business errors
	ErrUserExists   = errors.New("user already exists")
	ErrUserUnderAge = errors.New("user under age")
	ErrInvalidToken = errors.New("invalid token")
)
