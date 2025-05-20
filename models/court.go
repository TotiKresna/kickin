package models

import "gorm.io/gorm"

type Court struct {
	gorm.Model	// otomatis dapat ID, CreatedAt, UpdateAt, DeletedAt
	Name       string    `json:"name"`
	Location   string    `json:"location"`
	Price      int       `json:"price"`
	Image      string    `json:"image"`
	Status     string    `json:"status"`
	Bookings   []Booking `json:"bookings" gorm:"foreignKey:CourtID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
