package agent

import (
	"context"
	"fmt"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

type developer struct{}

func NewDeveloper() Agent { return &developer{} }

func (a *developer) Type() types.AgentType { return types.AgentDeveloper }

func (a *developer) Capabilities() []string {
	return []string{"attack_planning", "exploit_selection", "tool_recommendation", "strategy_generation"}
}

func (a *developer) Execute(ctx context.Context, input *Input) (*Output, error) {
	findings := input.Context["research_findings"]
	if findings == "" {
		findings = input.Instructions
	}

	actions := []types.Action{
		{
			ID:        fmt.Sprintf("%s-plan", input.SubTaskID),
			Name:      "Generate Attack Plan",
			Command:   "generate response plan based on threat analysis",
			RiskLevel: types.RiskMedium,
			Rationale: "Create structured response/attack plan with prioritized steps based on threat severity",
			Evidence: []types.Evidence{
				{Type: "planning", Source: "developer", Detail: findings},
			},
			Timeout: 30,
		},
		{
			ID:        fmt.Sprintf("%s-tools", input.SubTaskID),
			Name:      "Select Tools",
			Command:   "recommend tools based on attack vector",
			RiskLevel: types.RiskLow,
			Rationale: "Map attack vectors to appropriate tools and techniques",
			Evidence: []types.Evidence{
				{Type: "planning", Source: "developer", Detail: "Tool selection matrix"},
			},
			Timeout: 15,
		},
	}

	return &Output{
		Findings: map[string]string{
			"plan_steps":   "3",
			"priority":     "high",
			"tools":        "nmap,hydra,metasploit",
		},
		Actions:    actions,
		Confidence: 0.82,
		Summary:    fmt.Sprintf("Developer: Generated response plan with 3 phases based on %s", findings),
	}, nil
}
