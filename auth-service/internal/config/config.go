package config

import "github.com/Nurlan270/task-tracker-go/auth-service/internal/database"

type Config struct {
	App      AppConfig       `env-prefix:"APP_" yaml:"app"`
	Database database.Config `env-prefix:"DB_"  yaml:"database"`
}

type AppConfig struct {
	Name string `env:"NAME" env-default:"auth-service" yaml:"name"`
	Env  string `env:"ENV"  env-default:"dev"          yaml:"env"`
}

func (c Config) GetAppName() string {
	return c.App.Name
}

func (c Config) GetAppEnv() string {
	return c.App.Env
}
