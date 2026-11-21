package handlers

import (
	"net/http"
	"strings"

	"github.com/Nurlan270/task-tracker-go/common/validator"
	"github.com/gin-gonic/gin"

	ctxlib "github.com/Nurlan270/task-tracker-go/gateway/internal/context"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/request"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/entity"
	errs "github.com/Nurlan270/task-tracker-go/gateway/internal/errors"
	httplib "github.com/Nurlan270/task-tracker-go/gateway/internal/http"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/service"
)

type AuthHandler struct {
	validator *validator.Validate
	authSvc   *service.AuthService
}

func NewAuthHandler(validator *validator.Validate, authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{
		validator: validator,
		authSvc:   authSvc,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	body := validateRequest[request.Register](c, h.validator, nil)
	if body == nil {
		return
	}

	//	Register
	resp, err := h.authSvc.Register(c.Request.Context(), body)
	if err != nil {
		httplib.AbortWithErr(c, err)
		return
	}

	//	Flush old token cookies if any
	httplib.FlushTokenCookies(c)

	//	Set token cookies
	httplib.SetTokenCookies(c, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)

	c.Status(http.StatusCreated)
}

func (h *AuthHandler) Login(c *gin.Context) {
	body := validateRequest[request.Login](c, h.validator, nil)
	if body == nil {
		return
	}

	//	Login
	resp, err := h.authSvc.Login(c.Request.Context(), body)
	if err != nil {
		httplib.AbortWithErr(c, err)
		return
	}

	//	Flush old token cookies if any
	httplib.FlushTokenCookies(c)

	//	Set token cookies
	httplib.SetTokenCookies(c, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)

	c.Status(http.StatusOK)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	body := validateRequest[request.Logout](c, h.validator, func(r *request.Logout) error {
		//	Fetch refresh token from cookie
		token, err := c.Cookie(httplib.RefreshCookieName)
		if err != nil || strings.TrimSpace(token) == "" {
			return errs.ErrUnauthorized
		}

		r.RefreshToken = token

		return nil
	})
	if body == nil {
		return
	}

	//	Logout
	if err := h.authSvc.Logout(c.Request.Context(), body); err != nil {
		httplib.AbortWithErr(c, err)
		return
	}

	//	Flush token cookies
	httplib.FlushTokenCookies(c)

	c.Status(http.StatusOK)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	body := validateRequest[request.Refresh](c, h.validator, func(r *request.Refresh) error {
		//	Fetch refresh token from cookie
		token, err := c.Cookie(httplib.RefreshCookieName)
		if err != nil || strings.TrimSpace(token) == "" {
			//	Token does not exist or empty
			return errs.ErrUnauthorized
		}

		r.RefreshToken = token

		return nil
	})
	if body == nil {
		return
	}

	//	Refresh
	resp, err := h.authSvc.Refresh(c.Request.Context(), body)
	if err != nil {
		httplib.AbortWithErr(c, err)
		return
	}

	//	Flush old token cookies if any
	httplib.FlushTokenCookies(c)

	//	Set token cookies
	httplib.SetTokenCookies(c, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)

	c.Status(http.StatusCreated)
}

func (h *AuthHandler) User(c *gin.Context) {
	user := c.MustGet(ctxlib.UserCtxKey).(*entity.User)

	c.JSON(http.StatusOK, user)
}
