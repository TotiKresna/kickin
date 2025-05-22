package utils

import (
	"errors"
	"kickin/models"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtKey         = []byte(os.Getenv("JWT_SECRET"))
	refreshKey     = []byte(os.Getenv("JWT_REFRESH_SECRET"))
	refreshExpires = 7 * 24 * time.Hour
)

func GenerateTokens(user models.User) (accessToken string, refreshToken string, err error) {
	claims := jwt.MapClaims{
		"id":  user.ID,
		"email":  	user.Email,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
	}
	rtClaims := jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(refreshExpires).Unix(),
	}

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)

	accessToken, err = at.SignedString(jwtKey)
	if err != nil {
		return
	}
	refreshToken, err = rt.SignedString(refreshKey)
	return
}

func VerifyRefreshToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return refreshKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}
	return token.Claims.(jwt.MapClaims), nil
}

func GenerateTokensFromClaims(id uint, email string, username string, role string) (string, string, error) {
	user := models.User{Email: email, Username: username, Role: role}
	return GenerateTokens(user)
}