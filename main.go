package main

import (
	"kickin/config"
	"kickin/migrations"
	"kickin/routes"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := config.LoadConfig()
	db := config.ConnectDB(cfg)
	defer db.Close()

	// Jalankan migrasi otomatis
	migrations.RunMigrations(db)

	// Setup router + logger middleware
	router := routes.SetupRoutes(db, cfg)

	log.Printf("Server running on port %s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, router))
}
