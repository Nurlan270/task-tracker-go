package validator

import "fmt"

type ErrValidation struct {
	errors map[string]string
}

func (e *ErrValidation) Error() string {
	return fmt.Sprintf("validation errors: %s", e.errors)
}

func (e *ErrValidation) GetErrors() map[string]string {
	return e.errors
}
