# Go Backend

Minimal Go backend for the Arkham Localize RAG system (Arkham Horror LCG Consistent Content Translator).

## Prerequisites

- Go 1.21+
- PostgreSQL with pgvector (via docker-compose)

## Setup

1. Install dependencies:
   ```bash
   go mod tidy
   ```

2. Copy environment template:
   ```bash
   cp .env.example .env
   ```

3. Edit `.env` with your configuration.

## Running

```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:3001` (or PORT from .env).

## API Endpoints

### POST /translate

Translates English Arkham LCG text to multiple languages using RAG.

**Supported Languages:**
- `it` - Italian (default)
- `fr` - French
- `de` - German
- `es` - Spanish

**Request:**
```json
{
  "text": "You may spend [action] to investigate.",
  "language": "it"
}
```

**Response:**
```json
{
  "translation": "Puoi spendere [action] per investigare.",
  "context": [
    {
      "card_name": "Example Card",
      "english_text": "...",
      "translated_text": "..."
    }
  ]
}
```

**Notes:**
- If `language` is not provided, defaults to `it` (Italian)
- Only supported language codes are accepted (returns 400 for invalid languages)

