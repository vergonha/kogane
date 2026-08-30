package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

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

	if err := database.Init(db); err != nil {
		log.Fatal(err)
	}

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
		db,
		cfg.Development,
		cfg.BcryptCost,
		cfg.SessionDuration,
	)
	if err != nil {
		log.Fatal(err)
	}

	turnstileClient := turnstile.New(cfg.TurnstileSecretKey)

	go startSessionCleanup(db)

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

func startSessionCleanup(db *sql.DB) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := database.CleanupExpiredSessions(
			db,
			time.Now().Unix(),
		); err != nil {
			log.Printf("session cleanup: %v", err)
		}
	}
}
