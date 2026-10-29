package app

import (
	"github.com/Nurlan270/task-tracker-go/common/database"
	"github.com/Nurlan270/task-tracker-go/common/hasher"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/Nurlan270/task-tracker-go/auth-service/internal/config"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/jwt"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/repository"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/repository/postgres"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/service"
)

type diContainer struct {
	//	App configuration
	cfg *config.Config

	//	Core
	db           *gorm.DB
	hasher       hasher.Hasher
	tokenManager *jwt.TokenManager

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
		c.db = database.MustNew(c.cfg.Database)
	}

	return c.db
}

func (c *diContainer) GRPCServer() *grpc.Server {
	if c.grpcSrv == nil {
		//todo: add interceptors
		c.grpcSrv = grpc.NewServer(grpc.ConnectionTimeout(c.cfg.GRPC.Timeout))
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
		c.authSvc = service.NewAuthService(c.TokenManager(), c.Hasher(), c.DB(), c.UserRepo(), c.TokenRepo())
	}

	return c.authSvc
}

func (c *diContainer) Hasher() hasher.Hasher {
	if c.hasher == nil {
		c.hasher = hasher.NewArgonHasher()
	}

	return c.hasher
}

func (c *diContainer) TokenManager() *jwt.TokenManager {
	if c.tokenManager == nil {
		c.tokenManager = jwt.NewTokenManager(c.cfg.JWT)
	}

	return c.tokenManager
}
