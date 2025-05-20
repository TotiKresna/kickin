package models

import (
	"time"

	"gorm.io/gorm"
)

type Booking struct {
	gorm.Model
	CourtID   uint `json:"court_id"`
	UserID    uint `json:"user_id"`
	StartTime time.Time `json:"start_time"`
	EndTime time.Time `json:"end_time"`
	Status time.Time `json:"status"`
}
