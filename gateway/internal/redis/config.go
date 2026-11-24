package redis

import "fmt"

type Config struct {
	Host     string `env:"REDIS_HOST"     env-default:"redis" yaml:"host"`
	Port     uint16 `env:"REDIS_PORT"     env-default:"6379"  yaml:"port"`
	Password string `env:"REDIS_PASSWORD"                     yaml:"password" env-required:"true"`
	DB       int    `env:"REDIS_DB"       env-default:"0"     yaml:"db"`
}

func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
