package agent

import (
	"context"
	"fmt"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

type perceiver struct{}

func NewPerceiver() Agent { return &perceiver{} }

func (a *perceiver) Type() types.AgentType { return types.AgentPerceiver }

func (a *perceiver) Capabilities() []string {
	return []string{"log_collection", "traffic_analysis", "asset_discovery", "alert_aggregation"}
}

func (a *perceiver) Execute(ctx context.Context, input *Input) (*Output, error) {
	// Parse alert data from context
	alertSummary := input.Context["alerts"]
	if alertSummary == "" {
		alertSummary = input.Instructions
	}

	findings := map[string]string{
		"situation": fmt.Sprintf("Security event detected: %s", alertSummary),
		"scope":     "Initial assessment of affected systems",
		"artifacts": alertSummary,
	}

	// Generate a discovery action
	discoverAction := types.Action{
		ID:        fmt.Sprintf("%s-discovery", input.SubTaskID),
		Name:      "Initial Asset Discovery",
		Command:   "nmap -sn --script http-headers",
		Args:      nil,
		RiskLevel: types.RiskLow,
		Rationale: "Map the network perimeter to identify exposed assets related to the alert",
		Evidence: []types.Evidence{
			{Type: "alert", Source: "SIEM", Detail: alertSummary},
		},
		Sandbox: true,
		Timeout: 30,
		Status:  types.ActionPending,
	}

	logCollectAction := types.Action{
		ID:        fmt.Sprintf("%s-logs", input.SubTaskID),
		Name:      "Collect Relevant Logs",
		Command:   "journalctl --since '10 minutes ago' -u ssh -u auth",
		Args:      nil,
		RiskLevel: types.RiskLow,
		Rationale: "Gather authentication and service logs for the alert timeframe",
		Evidence: []types.Evidence{
			{Type: "alert", Source: "SIEM", Detail: alertSummary},
		},
		Sandbox: true,
		Timeout: 30,
		Status:  types.ActionPending,
	}

	return &Output{
		Findings:   findings,
		Actions:    []types.Action{discoverAction, logCollectAction},
		Confidence: 0.95,
		Summary:    fmt.Sprintf("Perceiver detected security event: %s. Initiated asset discovery and log collection.", alertSummary),
		Evidence: []types.Evidence{
			{Type: "observation", Source: "perceiver", Detail: alertSummary},
		},
	}, nil
}
