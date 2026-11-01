package database

import "fmt"

type Config struct {
	Host     string `env:"DB_HOST"     env-default:"localhost"    yaml:"host"`
	Port     uint16 `env:"DB_PORT"     env-default:"5432"         yaml:"port"`
	Name     string `env:"DB_NAME"     env-default:"task-tracker" yaml:"name"`
	User     string `env:"DB_USER"     env-default:"postgres"     yaml:"user"`
	Password string `env:"DB_PASSWORD"                            yaml:"password" env-required:"true"`
}

func (c Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name,
	)
}
