package config

import (
	"github.com/caarlos0/env/v6"
)

type Config struct {
	DB_HOST     string `env:"DB_HOST"`
	DB_PORT     string `env:"DB_PORT"`
	DB_NAME     string `env:"DB_NAME"`
	DB_LOGIN    string `env:"DB_LOGIN"`
	DB_PASS     string `env:"DB_PASS"`
	SERVER_PORT string `env:"SERVER_PORT"`
	SERVER_HOST string `env:"SERVER_HOST"`
	API_URL     string `env:"API_URL"`
}

func ParseConfigServer() (*Config, error) {
	config := &Config{}
	//считываем все переменны окружения в cfg
	if err := env.Parse(config); err != nil {
		return nil, err
	}

	return config, nil
}
