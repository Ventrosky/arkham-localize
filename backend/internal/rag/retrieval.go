package rag

import (
	"database/sql"
	"fmt"

	"github.com/pgvector/pgvector-go"
)

// ContextCard represents a card used as context for translation
type ContextCard struct {
	CardName       string `json:"card_name"`
	CardCode       string `json:"card_code"`
	IsBack         bool   `json:"is_back"`
	EnglishText    string `json:"english_text"`
	TranslatedText string `json:"translated_text"` // Text in the target language
}

// RetrieveSimilarCards retrieves the most similar cards from the database
// using vector similarity search, filtered by target language
// language is one of: "it", "fr", "de", "es"
func RetrieveSimilarCards(db *sql.DB, queryEmbedding []float32, limit int, language string) ([]ContextCard, error) {
	if len(queryEmbedding) == 0 {
		return nil, fmt.Errorf("query embedding is empty")
	}

	// Validate language
	validLanguages := map[string]string{
		"it": "it_text",
		"fr": "fr_text",
		"de": "de_text",
		"es": "es_text",
	}
	langColumn, ok := validLanguages[language]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s (supported: it, fr, de, es)", language)
	}

	vector := pgvector.NewVector(queryEmbedding)

	query := fmt.Sprintf(`
		SELECT card_code, card_name, is_back, english_text, COALESCE(%s, '') as translated_text
		FROM card_embeddings
		WHERE embedding IS NOT NULL AND card_code IS NOT NULL AND %s IS NOT NULL
		ORDER BY embedding <-> $1
		LIMIT $2
	`, langColumn, langColumn)

	rows, err := db.Query(query, vector, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar cards: %w", err)
	}
	defer rows.Close()

	cards := []ContextCard{} // Initialize as empty slice, not nil
	for rows.Next() {
		var card ContextCard
		if err := rows.Scan(
			&card.CardCode,
			&card.CardName,
			&card.IsBack,
			&card.EnglishText,
			&card.TranslatedText,
		); err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return cards, nil
}
