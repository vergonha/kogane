package main

import (
	"fmt"
	"log"
	"os"

	"kogane/internal/auth"
	"kogane/internal/config"
	"kogane/internal/database"
)

func main() {
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

	if err := database.Init(db); err != nil {
		log.Fatal(err)
	}

	authService, err := auth.NewService(
		db,
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
