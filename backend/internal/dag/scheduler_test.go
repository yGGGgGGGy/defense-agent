package dag_test

import (
	"testing"

	"github.com/gjy20/defense-agent/backend/internal/dag"
	"github.com/gjy20/defense-agent/backend/internal/types"
)

func TestSchedulerSimple(t *testing.T) {
	d := dag.NewDAG("d1", "t1", types.SceneIncidentResponse)

	// a -> b -> c
	d.AddNode(&dag.Node{ID: "a", AgentType: types.AgentPerceiver, State: types.NodePending})
	d.AddNode(&dag.Node{ID: "b", AgentType: types.AgentAnalyst, State: types.NodePending, Dependencies: []string{"a"}})
	d.AddNode(&dag.Node{ID: "c", AgentType: types.AgentResponder, State: types.NodePending, Dependencies: []string{"b"}})

	s := dag.NewScheduler(d)
	groups, err := s.Schedule()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if len(groups[0].Nodes) != 1 || groups[0].Nodes[0].ID != "a" {
		t.Errorf("group 0 should be [a]")
	}
	if len(groups[1].Nodes) != 1 || groups[1].Nodes[0].ID != "b" {
		t.Errorf("group 1 should be [b]")
	}
	if len(groups[2].Nodes) != 1 || groups[2].Nodes[0].ID != "c" {
		t.Errorf("group 2 should be [c]")
	}
}

func TestSchedulerParallel(t *testing.T) {
	d := dag.NewDAG("d2", "t2", types.SceneIncidentResponse)

	// a -> (b, c) -> d
	d.AddNode(&dag.Node{ID: "a", AgentType: types.AgentPerceiver, State: types.NodePending})
	d.AddNode(&dag.Node{ID: "b", AgentType: types.AgentAnalyst, State: types.NodePending, Dependencies: []string{"a"}})
	d.AddNode(&dag.Node{ID: "c", AgentType: types.AgentOperator, State: types.NodePending, Dependencies: []string{"a"}})
	d.AddNode(&dag.Node{ID: "d", AgentType: types.AgentResponder, State: types.NodePending, Dependencies: []string{"b", "c"}})

	s := dag.NewScheduler(d)
	groups, err := s.Schedule()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	// Group 0: [a]
	// Group 1: [b, c] (parallel)
	// Group 2: [d]
	if len(groups[1].Nodes) != 2 {
		t.Errorf("group 1 should have 2 parallel nodes, got %d", len(groups[1].Nodes))
	}
}

func TestSchedulerCycleDetection(t *testing.T) {
	d := dag.NewDAG("d3", "t3", types.SceneIncidentResponse)

	d.AddNode(&dag.Node{ID: "a", AgentType: types.AgentPerceiver, State: types.NodePending, Dependencies: []string{"b"}})
	d.AddNode(&dag.Node{ID: "b", AgentType: types.AgentAnalyst, State: types.NodePending, Dependencies: []string{"a"}})

	s := dag.NewScheduler(d)
	_, err := s.Schedule()
	if err == nil {
		t.Fatal("expected cycle detection error")
	}
}

func TestDAGSkipDownstream(t *testing.T) {
	d := dag.NewDAG("d4", "t4", types.SceneIncidentResponse)

	d.AddNode(&dag.Node{ID: "a", AgentType: types.AgentPerceiver, State: types.NodePending})
	d.AddNode(&dag.Node{ID: "b", AgentType: types.AgentAnalyst, State: types.NodePending, Dependencies: []string{"a"}})
	d.AddNode(&dag.Node{ID: "c", AgentType: types.AgentResponder, State: types.NodePending, Dependencies: []string{"a"}})
	d.AddNode(&dag.Node{ID: "d", AgentType: types.AgentOperator, State: types.NodePending, Dependencies: []string{"b"}})

	d.Nodes["a"].State = types.NodeFailed
	d.SkipDownstream("a")

	if d.Nodes["b"].State != types.NodeSkipped {
		t.Errorf("node b should be skipped, got %s", d.Nodes["b"].State)
	}
	if d.Nodes["c"].State != types.NodeSkipped {
		t.Errorf("node c should be skipped, got %s", d.Nodes["c"].State)
	}
	if d.Nodes["d"].State != types.NodeSkipped {
		t.Errorf("node d should be skipped, got %s", d.Nodes["d"].State)
	}
}

func TestDAGReadyNodes(t *testing.T) {
	d := dag.NewDAG("d5", "t5", types.SceneIncidentResponse)

	d.AddNode(&dag.Node{ID: "a", AgentType: types.AgentPerceiver, State: types.NodePending})
	d.AddNode(&dag.Node{ID: "b", AgentType: types.AgentAnalyst, State: types.NodePending, Dependencies: []string{"a"}})

	ready := d.ReadyNodes()
	if len(ready) != 1 || ready[0].ID != "a" {
		t.Errorf("only 'a' should be ready, got %d nodes", len(ready))
	}

	d.Nodes["a"].State = types.NodeSucceeded
	ready = d.ReadyNodes()
	if len(ready) != 1 || ready[0].ID != "b" {
		t.Errorf("only 'b' should be ready after a succeeds, got %d nodes", len(ready))
	}
}
