package orchestrator

import (
	"github.com/gjy20/defense-agent/backend/internal/dag"
	"github.com/gjy20/defense-agent/backend/internal/types"
)

// SceneTemplate defines the agent composition and DAG structure for a scenario
type SceneTemplate struct {
	Scene       types.SceneType
	Name        string
	Description string
	AgentOrder  []types.AgentType
	BuildDAG    func(taskID string) *dag.DAG
}

// SceneRouter maps scene types to their templates
type SceneRouter struct {
	scenes map[types.SceneType]*SceneTemplate
}

// NewSceneRouter creates a router with default templates
func NewSceneRouter() *SceneRouter {
	r := &SceneRouter{
		scenes: make(map[types.SceneType]*SceneTemplate),
	}

	// Incident Response template:
	// Perceiver → Analyst → (Responder ∥ Operator)
	irTemplate := &SceneTemplate{
		Scene:       types.SceneIncidentResponse,
		Name:        "Incident Response",
		Description: "Emergency response to security incidents with containment and recovery",
		AgentOrder:  []types.AgentType{types.AgentPerceiver, types.AgentAnalyst, types.AgentResponder, types.AgentOperator},
		BuildDAG: func(taskID string) *dag.DAG {
			d := dag.NewDAG("dag-"+taskID, taskID, types.SceneIncidentResponse)

			pNode := &dag.Node{
				ID:         "perceiver-" + taskID,
				AgentType:  types.AgentPerceiver,
				State:      types.NodePending,
				MaxRetries: 2,
				Timeout:    60,
			}
			aNode := &dag.Node{
				ID:           "analyst-" + taskID,
				AgentType:    types.AgentAnalyst,
				State:        types.NodePending,
				Dependencies: []string{"perceiver-" + taskID},
				MaxRetries:   2,
				Timeout:      60,
			}
			rNode := &dag.Node{
				ID:           "responder-" + taskID,
				AgentType:    types.AgentResponder,
				State:        types.NodePending,
				Dependencies: []string{"analyst-" + taskID},
				MaxRetries:   3,
				Timeout:      90,
			}
			oNode := &dag.Node{
				ID:           "operator-" + taskID,
				AgentType:    types.AgentOperator,
				State:        types.NodePending,
				Dependencies: []string{"analyst-" + taskID},
				MaxRetries:   2,
				Timeout:      60,
			}

			d.AddNode(pNode)
			d.AddNode(aNode)
			d.AddNode(rNode)
			d.AddNode(oNode)
			return d
		},
	}
	r.scenes[types.SceneIncidentResponse] = irTemplate

	// Ops Maintenance template:
	opsTemplate := &SceneTemplate{
		Scene:       types.SceneOpsMaintenance,
		Name:        "Operations Maintenance",
		Description: "Routine system health checks and configuration compliance",
		AgentOrder:  []types.AgentType{types.AgentOperator},
		BuildDAG: func(taskID string) *dag.DAG {
			d := dag.NewDAG("dag-"+taskID, taskID, types.SceneOpsMaintenance)
			oNode := &dag.Node{
				ID:         "operator-" + taskID,
				AgentType:  types.AgentOperator,
				State:      types.NodePending,
				MaxRetries: 1,
				Timeout:    120,
			}
			d.AddNode(oNode)
			return d
		},
	}
	r.scenes[types.SceneOpsMaintenance] = opsTemplate

	return r
}

// Route returns the template for a given scene
func (r *SceneRouter) Route(scene types.SceneType) (*SceneTemplate, bool) {
	t, ok := r.scenes[scene]
	return t, ok
}
