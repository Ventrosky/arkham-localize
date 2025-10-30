# Scripts Directory

This directory contains utility scripts for the Arkham LCG translation system.

## Scripts

- `download_data.sh` - Downloads the arkhamdb-json-data repository
- `setup_db.sql` - Database schema definition (auto-applied by ingestion tool)

## Data Ingestion

Data ingestion is handled by the Go tool in `cmd/ingest/`. See the main [README](../README.md) for instructions.

To run ingestion:
```bash
# Build the tool
go build -o bin/ingest ./cmd/ingest

# Run ingestion (set OPENAI_API_KEY first)
export OPENAI_API_KEY=your-key-here
./bin/ingest -clear -data .data/arkhamdb-json-data
```

Or use the Makefile:
```bash
make ingest
```

## Database Schema

The `card_embeddings` table stores:
- `card_code`: Card identifier
- `card_name`: Card name
- `is_back`: Boolean indicating if this is the back side of a card
- `english_text`: Original English card text
- `italian_text`: Official Italian translation
- `embedding`: Vector embedding (1536 dimensions for text-embedding-3-small)
