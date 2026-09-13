package config

import (
	"cmp"
	"errors"
	"os"
	"strings"
	"time"
)

const (
	DefaultDBDSN           = "file:manga.db?_journal_mode=WAL&_busy_timeout=5000"
	DefaultLibraryPath     = "library.json"
	DefaultTemplatesGlob   = "templates/*.html"
	DefaultBcryptCost      = 12
	DefaultSessionDuration = 2 * time.Hour
	DefaultPublicURL       = "http://localhost:8080"
)

type Config struct {
	Development        bool
	Addr               string
	PublicURL          string
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
		Addr:               cmp.Or(os.Getenv("KOGANE_SERVER_PORT"), ":8080"),
		PublicURL:          strings.TrimSuffix(cmp.Or(os.Getenv("KOGANE_PUBLIC_URL"), DefaultPublicURL), "/"),
		DBDSN:              DBDSN(),
		TurnstileSecretKey: os.Getenv("CLOUDFLARE_TURNSTILE_SECRET_KEY"),
		TurnstileSiteKey:   os.Getenv("CLOUDFLARE_TURNSTILE_SITE_KEY"),
		R2BucketName:       os.Getenv("R2_BUCKET_NAME"),
		R2AccountID:        os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyID:      os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey:  os.Getenv("R2_SECRET_ACCESS_KEY"),
		LibraryPath:        cmp.Or(os.Getenv("KOGANE_LIBRARY_PATH"), DefaultLibraryPath),
		TemplatesGlob:      cmp.Or(os.Getenv("KOGANE_TEMPLATES_GLOB"), DefaultTemplatesGlob),
		SessionDuration:    DefaultSessionDuration,
		BcryptCost:         DefaultBcryptCost,
	}

	if cfg.R2BucketName == "" ||
		cfg.R2AccountID == "" ||
		cfg.R2AccessKeyID == "" ||
		cfg.R2SecretAccessKey == "" {
		return Config{}, errors.New(
			"R2_BUCKET_NAME, R2_ACCOUNT_ID, R2_ACCESS_KEY_ID and R2_SECRET_ACCESS_KEY must be configured",
		)
	}

	// Every login goes through Turnstile, so missing keys must fail at startup
	// rather than on the first sign-in attempt.
	if cfg.TurnstileSecretKey == "" || cfg.TurnstileSiteKey == "" {
		return Config{}, errors.New(
			"CLOUDFLARE_TURNSTILE_SECRET_KEY and CLOUDFLARE_TURNSTILE_SITE_KEY must be configured",
		)
	}

	return cfg, nil
}

func DBDSN() string {
	return cmp.Or(os.Getenv("KOGANE_DB_DSN"), DefaultDBDSN)
}
