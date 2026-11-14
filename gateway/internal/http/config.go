package http

import (
	"fmt"
	"time"
)

type Config struct {
	Host         string        `env:"HTTP_HOST"          env-default:"0.0.0.0" yaml:"host"`
	Port         uint16        `env:"HTTP_PORT"          env-default:"80"      yaml:"port"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT"  env-default:"30s"     yaml:"read_timeout"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"30s"     yaml:"write_timeout"`
}

func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
