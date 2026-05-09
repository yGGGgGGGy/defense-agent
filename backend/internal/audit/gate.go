package audit

import (
	"context"
	"fmt"

	"github.com/gjy20/defense-agent/backend/internal/types"
	"github.com/rs/zerolog/log"
)

// GateResult holds the outcome of a security gate evaluation
type GateResult struct {
	Approved        bool   `json:"approved"`
	AutoApproved    bool   `json:"auto_approved"`
	RequiresHuman   bool   `json:"requires_human"`
	Message         string `json:"message,omitempty"`
}

// ChainAppender is the interface for audit chain operations
type ChainAppender interface {
	Append(ctx context.Context, r *Record) error
}

// Gate is the 4-tier security approval system
type Gate struct {
	chain     ChainAppender
	// PendingApprovals holds actions awaiting human review
	PendingApprovals map[string]*Record
}

// NewGate creates a security gate
func NewGate(chain ChainAppender) *Gate {
	return &Gate{
		chain:            chain,
		PendingApprovals: make(map[string]*Record),
	}
}

// Evaluate determines whether an action can proceed based on risk level
func (g *Gate) Evaluate(ctx context.Context, r *Record) (*GateResult, error) {
	switch r.RiskLevel {
	case types.RiskLow:
		// Low: always auto-approved
		r.Status = "approved"
		if err := g.chain.Append(ctx, r); err != nil {
			return nil, err
		}
		log.Info().Str("decision", r.Decision).Str("risk", "low").Msg("auto-approved (low risk)")
		return &GateResult{Approved: true, AutoApproved: true}, nil

	case types.RiskMedium:
		// Medium: auto-review by Auditor (auto-approve after recording)
		r.Status = "approved"
		if err := g.chain.Append(ctx, r); err != nil {
			return nil, err
		}
		log.Info().Str("decision", r.Decision).Str("risk", "medium").Msg("auto-approved (medium risk, auditor reviewed)")
		return &GateResult{Approved: true, AutoApproved: true}, nil

	case types.RiskHigh:
		// High: requires human approval, queues for review
		r.Status = "pending"
		r.HumanApprovalReq = true
		g.PendingApprovals[r.ID] = r
		if err := g.chain.Append(ctx, r); err != nil {
			return nil, err
		}
		log.Warn().Str("decision", r.Decision).Str("id", r.ID).Msg("requires human approval (high risk)")
		return &GateResult{
			Approved:      false,
			RequiresHuman: true,
			Message:       fmt.Sprintf("High-risk action requires human approval. Action ID: %s", r.ID),
		}, nil

	case types.RiskDangerous:
		// Dangerous: always blocked, mandatory human approval
		r.Status = "blocked"
		r.HumanApprovalReq = true
		g.PendingApprovals[r.ID] = r
		if err := g.chain.Append(ctx, r); err != nil {
			return nil, err
		}
		log.Warn().Str("decision", r.Decision).Str("id", r.ID).Msg("BLOCKED (dangerous risk, mandatory human review)")
		return &GateResult{
			Approved:      false,
			RequiresHuman: true,
			Message:       fmt.Sprintf("DANGEROUS operation blocked. Mandatory human approval required. Action ID: %s", r.ID),
		}, nil

	default:
		return nil, fmt.Errorf("unknown risk level: %s", r.RiskLevel)
	}
}

// ApproveHuman approves a previously queued action
func (g *Gate) ApproveHuman(ctx context.Context, recordID string) error {
	r, ok := g.PendingApprovals[recordID]
	if !ok {
		return fmt.Errorf("approval record %q not found in pending queue", recordID)
	}

	r.Status = "approved"
	delete(g.PendingApprovals, recordID)

	log.Info().Str("id", recordID).Msg("human approval granted")
	return g.chain.Append(ctx, r)
}

// Pending returns all actions awaiting human approval
func (g *Gate) Pending() []*Record {
	records := make([]*Record, 0, len(g.PendingApprovals))
	for _, r := range g.PendingApprovals {
		records = append(records, r)
	}
	return records
}
