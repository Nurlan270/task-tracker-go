package http

import (
	"fmt"
	"time"
)

type Config struct {
	Host         string        `yaml:"host" env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port         uint16        `yaml:"port" env:"HTTP_PORT" env-default:"80"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"HTTP_READ_TIMEOUT" env-default:"30s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"HTTP_WRITE_TIMEOUT" env-default:"30s"`
}

func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
