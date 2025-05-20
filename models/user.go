package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model	// otomatis dapat ID, CreatedAt, UpdateAt, DeletedAt
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"password" gorm:"not null"`
	Role      string         `json:"role" gorm:"not null"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=5"`
	Role     string `json:"role" validate:"required,oneof=user admin superadmin"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"` // Bisa username atau email
	Password   string `json:"password" validate:"required"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}