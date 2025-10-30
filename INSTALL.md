# Installation Guide

## Prerequisites

Before you begin, ensure you have the following installed:

- **Docker** and **Docker Compose** (for PostgreSQL with pgvector)
- **Go 1.21+** (for backend and ingestion)
- **Node.js 18+** (for frontend)
- **Make** (recommended, but optional - see manual setup below)

## Quick Installation (Using Makefile)

The easiest way to set up the project:

```bash
# 1. Clone the repository
git clone <repository-url>
cd arkham-localize

# 2. Run initial setup (downloads data, creates env files)
make setup

# 3. Edit scripts/.env and add your OPENAI_API_KEY
nano scripts/.env  # or use your preferred editor

# 4. Install all dependencies and run ingestion
make all
```

The `make all` command will:
1. Install Go dependencies  
2. Install Node.js dependencies
3. Build the ingestion tool (Go)
4. Start PostgreSQL database
5. Run the data ingestion pipeline

## Manual Installation (Without Make)

If you don't have Make installed, follow these steps:

### Step 1: Download Data

```bash
bash scripts/download_data.sh
```

### Step 2: Create Environment Files (Optional)

Environment variables can be set directly or via `.env` files:

```bash
# Set OpenAI API key
export OPENAI_API_KEY=your-key-here

# Optional: Create .env files for backend/frontend
cd backend
cp .env.example .env  # if exists

cd ../frontend
cp .env.example .env  # if exists
```

### Step 3: Install Dependencies

```bash
# Go dependencies
go mod tidy

# Node.js dependencies
cd frontend
npm install
```

### Step 4: Build Ingestion Tool

```bash
cd backend
go build -o ../bin/ingest ./cmd/ingest
cd ..
```

### Step 5: Start Database

```bash
docker-compose up -d postgres
```

### Step 6: Run Ingestion

```bash
export OPENAI_API_KEY=your-key-here
./bin/ingest -clear -data .data/arkhamdb-json-data
```

## Verification

After installation, verify everything works:

```bash
# Check setup status
make check  # or manually check:
# - Data directory exists: .data/arkhamdb-json-data
# - Environment files created
# - Dependencies installed
# - Database running

# Test extraction (no API calls)
make validate
```

## Troubleshooting

### "make: command not found"

Install Make:
- **Ubuntu/Debian**: `sudo apt-get install make`
- **macOS**: `xcode-select --install`
- **Windows**: Use WSL or install via Chocolatey

Or follow the [Manual Installation](#manual-installation-without-make) steps above.

### "Docker: command not found"

Install Docker Desktop:
- [Docker Desktop for Linux](https://docs.docker.com/desktop/install/linux-install/)
- [Docker Desktop for macOS](https://docs.docker.com/desktop/install/mac-install/)
- [Docker Desktop for Windows](https://docs.docker.com/desktop/install/windows-install/)

### "OPENAI_API_KEY not set"

1. Get your API key from [OpenAI Platform](https://platform.openai.com/api-keys)
2. Set it as an environment variable:
   ```bash
   export OPENAI_API_KEY=sk-your-actual-key-here
   ```
   Or use the `-openai-key` flag when running `./bin/ingest`

### "Data directory not found"

Run the download script:
```bash
bash scripts/download_data.sh
```

### PostgreSQL connection errors

1. Check if PostgreSQL is running:
   ```bash
   docker-compose ps postgres
   ```

2. Start PostgreSQL if not running:
   ```bash
   docker-compose up -d postgres
   ```

3. Verify connection settings (defaults are in the ingestion tool flags)

## Next Steps

After successful installation:

1. **Start development environment**:
   ```bash
   make dev  # Starts DB + backend + frontend
   ```

2. **Or run services separately**:
   ```bash
   make db-up      # Start PostgreSQL
   make backend    # Run Go backend
   make frontend   # Run React frontend
   ```

3. **Re-run ingestion** (if needed):
   ```bash
   make ingest
   ```

## Uninstallation

To clean up:

```bash
# Remove build artifacts
make clean

# Remove data directory
make clean-data

# Remove database volumes
make clean-db

# Remove everything
make clean-all
```

