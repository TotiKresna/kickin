package handlers

import (
	"database/sql"
	"encoding/json"
	"kickin/logger"
	"kickin/models"
	"kickin/utils"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u models.User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid input")
			return
		}

		hashed, _ := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
		_, err := db.Exec("INSERT INTO users (username, password, role) VALUES ($1, $2, $3)", u.Username, string(hashed), u.Role)
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		utils.RespondSuccess(w, "User registered successfully", nil)
	}
}

func Login(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u models.User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid input")
			return
		}

		var user models.User
		err := db.QueryRow("SELECT id, username, password, role FROM users WHERE username=$1", u.Username).
			Scan(&user.ID, &user.Username, &user.Password, &user.Role)
		if err != nil {
			utils.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password)) != nil {
			logger.LogError("Incorrect password")
			utils.RespondError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		accessToken, refreshToken, err := utils.GenerateTokens(user)
		if err != nil {
			logger.LogError(err.Error())
			utils.RespondError(w, http.StatusInternalServerError, "Could not generate token")
			return
		}

		csrfToken, _ := utils.GenerateCSRFToken()
		http.SetCookie(w, &http.Cookie{
			Name:     "csrf_token",
			Value:    csrfToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   true, // jika pakai HTTPS
			SameSite: http.SameSiteStrictMode,
		})
		w.Header().Set("X-CSRF-Token", csrfToken)

		// Simpan refresh token sebagai cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   true, // jika pakai HTTPS
			SameSite: http.SameSiteStrictMode,
			MaxAge:   60 * 60 * 24 * 7, // 7 hari
		})

		utils.RespondSuccess(w, "Login successful", map[string]string{"access_token": accessToken})
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// Remove cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	utils.RespondSuccess(w, "Logged out successfully", nil)
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "No refresh token")
		return
	}

	claims, err := utils.VerifyRefreshToken(cookie.Value)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	username := claims["username"].(string)

	// Fetch user from DB
	var user models.User
	err = r.Context().Value("db").(*sql.DB).
		QueryRow("SELECT username, role FROM users WHERE username=$1", username).
		Scan(&user.Username, &user.Role)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "User not found")
		return
	}

	at, rt, err := utils.GenerateTokens(user)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	csrfToken, _ := utils.GenerateCSRFToken()
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	w.Header().Set("X-CSRF-Token", csrfToken)

	// Set refresh token as HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    rt,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   60 * 60 * 24 * 7, // 7 hari
	})

	utils.RespondSuccess(w, "Token refreshed successfully", map[string]string{
		"access_token": at,
	})
}
