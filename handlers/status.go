package handlers

import (
	"kickin/utils"
	"net/http"

	"gorm.io/gorm"
)

func RootHandlerWithDB(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sqlDB, err := db.DB()
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to get database object")
			return
		}

		if err = sqlDB.Ping(); err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Database connection failed")
			return
		}
		utils.RespondSuccess(w, "Alhamdulillah", nil)
	}
}