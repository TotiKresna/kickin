package handlers

import (
	"fmt"
	"io"
	"kickin/models"
	"kickin/utils"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func GetCourt(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var courts []models.Court
		if err := db.Find(&courts).Error; err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch courts")
			return
		}

		if len(courts) == 0 {
			utils.RespondError(w, http.StatusNotFound, "Court not found")
			return
		}

		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}

		baseURL := fmt.Sprintf("%s://%s/image", scheme, r.Host)
		for i := range courts {
			if courts[i].Image != "" {
				courts[i].Image = fmt.Sprintf("%s/%s", baseURL, url.PathEscape(filepath.Base(courts[i].Image)))
			}
		}

		utils.RespondSuccess(w, "Courts retrieved successfully", courts)
	}
}

func CreateCourt(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		imageFolder := "assets/image"
		if err := os.MkdirAll(imageFolder, os.ModePerm); err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Unable to create uploads folder")
			return
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil { // max 10 mb
			utils.RespondError(w, http.StatusBadRequest, "Invalid form data")
			return
		}

		name := r.FormValue("name")
		location := r.FormValue("location")
		status := r.FormValue("status")
		priceStr := r.FormValue("price")

		priceInt, err := strconv.Atoi(priceStr)
		if err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid price format")
			return
		}

		file, handler, err := r.FormFile("image")
		if err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Image upload failed: "+err.Error())
			return
		}
		defer file.Close()

		cleanFileName := filepath.Base(handler.Filename)
		fileName := fmt.Sprintf("%d_%s", time.Now().Unix(), cleanFileName)
		filePath := filepath.Join(imageFolder, fileName)

		dst, err := os.Create(filePath)
		if err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to save image")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to save image")
			return
		}

		courts := models.Court{
			Name:     name,
			Location: location,
			Status:   status,
			Price:    priceInt,
			Image:    filePath,
		}

		if err := db.Create(&courts).Error; err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to Create Courts")
			return
		}

		utils.RespondSuccess(w, "Court created successfully", courts)
	}
}

func UpdateCourt(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParams := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idParams)
		if err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid court ID")
			return
		}

		var courts models.Court
		if err := db.First(&courts, id).Error; err != nil {
			utils.RespondError(w, http.StatusNotFound, "Court not found")
			return
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid form data")
			return
		}

		updateData := map[string]interface{}{}
		if name := r.FormValue("name"); name != "" {
			updateData["name"] = name
		}
		if location := r.FormValue("location"); location != "" {
			updateData["location"] = location
		}
		if status := r.FormValue("status"); status != "" {
			updateData["status"] = status
		}
		if priceStr := r.FormValue("price"); priceStr != "" {
			priceInt, err := strconv.Atoi(priceStr)
			if err == nil {
				updateData["price"] = priceInt
			}
		}

		file, handler, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			imageFolder := "assets/image"
			if err := os.MkdirAll(imageFolder, os.ModePerm); err != nil {
				utils.RespondError(w, http.StatusInternalServerError, "Unable to create image folder")
				return
			}

			cleanFileName := filepath.Base(handler.Filename)
			fileName := fmt.Sprintf("%d_%s", time.Now().Unix(), cleanFileName)
			filePath := filepath.Join(imageFolder, fileName)

			dst, err := os.Create(filePath)
			if err != nil {
				utils.RespondError(w, http.StatusInternalServerError, "Failed to save image")
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				utils.RespondError(w, http.StatusInternalServerError, "Failed to save image")
				return
			}

			// Hapus file lama
			if courts.Image != "" {
				_ = os.Remove(courts.Image)
			}

			updateData["image"] = filePath
		}

		if len(updateData) == 0 {
			utils.RespondError(w, http.StatusBadRequest, "No data to update")
			return
		}

		if err := db.Model(&courts).Updates(updateData).Error; err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to update court")
			return
		}

		db.First(&courts, id)
		utils.RespondSuccess(w, "Court updated successfully", courts)
	}
}

func DeleteCourt(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParams := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idParams)
		if err != nil {
			utils.RespondError(w, http.StatusBadRequest, "Invalid court ID")
			return
		}

		var courts models.Court
		if err := db.First(&courts, id).Error; err != nil {
			utils.RespondError(w, http.StatusNotFound, "Court not found")
			return
		}

		if courts.Image != "" {
			_ = os.Remove(courts.Image)
		}

		if err := db.Delete(&courts).Error; err != nil {
			utils.RespondError(w, http.StatusInternalServerError, "Failed to delete court")
			return
		}

		utils.RespondSuccess(w, "Court deleted successfully", nil)
	}
}
