package agent

import (
	"context"
	"fmt"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

type adviser struct {
	sameToolCount  map[string]int
	sameToolLimit  int
	totalToolLimit int
	totalToolCalls int
}

func NewAdviser() Agent {
	return &adviser{
		sameToolCount: make(map[string]int),
		sameToolLimit: 5,
		totalToolLimit: 10,
	}
}

func (a *adviser) Type() types.AgentType { return types.AgentAdviser }

func (a *adviser) Capabilities() []string {
	return []string{"execution_monitoring", "loop_detection", "progress_evaluation", "mentor_guidance"}
}

func (a *adviser) Execute(ctx context.Context, input *Input) (*Output, error) {
	lastTool := input.Context["last_tool_call"]
	totalCalls := input.Context["total_calls"]

	var warnings []string

	// Monitor for repeated tool calls (loop detection)
	if lastTool != "" {
		a.sameToolCount[lastTool]++
		if a.sameToolCount[lastTool] >= a.sameToolLimit {
			warnings = append(warnings, fmt.Sprintf("WARNING: Tool '%s' called %d times consecutively. Possible loop detected. Recommend trying alternative approach.", lastTool, a.sameToolCount[lastTool]))
		}
	}

	// Monitor total tool calls
	a.totalToolCalls++
	if a.totalToolCalls >= a.totalToolLimit {
		warnings = append(warnings, fmt.Sprintf("WARNING: Total tool calls (%d) approaching limit (%d). Consider wrapping up execution.", a.totalToolCalls, a.totalToolLimit))
	}

	// Reset counter when tool changes
	if lastTool != "" {
		for k := range a.sameToolCount {
			if k != lastTool {
				a.sameToolCount[k] = 0
			}
		}
	}

	analysis := "Agent progressing normally toward objectives."
	if len(warnings) > 0 {
		analysis = fmt.Sprintf("Adviser analysis:\n- %s", warnings[0])
		if len(warnings) > 1 {
			analysis += fmt.Sprintf("\n- %s", warnings[1])
		}
	}

	return &Output{
		Findings: map[string]string{
			"analysis":      analysis,
			"total_calls":   totalCalls,
			"loop_detected": fmt.Sprintf("%v", len(warnings) > 0),
		},
		Confidence: 0.90,
		Summary:    analysis,
	}, nil
}
