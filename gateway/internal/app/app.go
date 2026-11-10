package app

import (
	"context"
	"errors"
	"github.com/Nurlan270/task-tracker-go/common/closer"
	"github.com/Nurlan270/task-tracker-go/common/config"
	"github.com/Nurlan270/task-tracker-go/common/logger"
	appcfg "github.com/Nurlan270/task-tracker-go/gateway/internal/config"
	"go.uber.org/zap"
	"net/http"
	"os/signal"
	"syscall"
	"time"
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
		a.initRouter,
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

func (a *App) initRouter() {
	a.di.Router().InitRouter()
}

func (a *App) Run() {
	log := logger.Get().SetOp("Run")

	log.Info("Starting application")

	//	Catch signals
	sig, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := a.di.Server()
	srvErr := make(chan error, 1)

	go func() {
		log.Info("Starting HTTP server", zap.String("address", a.cfg.HTTP.GetAddr()))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
	}()

	select {
	case <-sig.Done():
		//	Received SIGINT/SIGTERM signal
		log.Info("Received SIGINT/SIGTERM signal. Shutting down application")

		//	Second signal will terminate app immediately
		stop()
	case err := <-srvErr:
		//	Server closed unexpectedly
		log.Error("Error while starting HTTP server", zap.Error(err))
	}

	srvCtx, srvCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer srvCancel()

	//	Shutdown HTTP server
	if err := srv.Shutdown(srvCtx); err != nil {
		log.Error("Failed to shutdown HTTP server", zap.Error(err))
	}

	closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer closeCancel()

	//	Close all opened resources
	closer.CloseAll(closeCtx)
}
