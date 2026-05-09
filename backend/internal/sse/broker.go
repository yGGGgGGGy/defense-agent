package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Broker manages Server-Sent Events subscribers
type Broker struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan Event]bool
}

// Event is a server-sent event
type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// NewBroker creates an SSE broker
func NewBroker() *Broker {
	return &Broker{
		subscribers: make(map[string]map[chan Event]bool),
	}
}

// Subscribe adds a subscriber for events on a specific instance
func (b *Broker) Subscribe(instanceID string) chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan Event, 64)
	if b.subscribers[instanceID] == nil {
		b.subscribers[instanceID] = make(map[chan Event]bool)
	}
	b.subscribers[instanceID][ch] = true
	return ch
}

// Unsubscribe removes a subscriber
func (b *Broker) Unsubscribe(instanceID string, ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if subs, ok := b.subscribers[instanceID]; ok {
		delete(subs, ch)
		close(ch)
		if len(subs) == 0 {
			delete(b.subscribers, instanceID)
		}
	}
}

// Publish sends an event to all subscribers of an instance
func (b *Broker) Publish(instanceID string, event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	subs, ok := b.subscribers[instanceID]
	if !ok {
		return
	}
	for ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
}

// SSEHandler serves the SSE stream
func (b *Broker) SSEHandler(w http.ResponseWriter, r *http.Request) {
	instanceID := r.URL.Query().Get("instance_id")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var ch chan Event
	if instanceID != "" {
		ch = b.Subscribe(instanceID)
	} else {
		ch = b.Subscribe("*")
	}
	defer b.Unsubscribe(instanceID, ch)

	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
	flusher.Flush()

	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(event.Data)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, string(data))
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
