package handlers

import (
	"kickin/middleware"
	"kickin/models"
	"kickin/utils"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// Handler untuk user melihat semua booking miliknya
func GetMyBookings(db *gorm.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
        if !ok {
            utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
            return
        }
        userID := uint(claims["id"].(float64))

        var bookings []models.Booking
        if err := db.Preload("Court").Preload("User").Where("user_id = ?", userID).Find(&bookings).Error; err != nil {
            utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch bookings")
            return
        }
        utils.RespondSuccess(w, "Your bookings fetched", bookings)
    }
}

// Handler untuk user melihat detail booking miliknya
func GetMyBookingByID(db *gorm.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
        if !ok {
            utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
            return
        }
        userID := uint(claims["id"].(float64))
        idStr := chi.URLParam(r, "id")
        id, err := strconv.Atoi(idStr)
        if err != nil {
            utils.RespondError(w, http.StatusBadRequest, "Invalid booking ID")
            return
        }

        var booking models.Booking
        if err := db.Preload("Court").Preload("User").First(&booking, id).Error; err != nil {
            utils.RespondError(w, http.StatusNotFound, "Booking not found")
            return
        }
        if booking.UserID != userID {
            utils.RespondError(w, http.StatusForbidden, "You are not allowed to access this booking")
            return
        }
        utils.RespondSuccess(w, "Your booking found", booking)
    }
}