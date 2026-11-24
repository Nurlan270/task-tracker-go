package router

import (
	"time"

	"github.com/gin-gonic/gin"

	errs "github.com/Nurlan270/task-tracker-go/gateway/internal/errors"
	httplib "github.com/Nurlan270/task-tracker-go/gateway/internal/http"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/handlers"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/middleware"
)

type Router struct {
	*gin.Engine

	//	Middlewares
	authMW    *middleware.AuthMiddleware
	limiterMW *middleware.RateLimiterMiddleware

	//	Handlers
	authH *handlers.AuthHandler
	taskH *handlers.TaskHandler
}

func New(
	r *gin.Engine,
	authMW *middleware.AuthMiddleware,
	limiterMW *middleware.RateLimiterMiddleware,
	authH *handlers.AuthHandler,
	taskH *handlers.TaskHandler,
) *Router {
	return &Router{
		Engine:    r,
		authMW:    authMW,
		limiterMW: limiterMW,
		authH:     authH,
		taskH:     taskH,
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
	guested := auth.Group(
		"/",
		middleware.RequireGuest(),
		//	Rate limit: 20 reqs/min
		r.limiterMW.Limit(middleware.GroupKey("guest"), 20, 20, time.Minute),
	)
	guested.POST("register", r.authH.Register)
	guested.POST("login", r.authH.Login)

	//	Semi-authorized routes (need valid refresh token)
	semiAuth := auth.Group("/",
		//	Rate limit: 10 reqs/min
		r.limiterMW.Limit(middleware.GroupKey("semiAuth"), 10, 10, time.Minute),
	)
	semiAuth.POST("logout", r.authH.Logout)
	semiAuth.POST("refresh", r.authH.Refresh)

	//	Authorized routes (need valid access token)
	authorized := auth.Group("/",
		r.authMW.RequireAuth(),
		//	Rate limit: 100 reqs every 3 minutes
		r.limiterMW.Limit(middleware.GroupKey("auth"), 100, 100, 3*time.Minute),
	)
	authorized.GET("user", r.authH.User)
}

func (r *Router) registerTaskV1Routes(v1 *gin.RouterGroup) {
	//	Task v1 Routes (authorized access only)
	task := v1.Group("/tasks",
		r.authMW.RequireAuth(),
		//	Rate limit: 100 reqs every 5 minutes
		r.limiterMW.Limit(middleware.GroupKey("tasks"), 100, 100, 5*time.Minute),
	)

	task.GET("/", r.taskH.List)
	task.POST("/", r.taskH.Create)
	task.PATCH("/:id", r.taskH.Update)
	task.DELETE("/:id", r.taskH.MoveToTrash)
	task.DELETE("/", r.taskH.DeletePermanently)
	task.PATCH("/:id/restore", r.taskH.RestoreFromTrash)
	task.PATCH("/:id/toggle-completion", r.taskH.ToggleCompletion)
}
