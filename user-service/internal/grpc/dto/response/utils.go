package response

import (
	"errors"

	"github.com/Nurlan270/task-tracker-go/common/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errdet "google.golang.org/genproto/googleapis/rpc/errdetails"
)

func ToGRPCValidationErr(err error) error {
	var validationErr *validator.ErrValidation

	if !errors.As(err, &validationErr) {
		return err
	}

	violations := make([]*errdet.BadRequest_FieldViolation, 0, len(validationErr.GetErrors()))

	for tag, msg := range validationErr.GetErrors() {
		violations = append(violations, &errdet.BadRequest_FieldViolation{
			Field:       tag,
			Description: msg,
		})
	}

	st := status.New(codes.InvalidArgument, "validation failed")

	st, detailErr := st.WithDetails(&errdet.BadRequest{
		FieldViolations: violations,
	})
	if detailErr != nil {
		return status.Error(codes.Internal, "failed to encode validation errors")
	}

	return st.Err()
}
