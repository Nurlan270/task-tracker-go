package database

import "fmt"

type Config struct {
	Host         string `env:"DB_HOST"           env-default:"localhost"    yaml:"host"`
	Port         uint16 `env:"DB_PORT"           env-default:"5432"         yaml:"port"`
	Name         string `env:"DB_NAME"           env-default:"task-tracker" yaml:"name"`
	User         string `env:"DB_USER"           env-default:"postgres"     yaml:"user"`
	Password     string `env:"DB_PASSWORD"                                  yaml:"password"       env-required:"true"`
	MaxOpenConns int    `env:"DB_MAX_OPEN_CONNS" env-default:"35"           yaml:"max_open_conns"`
	MaxIdleConns int    `env:"DB_MAX_IDLE_CONNS" env-default:"15"           yaml:"max_idle_conns"`
}

func (c Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name,
	)
}

func (c Config) GetMaxOpenConns() int {
	return c.MaxOpenConns
}

func (c Config) GetMaxIdleConns() int {
	return c.MaxIdleConns
}
