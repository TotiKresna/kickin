package handlers

import (
	"errors"
	"kickin/logger"
	"kickin/middleware"
	"kickin/models"
	"kickin/utils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func GetAllUsers(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
		if !ok {
			utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
			logger.LogError("Unauthorized access attempt to GetAllUsers")
			return
		}

		role := claims["role"].(string)
		if role != "superadmin" {
			utils.RespondError(w, http.StatusForbidden, "Forbidden: only superadmin can view all users")
			logger.LogWarning("Forbidden access to GetAllUsers by role: " + role)
			return
		}

		var users []models.User
		if err := db.Find(&users).Error; err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to retrieve users")
			logger.LogError("Failed to get all users: " + err.Error())
			return
		}

		// Sembunyikan password dari hasil
		var result []map[string]interface{}
		for _, user := range users {
			result = append(result, map[string]interface{}{
				"id":   user.ID,
				"username":  user.Username,
				"email":     user.Email,
				"role":      user.Role,
				"createdAt": user.CreatedAt,
				"updatedAt": user.UpdatedAt,
			})
		}

		utils.RespondSuccess(w, "List of users", result)
		logger.LogInfo("Successfully retrieved user list")
	}
}

func GetUserByID(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDParam := chi.URLParam(r, "id")
		userID, err := strconv.ParseUint(userIDParam, 10, 32)
		if err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid user ID")
			logger.LogWarning("Invalid user ID in GetUserByID: " + userIDParam)
			return
		}

		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.RespondError(w, http.StatusNotFound, "User not found")
				logger.LogWarning("User not found: ID " + userIDParam)
				return
			}
			utils.RespondError(w, http.StatusInternalServerError, "Failed to find user")
			logger.LogError("DB error in GetUserByID: " + err.Error())
			return
		}

		utils.RespondSuccess(w, "User found", map[string]interface{}{
			"id":   user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"role":      user.Role,
			"createdAt": user.CreatedAt,
			"updatedAt": user.UpdatedAt,
		})
		logger.LogInfo("Successfully retrieved user: ID " + userIDParam)
	}
}

func UpdateUser(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
		if !ok {
			utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
			logger.LogError("Unauthorized access attempt to UpdateUser")
			return
		}

		userIDParam := chi.URLParam(r, "id")
		requesterID := uint(claims["id"].(float64))
		requesterRole := claims["role"].(string)

		targetUserID, err := strconv.ParseUint(userIDParam, 10, 32)
		if err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid user ID")
			logger.LogWarning("Invalid user ID in UpdateUser: " + userIDParam)
			return
		}

		if requesterRole != "superadmin" && uint(targetUserID) != requesterID {
			utils.RespondError(w, http.StatusForbidden, "Forbidden")
			logger.LogWarning("User " + strconv.Itoa(int(requesterID)) + " attempted to modify another user")
			return
		}

		var input models.UpdateUserInput
		if err := utils.ParseJSON(r, &input); err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
			logger.LogWarning("Failed to parse JSON in UpdateUser")
			return
		}

		// Validasi input
		if err := validate.Struct(&input); err != nil {
			errors := utils.FormatValidationErrors(err)
			utils.RespondValidationError(w, "Validation failed", errors)
			logger.LogWarning("Validation failed in UpdateUser")
			return
		}

		var user models.User
		if err := db.First(&user, targetUserID).Error; err != nil {
			utils.RespondError(w, http.StatusNotFound, "User not found")
			logger.LogWarning("User not found: ID " + userIDParam)
			return
		}

		if err := user.UpdateProfile(db, input, requesterRole); err != nil {
			utils.RespondError(w, http.StatusBadRequest, err.Error())
			logger.LogWarning("UpdateProfile error: " + err.Error())
			return
		}

		logger.LogInfo("User updated successfully: ID " + userIDParam)
		utils.RespondSuccess(w, "User updated successfully", nil)
	}
}

func DeleteUser(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
		if !ok {
			utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
			logger.LogError("Unauthorized access attempt to DeleteUser")
			return
		}
		role := claims["role"].(string)
		if role != "superadmin" {
			utils.RespondError(w, http.StatusForbidden, "Forbidden: only superadmin can delete users")
			logger.LogWarning("Forbidden access to DeleteUser by role: " + role)
			return
		}

		// Ambil user ID dari URL
		userIDParam := chi.URLParam(r, "id")
		targetUserID, err := strconv.ParseUint(userIDParam, 10, 32)
		if err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid user ID")
			logger.LogWarning("Invalid user ID in DeleteUser: " + userIDParam)
			return
		}

		// Cek apakah user ditemukan
		var user models.User
		if err := db.First(&user, targetUserID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				utils.RespondError(w, http.StatusNotFound, "User not found")
				logger.LogWarning("User not found: ID " + userIDParam)
				return
			}
			utils.RespondError(w, http.StatusInternalServerError, "Failed to retrieve user")
			logger.LogError("DB error in DeleteUser: " + err.Error())
			return
		}

		// Soft delete user
		if err := db.Delete(&user).Error; err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to delete user")
			logger.LogError("Failed to delete user: " + err.Error())
			return
		}

		logger.LogInfo("User deleted successfully: ID " + userIDParam)
		utils.RespondSuccess(w, "User deleted successfully", nil)
	}
}
