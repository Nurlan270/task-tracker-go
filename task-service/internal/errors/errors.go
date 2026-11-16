package errors

import (
	"errors"
)

var (
	ErrTaskNotFound = errors.New("task not found")

	ErrInternal = errors.New("internal error")
)
