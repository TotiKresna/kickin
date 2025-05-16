package main

import (
	"kickin/config"
	"kickin/logger"
	"kickin/migrations"
	"kickin/routes"
	"log"
	"net/http"
)

func main() {
	cfg := config.LoadConfig()
	db := config.ConnectDB(cfg)
	defer db.Close()

	// Jalankan migrasi otomatis
	migrations.RunMigrations(db)

	// Setup router + logger middleware
	router := routes.SetupRoutes(db, cfg)
	handler := logger.RequestLogger(router)

	log.Printf("Server running on port %s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, handler))
}
