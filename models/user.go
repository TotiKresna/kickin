package models

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
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

type UpdateUserInput struct {
	Username    string `json:"username,omitempty" validate:"omitempty,min=3"`
	Email       string `json:"email,omitempty" validate:"omitempty,email"`
	OldPassword string `json:"old_password,omitempty"`
	NewPassword string `json:"new_password,omitempty" validate:"omitempty,min=5"`
	Role        string `json:"role,omitempty" validate:"omitempty,oneof=user admin superadmin"`
}

func (u *User) UpdateProfile(db *gorm.DB, input UpdateUserInput, requesterRole string) error {
	// Cek email unik (jika berubah)
	if input.Email != "" && input.Email != u.Email {
		var count int64
		db.Model(&User{}).Where("email = ? AND id != ?", input.Email, u.ID).Count(&count)
		if count > 0 {
			return errors.New("email is already in use")
		}
		u.Email = input.Email
	}

	// Cek username unik (jika berubah)
	if input.Username != "" && input.Username != u.Username {
		var count int64
		db.Model(&User{}).Where("username = ? AND id != ?", input.Username, u.ID).Count(&count)
		if count > 0 {
			return errors.New("username is already taken")
		}
		u.Username = input.Username
	}

	// Ganti password (jika diberikan)
	if input.NewPassword != "" {
		if requesterRole != "superadmin" {
			// User biasa harus masukkan password lama
			if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(input.OldPassword)); err != nil {
				return errors.New("old password is incorrect")
			}
		}
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
		u.Password = string(hashedPassword)
	}

	// Superadmin bisa ubah role
	if requesterRole == "superadmin" && input.Role != "" {
		u.Role = input.Role
	}

	return db.Save(u).Error
}
