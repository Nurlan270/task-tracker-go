package app

import (
	"github.com/Nurlan270/task-tracker-go/common/closer"
	"github.com/Nurlan270/task-tracker-go/common/database"
	"github.com/Nurlan270/task-tracker-go/common/hasher"
	"github.com/Nurlan270/task-tracker-go/common/logger"
	"github.com/Nurlan270/task-tracker-go/common/validator"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/Nurlan270/task-tracker-go/auth-service/internal/config"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/grpc/interceptors"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/jwt"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/repository"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/repository/postgres"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/service"
	cr "github.com/Nurlan270/task-tracker-go/auth-service/internal/validator"

	gpv "github.com/go-playground/validator/v10"
)

type diContainer struct {
	//	Configuration
	cfg *config.Config

	//	Core
	db           *gorm.DB
	hasher       hasher.Hasher
	tokenManager *jwt.TokenManager
	validator    *validator.Validate

	//	GRPC
	grpcSrv *grpc.Server

	//	Services
	authSvc *service.AuthService

	//	Repositories
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
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

		valid.RegisterCustomRules(cr.NameCustomRule())
		valid.RegisterCustomMessages(cr.Messages)

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

func (c *diContainer) UserRepo() repository.UserRepository {
	if c.userRepo == nil {
		c.userRepo = postgres.NewUserRepository(c.DB())
	}

	return c.userRepo
}

func (c *diContainer) TokenRepo() repository.TokenRepository {
	if c.tokenRepo == nil {
		c.tokenRepo = postgres.NewTokenRepository(c.DB())
	}

	return c.tokenRepo
}

func (c *diContainer) AuthService() *service.AuthService {
	if c.authSvc == nil {
		c.authSvc = service.NewAuthService(
			c.Validator(),
			c.TokenManager(),
			c.Hasher(),
			c.DB(),
			c.UserRepo(),
			c.TokenRepo(),
		)
	}

	return c.authSvc
}

func (c *diContainer) Hasher() hasher.Hasher {
	if c.hasher == nil {
		c.hasher = hasher.NewArgonHasher(nil)
	}

	return c.hasher
}

func (c *diContainer) TokenManager() *jwt.TokenManager {
	if c.tokenManager == nil {
		c.tokenManager = jwt.NewTokenManager(c.cfg.JWT)
	}

	return c.tokenManager
}
