package errors

import (
	"errors"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")

	ErrRefreshTokenNotFound = errors.New("refresh token not found")

	ErrInternal = errors.New("internal error")
)
