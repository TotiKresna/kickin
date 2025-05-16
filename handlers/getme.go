package handlers

import (
	"kickin/middleware"
	"kickin/utils"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func GetMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok {
		utils.RespondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userData := map[string]interface{}{
		"username": claims["username"],
		"role":     claims["role"],
	}

	utils.RespondSuccess(w, "Profile fetched successfully", userData)
}