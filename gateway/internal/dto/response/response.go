package response

import (
	"errors"
	"github.com/Nurlan270/task-tracker-go/common/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type Response struct {
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

func New(msg string) *Response {
	return &Response{Message: msg}
}

func (r *Response) SetValidationErrors(err error) *Response {
	var vErr *validator.ErrValidation

	if !errors.As(err, &vErr) {
		return r
	}

	r.Errors = vErr.GetErrors()

	return r
}

func (r *Response) SetGRPCErrors(err error) *Response {
	st, ok := status.FromError(err)
	if !ok {
		return nil
	}

	r.Errors = make(map[string]string, len(st.Details()))

	for _, detail := range st.Details() {
		badRequest, ok := detail.(*errdetails.BadRequest)
		if !ok {
			continue
		}

		for _, violation := range badRequest.FieldViolations {
			r.Errors[violation.Field] = violation.Description
		}
	}

	return r
}
