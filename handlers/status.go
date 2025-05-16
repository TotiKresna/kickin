package handlers

import (
	"database/sql"
	"kickin/utils"
	"net/http"
)

func RootHandlerWithDB(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Database connection failed")
			return
		}
		utils.RespondSuccess(w, "Connected to PostgreSQL Aiven with SSL", nil)
	}
}