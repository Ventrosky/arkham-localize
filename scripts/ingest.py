#!/usr/bin/env python3
"""
Data Ingestion Pipeline for Arkham LCG Translation System
Downloads arkhamdb-json-data, extracts English/Italian card pairs,
generates embeddings, and stores them in PostgreSQL with pgvector.
"""

import os
import json
import sys
from pathlib import Path
from typing import Dict, List, Optional, Tuple
import psycopg2
from psycopg2.extras import execute_values
from psycopg2 import sql
from dotenv import load_dotenv
from openai import OpenAI

# Load environment variables
load_dotenv()


def get_env_or_default(key: str, default: Optional[str] = None) -> str:
    """Get environment variable or raise error if required."""
    value = os.getenv(key, default)
    if value is None:
        raise ValueError(f"Required environment variable {key} is not set")
    return value


def connect_db() -> psycopg2.extensions.connection:
    """Connect to PostgreSQL database."""
    conn = psycopg2.connect(
        host=get_env_or_default("POSTGRES_HOST", "localhost"),
        port=int(get_env_or_default("POSTGRES_PORT", "5432")),
        user=get_env_or_default("POSTGRES_USER", "arkham"),
        password=get_env_or_default("POSTGRES_PASSWORD", "arkham"),
        database=get_env_or_default("POSTGRES_DB", "arkham_localize"),
    )
    return conn


def setup_database(conn: psycopg2.extensions.connection):
    """Create table and indexes if they don't exist."""
    with conn.cursor() as cur:
        # Enable pgvector extension
        cur.execute("CREATE EXTENSION IF NOT EXISTS vector;")
        
        # Create table
        cur.execute("""
            CREATE TABLE IF NOT EXISTS card_embeddings (
                id SERIAL PRIMARY KEY,
                card_name TEXT NOT NULL,
                english_text TEXT NOT NULL,
                italian_text TEXT NOT NULL,
                embedding vector(1536),
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            );
        """)
        
        # Create index for vector similarity search
        cur.execute("""
            CREATE INDEX IF NOT EXISTS card_embeddings_embedding_idx 
            ON card_embeddings 
            USING ivfflat (embedding vector_cosine_ops)
            WITH (lists = 100);
        """)
        
        # Create index for card name lookup
        cur.execute("""
            CREATE INDEX IF NOT EXISTS card_embeddings_card_name_idx 
            ON card_embeddings(card_name);
        """)
        
        conn.commit()
        print("✓ Database schema initialized")


def get_embedding(text: str, client: OpenAI, model: str) -> List[float]:
    """Generate embedding for English text."""
    response = client.embeddings.create(
        model=model,
        input=text,
    )
    return response.data[0].embedding


def extract_card_text(card_data: Dict) -> Optional[str]:
    """Extract the text field from a card, combining multiple text fields if needed."""
    text = card_data.get("text", "")
    if not text and "real_text" in card_data:
        text = card_data["real_text"]
    return text.strip() if text else None


def find_italian_translation(
    card_code: str, 
    card_name: str,
    data_dir: Path
) -> Optional[Dict[str, str]]:
    """Find the Italian translation for a card by code."""
    # Italian translations are in translations/it/pack/
    # Structure mirrors the pack/ directory
    italian_dir = data_dir / "translations" / "it" / "pack"
    
    if not italian_dir.exists():
        return None
    
    # Search through all subdirectories in translations/it/pack
    for pack_dir in italian_dir.iterdir():
        if not pack_dir.is_dir():
            continue
        
        json_file = pack_dir / f"{card_code}.json"
        if json_file.exists():
            try:
                with open(json_file, "r", encoding="utf-8") as f:
                    it_card = json.load(f)
                    it_text = extract_card_text(it_card)
                    it_name = it_card.get("name", card_name)
                    
                    if it_text:
                        return {
                            "text": it_text,
                            "name": it_name
                        }
            except Exception as e:
                print(f"  Warning: Error reading {json_file}: {e}")
                continue
    
    return None


def process_card_files(data_dir: Path) -> List[Tuple[str, str, str]]:
    """
    Process all card JSON files and extract English/Italian pairs.
    Returns list of (card_name, english_text, italian_text) tuples.
    """
    pack_dir = data_dir / "pack"
    
    if not pack_dir.exists():
        raise FileNotFoundError(
            f"Pack directory not found at {pack_dir}. "
            f"Please run scripts/download_data.sh first."
        )
    
    cards = []
    processed = 0
    skipped = 0
    
    print(f"Scanning card files in {pack_dir}...")
    
    # Iterate through all pack subdirectories
    for pack_subdir in pack_dir.iterdir():
        if not pack_subdir.is_dir():
            continue
        
        # Process each JSON file in the pack subdirectory
        for json_file in pack_subdir.glob("*.json"):
            try:
                with open(json_file, "r", encoding="utf-8") as f:
                    card_data = json.load(f)
                
                # Extract card code for finding Italian translation
                card_code = card_data.get("code")
                card_name = card_data.get("name", "")
                english_text = extract_card_text(card_data)
                
                if not english_text or not card_code:
                    skipped += 1
                    continue
                
                # Find Italian translation
                it_translation = find_italian_translation(
                    card_code, 
                    card_name,
                    data_dir
                )
                
                if not it_translation:
                    skipped += 1
                    continue
                
                cards.append((
                    card_name,
                    english_text,
                    it_translation["text"]
                ))
                processed += 1
                
                if processed % 100 == 0:
                    print(f"  Processed {processed} cards...")
                    
            except Exception as e:
                print(f"  Warning: Error processing {json_file}: {e}")
                skipped += 1
                continue
    
    print(f"✓ Extracted {processed} English/Italian card pairs (skipped {skipped})")
    return cards


def clear_existing_data(conn: psycopg2.extensions.connection):
    """Clear existing embeddings table."""
    with conn.cursor() as cur:
        cur.execute("TRUNCATE TABLE card_embeddings;")
        conn.commit()
        print("✓ Cleared existing data")


def ingest_cards(
    cards: List[Tuple[str, str, str]],
    client: OpenAI,
    embedding_model: str,
    conn: psycopg2.extensions.connection,
    batch_size: int = 50
):
    """Generate embeddings and store cards in database."""
    print(f"\nGenerating embeddings using {embedding_model}...")
    
    total = len(cards)
    inserted = 0
    
    # Process in batches
    for i in range(0, total, batch_size):
        batch = cards[i:i + batch_size]
        batch_data = []
        
        print(f"  Processing batch {i // batch_size + 1}/{(total + batch_size - 1) // batch_size}...")
        
        for card_name, english_text, italian_text in batch:
            try:
                # Generate embedding for English text
                embedding = get_embedding(english_text, client, embedding_model)
                batch_data.append((card_name, english_text, italian_text, embedding))
                
            except Exception as e:
                print(f"  Warning: Error generating embedding for '{card_name}': {e}")
                continue
        
        # Insert batch into database
        if batch_data:
            try:
                with conn.cursor() as cur:
                    execute_values(
                        cur,
                        """
                        INSERT INTO card_embeddings (card_name, english_text, italian_text, embedding)
                        VALUES %s
                        """,
                        batch_data,
                        template="(%s, %s, %s, %s::vector)",
                        page_size=batch_size
                    )
                    conn.commit()
                    inserted += len(batch_data)
                    
            except Exception as e:
                print(f"  Error inserting batch: {e}")
                conn.rollback()
                continue
    
    print(f"✓ Ingested {inserted} cards into database")


def main():
    """Main ingestion pipeline."""
    print("=" * 60)
    print("Arkham LCG Data Ingestion Pipeline")
    print("=" * 60)
    
    # Load configuration
    data_dir = Path(get_env_or_default("ARKHAM_DATA_DIR", ".data/arkhamdb-json-data"))
    api_key = get_env_or_default("OPENAI_API_KEY")
    embedding_model = get_env_or_default("EMBEDDING_MODEL", "text-embedding-3-small")
    
    # Validate data directory exists
    if not data_dir.exists():
        print(f"\n❌ Data directory not found: {data_dir}")
        print("   Please run: bash scripts/download_data.sh")
        sys.exit(1)
    
    # Initialize OpenAI client
    print(f"\nInitializing OpenAI client (model: {embedding_model})...")
    client = OpenAI(api_key=api_key)
    
    # Connect to database
    print("Connecting to PostgreSQL...")
    conn = connect_db()
    try:
        setup_database(conn)
        
        # Process card files
        print("\nExtracting card data...")
        cards = process_card_files(data_dir)
        
        if not cards:
            print("❌ No cards found to process. Exiting.")
            sys.exit(1)
        
        # Clear existing data (optional - comment out to append)
        print("\nClearing existing embeddings...")
        clear_existing_data(conn)
        
        # Generate embeddings and ingest
        ingest_cards(cards, client, embedding_model, conn)
        
        # Print summary
        with conn.cursor() as cur:
            cur.execute("SELECT COUNT(*) FROM card_embeddings;")
            count = cur.fetchone()[0]
        
        print("\n" + "=" * 60)
        print(f"✓ Ingestion complete! Total cards in database: {count}")
        print("=" * 60)
        
    finally:
        conn.close()


if __name__ == "__main__":
    main()

