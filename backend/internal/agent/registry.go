package agent

import (
	"fmt"
	"sync"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

// Registry holds all available agents
type Registry struct {
	mu      sync.RWMutex
	agents  map[types.AgentType]Agent
}

// NewRegistry creates an agent registry
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[types.AgentType]Agent),
	}
}

// Register adds an agent to the registry
func (r *Registry) Register(a Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agentType := a.Type()
	if _, ok := r.agents[agentType]; ok {
		return fmt.Errorf("agent type %q already registered", agentType)
	}
	r.agents[agentType] = a
	return nil
}

// Get retrieves an agent by type
func (r *Registry) Get(agentType types.AgentType) (Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, ok := r.agents[agentType]
	if !ok {
		return nil, fmt.Errorf("agent type %q not found", agentType)
	}
	return a, nil
}

// List returns all registered agent types
func (r *Registry) List() []types.AgentType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	typesList := make([]types.AgentType, 0, len(r.agents))
	for t := range r.agents {
		typesList = append(typesList, t)
	}
	return typesList
}

// DefaultRegistry returns a registry with all 11 agents pre-registered
func DefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(NewPerceiver())
	r.Register(NewAnalyst())
	r.Register(NewResponder())
	r.Register(NewOperator())
	r.Register(NewResearcher())
	r.Register(NewDeveloper())
	r.Register(NewExecutor())
	r.Register(NewAdviser())
	r.Register(NewReflector())
	r.Register(NewAuditor())
	r.Register(NewMemorist())
	return r
}
