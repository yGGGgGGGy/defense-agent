package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/gjy20/defense-agent/backend/internal/dag"
	"github.com/gjy20/defense-agent/backend/internal/types"
)

// Checkpoint represents a saved snapshot of task state
type Checkpoint struct {
	InstanceID string          `json:"instance_id"`
	DAGState   map[string]dagNodeSnap `json:"dag_state"`
	TaskStatus types.InstanceStatus `json:"task_status"`
	SavedAt    time.Time       `json:"saved_at"`
}

type dagNodeSnap struct {
	State      string `json:"state"`
	Output     string `json:"output"`
	RetryCount int    `json:"retry_count"`
	Error      string `json:"error,omitempty"`
}

// CheckpointManager handles state snapshots and rollback
type CheckpointManager struct {
	mu           sync.RWMutex
	checkpoints  map[string][]Checkpoint // instanceID -> checkpoints
	maxSnapshots int
}

// NewCheckpointManager creates a checkpoint manager
func NewCheckpointManager(maxSnapshots int) *CheckpointManager {
	return &CheckpointManager{
		checkpoints:  make(map[string][]Checkpoint),
		maxSnapshots: maxSnapshots,
	}
}

// Save creates a checkpoint of the current DAG state
func (cm *CheckpointManager) Save(instanceID string, d *dag.DAG, status types.InstanceStatus) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	snap := Checkpoint{
		InstanceID: instanceID,
		DAGState:   make(map[string]dagNodeSnap),
		TaskStatus: status,
		SavedAt:    time.Now().UTC(),
	}

	for id, node := range d.Nodes {
		snap.DAGState[id] = dagNodeSnap{
			State:      string(node.State),
			Output:     node.Output,
			RetryCount: node.RetryCount,
			Error:      node.Error,
		}
	}

	cm.checkpoints[instanceID] = append(cm.checkpoints[instanceID], snap)

	// Trim oldest if over limit
	if len(cm.checkpoints[instanceID]) > cm.maxSnapshots {
		cm.checkpoints[instanceID] = cm.checkpoints[instanceID][1:]
	}

	data, _ := json.Marshal(snap)
	log.Debug().Str("instance_id", instanceID).RawJSON("checkpoint", data).Msg("checkpoint saved")
	return nil
}

// Rollback restores DAG state from the most recent checkpoint
func (cm *CheckpointManager) Rollback(instanceID string, d *dag.DAG) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	snaps, ok := cm.checkpoints[instanceID]
	if !ok || len(snaps) == 0 {
		return fmt.Errorf("no checkpoints for instance %q", instanceID)
	}

	lastSnap := snaps[len(snaps)-1]
	for id, snap := range lastSnap.DAGState {
		node, ok := d.Nodes[id]
		if !ok {
			continue
		}
		node.State = types.NodeState(snap.State)
		node.Output = snap.Output
		node.RetryCount = snap.RetryCount
		node.Error = snap.Error
	}

	log.Info().Str("instance_id", instanceID).Msg("rolled back to last checkpoint")
	return nil
}

// LoadCheckpoint restores from the most recent checkpoint
func (cm *CheckpointManager) LoadCheckpoint(ctx context.Context, instanceID string) (*Checkpoint, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	snaps, ok := cm.checkpoints[instanceID]
	if !ok || len(snaps) == 0 {
		return nil, fmt.Errorf("no checkpoints for instance %q", instanceID)
	}
	snap := snaps[len(snaps)-1]
	return &snap, nil
}
