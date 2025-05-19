package tests

import (
	"bytes"
	"encoding/json"
	"kickin/config"
	"kickin/handlers"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegister(t *testing.T) {
	cfg := config.LoadConfig()
	db := config.ConnectDB(cfg)
	sqlDB, err := db.DB()
	if err != nil {
		t.Errorf("Failed to get sql.DB from gorm.DB: %v", err)
	}
	defer sqlDB.Close()

	payload := map[string]string{
		"username": "testuser",
		"password": "password123",
		"role":     "user",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler := handlers.Register(db)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}
}
