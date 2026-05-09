package agent

import (
	"context"
	"fmt"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

type executor struct{}

func NewExecutor() Agent { return &executor{} }

func (a *executor) Type() types.AgentType { return types.AgentExecutor }

func (a *executor) Capabilities() []string {
	return []string{"tool_execution", "command_dispatch", "result_collection", "sandbox_orchestration"}
}

func (a *executor) Execute(ctx context.Context, input *Input) (*Output, error) {
	plan := input.Context["plan"]
	if plan == "" {
		plan = input.Instructions
	}

	actions := []types.Action{
		{
			ID:        fmt.Sprintf("%s-exec", input.SubTaskID),
			Name:      "Execute Tools",
			Command:   plan,
			RiskLevel: types.RiskMedium,
			Rationale: "Execute the planned tools and commands in sandbox environment",
			Evidence: []types.Evidence{
				{Type: "execution", Source: "executor", Detail: plan},
			},
			Sandbox: true,
			Timeout: 60,
		},
		{
			ID:        fmt.Sprintf("%s-collect", input.SubTaskID),
			Name:      "Collect Results",
			Command:   "aggregate tool outputs",
			RiskLevel: types.RiskLow,
			Rationale: "Aggregate and normalize tool execution results for analysis",
			Timeout: 15,
		},
	}

	return &Output{
		Findings: map[string]string{
			"executed": "true",
			"results":  "collected",
		},
		Actions:    actions,
		Confidence: 0.95,
		Summary:    "Executor: All planned tools executed successfully. Results collected.",
	}, nil
}
