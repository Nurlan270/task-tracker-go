package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"time"
)

const (
	AccessCookieName  = "_access_token"
	RefreshCookieName = "_refresh_token"

	accessCookiePath  = "/"
	refreshCookiePath = "/api/v1/auth"
)

// SetTokenCookies sets token cookies with specified values
func SetTokenCookies(c *gin.Context, accessToken, refreshToken string, expiresIn time.Duration) {
	//	Set access token
	c.SetCookieData(&http.Cookie{
		Name:     AccessCookieName,
		Value:    url.QueryEscape(accessToken),
		Path:     accessCookiePath,
		Domain:   "",
		Expires:  time.Now().Add(expiresIn),
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	//	Set refresh token
	c.SetCookieData(&http.Cookie{
		Name:     RefreshCookieName,
		Value:    url.QueryEscape(refreshToken),
		Path:     refreshCookiePath,
		Domain:   "",
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// FlushTokenCookies deletes token cookies
func FlushTokenCookies(c *gin.Context) {
	for _, cookie := range []struct {
		name string
		path string
	}{
		{AccessCookieName, accessCookiePath},
		{RefreshCookieName, refreshCookiePath},
	} {
		c.SetCookieData(&http.Cookie{
			Name:     cookie.name,
			Value:    "",
			Path:     cookie.path,
			Domain:   "",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			Secure:   false,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
	}
}

// AbortWithErr attaches passed error to the context and aborts
func AbortWithErr(c *gin.Context, err error) {
	c.Error(err)
	c.Abort()
}
