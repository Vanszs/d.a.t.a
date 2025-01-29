package adapters

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db     *sql.DB
	dbPath string
}

func NewSQLiteStore(dbPath string) *SQLiteStore {
	return &SQLiteStore{
		dbPath: dbPath,
	}
}

func (s *SQLiteStore) Connect(ctx context.Context) error {
	db, err := sql.Open("sqlite3", s.dbPath)
	if err != nil {
		return fmt.Errorf("opening sqlite db: %w", err)
	}
	s.db = db

	if err := s.createTables(ctx); err != nil {
		return fmt.Errorf("creating tables: %w", err)
	}

	return nil
}

func (s *SQLiteStore) Close() error {
	return nil
}

func (s *SQLiteStore) createTables(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS memories (
            id TEXT PRIMARY KEY,
            type TEXT NOT NULL,
            content BLOB,
            embedding BLOB,
            metadata TEXT,
            tags TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )`,
		`CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type)`,
		`CREATE INDEX IF NOT EXISTS idx_memories_created ON memories(created_at)`,
	}

	for _, query := range queries {
		if _, err := s.db.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStore) CreateMemory(ctx context.Context, memory Memory) error {
	metadata, err := json.Marshal(memory.Metadata)
	if err != nil {
		return fmt.Errorf("marshaling metadata: %w", err)
	}

	tags, err := json.Marshal(memory.Tags)
	if err != nil {
		return fmt.Errorf("marshaling tags: %w", err)
	}

	embedding, err := json.Marshal(memory.Embedding)
	if err != nil {
		return fmt.Errorf("marshaling embedding: %w", err)
	}

	query := `
        INSERT INTO memories (id, type, content, embedding, metadata, tags, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `
	_, err = s.db.ExecContext(ctx, query,
		memory.ID,
		memory.Type,
		memory.Content,
		embedding,
		metadata,
		tags,
		memory.CreatedAt,
		memory.UpdatedAt,
	)
	return err
}

func (s *SQLiteStore) SearchByEmbedding(ctx context.Context, embedding []float64, opts SearchOptions) ([]Memory, error) {
	// Get all memories first
	query := `
        SELECT id, type, content, embedding, metadata, tags, created_at, updated_at
        FROM memories
    `
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying memories: %w", err)
	}
	defer rows.Close()

	var memories []Memory
	var similarities []float64

	// Calculate similarities and store with memories
	for rows.Next() {
		var m Memory
		var embeddingBlob []byte
		var metadataStr, tagsStr string

		err := rows.Scan(&m.ID, &m.Type, &m.Content, &embeddingBlob, &metadataStr, &tagsStr, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning memory: %w", err)
		}

		// Unmarshal embedding
		var memoryEmbedding []float64
		if err := json.Unmarshal(embeddingBlob, &memoryEmbedding); err != nil {
			return nil, fmt.Errorf("unmarshaling embedding: %w", err)
		}

		// Calculate similarity
		similarity := s.vectorOps.Similarity(embedding, memoryEmbedding)

		// Skip if below minimum similarity
		if similarity < opts.MinSimilarity {
			continue
		}

		// Unmarshal metadata and tags
		if err := json.Unmarshal([]byte(metadataStr), &m.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshaling metadata: %w", err)
		}
		if err := json.Unmarshal([]byte(tagsStr), &m.Tags); err != nil {
			return nil, fmt.Errorf("unmarshaling tags: %w", err)
		}

		memories = append(memories, m)
		similarities = append(similarities, similarity)
	}

	// Sort by similarity
	sortMemoriesBySimilarity(memories, similarities)

	// Return top k results
	if opts.Limit > 0 && opts.Limit < len(memories) {
		memories = memories[:opts.Limit]
	}

	return memories, nil
}

func sortMemoriesBySimilarity(memories []Memory, similarities []float64) {
	sort.Slice(memories, func(i, j int) bool {
		return similarities[i] > similarities[j]
	})
}
