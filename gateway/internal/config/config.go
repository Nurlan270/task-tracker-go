package config

import (
	"github.com/Nurlan270/task-tracker-go/common/database"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/grpc/auth_service"
	"github.com/Nurlan270/task-tracker-go/gateway/internal/http"
)

type Config struct {
	App         AppConfig           `yaml:"app"`
	Database    database.Config     `yaml:"database"`
	HTTP        http.Config         `yaml:"http_server"`
	AuthService auth_service.Config `yaml:"auth_service"`
}

type AppConfig struct {
	Name string `env:"APP_NAME" env-default:"auth-service" yaml:"name"`
	Env  string `env:"APP_ENV"  env-default:"dev"          yaml:"env"`
}

func (c *Config) GetAppName() string {
	return c.App.Name
}

func (c *Config) GetAppEnv() string {
	return c.App.Env
}
