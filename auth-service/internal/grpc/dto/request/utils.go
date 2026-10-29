package request

import (
	"errors"

	errs "github.com/Nurlan270/task-tracker-go/auth-service/internal/errors"

	valid "github.com/go-ozzo/ozzo-validation/v4"
)

// handleValidationErr wraps validation error with ErrInternal
// if validation fails with internal error of package.
func handleValidationErr(err error) error {
	if err != nil {
		var ie valid.InternalError
		if errors.As(err, &ie) {
			//	an internal error occurred
			return errors.Join(errs.ErrInternal, err)
		}

		return err
	}

	return nil
}
