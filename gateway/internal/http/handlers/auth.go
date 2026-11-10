package handlers

import (
	"errors"
	"github.com/Nurlan270/task-tracker-go/common/validator"
	ctxlib "github.com/Nurlan270/task-tracker-go/gateway/internal/context"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/request"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/entity"
	errs "github.com/Nurlan270/task-tracker-go/gateway/internal/errors"
	httplib "github.com/Nurlan270/task-tracker-go/gateway/internal/http"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
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
	var body request.Register

	//	Bind request body
	if err := c.ShouldBindJSON(&body); err != nil {
		httplib.AbortWithErr(c, errs.BadRequest)
		return
	}

	//	Validate request body
	if err := h.validator.Struct(&body); err != nil {
		var validationErr *validator.ErrValidation

		if errors.As(err, &validationErr) {
			httplib.AbortWithErr(c, errors.Join(errs.BadRequest, validationErr))
			return
		}

		httplib.AbortWithErr(c, errors.Join(errs.InternalServer, err))
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

func (h *AuthHandler) Login(c *gin.Context) {
	var body request.Login

	//	Bind request body
	if err := c.ShouldBindJSON(&body); err != nil {
		httplib.AbortWithErr(c, errs.BadRequest)
		return
	}

	//	Validate request body
	if err := h.validator.Struct(&body); err != nil {
		var validationErr *validator.ErrValidation

		if errors.As(err, &validationErr) {
			httplib.AbortWithErr(c, errors.Join(errs.BadRequest, validationErr))
			return
		}

		httplib.AbortWithErr(c, errors.Join(errs.InternalServer, err))
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
		httplib.AbortWithErr(c, errs.Unauthorized)
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
		httplib.AbortWithErr(c, errs.Unauthorized)
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
