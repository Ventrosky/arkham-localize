package main

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func setupTestHandlers() {
	// Set test OpenAI key if not set
	if os.Getenv("OPENAI_API_KEY") == "" {
		os.Setenv("OPENAI_API_KEY", "test-key")
	}
	openAIKey = os.Getenv("OPENAI_API_KEY")
	embeddingModel = "text-embedding-3-small"
}

func TestHealthHandler(t *testing.T) {
	setupTestHandlers()

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}
}

func TestTranslateHandler_MethodNotAllowed(t *testing.T) {
	setupTestHandlers()

	var db *sql.DB

	req, err := http.NewRequest("GET", "/translate", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := translateHandler(db)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, status)
	}
}

func TestTranslateHandler_EmptyBody(t *testing.T) {
	setupTestHandlers()

	var db *sql.DB

	req, err := http.NewRequest("POST", "/translate", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := translateHandler(db)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status %d for empty text, got %d", http.StatusBadRequest, status)
	}
}

func TestTranslateHandler_InvalidJSON(t *testing.T) {
	setupTestHandlers()

	var db *sql.DB

	req, err := http.NewRequest("POST", "/translate", bytes.NewBuffer([]byte("invalid json")))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := translateHandler(db)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, status)
	}
}
