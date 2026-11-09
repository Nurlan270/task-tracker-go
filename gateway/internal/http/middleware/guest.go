package middleware

import (
	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/response"
	httplib "github.com/Nurlan270/task-tracker-go/gateway/internal/http"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/message"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func RequireGuest() gin.HandlerFunc {
	return func(c *gin.Context) {
		//	Fetch access token from cookie
		if accessToken, err := c.Cookie(httplib.AccessCookieName); err == nil && strings.TrimSpace(accessToken) != "" {
			//	Token exist and not empty
			c.AbortWithStatusJSON(http.StatusForbidden, response.NewError(message.ErrForbidden))
			return
		}

		//	Fetch refresh token from cookie
		if refreshToken, err := c.Cookie(httplib.RefreshCookieName); err == nil && strings.TrimSpace(refreshToken) != "" {
			//	Token exist and not empty
			c.AbortWithStatusJSON(http.StatusForbidden, response.NewError(message.ErrForbidden))
			return
		}

		c.Next()
	}
}
