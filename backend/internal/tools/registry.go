package tools

import (
	"fmt"
	"sync"
)

// Tool defines a callable tool
type Tool struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Parameters  []Param  `json:"parameters,omitempty"`
	Sandbox     bool     `json:"sandbox"`
	RiskLevel   string   `json:"risk_level"`
	Handler     Handler  `json:"-"`
}

// Param describes a tool parameter
type Param struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// Handler is the function that executes the tool
type Handler func(args map[string]string) (string, error)

// Registry holds all available tools
type Registry struct {
	mu    sync.RWMutex
	tools map[string]*Tool
}

// NewRegistry creates a tool registry
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]*Tool)}
}

// Register adds a tool
func (r *Registry) Register(t *Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[t.Name] = t
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (*Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool %q not found", name)
	}
	return t, nil
}

// List returns all registered tool names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for n := range r.tools {
		names = append(names, n)
	}
	return names
}

// ToOpenAI converts registered tools to OpenAI function calling format
func (r *Registry) ToOpenAI() []map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fns := make([]map[string]any, 0, len(r.tools))
	for _, t := range r.tools {
		props := make(map[string]any)
		required := make([]string, 0)
		for _, p := range t.Parameters {
			props[p.Name] = map[string]any{
				"type":        p.Type,
				"description": p.Description,
			}
			if p.Required {
				required = append(required, p.Name)
			}
		}
		fns = append(fns, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"parameters": map[string]any{
					"type":       "object",
					"properties": props,
					"required":   required,
				},
			},
		})
	}
	return fns
}

// Execute runs a tool by name with given arguments
func (r *Registry) Execute(name string, args map[string]string) (string, error) {
	t, err := r.Get(name)
	if err != nil {
		return "", err
	}
	return t.Handler(args)
}

// DefaultRegistry returns a registry with standard tools
func DefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(searchTool())
	r.Register(terminalTool())
	r.Register(healthCheckTool())
	r.Register(fileReadTool())
	return r
}
