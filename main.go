package main

import (
	"kickin/config"
	"kickin/logger"
	"kickin/migrations"
	"kickin/routes"
	"kickin/utils"
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
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB from gorm.DB: %v", err)
	}
	defer sqlDB.Close()

	// Jalankan migrasi otomatis
	migrations.RunMigrations(db)

	// Setup router + logger middleware
	router := routes.SetupRoutes(db, cfg)
	handler := logger.RequestLogger(router)

	// Jalankan cronjob untuk update expired bookings	
	utils.StartBookingCronjob(db)

	log.Printf("Server running on port %s", cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.AppPort, handler))
}
