package agent

import (
	"context"
	"fmt"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

type reflector struct {
	failCount int
	maxFails  int
}

func NewReflector() Agent {
	return &reflector{maxFails: 3}
}

func (a *reflector) Type() types.AgentType { return types.AgentReflector }

func (a *reflector) Capabilities() []string {
	return []string{"failure_analysis", "guidance_generation", "error_recovery", "tool_suggestion"}
}

func (a *reflector) Execute(ctx context.Context, input *Input) (*Output, error) {
	errorMsg := input.Context["error"]
	a.failCount++

	var guidance string
	if a.failCount >= a.maxFails {
		guidance = fmt.Sprintf("CRITICAL: Agent failed %d times. Suggest using 'done' tool to gracefully exit or 'ask' tool to request human guidance.", a.failCount)
	} else {
		guidance = fmt.Sprintf("Reflector guidance (attempt %d/%d): %s\nConsider: 1) Check tool arguments 2) Try alternative approach 3) Verify input format.",
			a.failCount, a.maxFails, errorMsg)
	}

	return &Output{
		Findings: map[string]string{
			"guidance":     guidance,
			"fail_count":   fmt.Sprintf("%d", a.failCount),
			"should_abort": fmt.Sprintf("%v", a.failCount >= a.maxFails),
		},
		Confidence: 0.75,
		Summary:    guidance,
	}, nil
}
