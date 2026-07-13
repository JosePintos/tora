package config

import (
	"os"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	URL string
}

func Load() *Config {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	return &Config{
		Environment: os.Getenv("APP_ENV"),
		Server: ServerConfig{
			Port: port,
		},
		Database: DatabaseConfig{
			URL: os.Getenv("DATABASE_URL"),
		},
	}
}
