package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/gjy20/defense-agent/backend/internal/agent"
	"github.com/gjy20/defense-agent/backend/internal/sse"
	"github.com/gjy20/defense-agent/backend/internal/audit"
	"github.com/gjy20/defense-agent/backend/internal/comm"
	"github.com/gjy20/defense-agent/backend/internal/dag"
	"github.com/gjy20/defense-agent/backend/internal/graphiti"
	"github.com/gjy20/defense-agent/backend/internal/memory"
	"github.com/gjy20/defense-agent/backend/internal/types"
)

// Instance represents a single running task
type Instance struct {
	ID        string              `json:"id"`
	Task      *types.Task         `json:"task"`
	DAG       *dag.DAG            `json:"dag"`
	Status    types.InstanceStatus `json:"status"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	Error     string              `json:"error,omitempty"`
	cancel    context.CancelFunc
}

// Orchestrator is the main coordination engine
type Orchestrator struct {
	agentRegistry   *agent.Registry
	sceneRouter     *SceneRouter
	auditPool       *audit.Pool
	memoryStore     *memory.Store
	messageBus      *comm.Bus
	planner         *Planner
	checkpointMgr   *CheckpointManager
	graphClient     *graphiti.Client
	sseBroker       *sse.Broker
	instances       map[string]*Instance
	mu              sync.RWMutex
	maxInstances    int
}

// New creates an Orchestrator
func New(
	agentReg *agent.Registry,
	auditPool *audit.Pool,
	memStore *memory.Store,
	msgBus *comm.Bus,
	graphClient *graphiti.Client,
	sseBroker *sse.Broker,
	maxInstances int,
) *Orchestrator {
	o := &Orchestrator{
		agentRegistry: agentReg,
		sceneRouter:   NewSceneRouter(),
		auditPool:     auditPool,
		memoryStore:   memStore,
		messageBus:    msgBus,
		planner:       NewPlanner(agentReg),
		checkpointMgr: NewCheckpointManager(5),
		graphClient:   graphClient,
		sseBroker:     sseBroker,
		instances:     make(map[string]*Instance),
		maxInstances:  maxInstances,
	}
	o.sceneRouter.RegisterAllScenes()
	return o
}

// SubmitTask creates a new instance and begins execution
func (o *Orchestrator) SubmitTask(ctx context.Context, task *types.Task) (*Instance, error) {
	o.mu.Lock()
	if len(o.instances) >= o.maxInstances {
		o.mu.Unlock()
		return nil, fmt.Errorf("max instances reached (%d)", o.maxInstances)
	}

	instanceID := "inst-" + uuid.New().String()[:8]
	task.ID = instanceID
	task.CreatedAt = time.Now().UTC()
	task.UpdatedAt = time.Now().UTC()

	// Route to scene template and build DAG
	template, ok := o.sceneRouter.Route(task.Scene)
	if !ok {
		o.mu.Unlock()
		return nil, fmt.Errorf("unknown scene: %s", task.Scene)
	}

	d := template.BuildDAG(instanceID)
	// Propagate task context to nodes
	for _, n := range d.Nodes {
		n.Input = task.Input
	}

	inst := &Instance{
		ID:        instanceID,
		Task:      task,
		DAG:       d,
		Status:    types.InstancePending,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	o.instances[instanceID] = inst
	o.mu.Unlock()

	// Launch execution in background
	instCtx, cancel := context.WithCancel(ctx)
	inst.cancel = cancel
	go o.executeInstance(instCtx, inst)

	log.Info().
		Str("instance_id", instanceID).
		Str("scene", string(task.Scene)).
		Str("title", task.Title).
		Msg("task submitted")

	return inst, nil
}

// executeInstance runs the full DAG execution and audit lifecycle
func (o *Orchestrator) executeInstance(ctx context.Context, inst *Instance) {
	defer func() {
		o.UpdateInstanceStatus(inst.ID, inst.Status)
	}()

	o.UpdateInstanceStatus(inst.ID, types.InstanceRunning)

	// Create the agent runner that ties agents to audit+memory+graph
	runner := &agentRunner{
		registry:    o.agentRegistry,
		auditPool:   o.auditPool,
		memoryStore: o.memoryStore,
		graphClient: o.graphClient,
		sseBroker:   o.sseBroker,
		instanceID:  inst.ID,
	}

	executor := dag.NewExecutor(runner, 5)

	// Save initial checkpoint
	o.checkpointMgr.Save(inst.ID, inst.DAG, types.InstanceRunning)

	if err := executor.ExecuteDAG(ctx, inst.DAG); err != nil {
		log.Error().Err(err).Str("instance_id", inst.ID).Msg("dag execution error, attempting replan")

		// Attempt recovery
		for _, n := range inst.DAG.Nodes {
			if n.State == types.NodeFailed {
				replanned, replanErr := o.planner.Replan(ctx, inst.DAG, n.ID)
				if replanErr == nil && replanned != nil {
					log.Info().Str("node_id", n.ID).Msg("node replanned, retrying")
					// Save checkpoint and retry
					o.checkpointMgr.Save(inst.ID, inst.DAG, types.InstanceRunning)
					if execErr := executor.ExecuteDAG(ctx, inst.DAG); execErr == nil {
						break
					}
				}
			}
		}
	}

	// Save final checkpoint
	o.checkpointMgr.Save(inst.ID, inst.DAG, inst.Status)

	if inst.DAG.HasFailures() {
		// Try rollback and retry once
		if rollbackErr := o.checkpointMgr.Rollback(inst.ID, inst.DAG); rollbackErr == nil {
			log.Info().Str("instance_id", inst.ID).Msg("rolled back, retrying execution")
			if err := executor.ExecuteDAG(ctx, inst.DAG); err == nil && !inst.DAG.HasFailures() {
				inst.Status = types.InstanceDone
				inst.UpdatedAt = time.Now().UTC()
				return
			}
		}
		inst.Status = types.InstanceFailed
		inst.Error = "one or more nodes failed after retry"
	} else {
		inst.Status = types.InstanceDone
	}

	inst.UpdatedAt = time.Now().UTC()
	log.Info().Str("instance_id", inst.ID).Str("status", string(inst.Status)).Msg("instance complete")
}

// UpdateInstanceStatus thread-safely updates instance status
func (o *Orchestrator) UpdateInstanceStatus(id string, status types.InstanceStatus) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if inst, ok := o.instances[id]; ok {
		inst.Status = status
		inst.UpdatedAt = time.Now().UTC()
	}
}

// GetInstance returns an instance by ID
func (o *Orchestrator) GetInstance(id string) (*Instance, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	inst, ok := o.instances[id]
	return inst, ok
}

// ListInstances returns all instances
func (o *Orchestrator) ListInstances() []*Instance {
	o.mu.RLock()
	defer o.mu.RUnlock()
	insts := make([]*Instance, 0, len(o.instances))
	for _, inst := range o.instances {
		insts = append(insts, inst)
	}
	return insts
}

// CancelInstance cancels a running instance
func (o *Orchestrator) CancelInstance(id string) error {
	o.mu.RLock()
	inst, ok := o.instances[id]
	o.mu.RUnlock()
	if !ok {
		return fmt.Errorf("instance %q not found", id)
	}
	if inst.cancel != nil {
		inst.cancel()
	}
	o.UpdateInstanceStatus(id, types.InstanceFailed)
	return nil
}

// GetAuditTrail returns the audit records for a task
func (o *Orchestrator) GetAuditTrail(ctx context.Context, taskID string) ([]audit.Record, error) {
	return o.auditPool.GetTrail(ctx, taskID)
}

// PendingApprovals returns actions awaiting human review
func (o *Orchestrator) PendingApprovals() []*audit.Record {
	return o.auditPool.PendingApprovals()
}

// ApproveAction approves a specific pending action
func (o *Orchestrator) ApproveAction(ctx context.Context, recordID string) error {
	return o.auditPool.ApproveAction(ctx, recordID)
}

// ClearDoneInstances removes completed instances older than the given duration
func (o *Orchestrator) ClearDoneInstances() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	cleared := 0
	for id, inst := range o.instances {
		if inst.Status == types.InstanceDone || inst.Status == types.InstanceFailed {
			delete(o.instances, id)
			cleared++
		}
	}
	return cleared
}

// Shutdown gracefully stops the orchestrator
func (o *Orchestrator) Shutdown() {
	for _, inst := range o.ListInstances() {
		if inst.Status == types.InstanceRunning && inst.cancel != nil {
			inst.cancel()
		}
	}
}

// --- agentRunner implements dag.AgentRunner ---

type agentRunner struct {
	registry    *agent.Registry
	auditPool   *audit.Pool
	memoryStore *memory.Store
	graphClient *graphiti.Client
	sseBroker   *sse.Broker
	instanceID  string
}

func (r *agentRunner) RunAgent(ctx context.Context, node *dag.Node) (string, error) {
	a, err := r.registry.Get(node.AgentType)
	if err != nil {
		return "", fmt.Errorf("get agent: %w", err)
	}

	input := &agent.Input{
		TaskID:       r.instanceID,
		SubTaskID:    node.ID,
		Instructions: node.Input,
		Context:      map[string]string{"task_id": r.instanceID},
	}

	output, err := a.Execute(ctx, input)
	if err != nil {
		return "", fmt.Errorf("agent execute: %w", err)
	}

	// Get per-instance gate for audit
	gate := r.auditPool.GetGate(r.instanceID)

	// Write all records through the security gate
	for _, action := range output.Actions {
		rec := audit.NewRecord(
			r.instanceID,
			node.AgentType,
			action.Name,
			action.Rationale,
			action.Evidence,
			action.RiskLevel,
			output.Confidence,
		)
		rec.SubTaskID = node.ID

		result, err := gate.Evaluate(ctx, rec)
		if err != nil {
			log.Error().Err(err).Str("action", rec.Decision).Msg("security gate evaluation failed")
			continue
		}
		if !result.Approved {
			log.Warn().
				Str("action", rec.Decision).
				Str("reason", result.Message).
				Msg("action blocked by security gate")
		}
	}

	// Save findings to memory
	for k, v := range output.Findings {
		r.memoryStore.Save(ctx, &memory.Entry{
			TaskID:    r.instanceID,
			AgentType: string(node.AgentType),
			Key:       k,
			Value:     v,
		})
	}

	// Record to knowledge graph
	r.graphClient.RecordAgentAction(ctx, r.instanceID, string(node.AgentType), node.ID, output.Summary)

	// Publish SSE event with full thinking process
	r.sseBroker.Publish(r.instanceID, sse.Event{
		Type: "node_state",
		Data: map[string]any{
			"node_id":    node.ID,
			"agent_type": node.AgentType,
			"state":      node.State,
			"summary":    output.Summary,
			"confidence": output.Confidence,
			"findings":   output.Findings,
			"instance_id": r.instanceID,
		},
	})

	return output.Summary, nil
}
