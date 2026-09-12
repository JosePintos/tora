package config

import (
	"os"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	JWT         JWTConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	URL string
}

type JWTConfig struct {
	Secret string
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
		JWT: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
		},
	}
}
