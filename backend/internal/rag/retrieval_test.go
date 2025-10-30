package rag

import (
	"database/sql"
	"os"
	"testing"

	"github.com/pgvector/pgvector-go"
	"github.com/ventrosky/arkham-localize/backend/internal/db"
)

func TestRetrieveSimilarCards_EmptyEmbedding(t *testing.T) {
	var db *sql.DB
	emptyEmbedding := []float32{}

	cards, err := RetrieveSimilarCards(db, emptyEmbedding, 5)

	if err == nil {
		t.Error("Expected error for empty embedding, got nil")
	}

	if cards != nil {
		t.Errorf("Expected nil cards, got %v", cards)
	}

	expectedError := "query embedding is empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestRetrieveSimilarCards_RealDatabase(t *testing.T) {
	// Skip if DB_TEST environment variable is not set
	if os.Getenv("DB_TEST") == "" {
		t.Skip("Skipping integration test (set DB_TEST=1 to enable)")
	}

	// Connect to database
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := 5432
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "arkham"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "arkham"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "arkham_localize"
	}

	database, err := db.Connect(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Find Machete card and get its embedding
	var macheteCode string
	var macheteName string

	err = database.QueryRow(`
		SELECT card_code, card_name
		FROM card_embeddings
		WHERE LOWER(card_name) LIKE '%machete%'
		AND embedding IS NOT NULL
		AND card_code IS NOT NULL
		AND is_back = false
		LIMIT 1
	`).Scan(&macheteCode, &macheteName)
	if err != nil {
		t.Fatalf("Failed to find Machete card: %v", err)
	}

	t.Logf("Found Machete: %s (%s)", macheteName, macheteCode)

	// Get the embedding using pgvector
	var embeddingVector pgvector.Vector
	err = database.QueryRow(`
		SELECT embedding
		FROM card_embeddings
		WHERE card_code = $1 AND embedding IS NOT NULL
		LIMIT 1
	`, macheteCode).Scan(&embeddingVector)
	if err != nil {
		t.Fatalf("Failed to get Machete embedding: %v", err)
	}

	// Extract float32 slice from pgvector.Vector
	// pgvector.Vector implements a Slice() method that returns []float32
	embedding := embeddingVector.Slice()

	if len(embedding) == 0 {
		t.Fatalf("Embedding is empty")
	}

	t.Logf("Retrieved embedding with %d dimensions", len(embedding))

	// Test retrieval - search for cards similar to Machete
	limit := 5
	cards, err := RetrieveSimilarCards(database, embedding, limit)
	if err != nil {
		t.Fatalf("Failed to retrieve similar cards: %v", err)
	}

	if len(cards) == 0 {
		t.Fatal("Expected at least one card, got none")
	}

	if len(cards) > limit {
		t.Errorf("Expected at most %d cards, got %d", limit, len(cards))
	}

	// The first result should be Machete itself (using its own embedding)
	firstCard := cards[0]
	t.Logf("First retrieved card: %s (%s)", firstCard.CardName, firstCard.CardCode)
	t.Logf("Looking for: %s (%s)", macheteName, macheteCode)

	// Verify that the first card is Machete
	if firstCard.CardCode != macheteCode {
		t.Errorf("Expected first card to be Machete (%s), got: %s (%s)", macheteCode, firstCard.CardName, firstCard.CardCode)

		// Log all results for debugging
		for i, card := range cards {
			t.Logf("  Result %d: %s (%s)", i+1, card.CardName, card.CardCode)
		}
	}

	// Verify Machete is in the results
	found := false
	for i, card := range cards {
		if card.CardCode == macheteCode {
			t.Logf("Found Machete at position %d", i+1)
			found = true
			if i == 0 {
				t.Logf("✅ Machete is correctly the first result!")
			} else {
				t.Logf("⚠️  Machete is at position %d, expected first", i+1)
			}
			break
		}
	}

	if !found {
		t.Errorf("Expected Machete (%s) to be in results", macheteCode)
	}
}
