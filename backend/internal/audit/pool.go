package audit

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool manages per-instance audit chains
type Pool struct {
	dbPool *pgxpool.Pool
	chains map[string]*Chain // instanceID -> chain
	gates  map[string]*Gate  // instanceID -> gate
	mu     sync.RWMutex
}

// NewPool creates an audit pool
func NewPool(dbPool *pgxpool.Pool) *Pool {
	return &Pool{
		dbPool: dbPool,
		chains: make(map[string]*Chain),
		gates:  make(map[string]*Gate),
	}
}

// getOrCreateChain returns the chain for an instance (caller must hold lock)
func (p *Pool) getOrCreateChain(instanceID string) *Chain {
	c, ok := p.chains[instanceID]
	if !ok {
		c = NewChain(p.dbPool)
		p.chains[instanceID] = c
	}
	return c
}

// GetChain returns the chain for an instance, creating one if needed
func (p *Pool) GetChain(instanceID string) *Chain {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.getOrCreateChain(instanceID)
}

// GetGate returns the security gate for an instance
func (p *Pool) GetGate(instanceID string) *Gate {
	p.mu.Lock()
	defer p.mu.Unlock()
	g, ok := p.gates[instanceID]
	if !ok {
		g = NewGate(p.getOrCreateChain(instanceID))
		p.gates[instanceID] = g
	}
	return g
}

// GetTrail retrieves the audit trail for an instance
func (p *Pool) GetTrail(ctx context.Context, instanceID string) ([]Record, error) {
	chain := p.GetChain(instanceID)
	return chain.GetTrail(ctx, instanceID)
}

// PendingApprovals returns all pending actions across all instances
func (p *Pool) PendingApprovals() []*Record {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var all []*Record
	for _, g := range p.gates {
		all = append(all, g.Pending()...)
	}
	return all
}

// ApproveAction approves a pending action in any instance
func (p *Pool) ApproveAction(ctx context.Context, recordID string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, g := range p.gates {
		if err := g.ApproveHuman(ctx, recordID); err == nil {
			return nil
		}
	}
	return nil
}
