package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Chain is a tamper-evident linked chain of audit records
type Chain struct {
	pool       *pgxpool.Pool
	lastHash   string
	mu         sync.RWMutex
}

// NewChain creates a new audit chain
func NewChain(pool *pgxpool.Pool) *Chain {
	return &Chain{
		pool:     pool,
		lastHash: "0000000000000000000000000000000000000000000000000000000000000000", // genesis hash
	}
}

// Append adds a record to the chain, linking it to the previous record
func (c *Chain) Append(ctx context.Context, r *Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	prevHash := c.lastHash
	r.PrevHash = prevHash
	r.Timestamp = time.Now().UTC() // use insertion time for correct ordering

	// Recompute hash with linked prev_hash
	r.RecordHash = r.ComputeHash()
	r.PrevHash = prevHash // restore after ComputeHash clears it

	evidenceJSON, _ := json.Marshal(r.Evidence)

	_, err := c.pool.Exec(ctx,
		`INSERT INTO audit_records (task_id, sub_task_id, agent_type, agent_id, decision, rationale,
		 evidence_json, confidence, risk_level, status, timestamp, prev_hash, record_hash, requires_human)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		r.TaskID, r.SubTaskID, r.AgentType, r.AgentID, r.Decision, r.Rationale,
		string(evidenceJSON), r.Confidence, r.RiskLevel, r.Status, r.Timestamp,
		r.PrevHash, r.RecordHash, r.HumanApprovalReq,
	)
	if err != nil {
		return fmt.Errorf("append audit record: %w", err)
	}

	c.lastHash = r.RecordHash
	log.Debug().Str("task_id", r.TaskID).Str("decision", r.Decision).Msg("audit record appended")
	return nil
}

// GetTrail retrieves the full audit chain for a task
func (c *Chain) GetTrail(ctx context.Context, taskID string) ([]Record, error) {
	rows, err := c.pool.Query(ctx,
		`SELECT id::text, task_id, coalesce(sub_task_id,''), agent_type, coalesce(agent_id,''),
		 decision, rationale, coalesce(evidence_json::text,''), confidence, risk_level, status,
		 timestamp, prev_hash, record_hash, requires_human
		 FROM audit_records WHERE task_id = $1 ORDER BY timestamp ASC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		var evStr string
		if err := rows.Scan(&r.ID, &r.TaskID, &r.SubTaskID, &r.AgentType, &r.AgentID,
			&r.Decision, &r.Rationale, &evStr, &r.Confidence, &r.RiskLevel, &r.Status,
			&r.Timestamp, &r.PrevHash, &r.RecordHash, &r.HumanApprovalReq); err != nil {
			return nil, err
		}
		if evStr != "" {
			json.Unmarshal([]byte(evStr), &r.Evidence)
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// Verify validates the chain integrity by checking all hash links
func (c *Chain) Verify(ctx context.Context, taskID string) (bool, error) {
	records, err := c.GetTrail(ctx, taskID)
	if err != nil {
		return false, err
	}
	if len(records) == 0 {
		return true, nil
	}

	for i := 1; i < len(records); i++ {
		expected := records[i-1].RecordHash
		actual := records[i].PrevHash
		if expected != actual {
			log.Warn().
				Int("index", i).
				Str("expected", expected).
				Str("actual", actual).
				Msg("audit chain integrity broken")
			return false, nil
		}
	}
	return true, nil
}
