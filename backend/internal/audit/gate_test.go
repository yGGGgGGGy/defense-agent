package audit_test

import (
	"context"
	"testing"

	"github.com/gjy20/defense-agent/backend/internal/audit"
	"github.com/gjy20/defense-agent/backend/internal/types"
)

func TestGateLowRiskAutoApproved(t *testing.T) {
	chain := &fakeChain{}
	gate := audit.NewGate(chain)

	r := audit.NewRecord("t1", types.AgentPerceiver, "scan ports", "routine scan", nil, types.RiskLow, 0.95)
	result, err := gate.Evaluate(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Approved || !result.AutoApproved {
		t.Errorf("low risk should be auto-approved")
	}
}

func TestGateMediumRiskAutoApproved(t *testing.T) {
	chain := &fakeChain{}
	gate := audit.NewGate(chain)

	r := audit.NewRecord("t2", types.AgentResponder, "block IP", "suspicious activity", nil, types.RiskMedium, 0.85)
	result, err := gate.Evaluate(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Approved {
		t.Errorf("medium risk should be auto-approved")
	}
}

func TestGateHighRiskRequiresHuman(t *testing.T) {
	chain := &fakeChain{}
	gate := audit.NewGate(chain)

	r := audit.NewRecord("t3", types.AgentResponder, "isolate host", "critical threat", nil, types.RiskHigh, 0.90)
	result, err := gate.Evaluate(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Approved {
		t.Errorf("high risk should NOT be auto-approved")
	}
	if !result.RequiresHuman {
		t.Errorf("high risk should require human approval")
	}
}

func TestGateDangerousBlocked(t *testing.T) {
	chain := &fakeChain{}
	gate := audit.NewGate(chain)

	r := audit.NewRecord("t4", types.AgentOperator, "DROP TABLE", "cleanup", nil, types.RiskDangerous, 0.50)
	result, err := gate.Evaluate(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Approved {
		t.Errorf("dangerous risk should be blocked")
	}
}

// fakeChain implements audit.ChainInterface for testing
type fakeChain struct {
	records []*audit.Record
}

func (f *fakeChain) Append(ctx context.Context, r *audit.Record) error {
	f.records = append(f.records, r)
	return nil
}
