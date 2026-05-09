package dag

import (
	"context"
	"fmt"
	"sync"

	"github.com/gjy20/defense-agent/backend/internal/types"
	"github.com/rs/zerolog/log"
)

// AgentRunner is the interface the executor uses to run agents
type AgentRunner interface {
	RunAgent(ctx context.Context, node *Node) (string, error)
}

// Executor runs a DAG to completion using a worker pool
type Executor struct {
	runner       AgentRunner
	maxParallel  int
}

// NewExecutor creates a DAG executor
func NewExecutor(runner AgentRunner, maxParallel int) *Executor {
	if maxParallel <= 0 {
		maxParallel = 10
	}
	return &Executor{
		runner:      runner,
		maxParallel: maxParallel,
	}
}

// ExecuteDAG runs the full DAG, respecting dependencies and parallelism
func (e *Executor) ExecuteDAG(ctx context.Context, d *DAG) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	scheduler := NewScheduler(d)

	if err := scheduler.Validate(); err != nil {
		return fmt.Errorf("dag validation failed: %w", err)
	}

	groups, err := scheduler.Schedule()
	if err != nil {
		return err
	}

	for gi, group := range groups {
		log.Info().
			Int("group", gi).
			Int("depth", group.Depth).
			Int("nodes", len(group.Nodes)).
			Str("dag_id", d.ID).
			Msg("executing dag group")

		// Run all nodes in this group concurrently
		e.executeGroup(ctx, d, group)

		// After group completes, handle failures
		for _, n := range group.Nodes {
			if n.State == types.NodeFailed {
				log.Warn().
					Str("node_id", n.ID).
					Str("error", n.Error).
					Msg("node failed, skipping downstream")
				d.SkipDownstream(n.ID)
			}
		}

		// Check if we should stop
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if d.IsComplete() {
			break
		}
	}

	return nil
}

func (e *Executor) executeGroup(ctx context.Context, d *DAG, group ExecutionGroup) {
	sem := make(chan struct{}, e.maxParallel)
	var wg sync.WaitGroup

	for _, n := range group.Nodes {
		wg.Add(1)
		go func(node *Node) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			node.State = types.NodeRunning
			output, err := e.runner.RunAgent(ctx, node)
			if err != nil {
				node.State = types.NodeFailed
				node.Error = err.Error()
				return
			}
			node.State = types.NodeSucceeded
			node.Output = output
		}(n)
	}

	wg.Wait()
}
