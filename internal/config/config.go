package config

import (
	"fmt"
	"os"
	"time"
)

const (
	DefaultDBDSN           = "file:manga.db?_journal_mode=WAL&_busy_timeout=5000"
	DefaultLibraryPath     = "library.json"
	DefaultTemplatesGlob   = "templates/*.html"
	DefaultBcryptCost      = 12
	DefaultSessionDuration = 2 * time.Hour
)

type Config struct {
	Development        bool
	Addr               string
	DBDSN              string
	TurnstileSecretKey string
	TurnstileSiteKey   string
	R2BucketName       string
	R2AccountID        string
	R2AccessKeyID      string
	R2SecretAccessKey  string
	LibraryPath        string
	TemplatesGlob      string
	SessionDuration    time.Duration
	BcryptCost         int
}

func Load() (Config, error) {
	cfg := Config{
		Development:        os.Getenv("KOGANE_DEVELOPMENT") == "true",
		Addr:               envOr("KOGANE_SERVER_PORT", ":8080"),
		DBDSN:              DBDSN(),
		TurnstileSecretKey: os.Getenv("CLOUDFLARE_TURNSTILE_SECRET_KEY"),
		TurnstileSiteKey:   os.Getenv("CLOUDFLARE_TURNSTILE_SITE_KEY"),
		R2BucketName:       os.Getenv("R2_BUCKET_NAME"),
		R2AccountID:        os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyID:      os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey:  os.Getenv("R2_SECRET_ACCESS_KEY"),
		LibraryPath:        envOr("KOGANE_LIBRARY_PATH", DefaultLibraryPath),
		TemplatesGlob:      envOr("KOGANE_TEMPLATES_GLOB", DefaultTemplatesGlob),
		SessionDuration:    DefaultSessionDuration,
		BcryptCost:         DefaultBcryptCost,
	}

	if cfg.R2BucketName == "" ||
		cfg.R2AccountID == "" ||
		cfg.R2AccessKeyID == "" ||
		cfg.R2SecretAccessKey == "" {
		return Config{}, fmt.Errorf(
			"R2_BUCKET_NAME, R2_ACCOUNT_ID, R2_ACCESS_KEY_ID and R2_SECRET_ACCESS_KEY must be configured",
		)
	}

	return cfg, nil
}

func DBDSN() string {
	return envOr("KOGANE_DB_DSN", DefaultDBDSN)
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
