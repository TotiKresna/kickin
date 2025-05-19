package utils

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ParseJSON parses JSON request body into the target struct
func ParseJSON(r *http.Request, target interface{}) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	
	return decoder.Decode(target)
}

// FormatValidationErrors formats validation errors from validator.Validate
func FormatValidationErrors(err error) map[string]string {
	if err == nil {
		return nil
	}
	
	errors := make(map[string]string)
	
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		errors["general"] = err.Error()
		return errors
	}
	
	for _, e := range validationErrors {
		field := strings.ToLower(e.Field())
		switch e.Tag() {
		case "required":
			errors[field] = "This field is required"
		case "email":
			errors[field] = "Must be a valid email address"
		case "min":
			errors[field] = "Too short, minimum length is 5 characters" 
		case "oneof":
			errors[field] = "Invalid value"
		default:
			errors[field] = "Invalid value"
		}
	}
	
	return errors
}

func SetSecureCookie(w http.ResponseWriter, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	})
}

func ClearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}