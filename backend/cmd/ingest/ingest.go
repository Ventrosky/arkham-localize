package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

func setupDatabase(db *sql.DB) error {
	queries := []string{
		"CREATE EXTENSION IF NOT EXISTS vector",
		`CREATE TABLE IF NOT EXISTS card_embeddings (
			id SERIAL PRIMARY KEY,
			card_code TEXT NOT NULL,
			card_name TEXT NOT NULL,
			is_back BOOLEAN DEFAULT FALSE,
			english_text TEXT NOT NULL,
			it_text TEXT,
			fr_text TEXT,
			de_text TEXT,
			es_text TEXT,
			embedding vector(1536),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS card_embeddings_embedding_idx 
		 ON card_embeddings 
		 USING ivfflat (embedding vector_cosine_ops)
		 WITH (lists = 100)`,
		`CREATE INDEX IF NOT EXISTS card_embeddings_card_code_idx ON card_embeddings(card_code)`,
		`CREATE INDEX IF NOT EXISTS card_embeddings_card_name_idx ON card_embeddings(card_name)`,
		`CREATE INDEX IF NOT EXISTS card_embeddings_is_back_idx ON card_embeddings(is_back)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	fmt.Println("✓ Database schema initialized")
	return nil
}

func clearDatabase(db *sql.DB) error {
	if _, err := db.Exec("TRUNCATE TABLE card_embeddings"); err != nil {
		return fmt.Errorf("failed to clear database: %w", err)
	}
	fmt.Println("✓ Cleared existing data")
	return nil
}

func extractCardText(card Card, isBack bool) string {
	if isBack {
		return strings.TrimSpace(card.BackText)
	}
	if card.Text != "" {
		return strings.TrimSpace(card.Text)
	}
	return strings.TrimSpace(card.RealText)
}

type TranslationDict map[string]map[string]string

// Supported languages for translation
var supportedLanguages = []string{"it", "fr", "de", "es"}

func loadTranslations(dataPath, language string) (TranslationDict, error) {
	translationsDir := filepath.Join(dataPath, "translations", language, "pack")
	translations := make(TranslationDict)

	if _, err := os.Stat(translationsDir); os.IsNotExist(err) {
		return translations, nil
	}

	packDirs, err := filepath.Glob(filepath.Join(translationsDir, "*"))
	if err != nil {
		return nil, err
	}

	for _, packDir := range packDirs {
		if info, err := os.Stat(packDir); err != nil || !info.IsDir() {
			continue
		}

		jsonFiles, err := filepath.Glob(filepath.Join(packDir, "*.json"))
		if err != nil {
			continue
		}

		for _, jsonFile := range jsonFiles {
			data, err := os.ReadFile(jsonFile)
			if err != nil {
				continue
			}

			var cards []Card
			if err := json.Unmarshal(data, &cards); err != nil {
				continue
			}

			for _, card := range cards {
				if card.Code == "" {
					continue
				}

				if translations[card.Code] == nil {
					translations[card.Code] = make(map[string]string)
				}

				if card.Name != "" {
					translations[card.Code]["name"] = card.Name
				}

				if text := extractCardText(card, false); text != "" {
					translations[card.Code]["text"] = text
				}

				if backText := extractCardText(card, true); backText != "" {
					translations[card.Code]["back_text"] = backText
				}
			}
		}
	}

	return translations, nil
}

func findTranslation(code, name string, translations TranslationDict, isBack bool) (string, bool) {
	cardTrans, exists := translations[code]
	if !exists {
		return "", false
	}

	key := "back_text"
	if !isBack {
		key = "text"
	}

	text, exists := cardTrans[key]
	if !exists {
		return "", false
	}

	return text, true
}

func processCardFiles(dataPath string, allTranslations map[string]TranslationDict) ([]CardEntry, error) {
	packDir := filepath.Join(dataPath, "pack")
	var entries []CardEntry
	processed := 0
	skipped := 0

	packDirs, err := filepath.Glob(filepath.Join(packDir, "*"))
	if err != nil {
		return nil, err
	}

	fmt.Printf("Scanning card files in %s...\n", packDir)

	for _, packSubdir := range packDirs {
		if info, err := os.Stat(packSubdir); err != nil || !info.IsDir() {
			continue
		}

		jsonFiles, err := filepath.Glob(filepath.Join(packSubdir, "*.json"))
		if err != nil {
			continue
		}

		for _, jsonFile := range jsonFiles {
			data, err := os.ReadFile(jsonFile)
			if err != nil {
				continue
			}

			var cards []Card
			if err := json.Unmarshal(data, &cards); err != nil {
				continue
			}

			for _, card := range cards {
				if card.Code == "" {
					skipped++
					continue
				}

				// Process front text
				englishText := extractCardText(card, false)
				if englishText != "" {
					// Collect translations from all languages
					translationsMap := make(map[string]string)
					hasAnyTranslation := false
					for _, lang := range supportedLanguages {
						if transDict, ok := allTranslations[lang]; ok {
							if transText, found := findTranslation(card.Code, card.Name, transDict, false); found {
								translationsMap[lang] = transText
								hasAnyTranslation = true
							}
						}
					}
					if hasAnyTranslation {
						entries = append(entries, CardEntry{
							CardCode:     card.Code,
							CardName:     card.Name,
							IsBack:       false,
							EnglishText:  englishText,
							Translations: translationsMap,
						})
						processed++
					} else {
						skipped++
					}
				}

				// Process back text
				englishBackText := extractCardText(card, true)
				if englishBackText != "" {
					// Collect translations from all languages
					translationsMap := make(map[string]string)
					hasAnyTranslation := false
					for _, lang := range supportedLanguages {
						if transDict, ok := allTranslations[lang]; ok {
							if transText, found := findTranslation(card.Code, card.Name, transDict, true); found {
								translationsMap[lang] = transText
								hasAnyTranslation = true
							}
						}
					}
					if hasAnyTranslation {
						entries = append(entries, CardEntry{
							CardCode:     card.Code,
							CardName:     card.Name,
							IsBack:       true,
							EnglishText:  englishBackText,
							Translations: translationsMap,
						})
						processed++
					} else {
						skipped++
					}
				}

				if englishText == "" && englishBackText == "" {
					skipped++
				}

				if processed%100 == 0 {
					fmt.Printf("  Processed %d card entries...\n", processed)
				}
			}
		}
	}

	fmt.Printf("✓ Extracted %d card entries (skipped %d)\n", processed, skipped)
	return entries, nil
}

func getEmbedding(text, apiKey, model string) ([]float32, error) {
	// Simple HTTP request to OpenAI API
	url := "https://api.openai.com/v1/embeddings"

	// Properly escape JSON
	reqBody := struct {
		Model string `json:"model"`
		Input string `json:"input"`
	}{
		Model: model,
		Input: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	// Convert float64 to float32
	embedding := make([]float32, len(result.Data[0].Embedding))
	for i, v := range result.Data[0].Embedding {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

func ingestCards(db *sql.DB, entries []CardEntry, apiKey, model string, batchSize int) error {
	total := len(entries)
	inserted := 0

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		batch := entries[i:end]

		fmt.Printf("  Processing batch %d/%d...\n", i/batchSize+1, (total+batchSize-1)/batchSize)

		var wg sync.WaitGroup
		type batchItem struct {
			entry     CardEntry
			embedding []float32
			err       error
		}
		results := make([]batchItem, len(batch))

		// Generate embeddings in parallel
		for j, entry := range batch {
			wg.Add(1)
			go func(idx int, e CardEntry) {
				defer wg.Done()
				emb, err := getEmbedding(e.EnglishText, apiKey, model)
				results[idx] = batchItem{entry: e, embedding: emb, err: err}
			}(j, entry)
		}
		wg.Wait()

		// Insert batch
		batchData := make([][]interface{}, 0, len(batch))
		for _, result := range results {
			if result.err != nil {
				fmt.Printf("  Warning: Error generating embedding for '%s' (%s): %v\n",
					result.entry.CardName, map[bool]string{false: "front", true: "back"}[result.entry.IsBack], result.err)
				continue
			}

			vector := pgvector.NewVector(result.embedding)
			// Get translations for each language (NULL if not available)
			itText := result.entry.Translations["it"]
			frText := result.entry.Translations["fr"]
			deText := result.entry.Translations["de"]
			esText := result.entry.Translations["es"]
			batchData = append(batchData, []interface{}{
				result.entry.CardCode,
				result.entry.CardName,
				result.entry.IsBack,
				result.entry.EnglishText,
				itText,
				frText,
				deText,
				esText,
				vector,
			})
		}

		if len(batchData) > 0 {
			if err := insertBatch(db, batchData); err != nil {
				return fmt.Errorf("failed to insert batch: %w", err)
			}
			inserted += len(batchData)
		}
	}

	fmt.Printf("✓ Ingested %d card entries into database\n", inserted)
	return nil
}

func insertBatch(db *sql.DB, batchData [][]interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt := `INSERT INTO card_embeddings (card_code, card_name, is_back, english_text, it_text, fr_text, de_text, es_text, embedding)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	for _, row := range batchData {
		if _, err := tx.Exec(stmt, row...); err != nil {
			return err
		}
	}

	return tx.Commit()
}
