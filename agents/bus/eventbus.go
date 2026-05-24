package bus

import (
	"sync"
	"time"
)

// EventType identifies the kind of action that happened in the arena.
type EventType string

const (
	EvtArenaStart       EventType = "ARENA_START"
	EvtFindingsReady    EventType = "FINDINGS_READY"
	EvtExploitLaunched  EventType = "EXPLOIT_LAUNCHED"
	EvtExploitConfirmed EventType = "EXPLOIT_CONFIRMED"
	EvtPatchSubmitted   EventType = "PATCH_SUBMITTED"
	EvtRetestResult     EventType = "RETEST_RESULT"
	EvtVerdictRendered  EventType = "VERDICT_RENDERED"
	EvtBanter          EventType = "BANTER"
	EvtArenaOver        EventType = "ARENA_OVER"
)

// Event is the universal message passed between agents via the bus.
type Event struct {
	Type    EventType   `json:"type"`
	Agent   string      `json:"agent"`
	Payload interface{} `json:"payload"`
	Ts      int64       `json:"ts"`
}

func NewEvent(evtType EventType, agent string, payload interface{}) Event {
	return Event{
		Type:    evtType,
		Agent:   agent,
		Payload: payload,
		Ts:      time.Now().UnixMilli(),
	}
}

// EventBus is an in-memory pub/sub hub backed by Go channels.
// It allows agents to react to events without knowing about each other.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[EventType][]chan Event
	// wildcard receives every event regardless of type (used by the dashboard)
	wildcards []chan Event
}

func New() *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]chan Event),
	}
}

// Subscribe registers ch to receive all events of evtType.
func (b *EventBus) Subscribe(evtType EventType, ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[evtType] = append(b.subscribers[evtType], ch)
}

// SubscribeAll registers ch to receive every event, regardless of type.
// Used by the WebSocket dashboard to forward all events to browsers.
func (b *EventBus) SubscribeAll(ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.wildcards = append(b.wildcards, ch)
}

// Publish broadcasts evt to all matching subscribers and all wildcards.
// Non-blocking: if a subscriber channel is full the event is dropped for that consumer.
func (b *EventBus) Publish(evt Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subscribers[evt.Type] {
		select {
		case ch <- evt:
		default:
		}
	}
	for _, ch := range b.wildcards {
		select {
		case ch <- evt:
		default:
		}
	}
}
