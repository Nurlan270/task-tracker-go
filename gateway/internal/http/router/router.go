package router

import (
	"github.com/gin-gonic/gin"

	errs "github.com/Nurlan270/task-tracker-go/gateway/internal/errors"
	httplib "github.com/Nurlan270/task-tracker-go/gateway/internal/http"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/handlers"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/middleware"
)

type Router struct {
	*gin.Engine

	//	Middlewares
	authMW *middleware.AuthMiddleware

	//	Handlers
	authH *handlers.AuthHandler
	taskH *handlers.TaskHandler
}

func New(
	r *gin.Engine,
	authMW *middleware.AuthMiddleware,
	authH *handlers.AuthHandler,
	taskH *handlers.TaskHandler,
) *Router {
	return &Router{
		Engine: r,
		authMW: authMW,
		authH:  authH,
		taskH:  taskH,
	}
}

func (r *Router) InitRouter() {
	r.setGlobalMiddlewares()

	//	Custom 404 handling
	r.NoRoute(func(c *gin.Context) {
		httplib.AbortWithErr(c, errs.ErrNotFound)
	})

	//	Custom 405 handling
	r.NoMethod(func(c *gin.Context) {
		httplib.AbortWithErr(c, errs.ErrMethodNotAllowed)
	})

	r.registerRoutes()
}

func (r *Router) setGlobalMiddlewares() {
	//	Global middlewares
	r.Use(
		middleware.Logger(),
		middleware.ErrorHandler(),
		gin.CustomRecovery(middleware.Recover),
	)
}

func (r *Router) registerRoutes() {
	//	API v1 Routes
	apiV1 := r.Group("/api/v1")

	r.registerAuthV1Routes(apiV1)
	r.registerTaskV1Routes(apiV1)
}

func (r *Router) registerAuthV1Routes(v1 *gin.RouterGroup) {
	//	Auth v1 Routes
	auth := v1.Group("/auth")

	//	Guested routes (unauthorized access)
	guested := auth.Group("/", middleware.RequireGuest())
	guested.POST("register", r.authH.Register)
	guested.POST("login", r.authH.Login)

	//	Authorized routes (need valid access token)
	authorized := auth.Group("/", r.authMW.RequireAuth())
	authorized.GET("user", r.authH.User)

	//	Semi-authorized routes (need valid refresh token)
	auth.POST("logout", r.authH.Logout)
	auth.POST("refresh", r.authH.Refresh)
}

func (r *Router) registerTaskV1Routes(v1 *gin.RouterGroup) {
	//	Task v1 Routes (authorized access only)
	task := v1.Group("/task", r.authMW.RequireAuth())

	task.GET("/", r.taskH.List)
	task.POST("/", r.taskH.Create)
	task.PATCH("/:id", r.taskH.Update)
	task.DELETE("/:id", r.taskH.Delete)
	task.PATCH("/:id/completed", r.taskH.MarkTaskAsCompleted)
}
