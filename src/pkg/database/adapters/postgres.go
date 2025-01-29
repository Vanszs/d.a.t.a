package adapters

import (
	"context"
	"database/sql"
	"encoding/json"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connStr string) *PostgresStore {
	return &PostgresStore{}
}

func (s *PostgresStore) Connect(ctx context.Context) error {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	s.db = db
	return s.db.PingContext(ctx)
}

func (s *PostgresStore) CreateMemory(ctx context.Context, memory Memory) error {
	metadata, err := json.Marshal(memory.Metadata)
	if err != nil {
		return err
	}

	query := `
        INSERT INTO memories (id, type, content, embedding, metadata, tags, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err = s.db.ExecContext(ctx, query,
		memory.ID,
		memory.Type,
		memory.Content,
		memory.Embedding,
		metadata,
		memory.Tags,
		memory.CreatedAt,
		memory.UpdatedAt,
	)
	return err
}

func (s *PostgresStore) SearchByEmbedding(ctx context.Context, embedding []float64, opts SearchOptions) ([]Memory, error) {
	query := `
        SELECT id, type, content, embedding, metadata, tags, created_at, updated_at
        FROM memories
        ORDER BY embedding <-> $1
        LIMIT $2
    `
	rows, err := s.db.QueryContext(ctx, query, embedding, opts.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var m Memory
		var metadata []byte
		err := rows.Scan(&m.ID, &m.Type, &m.Content, &m.Embedding, &metadata, &m.Tags, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadata, &m.Metadata); err != nil {
			return nil, err
		}
		memories = append(memories, m)
	}
	return memories, nil
}
