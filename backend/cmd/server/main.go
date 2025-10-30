package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/ventrosky/arkham-localize/backend/internal/db"
	"github.com/ventrosky/arkham-localize/backend/internal/embeddings"
	"github.com/ventrosky/arkham-localize/backend/internal/rag"
)

type TranslateRequest struct {
	Text string `json:"text"`
}

type TranslateResponse struct {
	Translation string            `json:"translation"`
	Context     []rag.ContextCard `json:"context"`
}

var (
	openAIKey      string
	embeddingModel string
)

func init() {
	// Load .env file if exists
	godotenv.Load()

	// Get OpenAI API key
	openAIKey = os.Getenv("OPENAI_API_KEY")
	if openAIKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Get embedding model (default: text-embedding-3-small)
	embeddingModel = os.Getenv("EMBEDDING_MODEL")
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-small"
	}
}

func main() {
	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnvInt("DB_PORT", 5432)
	dbUser := getEnv("DB_USER", "arkham")
	dbPassword := getEnv("DB_PASSWORD", "arkham")
	dbName := getEnv("DB_NAME", "arkham_localize")

	database, err := db.Connect(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// HTTP handlers
	http.HandleFunc("/translate", translateHandler(database))
	http.HandleFunc("/health", healthHandler)

	// Start server
	port := getEnv("PORT", "3001")
	log.Printf("🚀 Server starting on http://localhost:%s", port)
	log.Printf("📝 POST /translate - Translate English text to Italian")
	log.Printf("💚 GET  /health - Health check")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// enableCORS sets CORS headers for all responses
func enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "3600")
}

// corsMiddleware wraps handlers with CORS support
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func translateHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req TranslateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if req.Text == "" {
			http.Error(w, "Text field is required", http.StatusBadRequest)
			return
		}

		// Step 1: Generate embedding for the query text
		queryEmbedding, err := embeddings.GetEmbedding(req.Text, openAIKey, embeddingModel)
		if err != nil {
			log.Printf("Error generating embedding: %v", err)
			http.Error(w, fmt.Sprintf("Failed to generate embedding: %v", err), http.StatusInternalServerError)
			return
		}

		// Step 2: Retrieve similar cards from database
		contextCards, err := rag.RetrieveSimilarCards(database, queryEmbedding, 5)
		if err != nil {
			log.Printf("Error retrieving similar cards: %v", err)
			http.Error(w, fmt.Sprintf("Failed to retrieve context: %v", err), http.StatusInternalServerError)
			return
		}

		// Step 3: Generate translation with context
		translation, err := rag.GenerateTranslation(req.Text, contextCards, openAIKey)
		if err != nil {
			log.Printf("Error generating translation: %v", err)
			http.Error(w, fmt.Sprintf("Failed to generate translation: %v", err), http.StatusInternalServerError)
			return
		}

		// Step 4: Return response
		response := TranslateResponse{
			Translation: translation,
			Context:     contextCards,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "arkham-localize-backend",
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
