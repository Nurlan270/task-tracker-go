package config

import (
	"github.com/Nurlan270/task-tracker-go/gateway/internal/grpc/user_service"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http"
)

type Config struct {
	App         AppConfig           `yaml:"app"`
	HTTP        http.Config         `yaml:"http_server"`
	UserService user_service.Config `yaml:"user-service"`
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
