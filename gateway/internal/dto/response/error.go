package response

import (
	"errors"
	"github.com/Nurlan270/task-tracker-go/common/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type Error struct {
	Message string `json:"message"`
}

func NewError(msg string) *Error {
	return &Error{Message: msg}
}

type ErrorMap struct {
	Errors map[string]string `json:"errors"`
}

func NewErrorMap(err error) *ErrorMap {
	var vErr *validator.ErrValidation

	if !errors.As(err, &vErr) {
		return nil
	}

	return &ErrorMap{
		Errors: vErr.GetErrors(),
	}
}

func NewErrorMapFromGRPC(err error) *ErrorMap {
	st, ok := status.FromError(err)
	if !ok {
		return nil
	}

	errorMap := &ErrorMap{
		Errors: make(map[string]string, len(st.Details())),
	}

	for _, detail := range st.Details() {
		badRequest, ok := detail.(*errdetails.BadRequest)
		if !ok {
			continue
		}

		for _, violation := range badRequest.FieldViolations {
			errorMap.Errors[violation.Field] = violation.Description
		}
	}

	if len(errorMap.Errors) == 0 {
		return nil
	}

	return errorMap
}
