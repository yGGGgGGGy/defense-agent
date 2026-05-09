package dag

import (
	"fmt"
	"sort"
)

// Scheduler computes execution groups using Kahn's algorithm
type Scheduler struct {
	dag *DAG
}

// NewScheduler creates a scheduler for the given DAG
func NewScheduler(d *DAG) *Scheduler {
	return &Scheduler{dag: d}
}

// ExecutionGroup is a set of nodes that can run concurrently
type ExecutionGroup struct {
	Nodes []*Node `json:"nodes"`
	Depth int     `json:"depth"`
}

// Schedule returns ordered groups where nodes within a group are parallel-safe
// Returns an error if a cycle is detected
func (s *Scheduler) Schedule() ([]ExecutionGroup, error) {
	// Compute in-degree for each node
	inDegree := make(map[string]int)
	for id := range s.dag.Nodes {
		inDegree[id] = 0
	}
	for _, n := range s.dag.Nodes {
		for _, depID := range n.Dependencies {
			inDegree[n.ID]++
			_ = depID
		}
	}

	// Recompute accurate in-degree
	for id := range s.dag.Nodes {
		inDegree[id] = 0
	}
	for _, n := range s.dag.Nodes {
		for _, depID := range n.Dependencies {
			if _, ok := s.dag.Nodes[depID]; ok {
				inDegree[n.ID]++
			}
		}
	}

	totalNodes := len(s.dag.Nodes)
	processed := 0
	depth := 0
	groups := make([]ExecutionGroup, 0)

	for processed < totalNodes {
		// Find all nodes with in-degree 0
		ready := make([]string, 0)
		for id, deg := range inDegree {
			if deg == 0 {
				ready = append(ready, id)
			}
		}

		if len(ready) == 0 {
			return nil, fmt.Errorf("cycle detected in DAG: %d nodes remaining unprocessed", totalNodes-processed)
		}

		// Sort for deterministic order
		sort.Strings(ready)

		group := ExecutionGroup{Depth: depth}
		for _, id := range ready {
			inDegree[id] = -1 // mark as processed
			group.Nodes = append(group.Nodes, s.dag.Nodes[id])
			processed++

			// Decrease in-degree of all successors
			for _, n := range s.dag.Nodes {
				for _, depID := range n.Dependencies {
					if depID == id {
						inDegree[n.ID]--
					}
				}
			}
		}

		groups = append(groups, group)
		depth++
	}

	return groups, nil
}

// Validate checks the DAG for structural issues
func (s *Scheduler) Validate() error {
	nodeIDs := make(map[string]bool)
	for id := range s.dag.Nodes {
		nodeIDs[id] = true
	}

	// Check all dependency references exist
	for _, n := range s.dag.Nodes {
		for _, depID := range n.Dependencies {
			if !nodeIDs[depID] {
				return fmt.Errorf("node %s references unknown dependency %s", n.ID, depID)
			}
		}
	}

	// Check for cycles via attempt to schedule
	if _, err := s.Schedule(); err != nil {
		return err
	}

	return nil
}
