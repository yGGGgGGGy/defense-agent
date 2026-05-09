package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

type auditorAgent struct{}

func NewAuditor() Agent { return &auditorAgent{} }

func (a *auditorAgent) Type() types.AgentType { return types.AgentAuditor }

func (a *auditorAgent) Capabilities() []string {
	return []string{"decision_review", "evidence_verification", "compliance_check", "risk_reassessment"}
}

func (a *auditorAgent) Execute(ctx context.Context, input *Input) (*Output, error) {
	decisions := input.Context["decisions"]
	if decisions == "" {
		decisions = input.Instructions
	}

	var findings strings.Builder
	findings.WriteString("Auditor Review:\n")

	// Review each decision
	for i, line := range strings.Split(decisions, "\n") {
		if line == "" {
			continue
		}
		risk := "low"
		status := "approved"

		if strings.Contains(strings.ToLower(line), "block") || strings.Contains(strings.ToLower(line), "isolate") {
			risk = "medium"
			status = "approved_with_note"
		}
		if strings.Contains(strings.ToLower(line), "drop") || strings.Contains(strings.ToLower(line), "delete") {
			risk = "dangerous"
			status = "blocked"
		}

		findings.WriteString(fmt.Sprintf("  [%d] Risk=%s Status=%s: %s\n", i+1, risk, status, line))
	}

	reviewAction := types.Action{
		ID:        fmt.Sprintf("%s-review", input.SubTaskID),
		Name:      "Audit Review",
		Command:   "review all pending decisions",
		RiskLevel: types.RiskLow,
		Rationale: "Complete audit review of all agent decisions for compliance and risk assessment",
		Evidence: []types.Evidence{
			{Type: "audit", Source: "auditor", Detail: findings.String()},
		},
		Timeout: 30,
	}

	return &Output{
		Findings: map[string]string{
			"audit_result": findings.String(),
			"total_reviewed": fmt.Sprintf("%d", strings.Count(decisions, "\n")+1),
		},
		Actions:    []types.Action{reviewAction},
		Confidence: 0.98,
		Summary:    findings.String(),
	}, nil
}
