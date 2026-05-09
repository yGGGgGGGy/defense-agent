package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

type analyst struct{}

func NewAnalyst() Agent { return &analyst{} }

func (a *analyst) Type() types.AgentType { return types.AgentAnalyst }

func (a *analyst) Capabilities() []string {
	return []string{"alert_correlation", "attack_chain_mapping", "root_cause_analysis", "impact_assessment", "attck_mapping"}
}

var attckMap = map[string]string{
	"SSH_BRUTE_FORCE":     "T1110 (Brute Force)",
	"SSH":                 "T1021.004 (Remote Services: SSH)",
	"PORT_SCAN":           "T1046 (Network Service Discovery)",
	"PRIVILEGE_ESCALATION": "T1068 (Exploitation for Privilege Escalation)",
	"DATA_EXFIL":          "T1048 (Exfiltration Over Alternative Protocol)",
}

func (a *analyst) Execute(ctx context.Context, input *Input) (*Output, error) {
	situation := input.Context["situation"]
	if situation == "" {
		situation = input.Instructions
	}

	// Extract alert types from context
	alertType := input.Context["alert_type"]
	if alertType == "" {
		alertType = "SSH_BRUTE_FORCE"
	}

	technique := attckMap[alertType]
	if technique == "" {
		technique = fmt.Sprintf("Unknown technique related to: %s", alertType)
	}

	// Determine severity
	severity := "HIGH"
	confidence := 0.85
	if strings.Contains(strings.ToLower(situation), "ssh") && strings.Contains(strings.ToLower(situation), "brute") {
		severity = "CRITICAL"
		confidence = 0.92
	}

	findings := map[string]string{
		"alert_type":     alertType,
		"attck_technique": technique,
		"severity":       severity,
		"impact":         "Potential unauthorized access to critical systems",
		"correlation":    "Multiple failed authentication attempts from single source IP indicate targeted attack",
	}

	// Analysis actions
	analyzeAction := types.Action{
		ID:        fmt.Sprintf("%s-analyze", input.SubTaskID),
		Name:      "Correlate Alerts and Map to ATT&CK",
		Command:   "attck-analyze --technique " + technique,
		RiskLevel: types.RiskLow,
		Rationale: fmt.Sprintf("Map observed activity to ATT&CK framework for structured threat assessment. Identified technique: %s", technique),
		Evidence: []types.Evidence{
			{Type: "alert", Source: "perceiver_output", Detail: situation},
			{Type: "intel", Source: "ATT&CK_KB", Detail: technique},
		},
		Sandbox: false,
		Timeout: 20,
		Status:  types.ActionPending,
	}

	recommendAction := types.Action{
		ID:        fmt.Sprintf("%s-recommend", input.SubTaskID),
		Name:      "Generate Response Recommendations",
		Command:   "recommend --severity " + severity,
		RiskLevel: types.RiskMedium,
		Rationale: fmt.Sprintf("Based on severity %s and technique %s, recommend immediate containment actions", severity, technique),
		Evidence: []types.Evidence{
			{Type: "analysis", Source: "analyst", Detail: fmt.Sprintf("Severity=%s, Technique=%s", severity, technique)},
		},
		Sandbox: false,
		Timeout: 15,
		Status:  types.ActionPending,
	}

	return &Output{
		Findings:   findings,
		Actions:    []types.Action{analyzeAction, recommendAction},
		Confidence: confidence,
		Summary:    fmt.Sprintf("Analyst assessment: %s | ATT&CK: %s | Severity: %s | Confidence: %.0f%%", situation, technique, severity, confidence*100),
		Evidence: []types.Evidence{
			{Type: "analysis", Source: "analyst", Detail: fmt.Sprintf("ATT&CK %s, severity %s", technique, severity)},
		},
	}, nil
}
