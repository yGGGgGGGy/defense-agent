package audit

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

// Record is a single decision entry in the tamper-evident audit chain
type Record struct {
	ID                 string           `json:"id"`
	TaskID             string           `json:"task_id"`
	SubTaskID          string           `json:"sub_task_id"`
	AgentType          types.AgentType  `json:"agent_type"`
	AgentID            string           `json:"agent_id"`
	Decision           string           `json:"decision"`
	Rationale          string           `json:"rationale"`
	Evidence           []types.Evidence `json:"evidence"`
	Confidence         float64          `json:"confidence"`
	RiskLevel          types.RiskLevel  `json:"risk_level"`
	Status             string           `json:"status"`
	Timestamp          time.Time        `json:"timestamp"`
	PrevHash           string           `json:"prev_hash"`
	RecordHash         string           `json:"record_hash"`
	HumanApprovalReq   bool             `json:"human_approval_required"`
	ApprovedBy         string           `json:"approved_by,omitempty"`
	ApprovedAt         *time.Time       `json:"approved_at,omitempty"`
}

// ComputeHash calculates SHA256 over all record fields except hashes
func (r *Record) ComputeHash() string {
	savedPrev := r.PrevHash
	r.PrevHash = ""
	r.RecordHash = ""

	data, _ := json.Marshal(r)
	h := sha256.Sum256(data)

	r.PrevHash = savedPrev
	return fmt.Sprintf("%x", h)
}

// NewRecord creates an audit record and computes its hash
func NewRecord(taskID string, agentType types.AgentType, decision, rationale string, evidence []types.Evidence, risk types.RiskLevel, confidence float64) *Record {
	r := &Record{
		ID:         fmt.Sprintf("audit-%s-%d", taskID, time.Now().UnixNano()),
		TaskID:     taskID,
		AgentType:  agentType,
		Decision:   decision,
		Rationale:  rationale,
		Evidence:   evidence,
		Confidence: confidence,
		RiskLevel:  risk,
		Status:     "pending",
		Timestamp:  time.Now().UTC(),
	}
	r.RecordHash = r.ComputeHash()
	return r
}
