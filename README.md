# Arkham Localize

A Retrieval-Augmented Generation (RAG) system for translating Arkham Horror: The Card Game fan content into consistent Italian using official translations as context.

**Arkham Horror LCG Consistent Content Translator** - Grounded translations using vector similarity for consistent terminology.

## 🚀 Quick Start (3 Commands)

```bash
git clone <repository-url>
cd arkham-localize
make setup          # Downloads data, creates env files
# Edit scripts/.env and add your OPENAI_API_KEY
make all            # Installs dependencies, starts DB, runs ingestion
```

See [Quick Start](#quick-start) below for details.

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
├── scripts/             # Utility scripts
│   └── download_data.sh # Download arkhamdb-json-data
├── docker-compose.yml   # PostgreSQL with pgvector
└── .data/               # Git-ignored: cached arkhamdb-json-data repo
```

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Make (optional but recommended)

### Easy Setup (Recommended)

Use the Makefile for a streamlined setup experience:

```bash
# 1. Complete initial setup (downloads data, creates env files)
make setup

# 2. Edit scripts/.env and add your OPENAI_API_KEY
# Then install all dependencies and run ingestion
make all
```

This will:
- Download the arkhamdb-json-data repository
- Create all necessary `.env` files
- Install Go and Node.js dependencies
- Build the ingestion tool (Go)
- Start PostgreSQL
- Run the data ingestion pipeline

### Manual Setup

If you prefer to set up manually:

#### 1. Setup Data

```bash
# Download arkhamdb-json-data repository
bash scripts/download_data.sh

# Start PostgreSQL
docker-compose up -d postgres

# Build ingestion tool (Go)
cd backend
go build -o ../bin/ingest ./backend/cmd/ingest
cd ..

# Set OpenAI API key
export OPENAI_API_KEY=your-key-here

# Run ingestion pipeline
./bin/ingest -clear -data .data/arkhamdb-json-data
```

#### 2. Setup Backend

```bash
cd backend
go mod tidy

# Configure environment
cp .env.example .env
# Edit .env with your configuration

# Run server
go run cmd/server/main.go
```

#### 3. Setup Frontend

```bash
cd frontend
npm install

# Configure environment (optional)
cp .env.example .env

# Run dev server
npm run dev
```

### Development

After initial setup, use Makefile shortcuts:

```bash
make check      # Validate your setup
make validate   # Test extraction without API calls
make db-up      # Start PostgreSQL
make db-down    # Stop PostgreSQL
make ingest     # Re-run ingestion (requires OPENAI_API_KEY)
make backend    # Run Go backend
make frontend   # Run React frontend
make dev        # Start all services (db + backend + frontend)
```

## Key Features

- **Symbol Preservation**: Game symbols (e.g., `[action]`, `[elder_sign]`) are preserved exactly as entered
- **Context-Aware**: Uses official Italian card translations to ensure terminology consistency
- **Transparent**: Shows which cards were used as context for each translation

## Documentation

- [Installation Guide](INSTALL.md) - Detailed setup instructions
- [Backend README](backend/README.md) - Backend development details
- [Frontend README](frontend/README.md) - Frontend development details

**Note:** Data ingestion is now handled by the Go tool in `backend/cmd/ingest`.

## Data Source

Card data is sourced from [arkhamdb-json-data](https://github.com/Kamalisk/arkhamdb-json-data).

## Acknowledgments

This software is not affiliated or endorsed by Arkham Horror: The Card Game, (c) 2016 Fantasy Flight Games.
