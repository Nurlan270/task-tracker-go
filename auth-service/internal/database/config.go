package database

import "fmt"

type Config struct {
	Host     string `env:"HOST"     env-default:"localhost"    yaml:"host"`
	Port     uint16 `env:"PORT"     env-default:"5432"         yaml:"port"`
	Name     string `env:"NAME"     env-default:"task-tracker" yaml:"name"`
	User     string `env:"USER"     env-default:"postgres"     yaml:"user"`
	Password string `env:"PASSWORD"                            yaml:"password" env-required:"true"`
}

func (c Config) GetConnString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable pool_max_conns=25",
		c.Host, c.Port, c.User, c.Password, c.Name,
	)
}
