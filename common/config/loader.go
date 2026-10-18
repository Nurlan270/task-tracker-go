package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config interface {
	GetAppEnv() string
	GetAppName() string
}

// MustLoad fills Config with data from configuration file and .env file.
func MustLoad[C Config]() *C {
	if err := godotenv.Load(); err != nil {
		panic(fmt.Sprintf("failed to load env file: %s", err))
	}

	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		panic("CONFIG_PATH environment variable not set")
	}

	var cfg C

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic(fmt.Sprintf("failed to read config: %s", err))
	}

	return &cfg
}
