package handlers

import (
	"errors"
	"github.com/Nurlan270/task-tracker-go/common/logger"
	"github.com/Nurlan270/task-tracker-go/common/validator"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/request"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/response"
	httplib "github.com/Nurlan270/task-tracker-go/gateway/internal/http"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/message"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
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
	log := logger.Get().SetOp("Register")

	var body request.Register

	//	Bind request body
	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, response.NewError(message.ErrBadRequest))
		return
	}

	//	Validate request body
	if err := h.validator.Struct(&body); err != nil {
		var validationErr *validator.ErrValidation

		if errors.As(err, &validationErr) {
			c.AbortWithStatusJSON(http.StatusBadRequest, response.NewErrorMap(validationErr))
			return
		}

		log.Error("failed to validate request", zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(message.ErrInternal))
		return
	}

	//	Register
	resp, err := h.authSvc.Register(c, &body)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok || st.Code() == codes.Internal {
			log.Error("failed to register user", zap.Error(err))

			c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(message.ErrInternal))
			return
		}

		switch st.Code() {
		case codes.InvalidArgument:
			c.AbortWithStatusJSON(http.StatusBadRequest, response.NewErrorMapFromGRPC(st.Err()))
		case codes.AlreadyExists:
			c.AbortWithStatusJSON(http.StatusConflict, response.NewError(st.Message()))
		default:
			log.Error("failed to register user", zap.Error(err))

			c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(message.ErrInternal))
		}

		return
	}

	//	Flush old token cookies if any
	httplib.FlushTokenCookies(c)

	//	Set token cookies
	httplib.SetTokenCookies(c, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)

	c.Status(http.StatusCreated)
}

func (h *AuthHandler) Login(c *gin.Context) {
	log := logger.Get().SetOp("Login")

	var body request.Login

	//	Bind request body
	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, response.NewError(message.ErrBadRequest))
		return
	}

	//	Validate request body
	if err := h.validator.Struct(&body); err != nil {
		var validationErr *validator.ErrValidation

		if errors.As(err, &validationErr) {
			c.AbortWithStatusJSON(http.StatusBadRequest, response.NewErrorMap(validationErr))
			return
		}

		log.Error("failed to validate request", zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(message.ErrInternal))
		return
	}

	//	Login
	resp, err := h.authSvc.Login(c, &body)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok || st.Code() == codes.Internal {
			log.Error("failed to register user", zap.Error(err))

			c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(message.ErrInternal))
			return
		}

		switch st.Code() {
		case codes.InvalidArgument:
			c.AbortWithStatusJSON(http.StatusBadRequest, response.NewErrorMapFromGRPC(st.Err()))
		case codes.Unauthenticated:
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewError(st.Message()))
		default:
			log.Error("failed to register user", zap.Error(err))

			c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(message.ErrInternal))
		}

		return
	}

	//	Flush old token cookies if any
	httplib.FlushTokenCookies(c)

	//	Set token cookies
	httplib.SetTokenCookies(c, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)

	c.Status(http.StatusOK)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	log := logger.Get().SetOp("Logout")

	//	Fetch refresh token from cookie
	refresh, err := c.Cookie(httplib.RefreshCookieName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewError(message.ErrUnauthorized))
		return
	}

	body := request.Logout{
		RefreshToken: refresh,
	}

	//	Bind request body
	if err = c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, response.NewError(message.ErrBadRequest))
		return
	}

	//	Validate request body
	if err = h.validator.Struct(&body); err != nil {
		var validationErr *validator.ErrValidation

		if errors.As(err, &validationErr) {
			c.AbortWithStatusJSON(http.StatusBadRequest, response.NewErrorMap(validationErr))
			return
		}

		log.Error("failed to validate request", zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(message.ErrInternal))
		return
	}

	//	Logout
	if err = h.authSvc.Logout(c, &body); err != nil {
		st, ok := status.FromError(err)
		if !ok || st.Code() == codes.Internal {
			log.Error("failed to register user", zap.Error(err))

			c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(message.ErrInternal))
			return
		}

		switch st.Code() {
		case codes.InvalidArgument:
			c.AbortWithStatusJSON(http.StatusBadRequest, response.NewErrorMapFromGRPC(st.Err()))
		case codes.Unauthenticated:
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewError(st.Message()))
		default:
			log.Error("failed to register user", zap.Error(err))

			c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(message.ErrInternal))
		}

		return
	}

	//	Flush token cookies
	httplib.FlushTokenCookies(c)

	c.Status(http.StatusOK)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	log := logger.Get().SetOp("Refresh")

	//	Fetch refresh token from cookie
	refresh, err := c.Cookie(httplib.RefreshCookieName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewError(message.ErrUnauthorized))
		return
	}

	body := request.Refresh{
		RefreshToken: refresh,
	}

	//	Bind request body
	if err = c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, response.NewError(message.ErrBadRequest))
		return
	}

	//	Validate request body
	if err = h.validator.Struct(&body); err != nil {
		var validationErr *validator.ErrValidation

		if errors.As(err, &validationErr) {
			c.AbortWithStatusJSON(http.StatusBadRequest, response.NewErrorMap(validationErr))
			return
		}

		log.Error("failed to validate request", zap.Error(err))

		c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(message.ErrInternal))
		return
	}

	//	Refresh
	resp, err := h.authSvc.Refresh(c, &body)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok || st.Code() == codes.Internal {
			log.Error("failed to register user", zap.Error(err))

			c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(message.ErrInternal))
			return
		}

		switch st.Code() {
		case codes.InvalidArgument:
			c.AbortWithStatusJSON(http.StatusBadRequest, response.NewErrorMapFromGRPC(st.Err()))
		case codes.Unauthenticated:
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.NewError(st.Message()))
		default:
			log.Error("failed to register user", zap.Error(err))

			c.AbortWithStatusJSON(http.StatusInternalServerError, response.NewError(message.ErrInternal))
		}

		return
	}

	//	Flush old token cookies if any
	httplib.FlushTokenCookies(c)

	//	Set token cookies
	httplib.SetTokenCookies(c, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)

	c.Status(http.StatusCreated)
}
