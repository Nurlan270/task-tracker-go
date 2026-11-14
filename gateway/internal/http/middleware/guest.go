package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	errs "github.com/Nurlan270/task-tracker-go/gateway/internal/errors"
	httplib "github.com/Nurlan270/task-tracker-go/gateway/internal/http"
)

func RequireGuest() gin.HandlerFunc {
	return func(c *gin.Context) {
		//	Fetch access token from cookie
		if accessToken, err := c.Cookie(
			httplib.AccessCookieName,
		); err == nil &&
			strings.TrimSpace(accessToken) != "" {
			//	Token exist and not empty
			httplib.AbortWithErr(c, errs.ErrForbidden)
			return
		}

		//	Fetch refresh token from cookie
		if refreshToken, err := c.Cookie(
			httplib.RefreshCookieName,
		); err == nil &&
			strings.TrimSpace(refreshToken) != "" {
			//	Token exist and not empty
			httplib.AbortWithErr(c, errs.ErrForbidden)
			return
		}

		c.Next()
	}
}
