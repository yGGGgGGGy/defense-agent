package comm

import (
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"

	"github.com/gjy20/defense-agent/backend/internal/types"
)

// Bus is a NATS-backed message bus for inter-agent communication
type Bus struct {
	nc          *nats.Conn
	js          nats.JetStreamContext
	subs        map[string]*nats.Subscription
	mu          sync.RWMutex
}

// NewBus creates a message bus connected to NATS
func NewBus(url string) (*Bus, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("nats jetstream: %w", err)
	}

	// Create the main task stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "AGENT_TASKS",
		Subjects: []string{"agent.>"},
		Storage:  nats.MemoryStorage,
	})
	if err != nil {
		log.Warn().Err(err).Msg("stream may already exist")
	}

	log.Info().Str("url", url).Msg("connected to NATS")

	return &Bus{
		nc:   nc,
		js:   js,
		subs: make(map[string]*nats.Subscription),
	}, nil
}

// PublishTask sends a task assignment to an agent
func (b *Bus) PublishTask(agentType types.AgentType, taskID string, payload []byte) error {
	subject := fmt.Sprintf("agent.%s.task", agentType)
	_, err := b.js.Publish(subject, payload)
	return err
}

// SubscribeTasks listens for tasks assigned to a specific agent
func (b *Bus) SubscribeTasks(agentType types.AgentType, handler func(taskID string, data []byte)) error {
	subject := fmt.Sprintf("agent.%s.task", agentType)

	sub, err := b.js.Subscribe(subject, func(msg *nats.Msg) {
		// Extract task ID from header or subject
		taskID := msg.Header.Get("X-Task-ID")
		handler(taskID, msg.Data)
	}, nats.DeliverAll())
	if err != nil {
		return fmt.Errorf("subscribe tasks: %w", err)
	}

	b.mu.Lock()
	b.subs[subject] = sub
	b.mu.Unlock()

	return nil
}

// PublishResult sends an agent's result back
func (b *Bus) PublishResult(agentType types.AgentType, taskID string, payload []byte) error {
	subject := fmt.Sprintf("agent.%s.result", agentType)
	msg := nats.NewMsg(subject)
	msg.Header.Set("X-Task-ID", taskID)
	msg.Data = payload
	_, err := b.js.PublishMsg(msg)
	return err
}

// Close shuts down the NATS connection
func (b *Bus) Close() {
	b.nc.Drain()
	b.nc.Close()
}
