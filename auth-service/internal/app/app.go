package app

import (
	"context"
	"net"
	"os/signal"
	"syscall"

	"github.com/Nurlan270/task-tracker-go/common/config"
	"github.com/Nurlan270/task-tracker-go/common/logger"
	"go.uber.org/zap"

	appcfg "github.com/Nurlan270/task-tracker-go/auth-service/internal/config"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/models"

	authV1 "github.com/Nurlan270/task-tracker-go/protos/gen/proto/auth/v1"
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
		a.migrateDB,
		a.initGRPC,
	}

	for _, init := range deps {
		init()
	}
}

func (a *App) initLogger() {
	logger.Init(a.cfg)
}

func (a *App) migrateDB() {
	log := logger.Get().SetOp("migrateDB")

	err := a.di.DB().AutoMigrate(
		&models.User{}, &models.RefreshToken{},
	)
	if err != nil {
		log.Fatal("failed to migrate database", zap.Error(err))
	}
}

func (a *App) initGRPC() {
	grpcSrv := a.di.GRPCServer()

	authV1.RegisterAuthServiceServer(grpcSrv, a.di.AuthService())
}

func (a *App) Run() {
	log := logger.Get().SetOp("Run")

	listener, err := net.Listen("tcp", a.cfg.GRPC.GetAddr())
	if err != nil {
		log.Fatal("failed to listen on tcp", zap.Error(err))
	}

	grpcSrv := a.di.GRPCServer()

	grpcSrvErr := make(chan error, 1)

	go func() {
		log.Info("starting gRPC server", zap.String("address", a.cfg.GRPC.GetAddr()))

		if err = grpcSrv.Serve(listener); err != nil {
			grpcSrvErr <- err
			return
		}

		grpcSrvErr <- nil
	}()

	//	Graceful shutdown handling
	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-signalCtx.Done():
		log.Info("received SIGINT/SIGTERM, shutting down gRPC server")

		//	Double Ctrl+C pattern.
		//	Second Ctrl+C will close server immediately.
		stop()
	case err = <-grpcSrvErr:
		log.Error("failed to serve gRPC", zap.Error(err))
	}

	//	Shutting down gRPC server
	grpcSrv.GracefulStop()

	//	here we'll close all other services
}
