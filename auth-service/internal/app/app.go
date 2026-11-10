package app

import (
	"context"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nurlan270/task-tracker-go/common/closer"
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

	closer.Add("Logger", func() error {
		return logger.Get().Close()
	})
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

func (a *App) startCleanupWorker(ctx context.Context) {
	log := logger.Get().SetOp("Start Cleanup Worker")

	//	Worker will try to remove all expired refresh tokens every 24 hours
	ticker := time.NewTicker(24 * time.Hour)

	go func() {
		defer ticker.Stop()

		log.Info("Starting cleanup worker")

		for {
			select {
			case <-ctx.Done():
				log.Info("Shutting down cleanup worker")
				return

			case <-ticker.C:
				deleted, err := a.di.TokenRepo().DeleteExpiredOrRevokedTokens(ctx)
				if err != nil {
					log.Error("failed to delete expired tokens", zap.Error(err))
					continue
				}

				log.Info("deleted expired tokens", zap.Int("count", deleted))
			}
		}
	}()
}

func (a *App) startGRPCServer(ctx context.Context) {
	log := logger.Get().SetOp("Start gRPC Server")

	listener, err := net.Listen("tcp", a.cfg.GRPC.GetAddr())
	if err != nil {
		log.Fatal("failed to listen on tcp", zap.Error(err))
	}

	grpcSrv := a.di.GRPCServer()

	grpcSrvErr := make(chan error, 1)

	go func() {
		log.Info("Starting gRPC server", zap.String("address", a.cfg.GRPC.GetAddr()))

		if err = grpcSrv.Serve(listener); err != nil {
			grpcSrvErr <- err
			return
		}

		grpcSrvErr <- nil
	}()

	select {
	case <-ctx.Done():
		log.Info("Shutting down gRPC server")

		//	Shutting down gRPC server
		grpcSrv.GracefulStop()
	case err = <-grpcSrvErr:
		log.Error("failed to serve gRPC", zap.Error(err))
	}
}

func (a *App) Run() {
	log := logger.Get().SetOp("Run")

	log.Info("Starting application")

	//	Catch signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	//	Start components
	a.startCleanupWorker(ctx)
	a.startGRPCServer(ctx)

	//	Close all opened resources
	closer.CloseAll(ctx)
}
