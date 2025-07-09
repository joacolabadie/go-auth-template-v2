package main

import (
	"log"

	"github.com/joacolabadie/go-auth-template-v2/internal/config"
	"github.com/joacolabadie/go-auth-template-v2/internal/database"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.ConnectDatabase(cfg.Database); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDatabase()
}
