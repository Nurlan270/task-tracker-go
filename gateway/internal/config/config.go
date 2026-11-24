package config

import (
	"github.com/Nurlan270/task-tracker-go/gateway/internal/grpc/taskservice"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/grpc/userservice"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/redis"
)

type Config struct {
	App         AppConfig          `yaml:"app"`
	HTTP        http.Config        `yaml:"http_server"`
	Redis       redis.Config       `yaml:"redis"`
	UserService userservice.Config `yaml:"user_service"`
	TaskService taskservice.Config `yaml:"task_service"`
}

type AppConfig struct {
	Name string `env:"APP_NAME" env-default:"user-service" yaml:"name"`
	Env  string `env:"APP_ENV"  env-default:"dev"          yaml:"env"`
}

func (c *Config) GetAppName() string {
	return c.App.Name
}

func (c *Config) GetAppEnv() string {
	return c.App.Env
}
