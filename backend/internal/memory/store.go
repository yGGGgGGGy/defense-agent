package memory

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store is the memory storage backed by PostgreSQL with pgvector
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates a memory store
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Entry is a memory record
type Entry struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	AgentType string    `json:"agent_type"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Embedding []float32 `json:"embedding,omitempty"`
}

// Save persists a memory entry
func (s *Store) Save(ctx context.Context, e *Entry) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO memory_entries (task_id, agent_type, key, value, embedding)
		 VALUES ($1, $2, $3, $4, $5)`,
		e.TaskID, e.AgentType, e.Key, e.Value, e.Embedding,
	)
	return err
}

// Query retrieves memory entries for a task
func (s *Store) Query(ctx context.Context, taskID string) ([]Entry, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id::text, task_id, agent_type, key, value FROM memory_entries
		 WHERE task_id = $1 ORDER BY created_at DESC LIMIT 50`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.TaskID, &e.AgentType, &e.Key, &e.Value); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// SearchSimilar finds entries similar to the given embedding via cosine distance
func (s *Store) SearchSimilar(ctx context.Context, embedding []float32, limit int) ([]Entry, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id::text, task_id, agent_type, key, value,
		        1 - (embedding <=> $1) AS similarity
		 FROM memory_entries
		 WHERE embedding IS NOT NULL
		 ORDER BY embedding <=> $1
		 LIMIT $2`, embedding, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		var sim float32
		if err := rows.Scan(&e.ID, &e.TaskID, &e.AgentType, &e.Key, &e.Value, &sim); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
