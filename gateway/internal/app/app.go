package app

import (
	"context"
	"errors"
	"github.com/Nurlan270/task-tracker-go/common/closer"
	"github.com/Nurlan270/task-tracker-go/common/config"
	"github.com/Nurlan270/task-tracker-go/common/logger"
	appcfg "github.com/Nurlan270/task-tracker-go/gateway/internal/config"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/dto/response"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/message"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"os/signal"
	"syscall"
)

type App struct {
	cfg *appcfg.Config
	di  diContainer
}

func New() *App {
	app := &App{
		cfg: config.MustLoad[appcfg.Config](),
	}

	app.di = newDIContainer(app.cfg)

	app.initDeps()

	return app
}

func (a *App) initDeps() {
	deps := []func(){
		a.initLogger,
		a.registerRoutes,
	}

	for _, init := range deps {
		init()
	}
}

func (a *App) initLogger() {
	logger.Init(a.cfg)

	closer.Add("Logger", func() error {
		return logger.Get().Close()
	})
}

func (a *App) registerRoutes() {
	router := a.di.Gin()

	//	Custom 404 handling
	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, response.NewError(message.ErrNotFound))
	})

	//	API
	api := router.Group("/api")
	{
		//	API v1
		v1 := api.Group("/v1")
		{
			authHandler := a.di.AuthHandler()

			//	Auth
			auth := v1.Group("/auth")
			auth.POST("register", authHandler.Register)
			auth.POST("login", authHandler.Login)
			auth.POST("logout", authHandler.Logout)
			auth.POST("refresh", authHandler.Refresh)
		}
	}
}

func (a *App) Run() {
	log := logger.Get().SetOp("Run")

	log.Info("Starting application")

	//	Catch signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("Starting HTTP server", zap.String("addr", a.cfg.HTTP.GetAddr()))

	if err := a.di.Server().ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("Error starting server", zap.Error(err))
	}

	//	Close all opened resources
	closer.CloseAll(ctx)
}
