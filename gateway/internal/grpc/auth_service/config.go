package auth_service

import (
	"fmt"
)

type Config struct {
	Host string `yaml:"host" env:"AUTH_SERVICE_HOST" env-default:"auth_service"`
	Port uint16 `yaml:"port" env:"AUTH_SERVICE_PORT" env-default:"7070"`
}

func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
