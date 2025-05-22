package migrations

import (
	"kickin/models"
	"log"

	"gorm.io/gorm"
)

// RunMigrations performs automatic database migrations
func RunMigrations(db *gorm.DB) {
	log.Println("Running database migrations...")

	// Auto migrate all models
	err := db.AutoMigrate(&models.User{}, &models.Court{}, &models.Booking{})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Database migrated successfully")
}
