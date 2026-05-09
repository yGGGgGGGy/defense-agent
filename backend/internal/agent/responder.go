package agent

import (
	"context"
	"fmt"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

type responder struct{}

func NewResponder() Agent { return &responder{} }

func (a *responder) Type() types.AgentType { return types.AgentResponder }

func (a *responder) Capabilities() []string {
	return []string{"ip_blocking", "host_isolation", "patch_deployment", "service_recovery"}
}

func (a *responder) Execute(ctx context.Context, input *Input) (*Output, error) {
	severity := input.Context["severity"]
	if severity == "" {
		severity = "HIGH"
	}

	sourceIP := input.Context["source_ip"]
	if sourceIP == "" {
		sourceIP = "0.0.0.0"
	}

	findings := make(map[string]string)
	actions := make([]types.Action, 0)

	// Block source IP action
	blockAction := types.Action{
		ID:        fmt.Sprintf("%s-block", input.SubTaskID),
		Name:      "Block Source IP",
		Command:   fmt.Sprintf("iptables -A INPUT -s %s -j DROP", sourceIP),
		Args:      []string{sourceIP},
		RiskLevel: types.RiskMedium,
		Rationale: fmt.Sprintf("Block malicious source IP %s to stop ongoing attack. Severity: %s", sourceIP, severity),
		Evidence: []types.Evidence{
			{Type: "analysis", Source: "analyst_output", Detail: fmt.Sprintf("Severity=%s, SourceIP=%s", severity, sourceIP)},
		},
		Sandbox: true,
		Timeout: 15,
		Status:  types.ActionPending,
	}
	actions = append(actions, blockAction)
	findings["blocked_ip"] = sourceIP

	// For critical severity, add host isolation
	if severity == "CRITICAL" {
		isolateAction := types.Action{
			ID:        fmt.Sprintf("%s-isolate", input.SubTaskID),
			Name:      "Isolate Affected Host",
			Command:   "iptables -A INPUT -p tcp --dport 22 -j DROP",
			RiskLevel: types.RiskHigh,
			Rationale: "CRITICAL severity: Isolate affected systems to prevent lateral movement. Temporarily restrict SSH access.",
			Evidence: []types.Evidence{
				{Type: "analysis", Source: "analyst_output", Detail: "Severity=CRITICAL, potential lateral movement risk"},
			},
			Sandbox: true,
			Timeout: 15,
			Status:  types.ActionPending,
		}
		actions = append(actions, isolateAction)
		findings["host_isolated"] = "true"
	}

	// Service recovery verification
	verifyAction := types.Action{
		ID:        fmt.Sprintf("%s-verify", input.SubTaskID),
		Name:      "Verify Service Integrity",
		Command:   "systemctl status sshd --no-pager",
		RiskLevel: types.RiskLow,
		Rationale: "Check that critical services remain operational after containment actions",
		Evidence: []types.Evidence{
			{Type: "observation", Source: "responder", Detail: "Post-containment service verification"},
		},
		Sandbox: true,
		Timeout: 10,
		Status:  types.ActionPending,
	}
	actions = append(actions, verifyAction)

	return &Output{
		Findings:   findings,
		Actions:    actions,
		Confidence: 0.90,
		Summary:    fmt.Sprintf("Responder executed %d containment actions: blocked IP %s, severity %s", len(actions), sourceIP, severity),
		Evidence: []types.Evidence{
			{Type: "action", Source: "responder", Detail: fmt.Sprintf("Blocked %s, %d actions taken", sourceIP, len(actions))},
		},
	}, nil
}
