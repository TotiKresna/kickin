package models

import (
	"time"

	"gorm.io/gorm"
)

type Booking struct {
	gorm.Model
	UserID    	uint 			 `json:"user_id"`
	User        User       		 `json:"user" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CourtID   	uint 			 `json:"court_id"`
	Court       Court      		 `json:"court" gorm:"foreignKey:CourtID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	StartTime   time.Time  		 `json:"start_time"`
	EndTime     time.Time        `json:"end_time"`
	TotalHours  int              `json:"total_hours"`
	TotalPrice  int          	 `json:"total_price"`
	Status      string           `json:"status"`       // pending, paid, cancelled, expired
	PaymentType string           `json:"payment_type"` // dp, full
	IsManual    bool             `json:"is_manual"`	 // true kalau admin/superadmin
	ExpiresAt   *time.Time 	 	 `json:"expires_at,omitempty"` // bisa null kalau admin/superadmin
}
