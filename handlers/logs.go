package handlers

import (
	"net/http"
	"os"

	"kickin/utils"
)

func ViewLogs(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("app.log")
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to read logs")
		return
	}

	utils.RespondSuccess(w, "Log fetched", map[string]string{
		"content": string(data),
	})
}

func ClearLogs(w http.ResponseWriter, r *http.Request) {
	err := os.WriteFile("app.log", []byte{}, 0644) // Kosongkan file
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to clear logs")
		return
	}
	utils.RespondSuccess(w, "Log cleared", nil)
}