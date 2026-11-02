package interceptors

import (
	"runtime/debug"

	"github.com/Nurlan270/task-tracker-go/common/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Nurlan270/task-tracker-go/auth-service/internal/grpc/message"
)

func PanicRecoveryHandler(p any) error {
	//	Log panic
	logger.Get().SetOp("PanicRecoveryHandler").
		Error("recovered from panic",
			zap.Any("panic", p), zap.ByteString("stack", debug.Stack()),
		)

	return status.Errorf(codes.Internal, "%s", message.ErrInternal)
}
