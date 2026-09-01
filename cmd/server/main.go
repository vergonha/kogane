package main

import (
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
)

func main() {
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
	go database.StartSessionCleanup(repository.Session)

	renderer, err := apptemplates.New(
		cfg.TemplatesGlob,
		cfg.Development,
	)
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

	turnstileClient := turnstile.New(cfg.TurnstileSecretKey)

	h := handlers.New(
		cfg,
		authService,
		renderer,
		turnstileClient,
		r2,
		lib,
	)

	router := apphttp.NewRouter(h, authService)

	log.Printf("Servidor em %s", cfg.Addr)
	log.Printf("Modo desenvolvimento: %v", cfg.Development)
	log.Fatal(http.ListenAndServe(cfg.Addr, router))
}
