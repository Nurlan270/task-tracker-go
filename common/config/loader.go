package config

import (
	"errors"
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
//
// It will panic if there's an error while loading the config.
func MustLoad[C Config]() *C {
	cfg, err := Load[C]()
	if err != nil {
		panic(err)
	}

	return cfg
}

// Load fills Config with data from configuration file and .env file.
func Load[C Config]() (*C, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("failed to load env file: %w", err)
	}

	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		return nil, errors.New("CONFIG_PATH environment variable not set")
	}

	var cfg C

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &cfg, nil
}
