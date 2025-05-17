package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"kickin/models"
	"kickin/utils"
	"kickin/logger"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u models.User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		hashed, _ := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
		_, err := db.Exec("INSERT INTO users (username, password, role) VALUES ($1, $2, $3)", u.Username, string(hashed), u.Role)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, "User registered")
	}
}

func Login(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u models.User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		var user models.User
		err := db.QueryRow("SELECT id, username, password, role FROM users WHERE username=$1", u.Username).
			Scan(&user.ID, &user.Username, &user.Password, &user.Role)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password)) != nil {
			logger.LogError( "Incorrect password")
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		accessToken, refreshToken, err := utils.GenerateTokens(user)
		if err != nil {
			logger.LogError(err.Error())
			http.Error(w, "Could not generate token", http.StatusInternalServerError)
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

		json.NewEncoder(w).Encode(map[string]string{"access_token": accessToken})
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
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Logged out")
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "No refresh token", http.StatusUnauthorized)
		return
	}

	claims, err := utils.VerifyRefreshToken(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	username := claims["username"].(string)

	// Fetch user from DB
	var user models.User
	err = r.Context().Value("db").(*sql.DB).
		QueryRow("SELECT username, role FROM users WHERE username=$1", username).
		Scan(&user.Username, &user.Role)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	at, rt, err := utils.GenerateTokens(user)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
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

	json.NewEncoder(w).Encode(map[string]string{
		"access_token": at,
	})
}
