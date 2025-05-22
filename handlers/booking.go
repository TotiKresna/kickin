package handlers

import (
    "kickin/logger"
    "kickin/middleware"
    "kickin/models"
    "kickin/utils"
    "net/http"
    "strconv"

    "github.com/golang-jwt/jwt/v5"
    "github.com/go-chi/chi/v5"
    "gorm.io/gorm"
)

// Helper untuk cek role admin/superadmin
func isAdminOrSuperadmin(r *http.Request) bool {
    claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
    if !ok {
        return false
    }
    role, ok := claims["role"].(string)
    return ok && (role == "admin" || role == "superadmin")
}

func CreateBooking(db *gorm.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var input struct {
            CourtID     uint             `json:"court_id"`
            StartTime   utils.CustomTime `json:"start_time"`
            EndTime     utils.CustomTime `json:"end_time"`
            PaymentType string           `json:"payment_type"` // "dp" or "full"
            Status      string           `json:"status"`       // only for admin/superadmin e.g. "pending", "paid", "cancelled"
        }

        if err := utils.ParseJSON(r, &input); err != nil {
            logger.LogError("Failed to parse JSON: " + err.Error())
            utils.RespondError(w, http.StatusBadRequest, "Invalid input")
            return
        }

        start := input.StartTime.Time
        end := input.EndTime.Time

        if end.Before(start) || !end.After(start) {
            logger.LogError("Invalid time range: start time is after end time")
            utils.RespondError(w, http.StatusBadRequest, "Invalid time range")
            return
        }

        duration := int(end.Sub(start).Hours())
        if duration <= 0 {
            utils.RespondError(w, http.StatusBadRequest, "Minimum booking is 1 hour")
            return
        }

        // Get claims from JWT
        claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
        if !ok {
            utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
            return
        }

        role := claims["role"].(string)
        userID := uint(claims["id"].(float64))

        // Ambil data lapangan
        var court models.Court
        if err := db.First(&court, input.CourtID).Error; err != nil {
            logger.LogError("Court not found: " + err.Error())
            utils.RespondError(w, http.StatusNotFound, "Court not found")
            return
        }

        // Validasi tabrakan waktu untuk SEMUA role
        var count int64
        db.Model(&models.Booking{}).
            Where("court_id = ? AND status IN ? AND start_time < ? AND end_time > ?",
                input.CourtID, []string{"pending", "paid"}, end, start).
            Count(&count)

        if count > 0 {
            utils.RespondError(w, http.StatusConflict, "Slot already booked")
            return
        }

        isManual := role == "admin" || role == "superadmin"

        // Buat booking
        booking := models.Booking{
            UserID:      userID,
            CourtID:     input.CourtID,
            StartTime:   start,
            EndTime:     end,
            TotalHours:  duration,
            TotalPrice:  int(duration) * court.Price,
            PaymentType: input.PaymentType,
            IsManual:    isManual,
        }

        if isManual {
            booking.Status = input.Status
        } else {
            booking.Status = "pending"
            booking.ExpiresAt = utils.NowPlusMinutes(15)
        }

        if err := db.Create(&booking).Error; err != nil {
            utils.RespondError(w, http.StatusInternalServerError, "Failed to create booking")
            return
        }

        utils.RespondSuccess(w, "Booking created", booking)
    }
}

func GetAllBookings(db *gorm.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !isAdminOrSuperadmin(r) {
            logger.LogError("Forbidden: user is not admin/superadmin")
            utils.RespondError(w, http.StatusForbidden, "Forbidden")
            return
        }
        var bookings []models.Booking
        if err := db.Preload("Court").Preload("User").Find(&bookings).Error; err != nil {
            logger.LogError("Failed to fetch bookings: " + err.Error())
            utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch bookings")
            return
        }
        utils.RespondSuccess(w, "Bookings fetched", bookings)
    }
}

func GetBookingByID(db *gorm.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !isAdminOrSuperadmin(r) {
            logger.LogError("Forbidden: user is not admin/superadmin")
            utils.RespondError(w, http.StatusForbidden, "Forbidden")
            return
        }
        idStr := chi.URLParam(r, "id")
        id, err := strconv.Atoi(idStr)
        if err != nil {
            logger.LogError("Invalid booking ID: " + err.Error())
            utils.RespondError(w, http.StatusBadRequest, "Invalid booking ID")
            return
        }
        var booking models.Booking
        if err := db.Preload("Court").Preload("User").First(&booking, id).Error; err != nil {
            logger.LogError("Booking not found: " + err.Error())
            utils.RespondError(w, http.StatusNotFound, "Booking not found")
            return
        }
        utils.RespondSuccess(w, "Booking found", booking)
    }
}

func UpdateBooking(db *gorm.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !isAdminOrSuperadmin(r) {
            logger.LogError("Forbidden: user is not admin/superadmin")
            utils.RespondError(w, http.StatusForbidden, "Forbidden")
            return
        }
        idStr := chi.URLParam(r, "id")
        id, err := strconv.Atoi(idStr)
        if err != nil {
            logger.LogError("Invalid booking ID: " + err.Error())
            utils.RespondError(w, http.StatusBadRequest, "Invalid booking ID")
            return
        }
        var input struct {
            Status      string `json:"status"`
            PaymentType string `json:"payment_type"`
        }
        if err := utils.ParseJSON(r, &input); err != nil {
            logger.LogError("Failed to parse JSON: " + err.Error())
            utils.RespondError(w, http.StatusBadRequest, "Invalid input")
            return
        }
        var booking models.Booking
        if err := db.First(&booking, id).Error; err != nil {
            logger.LogError("Booking not found: " + err.Error())
            utils.RespondError(w, http.StatusNotFound, "Booking not found")
            return
        }
        if input.Status != "" {
            booking.Status = input.Status
        }
        if input.PaymentType != "" {
            booking.PaymentType = input.PaymentType
        }
        if err := db.Save(&booking).Error; err != nil {
            logger.LogError("Failed to update booking: " + err.Error())
            utils.RespondError(w, http.StatusInternalServerError, "Failed to update booking")
            return
        }
        utils.RespondSuccess(w, "Booking updated", booking)
    }
}

func DeleteBooking(db *gorm.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if !isAdminOrSuperadmin(r) {
            logger.LogError("Forbidden: user is not admin/superadmin")
            utils.RespondError(w, http.StatusForbidden, "Forbidden")
            return
        }
        idStr := chi.URLParam(r, "id")
        id, err := strconv.Atoi(idStr)
        if err != nil {
            logger.LogError("Invalid booking ID: " + err.Error())
            utils.RespondError(w, http.StatusBadRequest, "Invalid booking ID")
            return
        }
        if err := db.Delete(&models.Booking{}, id).Error; err != nil {
            logger.LogError("Failed to delete booking: " + err.Error())
            utils.RespondError(w, http.StatusInternalServerError, "Failed to delete booking")
            return
        }
        utils.RespondSuccess(w, "Booking deleted", nil)
    }
}