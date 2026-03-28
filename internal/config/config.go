package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://cristianbutiri@localhost:5432/property_manager?sslmode=disable"
	}

	cfg := &Config{
		Port:        port,
		DatabaseURL: dbURL,
	}

	return cfg, nil
}

func (c *Config) Addr() string {
	return fmt.Sprintf(":%s", c.Port)
}
