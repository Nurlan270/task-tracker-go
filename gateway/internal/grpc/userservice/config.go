package userservice

import (
	"fmt"
)

type Config struct {
	Host string `env:"USER_SERVICE_HOST" env-default:"user-service" yaml:"host"`
	Port uint16 `env:"USER_SERVICE_PORT" env-default:"7070"         yaml:"port"`
}

func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
