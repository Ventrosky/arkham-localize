# Arkham LCG Consistent Content Translator (LCCCT)

A Retrieval-Augmented Generation (RAG) system for translating Arkham Horror: The Card Game fan content into consistent Italian using official translations as context.

## Architecture

- **Backend**: Go with PostgreSQL + pgvector
- **Frontend**: React + Vite + Tailwind CSS
- **Vector Database**: PostgreSQL with pgvector extension
- **Embeddings**: OpenAI text-embedding-3-small
- **LLM**: OpenAI GPT-4o

## Project Structure

```
arkham-localize/
├── backend/              # Go backend
│   ├── cmd/
│   │   └── server/      # Main server entry point
│   ├── internal/
│   │   ├── rag/         # RAG logic (retrieval, prompt construction)
│   │   ├── embeddings/  # Embedding generation
│   │   └── db/          # PostgreSQL client
│   └── go.mod
├── frontend/            # React + Vite frontend
│   ├── src/
│   │   ├── components/  # UI components
│   │   └── lib/         # API client
│   └── package.json
├── scripts/             # Data ingestion pipeline
│   ├── ingest.py        # Main ingestion script
│   ├── download_data.sh # Download arkhamdb-json-data
│   └── requirements.txt
├── docker-compose.yml   # PostgreSQL with pgvector
└── .data/               # Git-ignored: cached arkhamdb-json-data repo
```

## Quick Start

### 1. Prerequisites

- Go 1.21+
- Python 3.8+
- Node.js 18+
- Docker & Docker Compose

### 2. Setup Data

```bash
# Download arkhamdb-json-data repository
bash scripts/download_data.sh

# Start PostgreSQL
docker-compose up -d postgres

# Install Python dependencies
cd scripts
pip install -r requirements.txt

# Configure environment (create .env from .env.example)
cp .env.example .env
# Edit .env with your OpenAI API key

# Run ingestion pipeline
python ingest.py
```

### 3. Setup Backend

```bash
cd backend
go mod tidy

# Configure environment
cp .env.example .env
# Edit .env with your configuration

# Run server
go run cmd/server/main.go
```

### 4. Setup Frontend

```bash
cd frontend
npm install

# Configure environment (optional)
cp .env.example .env

# Run dev server
npm run dev
```

## Key Features

- **Symbol Preservation**: Game symbols (e.g., `[action]`, `[elder_sign]`) are preserved exactly as entered
- **Context-Aware**: Uses official Italian card translations to ensure terminology consistency
- **Transparent**: Shows which cards were used as context for each translation

## Development

See individual README files in each directory for more details:
- [Backend README](backend/README.md)
- [Frontend README](frontend/README.md)
- [Scripts README](scripts/README.md)

## Data Source

Card data is sourced from [arkhamdb-json-data](https://github.com/Kamalisk/arkhamdb-json-data).
