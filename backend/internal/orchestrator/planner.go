package orchestrator

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/gjy20/defense-agent/backend/internal/agent"
	"github.com/gjy20/defense-agent/backend/internal/dag"
	"github.com/gjy20/defense-agent/backend/internal/types"
)

// Planner handles dynamic replanning when DAG nodes fail
type Planner struct {
	registry *agent.Registry
}

// NewPlanner creates a planner
func NewPlanner(reg *agent.Registry) *Planner {
	return &Planner{registry: reg}
}

// Replan attempts to recover from a failed node by inserting substitute nodes
func (p *Planner) Replan(ctx context.Context, d *dag.DAG, failedNodeID string) (*dag.Node, error) {
	node := d.Nodes[failedNodeID]
	if node == nil {
		return nil, fmt.Errorf("node %q not found", failedNodeID)
	}

	log.Info().
		Str("node_id", failedNodeID).
		Str("agent_type", string(node.AgentType)).
		Int("retry_count", node.RetryCount).
		Int("max_retries", node.MaxRetries).
		Msg("attempting replan")

	// Try retry first
	if node.RetryCount < node.MaxRetries {
		node.RetryCount++
		node.State = types.NodePending
		node.Error = ""
		log.Info().Str("node_id", failedNodeID).Int("retry", node.RetryCount).Msg("retrying node")
		return node, nil
	}

	// Try substitute agent
	substitute := p.findSubstitute(node.AgentType)
	if substitute != types.AgentType("") {
		log.Info().Str("original", string(node.AgentType)).Str("substitute", string(substitute)).Msg("substituting agent")
		node.AgentType = substitute
		node.RetryCount = 0
		node.State = types.NodePending
		node.Error = ""
		return node, nil
	}

	// Cannot recover - mark as failed
	log.Warn().Str("node_id", failedNodeID).Msg("cannot replan, node is unrecoverable")
	return nil, fmt.Errorf("unrecoverable node %q", failedNodeID)
}

func (p *Planner) findSubstitute(at types.AgentType) types.AgentType {
	substitutes := map[types.AgentType]types.AgentType{
		types.AgentPerceiver:  types.AgentResearcher,
		types.AgentAnalyst:    types.AgentAuditor,
		types.AgentResponder:  types.AgentExecutor,
		types.AgentOperator:   types.AgentExecutor,
		types.AgentResearcher: types.AgentPerceiver,
		types.AgentDeveloper:  types.AgentAdviser,
	}
	return substitutes[at]
}
