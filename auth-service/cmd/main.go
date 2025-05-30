package main

import (
	"context"
	"github.com/joho/godotenv"
	"log"

	"auth_svc/config"
	"auth_svc/internal/app"
)

func main() {
	ctx := context.Background()
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	log.Printf("Config loaded: %+v\n", cfg)

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Println("failed to setup application:", err)

		return
	}

	err = application.Run()
	if err != nil {
		log.Println("failed to run application: ", err)

		return
	}
}
