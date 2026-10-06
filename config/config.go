package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
)

type ServerStruct struct {
	Host string `validate:"required"`
}

type PostgresStruct struct {
	Host     string `validate:"required"`
	Port     string `validate:"required"`
	Username string `validate:"required"`
	Password string `validate:"required"`
	Database string `validate:"required"`
	SSLMode  string `validate:"required"`
}

type RedisStruct struct {
	Addr     string `validate:"required"`
	Password string
	DB       int
	CacheTTL int `validate:"required"`
}

type AuthStruct struct {
	JWTSecret string `validate:"required"`
}

type Config struct {
	Server   ServerStruct   `validate:"required"`
	Postgres PostgresStruct `validate:"required"`
	Redis    RedisStruct    `validate:"required"`
	Auth     AuthStruct     `validate:"required"`
}

const path = "config/config.json"

func Load() (cfg *Config, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config.Load.Open: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	cfg = &Config{}
	if err = json.NewDecoder(file).Decode(cfg); err != nil {
		return nil, fmt.Errorf("config.Load.Decode: %w", err)
	}

	if err = validator.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("config.Load.Validate: %w", err)
	}

	return cfg, nil
}
