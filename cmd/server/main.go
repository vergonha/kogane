package main

import (
	"context"
	"log"
	"net/http"

	"kogane/internal/auth"
	"kogane/internal/config"
	"kogane/internal/database"
	apphttp "kogane/internal/http"
	"kogane/internal/http/handlers"
	"kogane/internal/library"
	"kogane/internal/storage"
	apptemplates "kogane/internal/templates"
	"kogane/internal/turnstile"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("load .env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Open(cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repository := database.NewRepository(db)

	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	go database.StartSessionCleanup(ctx, repository.Session)

	renderer, err := apptemplates.New(cfg.TemplatesGlob, cfg.Development)
	if err != nil {
		log.Fatal(err)
	}

	lib, err := library.Load(cfg.LibraryPath)
	if err != nil {
		log.Fatal(err)
	}

	r2, err := storage.New(
		cfg.R2BucketName,
		cfg.R2AccountID,
		cfg.R2AccessKeyID,
		cfg.R2SecretAccessKey,
	)
	if err != nil {
		log.Fatal(err)
	}

	authService, err := auth.NewService(
		repository,
		cfg.Development,
		cfg.BcryptCost,
		cfg.SessionDuration,
	)
	if err != nil {
		log.Fatal(err)
	}

	h := &handlers.Handler{
		Config:     cfg,
		Auth:       authService,
		Renderer:   renderer,
		Turnstile:  turnstile.New(cfg.TurnstileSecretKey),
		Storage:    r2,
		Library:    lib,
		Repository: repository,
	}

	log.Printf("Server running on %s (development: %v)", cfg.Addr, cfg.Development)
	log.Fatal(http.ListenAndServe(cfg.Addr, apphttp.NewRouter(h, authService)))
}
