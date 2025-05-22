package utils

import (
    "log"
    "time"
    "gorm.io/gorm"
    "kickin/models"
)

// updateExpiredBookings mengubah status booking dari pending ke expired jika sudah melewati expires_at
func updateExpiredBookings(db *gorm.DB) {
    now := time.Now()
    result := db.Model(&models.Booking{}).
        Where("status = ? AND expires_at < ?", "pending", now).
        Update("status", "expired")
    if result.Error != nil {
        log.Println("Error update expired bookings:", result.Error)
    } else if result.RowsAffected > 0 {
        log.Printf("Expired %d bookings", result.RowsAffected)
    }
}

// StartBookingCronjob menjalankan cronjob update expired bookings setiap 1 menit
func StartBookingCronjob(db *gorm.DB) {
    go func() {
        ticker := time.NewTicker(1 * time.Minute)
        defer ticker.Stop()
        for {
            updateExpiredBookings(db)
            <-ticker.C
        }
    }()
}