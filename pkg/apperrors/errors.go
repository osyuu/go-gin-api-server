package apperrors

import "errors"

var (
	// common errors
	ErrNotFound     = errors.New("resource not found")
	ErrValidation   = errors.New("validation error")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")

	// user errors
	ErrUserExists   = errors.New("user already exists")
	ErrUserUnderAge = errors.New("user under age")

	// auth errors
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")

	// post errors
	ErrPostContentTooLong        = errors.New("post content too long")
	ErrPostContentTooShort       = errors.New("post content too short")
	ErrPostContentSensitiveWords = errors.New("post content contains sensitive words")
)
