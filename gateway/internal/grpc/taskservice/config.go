package taskservice

import (
	"fmt"
)

type Config struct {
	Host string `env:"TASK_SERVICE_HOST" env-default:"task-service" yaml:"host"`
	Port uint16 `env:"TASK_SERVICE_PORT" env-default:"6060"         yaml:"port"`
}

func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
