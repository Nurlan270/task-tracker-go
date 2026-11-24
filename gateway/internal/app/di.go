package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Nurlan270/task-tracker-go/common/closer"
	"github.com/Nurlan270/task-tracker-go/common/validator"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Nurlan270/task-tracker-go/gateway/internal/config"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/handlers"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/middleware"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http/router"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/ratelimiter"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/service"
	cv "github.com/Nurlan270/task-tracker-go/gateway/internal/validator"

	taskV1 "github.com/Nurlan270/task-tracker-go/protos/gen/proto/task/v1"
	userV1 "github.com/Nurlan270/task-tracker-go/protos/gen/proto/user/v1"
	gpv "github.com/go-playground/validator/v10"
)

type diContainer struct {
	//	Configuration
	cfg *config.Config

	//	Gin
	gin *gin.Engine

	//	Redis
	rdb *redis.Client

	//	Rate limiter
	rateLimiter *ratelimiter.TokenBucket

	//	Validator
	validator *validator.Validate

	//	HTTP Core
	srv    *http.Server
	router *router.Router

	//	HTTP Handlers
	authHandler *handlers.AuthHandler
	taskHandler *handlers.TaskHandler

	//	HTTP Middlewares
	authMW    *middleware.AuthMiddleware
	limiterMW *middleware.RateLimiterMiddleware

	//	gRPC Core
	userGRPCConn *grpc.ClientConn
	taskGRPCConn *grpc.ClientConn

	//	gRPC Clients
	authGRPCClient userV1.AuthServiceClient
	taskGRPCClient taskV1.TaskServiceClient

	//	Services
	authSvc *service.AuthService
	taskSvc *service.TaskService
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

func (c *diContainer) TaskGRPCConn() *grpc.ClientConn {
	if c.taskGRPCConn == nil {
		conn, err := grpc.NewClient(
			c.cfg.TaskService.GetAddr(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to connect to task service: %s", err))
		}

		closer.Add("gRPC / Task Service Client", func() error {
			return conn.Close()
		})

		c.taskGRPCConn = conn
	}

	return c.taskGRPCConn
}

func (c *diContainer) AuthGRPCClient() userV1.AuthServiceClient {
	if c.authGRPCClient == nil {
		c.authGRPCClient = userV1.NewAuthServiceClient(c.UserGRPCConn())
	}

	return c.authGRPCClient
}

func (c *diContainer) TaskGRPCClient() taskV1.TaskServiceClient {
	if c.taskGRPCClient == nil {
		c.taskGRPCClient = taskV1.NewTaskServiceClient(c.TaskGRPCConn())
	}

	return c.taskGRPCClient
}

func (c *diContainer) Validator() *validator.Validate {
	if c.validator == nil {
		v := validator.New(gpv.WithRequiredStructEnabled())

		v.RegisterCustomRules(cv.TaskTypeCustomRule())
		v.RegisterCustomMessages(cv.Messages)

		c.validator = v
	}

	return c.validator
}

func (c *diContainer) AuthService() *service.AuthService {
	if c.authSvc == nil {
		c.authSvc = service.NewAuthService(c.AuthGRPCClient())
	}

	return c.authSvc
}

func (c *diContainer) TaskService() *service.TaskService {
	if c.taskSvc == nil {
		c.taskSvc = service.NewTaskService(c.TaskGRPCClient())
	}

	return c.taskSvc
}

func (c *diContainer) AuthHandler() *handlers.AuthHandler {
	if c.authHandler == nil {
		c.authHandler = handlers.NewAuthHandler(c.Validator(), c.AuthService())
	}

	return c.authHandler
}

func (c *diContainer) TaskHandler() *handlers.TaskHandler {
	if c.taskHandler == nil {
		c.taskHandler = handlers.NewTaskHandler(c.Validator(), c.TaskService())
	}

	return c.taskHandler
}

func (c *diContainer) AuthMW() *middleware.AuthMiddleware {
	if c.authMW == nil {
		c.authMW = middleware.NewAuthMiddleware(c.AuthService())
	}

	return c.authMW
}

func (c *diContainer) RateLimiter() *ratelimiter.TokenBucket {
	if c.rateLimiter == nil {
		c.rateLimiter = ratelimiter.NewTokenBucket(c.RDB())
	}

	return c.rateLimiter
}

func (c *diContainer) LimiterMW() *middleware.RateLimiterMiddleware {
	if c.limiterMW == nil {
		c.limiterMW = middleware.NewRateLimiterMiddleware(c.RateLimiter())
	}

	return c.limiterMW
}

func (c *diContainer) RDB() *redis.Client {
	if c.rdb == nil {
		rdb := redis.NewClient(&redis.Options{
			Addr:     c.cfg.Redis.GetAddr(),
			Password: c.cfg.Redis.Password,
			DB:       c.cfg.Redis.DB,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		//	Ping redis server to ensure its connected successfully
		if _, err := rdb.Ping(ctx).Result(); err != nil {
			panic(fmt.Sprintf("failed to connect to redis server: %s", err))
		}

		closer.Add("Redis client", func() error {
			return rdb.Close()
		})

		c.rdb = rdb
	}

	return c.rdb
}

func (c *diContainer) Router() *router.Router {
	if c.router == nil {
		c.router = router.New(c.Gin(), c.AuthMW(), c.LimiterMW(), c.AuthHandler(), c.TaskHandler())
	}

	return c.router
}
