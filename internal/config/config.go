package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DatabaseURL        string
	BaseCurrency       string
	AnthropicAPIKey    string
	JWTSecret          string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	SyncInterval       time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://cristianbutiri@localhost:5432/property_manager?sslmode=disable"
	}

	baseCurrency := os.Getenv("BASE_CURRENCY")
	if baseCurrency == "" {
		baseCurrency = "EUR"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-me"
	}

	googleRedirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	if googleRedirectURL == "" {
		googleRedirectURL = "http://localhost:8080/api/auth/google/callback"
	}

	syncInterval := 6 * time.Hour
	if si := os.Getenv("SYNC_INTERVAL"); si != "" {
		if d, err := time.ParseDuration(si); err == nil {
			syncInterval = d
		}
	}

	cfg := &Config{
		Port:               port,
		DatabaseURL:        dbURL,
		BaseCurrency:       baseCurrency,
		AnthropicAPIKey:    os.Getenv("ANTHROPIC_API_KEY"),
		JWTSecret:          jwtSecret,
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  googleRedirectURL,
		SyncInterval:       syncInterval,
	}

	return cfg, nil
}

func (c *Config) Addr() string {
	return fmt.Sprintf(":%s", c.Port)
}
