package middleware

import (
	ctxlib "github.com/Nurlan270/task-tracker-go/gateway/internal/context"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/request"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/entity"
	errs "github.com/Nurlan270/task-tracker-go/gateway/internal/errors"
	httplib "github.com/Nurlan270/task-tracker-go/gateway/internal/http"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/service"
	"github.com/gin-gonic/gin"
	"strings"
)

type AuthMiddleware struct {
	authSvc *service.AuthService
}

func NewAuthMiddleware(authSvc *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authSvc: authSvc}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		//	Fetch access token from cookie
		token, err := c.Cookie(httplib.AccessCookieName)
		if err != nil || strings.TrimSpace(token) == "" {
			//	Token does not exist or empty
			httplib.AbortWithErr(c, errs.Unauthorized)
			return
		}

		//	Get user using access token
		resp, err := m.authSvc.GetUser(c.Request.Context(), &request.GetUser{
			AccessToken: token,
		})
		if err != nil {
			httplib.AbortWithErr(c, err)
			return
		}

		//	Set user into context
		c.Set(ctxlib.UserCtxKey, &entity.User{
			ID:    resp.ID,
			Name:  resp.Name,
			Email: resp.Email,
		})

		//	Call next handler
		c.Next()
	}
}
