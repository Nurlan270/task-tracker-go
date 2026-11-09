package router

import (
	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/response"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/handlers"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/message"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Router struct {
	*gin.Engine

	//	Middlewares
	authMW *middleware.AuthMiddleware

	//	Handlers
	authH *handlers.AuthHandler
}

func New(r *gin.Engine, authMW *middleware.AuthMiddleware, authH *handlers.AuthHandler) *Router {
	return &Router{
		Engine: r,
		authMW: authMW,
		authH:  authH,
	}
}

func (r *Router) InitRouter() {
	r.setGlobalMiddlewares()

	//	Custom 404 handling
	r.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, response.NewError(message.ErrNotFound))
	})

	//	Custom 405 handling
	r.NoMethod(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, response.NewError(message.ErrMethodNotAllowed))
	})

	//	API v1 Routes
	apiV1 := r.Group("/api/v1")

	r.registerAuthV1Routes(apiV1)
}

func (r *Router) setGlobalMiddlewares() {
	//	Global middlewares
	r.Use(
		middleware.Logger(),
		gin.CustomRecovery(middleware.Recover),
	)
}

func (r *Router) registerAuthV1Routes(v1 *gin.RouterGroup) {
	//	Auth v1 Routes
	auth := v1.Group("/auth")

	//	Guested routes (unauthorized access)
	guested := auth.Group("/", middleware.RequireGuest())
	guested.POST("register", r.authH.Register)
	guested.POST("login", r.authH.Login)

	//	Authorized routes
	authorized := auth.Group("/", r.authMW.RequireAuth())
	authorized.GET("user", r.authH.User)
	authorized.POST("logout", r.authH.Logout)
	authorized.POST("refresh", r.authH.Refresh)
}
