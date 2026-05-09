package dag

import "github.com/gjy20/defense-agent/backend/internal/types"

// Node is a single DAG node representing one sub-task
type Node struct {
	ID           string            `json:"id"`
	AgentType    types.AgentType   `json:"agent_type"`
	State        types.NodeState   `json:"state"`
	Dependencies []string          `json:"dependencies"`
	Input        string            `json:"input"`  // instructions/context for the agent
	Output       string            `json:"output"` // agent execution result
	Error        string            `json:"error,omitempty"`
	RetryCount   int               `json:"retry_count"`
	MaxRetries   int               `json:"max_retries"`
	Timeout      int               `json:"timeout"` // execution timeout in seconds
}

// DAG is a directed acyclic graph of nodes
type DAG struct {
	ID        string           `json:"id"`
	TaskID    string           `json:"task_id"`
	Scene     types.SceneType  `json:"scene"`
	Nodes     map[string]*Node `json:"nodes"`
	RootNodes []string         `json:"root_nodes"`
}

// NewDAG creates an empty DAG
func NewDAG(id, taskID string, scene types.SceneType) *DAG {
	return &DAG{
		ID:     id,
		TaskID: taskID,
		Scene:  scene,
		Nodes:  make(map[string]*Node),
	}
}

// AddNode inserts a node and updates root tracking
func (d *DAG) AddNode(n *Node) {
	d.Nodes[n.ID] = n
	d.recomputeRoots()
}

// recomputeRoots finds nodes with no incoming dependencies
func (d *DAG) recomputeRoots() {
	d.RootNodes = nil
	for id, n := range d.Nodes {
		if len(n.Dependencies) == 0 {
			d.RootNodes = append(d.RootNodes, id)
		}
	}
}

// ReadyNodes returns nodes whose dependencies are all satisfied
func (d *DAG) ReadyNodes() []*Node {
	ready := make([]*Node, 0)
	for _, n := range d.Nodes {
		if n.State != types.NodePending {
			continue
		}
		allReady := true
		for _, depID := range n.Dependencies {
			dep, ok := d.Nodes[depID]
			if !ok || dep.State != types.NodeSucceeded {
				allReady = false
				break
			}
		}
		if allReady {
			ready = append(ready, n)
		}
	}
	return ready
}

// IsComplete returns true when all nodes have reached a terminal state
func (d *DAG) IsComplete() bool {
	for _, n := range d.Nodes {
		switch n.State {
		case types.NodeSucceeded, types.NodeFailed, types.NodeSkipped:
			continue
		default:
			return false
		}
	}
	return true
}

// HasFailures returns true if any node failed
func (d *DAG) HasFailures() bool {
	for _, n := range d.Nodes {
		if n.State == types.NodeFailed {
			return true
		}
	}
	return false
}

// SkipDownstream marks all downstream nodes from failed nodes as skipped
func (d *DAG) SkipDownstream(failedNodeID string) {
	// Collect all nodes that transitively depend on the failed node
	affected := make(map[string]bool)
	var walk func(id string)
	walk = func(id string) {
		for _, n := range d.Nodes {
			for _, depID := range n.Dependencies {
				if depID == id {
					if !affected[n.ID] {
						affected[n.ID] = true
						walk(n.ID)
					}
				}
			}
		}
	}
	walk(failedNodeID)

	for _, n := range d.Nodes {
		if affected[n.ID] && n.State == types.NodePending {
			n.State = types.NodeSkipped
			n.Error = "upstream dependency failed"
		}
	}
}
