package response

import (
	"errors"

	"github.com/Nurlan270/task-tracker-go/common/logger"
	"github.com/Nurlan270/task-tracker-go/common/validator"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Nurlan270/task-tracker-go/user-service/internal/grpc/message"

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
		logger.Get().SetOp("grpc.dto.response.utils.ToGRPCValidationErr").
			Error("failed to encode validation errors", zap.Error(detailErr))

		return status.Error(codes.Internal, message.ErrInternal)
	}

	return st.Err()
}
