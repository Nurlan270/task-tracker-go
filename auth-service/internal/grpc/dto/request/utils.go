package request

import (
	"errors"

	errs "github.com/Nurlan270/task-tracker-go/auth-service/internal/errors"

	valid "github.com/go-ozzo/ozzo-validation/v4"
)

// handleValidationErr wraps error into ErrValidation if error is validation related.
// Otherwise, if error is due to internal error of validator package error will be returned as is.
func handleValidationErr(err error) error {
	if err != nil {
		var ie valid.InternalError
		if errors.As(err, &ie) {
			//	an internal error occurred
			return err
		}

		//	validation error
		return errs.ErrValidation{
			Message: err.Error(),
		}
	}

	return nil
}
