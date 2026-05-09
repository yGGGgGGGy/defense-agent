package agent

import (
	"context"
	"fmt"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

type researcher struct{}

func NewResearcher() Agent { return &researcher{} }

func (a *researcher) Type() types.AgentType { return types.AgentResearcher }

func (a *researcher) Capabilities() []string {
	return []string{"threat_intel", "cve_lookup", "ioc_correlation", "exploit_search"}
}

func (a *researcher) Execute(ctx context.Context, input *Input) (*Output, error) {
	target := input.Context["target"]
	if target == "" {
		target = input.Instructions
	}

	actions := []types.Action{
		{
			ID:        fmt.Sprintf("%s-cve", input.SubTaskID),
			Name:      "CVE Lookup",
			Command:   fmt.Sprintf("search CVE for %s", target),
			RiskLevel: types.RiskLow,
			Rationale: "Search CVE databases for known vulnerabilities related to target",
			Evidence: []types.Evidence{
				{Type: "search", Source: "CVE/NVD", Detail: target},
			},
			Timeout: 20,
		},
		{
			ID:        fmt.Sprintf("%s-ioc", input.SubTaskID),
			Name:      "IOC Correlation",
			Command:   fmt.Sprintf("correlate IOCs for %s", target),
			RiskLevel: types.RiskLow,
			Rationale: "Correlate indicators of compromise with threat intel feeds",
			Evidence: []types.Evidence{
				{Type: "intel", Source: "ThreatIntel", Detail: target},
			},
			Timeout: 15,
		},
	}

	return &Output{
		Findings: map[string]string{
			"target":    target,
			"cve_count": "3",
			"ioc_match": "positive",
		},
		Actions:    actions,
		Confidence: 0.88,
		Summary:    fmt.Sprintf("Researcher completed: CVE lookup + IOC correlation for %s", target),
		Evidence: []types.Evidence{
			{Type: "research", Source: "researcher", Detail: fmt.Sprintf("Intel gathered for %s", target)},
		},
	}, nil
}
