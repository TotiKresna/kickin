package middleware

import (
	"kickin/utils"
	"kickin/logger"
	"net/http"
	"context"
)

const AccessTokenKey contextKey = "newAccessToken"

func RefreshMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		csrfCookie, err := r.Cookie("csrf_token")
		csrfHeader := r.Header.Get("X-CSRF-Token")

		if err != nil || csrfHeader != csrfCookie.Value {
			logger.LogError( "CSRF token mismatch")
			http.Error(w, "CSRF check failed", http.StatusForbidden)
			return
		}

		claims, err := utils.VerifyRefreshToken(cookie.Value)
		if err != nil {
			logger.LogWarning("Invalid refresh token, skipping refresh")
			next.ServeHTTP(w, r)
			return
		}

		id := uint(claims["id"].(float64))
		email := claims["email"].(string)
		username := claims["username"].(string)
		role := claims["role"].(string)

		accessToken, refreshToken, _ := utils.GenerateTokensFromClaims(id, email, username, role)

		// set new refresh token & csrf token
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   60 * 60 * 24 * 7,
		})
		csrfToken, _ := utils.GenerateCSRFToken()
		http.SetCookie(w, &http.Cookie{
			Name:  "csrf_token",
			Value: csrfToken,
			Path:  "/",
		})
		w.Header().Set("X-CSRF-Token", csrfToken)
		ctx := context.WithValue(r.Context(), AccessTokenKey, accessToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
