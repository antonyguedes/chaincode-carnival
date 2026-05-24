package dashboard

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/antonioforte/chaincode-carnival/agents/bus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub manages all active WebSocket connections and broadcasts events to them.
type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
	send    chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
		send:    make(chan []byte, 256),
	}
}

// ForwardEvents wires the hub to the event bus — every event goes to all browsers.
func (h *Hub) ForwardEvents(b *bus.EventBus) {
	allCh := make(chan bus.Event, 128)
	b.SubscribeAll(allCh)

	go func() {
		for evt := range allCh {
			data, err := json.Marshal(evt)
			if err == nil {
				select {
				case h.send <- data:
				default:
				}
			}
		}
	}()
}

// Run starts the broadcast loop — must be called in a goroutine.
func (h *Hub) Run() {
	for msg := range h.send {
		h.mu.RLock()
		for conn := range h.clients {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				conn.Close()
			}
		}
		h.mu.RUnlock()
	}
}

// ServeWS upgrades an HTTP connection to WebSocket and registers it with the hub.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	// Drain incoming messages (ping/pong) and clean up on disconnect
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.clients, conn)
			h.mu.Unlock()
			conn.Close()
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
