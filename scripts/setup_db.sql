-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create table for card embeddings
CREATE TABLE IF NOT EXISTS card_embeddings (
    id SERIAL PRIMARY KEY,
    card_code TEXT NOT NULL,
    card_name TEXT NOT NULL,
    is_back BOOLEAN DEFAULT FALSE,
    english_text TEXT NOT NULL,
    italian_text TEXT NOT NULL,
    embedding vector(1536),  -- text-embedding-3-small dimension
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for vector similarity search
CREATE INDEX IF NOT EXISTS card_embeddings_embedding_idx 
ON card_embeddings 
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- Indexes for card lookup
CREATE INDEX IF NOT EXISTS card_embeddings_card_code_idx ON card_embeddings(card_code);
CREATE INDEX IF NOT EXISTS card_embeddings_card_name_idx ON card_embeddings(card_name);
CREATE INDEX IF NOT EXISTS card_embeddings_is_back_idx ON card_embeddings(is_back);

