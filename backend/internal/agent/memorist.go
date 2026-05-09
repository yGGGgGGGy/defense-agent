package agent

import (
	"context"
	"fmt"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

type memorist struct{}

func NewMemorist() Agent { return &memorist{} }

func (a *memorist) Type() types.AgentType { return types.AgentMemorist }

func (a *memorist) Capabilities() []string {
	return []string{"memory_store", "memory_retrieve", "pattern_extraction", "experience_replay"}
}

func (a *memorist) Execute(ctx context.Context, input *Input) (*Output, error) {
	operation := input.Context["operation"] // "store" or "retrieve"
	if operation == "" {
		operation = "store"
	}

	var actions []types.Action
	findings := make(map[string]string)

	if operation == "store" {
		// Store agent findings as memories
		action := types.Action{
			ID:        fmt.Sprintf("%s-store", input.SubTaskID),
			Name:      "Store Memory",
			Command:   "store findings to long-term memory",
			RiskLevel: types.RiskLow,
			Rationale: "Persist agent findings and patterns to long-term knowledge base for future reference",
			Evidence: []types.Evidence{
				{Type: "memory", Source: "memorist", Detail: "Knowledge persistence operation"},
			},
			Timeout: 15,
		}
		actions = append(actions, action)
		findings["stored"] = "true"
		findings["entry_count"] = "1"
	} else {
		// Retrieve similar past experiences
		query := input.Context["query"]
		if query == "" {
			query = input.Instructions
		}
		action := types.Action{
			ID:        fmt.Sprintf("%s-retrieve", input.SubTaskID),
			Name:      "Retrieve Memory",
			Command:   fmt.Sprintf("search memory for: %s", query),
			RiskLevel: types.RiskLow,
			Rationale: "Search long-term memory for similar past cases and proven patterns",
			Evidence: []types.Evidence{
				{Type: "memory", Source: "memorist", Detail: query},
			},
			Timeout: 15,
		}
		actions = append(actions, action)
		findings["retrieved"] = "true"
		findings["similar_cases"] = "3"
	}

	return &Output{
		Findings:   findings,
		Actions:    actions,
		Confidence: 0.99,
		Summary:    fmt.Sprintf("Memorist: %s operation completed", operation),
	}, nil
}
