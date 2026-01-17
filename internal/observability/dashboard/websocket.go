package dashboard

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dashboard
	},
}

// MessageType represents the type of WebSocket message.
type MessageType string

const (
	MessageTypeStats        MessageType = "stats"
	MessageTypeWorkerAdded  MessageType = "worker_added"
	MessageTypeWorkerRemove MessageType = "worker_removed"
	MessageTypeTaskStarted  MessageType = "task_started"
	MessageTypeTaskComplete MessageType = "task_completed"
	MessageTypePing         MessageType = "ping"
	MessageTypePong         MessageType = "pong"
)

// Message represents a WebSocket message.
type Message struct {
	Type      MessageType `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// Client represents a WebSocket client connection.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// Hub manages WebSocket client connections.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	done       chan struct{}
	mu         sync.RWMutex
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
	}
}

// Run starts the hub's event loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Debug().Int("clients", len(h.clients)).Msg("WebSocket client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Debug().Int("clients", len(h.clients)).Msg("WebSocket client disconnected")

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client send buffer full, disconnect
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()

		case <-h.done:
			return
		}
	}
}

// Stop stops the hub.
func (h *Hub) Stop() {
	close(h.done)
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(msg *Message) {
	msg.Timestamp = time.Now().Unix()
	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal WebSocket message")
		return
	}

	select {
	case h.broadcast <- data:
	default:
		log.Warn().Msg("WebSocket broadcast channel full")
	}
}

// BroadcastStats sends stats update to all clients.
func (h *Hub) BroadcastStats(stats *Stats) {
	h.Broadcast(&Message{
		Type: MessageTypeStats,
		Data: stats,
	})
}

// BroadcastWorkerAdded notifies clients of a new worker.
func (h *Hub) BroadcastWorkerAdded(worker *WorkerInfo) {
	h.Broadcast(&Message{
		Type: MessageTypeWorkerAdded,
		Data: worker,
	})
}

// BroadcastWorkerRemoved notifies clients of a removed worker.
func (h *Hub) BroadcastWorkerRemoved(workerID string) {
	h.Broadcast(&Message{
		Type: MessageTypeWorkerRemove,
		Data: map[string]string{"worker_id": workerID},
	})
}

// BroadcastTaskStarted notifies clients of a started task.
func (h *Hub) BroadcastTaskStarted(task *TaskInfo) {
	h.Broadcast(&Message{
		Type: MessageTypeTaskStarted,
		Data: task,
	})
}

// BroadcastTaskCompleted notifies clients of a completed task.
func (h *Hub) BroadcastTaskCompleted(task *TaskInfo) {
	h.Broadcast(&Message{
		Type: MessageTypeTaskComplete,
		Data: task,
	})
}

// handleWebSocket handles WebSocket upgrade requests.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("WebSocket upgrade failed")
		return
	}

	client := &Client{
		hub:  s.hub,
		conn: conn,
		send: make(chan []byte, 256),
	}

	s.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// readPump reads messages from the WebSocket connection.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Msg("WebSocket read error")
			}
			break
		}

		// Handle ping messages
		var msg Message
		if json.Unmarshal(message, &msg) == nil && msg.Type == MessageTypePing {
			pong := &Message{Type: MessageTypePong, Timestamp: time.Now().Unix()}
			if data, err := json.Marshal(pong); err == nil {
				c.send <- data
			}
		}
	}
}

// writePump writes messages to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Batch queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
