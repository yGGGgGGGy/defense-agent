package agent

import (
	"context"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

// Input carries everything an agent needs to perform its job
type Input struct {
	TaskID       string            `json:"task_id"`
	SubTaskID    string            `json:"sub_task_id"`
	Scene        types.SceneType   `json:"scene"`
	Instructions string            `json:"instructions"`
	Context      map[string]string `json:"context"`
	Artifacts    []types.Evidence  `json:"artifacts,omitempty"`
}

// Output is what an agent produces after execution
type Output struct {
	Findings    map[string]string `json:"findings"`
	Actions     []types.Action    `json:"actions"`
	Confidence  float64           `json:"confidence"`
	Summary     string            `json:"summary"`
	Evidence    []types.Evidence  `json:"evidence,omitempty"`
}

// Agent is the interface all agents must implement
type Agent interface {
	// Type returns the agent's role identifier
	Type() types.AgentType

	// Execute runs the agent with the given input and returns output
	Execute(ctx context.Context, input *Input) (*Output, error)

	// Capabilities lists what this agent can do
	Capabilities() []string
}
