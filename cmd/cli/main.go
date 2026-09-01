package main

import (
	"fmt"
	"log"
	"os"

	"kogane/internal/auth"
	"kogane/internal/config"
	"kogane/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if len(os.Args) != 4 || os.Args[1] != "create-user" {
		log.Fatalf("uso: %s create-user <usuario> <senha>", os.Args[0])
	}

	username := os.Args[2]
	password := os.Args[3]

	db, err := database.Open(config.DBDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repository := database.NewRepository(db)
	go database.StartSessionCleanup(repository.Session)

	authService, err := auth.NewService(
		repository,
		false,
		config.DefaultBcryptCost,
		config.DefaultSessionDuration,
	)
	if err != nil {
		log.Fatal(err)
	}

	adminExists, err := authService.AdminExists()
	if err != nil {
		log.Fatal(err)
	}

	if !adminExists {
		log.Fatal(
			"nenhum admin existe ainda; abra /login no server e crie o admin inicial primeiro",
		)
	}

	if err := authService.CreateUser(username, password, false); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("usuário %q criado com sucesso\n", username)
}
