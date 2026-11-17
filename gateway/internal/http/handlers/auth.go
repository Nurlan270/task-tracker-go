package handlers

import (
	"errors"
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

//nolint:dupl
func (h *AuthHandler) Register(c *gin.Context) {
	var body request.Register

	//	Bind request body
	if err := c.ShouldBindJSON(&body); err != nil {
		httplib.AbortWithErr(c, errs.ErrBadRequest)
		return
	}

	//	Validate request body
	if err := h.validator.Struct(&body); err != nil {
		var validationErr *validator.ErrValidation

		if errors.As(err, &validationErr) {
			httplib.AbortWithErr(c, errors.Join(errs.ErrBadRequest, validationErr))
			return
		}

		httplib.AbortWithErr(c, errors.Join(errs.ErrInternalServer, err))

		return
	}

	//	Register
	resp, err := h.authSvc.Register(c.Request.Context(), &body)
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

//nolint:dupl
func (h *AuthHandler) Login(c *gin.Context) {
	var body request.Login

	//fixme: move duplicating validation code to separate function

	//	Bind request body
	if err := c.ShouldBindJSON(&body); err != nil {
		httplib.AbortWithErr(c, errs.ErrBadRequest)
		return
	}

	//	Validate request body
	if err := h.validator.Struct(&body); err != nil {
		var validationErr *validator.ErrValidation

		if errors.As(err, &validationErr) {
			httplib.AbortWithErr(c, errors.Join(errs.ErrBadRequest, validationErr))
			return
		}

		httplib.AbortWithErr(c, errors.Join(errs.ErrInternalServer, err))

		return
	}

	//	Login
	resp, err := h.authSvc.Login(c.Request.Context(), &body)
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
	//	Fetch refresh token from cookie
	token, err := c.Cookie(httplib.RefreshCookieName)
	if err != nil || strings.TrimSpace(token) == "" {
		//	Token does not exist or empty
		httplib.AbortWithErr(c, errs.ErrUnauthorized)
		return
	}

	body := request.Logout{
		RefreshToken: token,
	}

	//	Logout
	err = h.authSvc.Logout(c.Request.Context(), &body)
	if err != nil {
		httplib.AbortWithErr(c, err)
		return
	}

	//	Flush token cookies
	httplib.FlushTokenCookies(c)

	c.Status(http.StatusOK)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	//	Fetch refresh token from cookie
	token, err := c.Cookie(httplib.RefreshCookieName)
	if err != nil || strings.TrimSpace(token) == "" {
		//	Token does not exist or empty
		httplib.AbortWithErr(c, errs.ErrUnauthorized)
		return
	}

	body := request.Refresh{
		RefreshToken: token,
	}

	//	Refresh
	resp, err := h.authSvc.Refresh(c.Request.Context(), &body)
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
