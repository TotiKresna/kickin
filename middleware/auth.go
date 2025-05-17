package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"
	"kickin/logger"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := ""

		// Ambil access token dari context dulu (jika diperbarui)
		if tokenFromCtx, ok := r.Context().Value("newAccessToken").(string); ok {
			tokenStr = tokenFromCtx
		} else {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				logger.LogError("missing or invalid Authorization header")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			logger.LogError("invalid or expired token")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.LogError("Token claims conversion failed")
			http.Error(w, "Invalid claims", http.StatusUnauthorized)
			return
			
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RoleMiddleware(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(jwt.MapClaims)
			if !ok || claims["role"] != role {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
