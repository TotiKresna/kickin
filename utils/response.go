package utils

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func RespondJSON(w http.ResponseWriter, statusCode int, status string, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}

	_ = json.NewEncoder(w).Encode(response)
}

func RespondSuccess(w http.ResponseWriter, message string, data interface{}) {
	RespondJSON(w, http.StatusOK, "success", message, data)
}

func RespondError(w http.ResponseWriter, code int, message string) {
	RespondJSON(w, code, "error", message, nil)
}