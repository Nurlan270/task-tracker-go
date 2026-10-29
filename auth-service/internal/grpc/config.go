package grpc

import (
	"strconv"
	"time"
)

type Config struct {
	Port    uint16        `env:"GRPC_PORT"    env-default:"55055" yaml:"port"`
	Timeout time.Duration `env:"GRPC_TIMEOUT" env-default:"15s"   yaml:"timeout"`
}

func (c Config) GetPort() string {
	return strconv.FormatInt(int64(c.Port), 10)
}

func (c Config) GetAddr() string {
	return ":" + c.GetPort()
}
