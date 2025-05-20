package handlers

import (
	"kickin/utils"
	"kickin/middleware"

	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func GetMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Return user information from JWT claims
	userInfo := map[string]interface{}{
		"id":       claims["user_id"],
		"username": claims["username"],
		"email":    claims["email"],
		"role":     claims["role"],
	}

	utils.RespondSuccess(w, "User profile retrieved", userInfo)
}