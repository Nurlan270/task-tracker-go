package app

import (
	"github.com/Nurlan270/task-tracker-go/common/closer"
	"github.com/Nurlan270/task-tracker-go/common/database"
	"github.com/Nurlan270/task-tracker-go/common/logger"
	"github.com/Nurlan270/task-tracker-go/common/validator"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/Nurlan270/task-tracker-go/task-service/internal/config"
	"github.com/Nurlan270/task-tracker-go/task-service/internal/grpc/interceptors"
	"github.com/Nurlan270/task-tracker-go/task-service/internal/repository"
	"github.com/Nurlan270/task-tracker-go/task-service/internal/repository/postgres"
	"github.com/Nurlan270/task-tracker-go/task-service/internal/service"

	gpv "github.com/go-playground/validator/v10"
)

type diContainer struct {
	//	Configuration
	cfg *config.Config

	//	Core
	db        *gorm.DB
	validator *validator.Validate

	//	GRPC
	grpcSrv *grpc.Server

	//	Services
	taskSvc *service.TaskService

	//	Repositories
	taskRepo repository.TaskRepository
}

func newDIContainer(cfg *config.Config) diContainer {
	return diContainer{cfg: cfg}
}

func (c *diContainer) DB() *gorm.DB {
	if c.db == nil {
		db, sqlDB := database.MustNew(c.cfg.Database)

		c.db = db

		closer.Add("Database", func() error {
			return sqlDB.Close()
		})
	}

	return c.db
}

func (c *diContainer) Validator() *validator.Validate {
	if c.validator == nil {
		valid := validator.New(gpv.WithRequiredStructEnabled())

		c.validator = valid
	}

	return c.validator
}

func (c *diContainer) GRPCServer() *grpc.Server {
	if c.grpcSrv == nil {
		c.grpcSrv = grpc.NewServer(
			grpc.ConnectionTimeout(c.cfg.GRPC.Timeout),
			//	Interceptors
			grpc.ChainUnaryInterceptor(
				//	Logger
				logging.UnaryServerInterceptor(
					interceptors.LoggerInterceptor(logger.Get()),
				),
				//	Recovery
				recovery.UnaryServerInterceptor(
					recovery.WithRecoveryHandler(interceptors.PanicRecoveryHandler),
				),
			),
		)
	}

	return c.grpcSrv
}

func (c *diContainer) TaskRepo() repository.TaskRepository {
	if c.taskRepo == nil {
		c.taskRepo = postgres.NewTaskRepository(c.DB())
	}

	return c.taskRepo
}

func (c *diContainer) TaskService() *service.TaskService {
	if c.taskSvc == nil {
		c.taskSvc = service.NewTaskService(c.Validator(), c.TaskRepo())
	}

	return c.taskSvc
}
