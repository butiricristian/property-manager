package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Production         bool
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

	production := os.Getenv("APP_ENV") == "production"
	if production {
		if err := checkProductionEnv(); err != nil {
			return nil, err
		}
	}

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
		Production:         production,
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

// checkProductionEnv refuses to start with the local-dev fallbacks, which would
// silently point at a localhost database or sign tokens with a public secret.
func checkProductionEnv() error {
	if os.Getenv("DATABASE_URL") == "" {
		return fmt.Errorf("DATABASE_URL is required when APP_ENV=production")
	}
	if len(os.Getenv("JWT_SECRET")) < 32 {
		return fmt.Errorf("JWT_SECRET must be set to at least 32 characters when APP_ENV=production")
	}
	if os.Getenv("GOOGLE_CLIENT_ID") != "" && !strings.HasPrefix(os.Getenv("GOOGLE_REDIRECT_URL"), "https://") {
		return fmt.Errorf("GOOGLE_REDIRECT_URL must be an https URL when Google OAuth is enabled and APP_ENV=production")
	}
	return nil
}

func (c *Config) Addr() string {
	return fmt.Sprintf(":%s", c.Port)
}
