package handlers

import (
	"kickin/logger"
	"kickin/models"
	"kickin/utils"
	"net/http"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var validate = validator.New()

func Register(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterRequest
		if err := utils.ParseJSON(r, &req); err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Validate input
		if err := validate.Struct(req); err != nil {
			utils.RespondValidationError(w, "Validation failed", utils.FormatValidationErrors(err))
			return
		}
		// Check if username already exists
		var count int64
		db.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
		if count > 0 {
			utils.RespondError(w, http.StatusConflict, "Username already exists")
			return
		}

		// Check if email already exists
		db.Model(&models.User{}).Where("email = ?", req.Email).Count(&count)
		if count > 0 {
			utils.RespondError(w, http.StatusConflict, "Email already exists")
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			logger.LogError("Failed to hash password: " + err.Error())
			utils.RespondError(w, http.StatusInternalServerError, "Failed to process registration")
			return
		}

		// Create user
		user := models.User{
			Username: req.Username,
			Email:    req.Email,
			Password: string(hashedPassword),
			Role:     req.Role,
		}

		if result := db.Create(&user); result.Error != nil {
			logger.LogError("Failed to create user: " + result.Error.Error())
			utils.RespondError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		utils.RespondSuccess(w, "User registered successfully", nil)
	}
}

func Login(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginRequest
		if err := utils.ParseJSON(r, &req); err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Validate input
		if err := validate.Struct(req); err != nil {
			utils.RespondValidationError(w, "Validation failed", utils.FormatValidationErrors(err))
			return
		}

		// Find user by username or email
		var user models.User
		if err := db.Where("username = ? OR email = ?", req.Identifier, req.Identifier).First(&user).Error; err != nil {
			logger.LogError("User not found: " + req.Identifier)
			utils.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			logger.LogError("Incorrect password for user: " + req.Identifier)
			utils.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		// Generate tokens
		accessToken, refreshToken, err := utils.GenerateTokens(user)
		if err != nil {
			logger.LogError("Failed to generate tokens: " + err.Error())
			utils.RespondError(w, http.StatusInternalServerError, "Authentication failed")
			return
		}

		// Generate CSRF token
		csrfToken, err := utils.GenerateCSRFToken()
		if err != nil {
			logger.LogError("Failed to generate CSRF token: " + err.Error())
			utils.RespondError(w, http.StatusInternalServerError, "Authentication failed")
			return
		}

		// Set cookies
		utils.SetSecureCookie(w, "csrf_token", csrfToken, 24*60*60)         // 24 hours
		utils.SetSecureCookie(w, "refresh_token", refreshToken, 7*24*60*60) // 7 days

		// Set header for CSRF token
		w.Header().Set("X-CSRF-Token", csrfToken)

		// Return response
		utils.RespondSuccess(w, "Login successful", models.TokenResponse{
			AccessToken: accessToken,
		})
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// Clear cookies
	utils.ClearCookie(w, "refresh_token")
	utils.ClearCookie(w, "csrf_token")

	utils.RespondSuccess(w, "Logged out successfully", nil)
}

func RefreshToken(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get refresh token from cookie
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			utils.RespondError(w, http.StatusUnauthorized, "No refresh token")
			return
		}

		// Verify refresh token
		claims, err := utils.VerifyRefreshToken(cookie.Value)
		if err != nil {
			utils.RespondError(w, http.StatusUnauthorized, "Invalid refresh token")
			return
		}

		// Get username from claims
		username, ok := claims["username"].(string)
		if !ok {
			utils.RespondError(w, http.StatusUnauthorized, "Invalid token claims")
			return
		}

		// Find user
		var user models.User
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			utils.RespondError(w, http.StatusUnauthorized, "User not found")
			return
		}

		// Generate new tokens
		accessToken, refreshToken, err := utils.GenerateTokens(user)
		if err != nil {
			logger.LogError("Failed to generate tokens: " + err.Error())
			utils.RespondError(w, http.StatusInternalServerError, "Failed to refresh tokens")
			return
		}

		// Generate new CSRF token
		csrfToken, err := utils.GenerateCSRFToken()
		if err != nil {
			logger.LogError("Failed to generate CSRF token: " + err.Error())
			utils.RespondError(w, http.StatusInternalServerError, "Failed to refresh tokens")
			return
		}

		// Set cookies
		utils.SetSecureCookie(w, "csrf_token", csrfToken, 24*60*60)         // 24 hours
		utils.SetSecureCookie(w, "refresh_token", refreshToken, 7*24*60*60) // 7 days

		// Set header for CSRF token
		w.Header().Set("X-CSRF-Token", csrfToken)

		// Return response
		utils.RespondSuccess(w, "Token refreshed successfully", models.TokenResponse{
			AccessToken: accessToken,
		})
	}
}
