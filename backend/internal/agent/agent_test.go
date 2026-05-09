package agent_test

import (
	"context"
	"testing"

	"github.com/gjy20/defense-agent/backend/internal/agent"
	"github.com/gjy20/defense-agent/backend/internal/types"
)

func TestRegistry(t *testing.T) {
	r := agent.DefaultRegistry()
	agents := r.List()

	if len(agents) != 11 {
		t.Errorf("expected 11 agents, got %d", len(agents))
	}

	expectedTypes := map[types.AgentType]bool{
		types.AgentPerceiver: false,
		types.AgentAnalyst:   false,
		types.AgentResponder: false,
		types.AgentOperator:  false,
	}
	for _, at := range agents {
		expectedTypes[at] = true
	}
	for at, found := range expectedTypes {
		if !found {
			t.Errorf("missing agent type: %s", at)
		}
	}
}

func TestPerceiverExecute(t *testing.T) {
	a := agent.NewPerceiver()
	output, err := a.Execute(context.Background(), &agent.Input{
		TaskID:       "test-1",
		SubTaskID:    "sub-1",
		Scene:        types.SceneIncidentResponse,
		Instructions: "SSH brute force detected",
		Context:      map[string]string{"alerts": "SSH_BRUTE_FORCE from 10.0.0.50"},
	})
	if err != nil {
		t.Fatalf("perceiver execute failed: %v", err)
	}
	if output.Summary == "" {
		t.Error("expected summary output")
	}
	if len(output.Actions) == 0 {
		t.Error("expected at least one action")
	}
	if output.Actions[0].RiskLevel != types.RiskLow {
		t.Errorf("expected low risk for discovery, got %s", output.Actions[0].RiskLevel)
	}
}

func TestAnalystExecute(t *testing.T) {
	a := agent.NewAnalyst()
	output, err := a.Execute(context.Background(), &agent.Input{
		TaskID:       "test-2",
		SubTaskID:    "sub-2",
		Scene:        types.SceneIncidentResponse,
		Instructions: "SSH brute force attack detected",
		Context:      map[string]string{"situation": "SSH brute force", "alert_type": "SSH_BRUTE_FORCE"},
	})
	if err != nil {
		t.Fatalf("analyst execute failed: %v", err)
	}
	if output.Findings["severity"] != "CRITICAL" {
		t.Errorf("expected CRITICAL severity, got %s", output.Findings["severity"])
	}
	if output.Findings["attck_technique"] == "" {
		t.Error("expected ATT&CK technique mapping")
	}
}

func TestResponderExecute(t *testing.T) {
	a := agent.NewResponder()
	output, err := a.Execute(context.Background(), &agent.Input{
		TaskID:       "test-3",
		SubTaskID:    "sub-3",
		Scene:        types.SceneIncidentResponse,
		Instructions: "Block malicious IP",
		Context:      map[string]string{"severity": "CRITICAL", "source_ip": "10.0.0.50"},
	})
	if err != nil {
		t.Fatalf("responder execute failed: %v", err)
	}
	if len(output.Actions) < 2 {
		t.Errorf("expected at least 2 actions for CRITICAL, got %d", len(output.Actions))
	}
}

func TestOperatorExecute(t *testing.T) {
	a := agent.NewOperator()
	output, err := a.Execute(context.Background(), &agent.Input{
		TaskID:       "test-4",
		SubTaskID:    "sub-4",
		Scene:        types.SceneOpsMaintenance,
		Instructions: "Perform health check",
	})
	if err != nil {
		t.Fatalf("operator execute failed: %v", err)
	}
	if len(output.Actions) < 3 {
		t.Errorf("expected at least 3 maintenance actions, got %d", len(output.Actions))
	}
}
