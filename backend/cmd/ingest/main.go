package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Card struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Text     string `json:"text"`
	RealText string `json:"real_text"`
	BackText string `json:"back_text"`
}

type CardEntry struct {
	CardCode    string
	CardName    string
	IsBack      bool
	EnglishText string
	ItalianText string
}

var (
	dataDir        = flag.String("data", ".data/arkhamdb-json-data", "Path to arkhamdb-json-data directory")
	openAIKey      = flag.String("openai-key", "", "OpenAI API key (or use OPENAI_API_KEY env var)")
	embeddingModel = flag.String("embedding-model", "text-embedding-3-small", "OpenAI embedding model")
	batchSize      = flag.Int("batch-size", 50, "Batch size for embeddings")
	clearDB        = flag.Bool("clear", false, "Clear existing data before ingestion")
	limitEntries   = flag.Int("limit", 0, "Limit number of entries to process (0 = all, useful for testing)")
	dbHost         = flag.String("db-host", "localhost", "PostgreSQL host")
	dbPort         = flag.Int("db-port", 5432, "PostgreSQL port")
	dbUser         = flag.String("db-user", "arkham", "PostgreSQL user")
	dbPassword     = flag.String("db-password", "arkham", "PostgreSQL password")
	dbName         = flag.String("db-name", "arkham_localize", "PostgreSQL database name")
)

func main() {
	flag.Parse()

	// Load .env file if exists
	godotenv.Load()

	// Get OpenAI key from flag or env
	apiKey := *openAIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		log.Fatal("OpenAI API key required. Set OPENAI_API_KEY env var or use -openai-key flag")
	}

	// Resolve data directory
	dataPath, err := filepath.Abs(*dataDir)
	if err != nil {
		log.Fatalf("Failed to resolve data directory: %v", err)
	}

	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Println("Arkham Localize - Data Ingestion Pipeline (Go)")
	fmt.Println("=" + strings.Repeat("=", 59))
	fmt.Printf("\nData directory: %s\n", dataPath)
	fmt.Printf("Embedding model: %s\n", *embeddingModel)
	fmt.Printf("Batch size: %d\n", *batchSize)

	// Validate data directory
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		log.Fatalf("Data directory not found: %s\nRun: bash scripts/download_data.sh", dataPath)
	}

	// Connect to database
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		*dbUser, *dbPassword, *dbHost, *dbPort, *dbName)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Setup database schema
	if err := setupDatabase(db); err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Clear existing data if requested
	if *clearDB {
		if err := clearDatabase(db); err != nil {
			log.Fatalf("Failed to clear database: %v", err)
		}
	}

	// Load Italian translations
	fmt.Println("\nLoading Italian translations...")
	translations, err := loadItalianTranslations(dataPath)
	if err != nil {
		log.Fatalf("Failed to load Italian translations: %v", err)
	}
	fmt.Printf("✓ Loaded %d card translations\n", len(translations))

	// Process card files
	fmt.Println("\nExtracting card data...")
	entries, err := processCardFiles(dataPath, translations)
	if err != nil {
		log.Fatalf("Failed to process card files: %v", err)
	}

	if len(entries) == 0 {
		log.Fatal("No cards found to process")
	}

	fmt.Printf("✓ Extracted %d card entries\n", len(entries))

	// Limit entries if requested (useful for testing)
	if *limitEntries > 0 && *limitEntries < len(entries) {
		entries = entries[:*limitEntries]
		fmt.Printf("⚠️  Limited to first %d entries for testing\n", *limitEntries)
	}

	// Generate embeddings and ingest
	fmt.Printf("\nGenerating embeddings using %s...\n", *embeddingModel)
	if err := ingestCards(db, entries, apiKey, *embeddingModel, *batchSize); err != nil {
		log.Fatalf("Failed to ingest cards: %v", err)
	}

	// Print summary
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM card_embeddings").Scan(&count); err != nil {
		log.Printf("Warning: Failed to count entries: %v", err)
	} else {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Printf("✓ Ingestion complete! Total entries in database: %d\n", count)
		fmt.Println(strings.Repeat("=", 60))
	}
}
