package middleware

import (
	"errors"
	"github.com/Nurlan270/task-tracker-go/common/logger"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/response"
	errs "github.com/Nurlan270/task-tracker-go/gateway/internal/errors"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/message"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() //	Process the request first

		ctx := c.Request.Context()
		if ctx != nil && ctx.Err() != nil {
			logger.Get().
				Error("Request context has error. Nothing will be written to response",
					zap.Error(ctx.Err()))
			return
		}

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		st, ok := status.FromError(err)
		if ok {
			//	gRPC error
			handleGRPC(c, st)
		} else {
			//	HTTP error
			handleHTTP(c, err)
		}
	}
}

func handleHTTP(c *gin.Context, err error) {
	log := logger.Get().SetOp("Handle HTTP")

	switch {
	case errors.Is(err, errs.Unauthorized):
		c.AbortWithStatusJSON(http.StatusUnauthorized, response.New(message.ErrUnauthorized))
	case errors.Is(err, errs.Forbidden):
		c.AbortWithStatusJSON(http.StatusForbidden, response.New(message.ErrForbidden))
	case errors.Is(err, errs.BadRequest):
		c.AbortWithStatusJSON(http.StatusBadRequest, response.New(message.ErrBadRequest).SetValidationErrors(err))
	case errors.Is(err, errs.NotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, response.New(message.ErrNotFound))
	case errors.Is(err, errs.MethodNotAllowed):
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, response.New(message.ErrMethodNotAllowed))
	case errors.Is(err, errs.InternalServer):
		log.Error("Internal server error", zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, response.New(message.ErrInternal))
	default:
		log.Error("Unexpected HTTP error", zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, response.New(message.ErrInternal))
	}
}

func handleGRPC(c *gin.Context, st *status.Status) {
	log := logger.Get().SetOp("Handle gRPC")

	switch st.Code() {
	case codes.InvalidArgument:
		c.AbortWithStatusJSON(http.StatusBadRequest, response.New(message.ErrBadRequest).SetGRPCErrors(st.Err()))
	case codes.AlreadyExists:
		c.AbortWithStatusJSON(http.StatusConflict, response.New(st.Message()))
	case codes.Unauthenticated:
		c.AbortWithStatusJSON(http.StatusUnauthorized, response.New(st.Message()))
	default:
		log.Error("Internal gRPC error", zap.Error(st.Err()))

		c.AbortWithStatusJSON(http.StatusInternalServerError, response.New(message.ErrInternal))
	}
}
