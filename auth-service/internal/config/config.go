package config

import (
	"github.com/Nurlan270/task-tracker-go/common/database"

	"github.com/Nurlan270/task-tracker-go/auth-service/internal/grpc"
	"github.com/Nurlan270/task-tracker-go/auth-service/internal/jwt"
)

type Config struct {
	App      AppConfig       `yaml:"app"`
	Database database.Config `yaml:"database"`
	GRPC     grpc.Config     `yaml:"grpc"`
	JWT      jwt.Config      `yaml:"jwt"`
}

type AppConfig struct {
	Name string `env:"APP_NAME" env-default:"auth-service" yaml:"name"`
	Env  string `env:"APP_ENV"  env-default:"dev"          yaml:"env"`
}

func (c Config) GetAppName() string {
	return c.App.Name
}

func (c Config) GetAppEnv() string {
	return c.App.Env
}
