package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	GRPCServerHost string `env:"GRPC_SERVER_HOST" env-default:""`
	GRPCServerPort int    `env:"GRPC_SERVER_PORT" env-default:"9090"`
	RESTServerHost string `env:"REST_SERVER_HOST" env-default:""`
	RESTServerPort int    `env:"REST_SERVER_PORT" env-default:"8080"`
}

func New() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	cfg := Config{}
	err := cleanenv.ReadEnv(&cfg)

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
