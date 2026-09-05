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
		log.Fatalf("usage: %s create-user <username> <password>", os.Args[0])
	}

	username := os.Args[2]
	password := os.Args[3]

	db, err := database.Open(config.DBDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repository := database.NewRepository(db)

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
			"no admin exists yet; open /login on the server and create the initial admin first",
		)
	}

	if err := authService.CreateUser(username, password, false); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("user %q created successfully\n", username)
}
