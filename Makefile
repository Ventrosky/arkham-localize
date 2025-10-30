.PHONY: help setup install test clean all dev db-up db-down ingest backend frontend validate

# Default target
help:
	@echo "Arkham LCG Translation System - Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make setup         - Complete initial setup (downloads data, creates env files)"
	@echo "  make install       - Install all dependencies (Go, Node)"
	@echo "  make db-up         - Start PostgreSQL with pgvector"
	@echo "  make db-down       - Stop PostgreSQL"
	@echo "  make ingest        - Run data ingestion pipeline"
	@echo "  make validate      - Validate setup without running ingestion"
	@echo "  make backend       - Build and run Go backend"
	@echo "  make frontend      - Start React frontend dev server"
	@echo "  make dev           - Start all services (db + backend + frontend)"
	@echo "  make clean         - Clean build artifacts and caches"
	@echo "  make all           - Full setup and ingestion (setup + install + db-up + ingest)"
	@echo ""

# Variables
SCRIPTS_DIR := scripts
BACKEND_DIR := backend
FRONTEND_DIR := frontend
DATA_DIR := .data/arkhamdb-json-data
ENV_EXAMPLE := $(SCRIPTS_DIR)/.env.example
ENV_FILE := $(SCRIPTS_DIR)/.env

# Complete initial setup
setup:
	@echo "🚀 Setting up Arkham LCG Translation System..."
	@echo ""
	@echo "📦 Downloading arkhamdb-json-data..."
	@bash $(SCRIPTS_DIR)/download_data.sh || true
	@echo ""
	@echo "📝 Creating environment files..."
	@if [ ! -f $(ENV_FILE) ]; then \
		if [ -f $(ENV_EXAMPLE) ]; then \
			cp $(ENV_EXAMPLE) $(ENV_FILE); \
			echo "✅ Created $(ENV_FILE)"; \
			echo "⚠️  Please edit $(ENV_FILE) and add your OPENAI_API_KEY"; \
		else \
			echo "⚠️  $(ENV_EXAMPLE) not found, creating minimal $(ENV_FILE)"; \
			echo "# OpenAI API" > $(ENV_FILE); \
			echo "OPENAI_API_KEY=your_openai_api_key_here" >> $(ENV_FILE); \
			echo "# PostgreSQL connection" >> $(ENV_FILE); \
			echo "POSTGRES_HOST=localhost" >> $(ENV_FILE); \
			echo "POSTGRES_PORT=5432" >> $(ENV_FILE); \
			echo "POSTGRES_USER=arkham" >> $(ENV_FILE); \
			echo "POSTGRES_PASSWORD=arkham" >> $(ENV_FILE); \
			echo "POSTGRES_DB=arkham_localize" >> $(ENV_FILE); \
			echo "EMBEDDING_MODEL=text-embedding-3-small" >> $(ENV_FILE); \
			echo "ARKHAM_DATA_DIR=../.data/arkhamdb-json-data" >> $(ENV_FILE); \
			echo "✅ Created $(ENV_FILE)"; \
			echo "⚠️  Please edit $(ENV_FILE) and add your OPENAI_API_KEY"; \
		fi; \
	else \
		echo "✅ $(ENV_FILE) already exists"; \
	fi
	@if [ ! -f $(BACKEND_DIR)/.env ]; then \
		if [ -f $(BACKEND_DIR)/.env.example ]; then \
			cp $(BACKEND_DIR)/.env.example $(BACKEND_DIR)/.env; \
			echo "✅ Created $(BACKEND_DIR)/.env"; \
		fi; \
	fi
	@if [ ! -f $(FRONTEND_DIR)/.env ]; then \
		if [ -f $(FRONTEND_DIR)/.env.example ]; then \
			cp $(FRONTEND_DIR)/.env.example $(FRONTEND_DIR)/.env; \
			echo "✅ Created $(FRONTEND_DIR)/.env"; \
		fi; \
	fi
	@echo ""
	@echo "📝 Next steps:"
	@echo "   1. Set OPENAI_API_KEY: export OPENAI_API_KEY=your-key-here"
	@echo "   2. Run: make install"
	@echo "   3. Run: make all (or make db-up && make ingest separately)"

# Install all dependencies
install: install-go install-frontend

install-go:
	@echo "🔷 Setting up Go dependencies..."
	@cd $(BACKEND_DIR) && go mod tidy
	@echo "✅ Go dependencies installed"

install-frontend:
	@echo "📦 Setting up Node.js dependencies..."
	@cd $(FRONTEND_DIR) && npm install
	@echo "✅ Node.js dependencies installed"

# Database management
db-up:
	@echo "🐘 Starting PostgreSQL with pgvector..."
	@docker-compose up -d postgres
	@echo "⏳ Waiting for PostgreSQL to be ready..."
	@sleep 3
	@docker-compose ps postgres
	@echo "✅ PostgreSQL is ready"

db-down:
	@echo "🐘 Stopping PostgreSQL..."
	@docker-compose stop postgres
	@echo "✅ PostgreSQL stopped"

db-status:
	@docker-compose ps postgres

# Data ingestion
ingest: db-up
	@echo "📊 Running data ingestion pipeline (Go)..."
	@if [ ! -d bin ]; then \
		echo "🔨 Building ingest tool..."; \
		go build -o bin/ingest ./cmd/ingest; \
	fi
	@./bin/ingest -clear -data .data/arkhamdb-json-data
	@echo "✅ Ingestion complete!"

# Validation (without API calls)
validate: db-up
	@echo "✅ Validating setup..."
	@echo "Checking data directory..."
	@if [ ! -d .data/arkhamdb-json-data ]; then \
		echo "❌ Data directory not found. Run 'make setup' first."; \
		exit 1; \
	fi
	@echo "✅ Data directory exists"
	@echo ""
	@echo "Checking database connection..."
	@docker exec arkham-localize-db psql -U arkham -d arkham_localize -c "SELECT 1" > /dev/null 2>&1 && echo "✅ Database connection successful" || echo "❌ Database connection failed"
	@echo ""
	@echo "Checking OpenAI API key..."
	@if [ -z "$$OPENAI_API_KEY" ]; then \
		echo "⚠️  OPENAI_API_KEY not set in environment"; \
	else \
		echo "✅ OPENAI_API_KEY is set"; \
	fi

# Backend
backend:
	@echo "🔷 Starting Go backend..."
	@cd $(BACKEND_DIR) && go run cmd/server/main.go

backend-build:
	@echo "🔷 Building Go backend..."
	@cd $(BACKEND_DIR) && go build -o arkham-localize cmd/server/main.go
	@echo "✅ Backend built"

# Frontend
frontend:
	@echo "⚛️  Starting React frontend..."
	@cd $(FRONTEND_DIR) && npm run dev

frontend-build:
	@echo "⚛️  Building React frontend..."
	@cd $(FRONTEND_DIR) && npm run build
	@echo "✅ Frontend built"

# Development (start all services)
dev: db-up
	@echo "🚀 Starting development environment..."
	@echo "   Backend: http://localhost:3001"
	@echo "   Frontend: http://localhost:5173"
	@echo ""
	@echo "⚠️  Press Ctrl+C to stop all services"
	@trap 'docker-compose stop postgres' INT TERM; \
		make -j2 backend frontend

# Full setup and ingestion
all: setup install db-up ingest
	@echo ""
	@echo "🎉 Complete! Your system is ready."
	@echo ""
	@echo "Next steps:"
	@echo "  - Run 'make dev' to start backend and frontend"
	@echo "  - Or run 'make backend' and 'make frontend' in separate terminals"

# Cleanup
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@cd $(BACKEND_DIR) && go clean
	@cd $(FRONTEND_DIR) && rm -rf node_modules dist
	@echo "✅ Cleanup complete"

clean-data:
	@echo "🧹 Cleaning data directory..."
	@rm -rf $(DATA_DIR)
	@echo "✅ Data directory cleaned (run 'make setup' to re-download)"

clean-db:
	@echo "🧹 Cleaning database volumes..."
	@docker-compose down -v
	@echo "✅ Database volumes removed"

clean-all: clean clean-data clean-db
	@echo "✅ All artifacts cleaned"

# Quick validation check
check:
	@echo "🔍 Quick setup check..."
	@echo ""
	@echo "Data directory:"
	@if [ -d $(DATA_DIR) ]; then \
		echo "  ✅ $(DATA_DIR) exists"; \
	else \
		echo "  ❌ $(DATA_DIR) missing (run 'make setup')"; \
	fi
	@echo ""
	@echo "Environment files:"
	@if [ -f $(ENV_FILE) ]; then \
		echo "  ✅ $(ENV_FILE) exists"; \
		if grep -q "your_openai_api_key_here" $(ENV_FILE); then \
			echo "  ⚠️  OPENAI_API_KEY not set"; \
		else \
			echo "  ✅ OPENAI_API_KEY configured"; \
		fi; \
	else \
		echo "  ❌ $(ENV_FILE) missing (run 'make setup')"; \
	fi
	@echo ""
	@echo "Go tool:"
	@if [ -f bin/ingest ]; then \
		echo "  ✅ Ingest tool built"; \
	else \
		echo "  ⚠️  Ingest tool not built (will build on first run)"; \
	fi
	@echo ""
	@echo "Dependencies:"
	@command -v docker >/dev/null 2>&1 && echo "  ✅ Docker installed" || echo "  ❌ Docker not found"
	@command -v go >/dev/null 2>&1 && echo "  ✅ Go installed" || echo "  ❌ Go not found"
	@command -v npm >/dev/null 2>&1 && echo "  ✅ npm installed" || echo "  ❌ npm not found"
	@echo ""
	@echo "Note: Python is no longer required (ingestion now uses Go)"
	@echo ""
	@echo "Database:"
	@docker-compose ps postgres 2>/dev/null | grep -q "Up" && echo "  ✅ PostgreSQL running" || echo "  ❌ PostgreSQL not running (run 'make db-up')"

