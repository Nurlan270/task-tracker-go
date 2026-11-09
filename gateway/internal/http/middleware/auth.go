package middleware

import (
	"github.com/Nurlan270/task-tracker-go/common/logger"
	"github.com/Nurlan270/task-tracker-go/gateway/entity"
	ctxlib "github.com/Nurlan270/task-tracker-go/gateway/internal/context"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/response"
	httplib "github.com/Nurlan270/task-tracker-go/gateway/internal/http"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/message"
	authV1 "github.com/Nurlan270/task-tracker-go/protos/gen/proto/auth/v1"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"strings"
)

type AuthMiddleware struct {
	//fixme: inject auth service NOT client directly
	client authV1.AuthServiceClient
}

func NewAuthMiddleware(client authV1.AuthServiceClient) *AuthMiddleware {
	return &AuthMiddleware{client: client}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	log := logger.Get().SetOp("Require Auth")

	return func(c *gin.Context) {
		//	Fetch access token from cookie
		accessToken, err := c.Cookie(httplib.AccessCookieName)
		if err != nil || strings.TrimSpace(accessToken) == "" {
			//	Token does not exist or empty
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewError(message.ErrUnauthorized))
			return
		}

		//	Fetch refresh token from cookie
		refreshToken, err := c.Cookie(httplib.RefreshCookieName)
		if err != nil || strings.TrimSpace(refreshToken) == "" {
			//	Token does not exist or empty
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewError(message.ErrUnauthorized))
			return
		}

		//	Call gRPC client and send token
		resp, err := m.client.User(c, &authV1.UserRequest{
			AccessToken: accessToken,
		})
		if err != nil {
			//fixme: we should convert any gRPC error into some internal error and pass to handlers that way
			st, ok := status.FromError(err)
			if !ok || st.Code() == codes.Internal {
				log.Error("failed to get user", zap.Error(err))

				c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(message.ErrInternal))
				return
			}

			switch st.Code() {
			case codes.InvalidArgument:
				c.AbortWithStatusJSON(http.StatusBadRequest, response.NewErrorMapFromGRPC(st.Err()))
			case codes.Unauthenticated:
				c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewError(st.Message()))
			default:
				log.Error("failed to get user", zap.Error(err))

				c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(message.ErrInternal))
			}

			return
		}

		//	Set user into context
		c.Set(ctxlib.UserCtxKey, &entity.User{
			ID:    uint(resp.Id),
			Name:  resp.Name,
			Email: resp.Email,
		})

		//	Set refresh token into context
		c.Set(ctxlib.RefreshTokenCtxKey, refreshToken)

		//	Call next handler
		c.Next()
	}
}
