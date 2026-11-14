package app

import (
	"fmt"
	"net/http"

	"github.com/Nurlan270/task-tracker-go/common/closer"
	"github.com/Nurlan270/task-tracker-go/common/validator"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Nurlan270/task-tracker-go/gateway/internal/config"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/handlers"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/middleware"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/router"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/service"

	userV1 "github.com/Nurlan270/task-tracker-go/protos/gen/proto/user/v1"
	gpv "github.com/go-playground/validator/v10"
)

type diContainer struct {
	//	Configuration
	cfg *config.Config

	//	HTTP Core
	srv    *http.Server
	router *router.Router

	//	HTTP Handlers
	authHandler *handlers.AuthHandler

	//	HTTP Middlewares
	authMW *middleware.AuthMiddleware

	//	Gin
	gin *gin.Engine

	//	Validator
	validator *validator.Validate

	//	gRPC Core
	userGRPCConn *grpc.ClientConn

	//	gRPC Clients
	authGRPCClient userV1.AuthServiceClient

	//	Services
	authSvc *service.AuthService
}

func newDIContainer(cfg *config.Config) diContainer {
	return diContainer{cfg: cfg}
}

func (c *diContainer) Server() *http.Server {
	if c.srv == nil {
		c.srv = &http.Server{
			Addr:         c.cfg.HTTP.GetAddr(),
			Handler:      c.Gin(),
			ReadTimeout:  c.cfg.HTTP.ReadTimeout,
			WriteTimeout: c.cfg.HTTP.WriteTimeout,
		}
	}

	return c.srv
}

func (c *diContainer) Gin() *gin.Engine {
	if c.gin == nil {
		engine := gin.New()
		engine.HandleMethodNotAllowed = true

		c.gin = engine
	}

	return c.gin
}

func (c *diContainer) UserGRPCConn() *grpc.ClientConn {
	if c.userGRPCConn == nil {
		conn, err := grpc.NewClient(
			c.cfg.UserService.GetAddr(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to connect to user service: %s", err))
		}

		closer.Add("gRPC / User Service Client", func() error {
			return conn.Close()
		})

		c.userGRPCConn = conn
	}

	return c.userGRPCConn
}

func (c *diContainer) AuthGRPCClient() userV1.AuthServiceClient {
	if c.authGRPCClient == nil {
		c.authGRPCClient = userV1.NewAuthServiceClient(c.UserGRPCConn())
	}

	return c.authGRPCClient
}

func (c *diContainer) Validator() *validator.Validate {
	if c.validator == nil {
		c.validator = validator.New(gpv.WithRequiredStructEnabled())
	}

	return c.validator
}

func (c *diContainer) AuthService() *service.AuthService {
	if c.authSvc == nil {
		c.authSvc = service.NewAuthService(c.AuthGRPCClient())
	}

	return c.authSvc
}

func (c *diContainer) AuthHandler() *handlers.AuthHandler {
	if c.authHandler == nil {
		c.authHandler = handlers.NewAuthHandler(c.Validator(), c.AuthService())
	}

	return c.authHandler
}

func (c *diContainer) AuthMW() *middleware.AuthMiddleware {
	if c.authMW == nil {
		c.authMW = middleware.NewAuthMiddleware(c.AuthService())
	}

	return c.authMW
}

func (c *diContainer) Router() *router.Router {
	if c.router == nil {
		c.router = router.New(c.Gin(), c.AuthMW(), c.AuthHandler())
	}

	return c.router
}
