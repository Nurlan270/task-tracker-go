package errors

import (
	"errors"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")

	ErrInternal = errors.New("internal error")
)

type ErrValidation struct {
	Message string
}

func (e ErrValidation) Error() string {
	return e.Message
}

// IsValidation checks whether error is validation related
// if yes it will return true, otherwise false.
func IsValidation(err error) bool {
	var vErr ErrValidation
	return errors.As(err, &vErr)
}
