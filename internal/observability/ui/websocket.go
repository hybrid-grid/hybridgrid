package ui

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// MessageType represents the type of a WebSocket frame. These are the
// only types the embedded dashboard ever broadcast; worker
// add/remove were dead-code paths (the coordinator never emitted
// them) and the SPA polls /api/v1/workers instead, so they are not
// relayed by the standalone server.
type MessageType string

const (
	MessageTypeStats        MessageType = "stats"
	MessageTypeTaskStarted  MessageType = "task_started"
	MessageTypeTaskComplete MessageType = "task_completed"
	MessageTypePing         MessageType = "ping"
	MessageTypePong         MessageType = "pong"
)

// Message is the WebSocket frame shape. Data serializes as parsed
// payload_json from the upstream TelemetryEvent, so the already-decoded
// object reaches the SPA instead of a string.
type Message struct {
	Type      MessageType `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// wsClient is one connected WebSocket session.
type wsClient struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// Hub fans broadcasts out to every connected WebSocket client, holds a
// bounded ring of recent task events to replay to late joiners, and
// exposes the same surface the embedded dashboard's hub did so the REST
// /events handler can return the translated frames.
type Hub struct {
	clients    map[*wsClient]bool
	broadcast  chan []byte
	register   chan *wsClient
	unregister chan *wsClient
	done       chan struct{}

	mu sync.RWMutex

	recentMu     sync.RWMutex
	maxEvents    int
	recentEvents [][]byte
}

// NewHub constructs a hub with the same 100-event replay bound the
// embedded dashboard always used.
func NewHub() *Hub {
	return &Hub{
		clients:      make(map[*wsClient]bool),
		broadcast:    make(chan []byte, 256),
		register:     make(chan *wsClient),
		unregister:   make(chan *wsClient),
		done:         make(chan struct{}),
		maxEvents:    100,
		recentEvents: make([][]byte, 0, 100),
	}
}

// Run is the hub's event loop. It must run until Stop. New clients
// receive the replay-on-connect ring immediately on registration.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Debug().Int("clients", h.clientCountLocked()).Msg("WebSocket client connected")

			h.recentMu.RLock()
			for _, event := range h.recentEvents {
				select {
				case client.send <- event:
				default:
					// Client buffer full; skip this frame.
				}
			}
			h.recentMu.RUnlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Debug().Int("clients", h.clientCountLocked()).Msg("WebSocket client disconnected")

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Slow consumer: drop it so the producer side
					// does not block on a dead client.
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()

		case <-h.done:
			return
		}
	}
}

func (h *Hub) clientCountLocked() int {
	return len(h.clients)
}

// Stop terminates the event loop. Safe to call once.
func (h *Hub) Stop() {
	close(h.done)
}

// ClientCount reports the number of currently connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetRecentEvents returns the translated recent-event frames as parsed
// JSON, the same slice the embedded dashboard's /events reader
// returned so the SPA activity pane keeps working unchanged.
func (h *Hub) GetRecentEvents() []json.RawMessage {
	h.recentMu.RLock()
	defer h.recentMu.RUnlock()
	out := make([]json.RawMessage, len(h.recentEvents))
	for i, data := range h.recentEvents {
		out[i] = json.RawMessage(data)
	}
	return out
}

// TranslateAndBroadcast converts one TelemetryEvent into the
// {"type","timestamp","data"} frame the SPA expects and fans it out.
// Data is the parsed payload_json so a JSON-decoding browser client
// receives the rich object directly. Task frames are recorded in the
// bounded replay ring so late joiners catch up; stats frames are not,
// matching the embedded dashboard's "store task events only" policy.
func (h *Hub) TranslateAndBroadcast(ev *pb.TelemetryEvent) {
	if ev == nil {
		return
	}
	evType := newEventType(ev.GetType())
	if evType == "" {
		// Unknown types: the contract pins the set to stats/
		// task_started/task_completed/ping. Drop anything else
		// rather than invent a new node.
		return
	}

	var data interface{}
	if payload := ev.GetPayloadJson(); payload != "" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(payload), &parsed); err == nil {
			data = parsed
		}
		// On decode failure data stays nil; encode drops the field
		// via the omitempty tag, same as an absent payload.
	}

	frame, err := json.Marshal(&Message{
		Type:      evType,
		Timestamp: ev.GetTimestamp(),
		Data:      data,
	})
	if err != nil {
		log.Error().Err(err).Msg("ui: failed to marshal WebSocket frame")
		return
	}

	if evType == MessageTypeTaskStarted || evType == MessageTypeTaskComplete {
		h.recentMu.Lock()
		h.recentEvents = append(h.recentEvents, frame)
		if len(h.recentEvents) > h.maxEvents {
			h.recentEvents = h.recentEvents[len(h.recentEvents)-h.maxEvents:]
		}
		h.recentMu.Unlock()
	}

	select {
	case h.broadcast <- frame:
	default:
		log.Warn().Msg("ui: WebSocket broadcast channel full")
	}
}

func newEventType(t string) MessageType {
	switch t {
	case "stats":
		return MessageTypeStats
	case "task_started":
		return MessageTypeTaskStarted
	case "task_completed":
		return MessageTypeTaskComplete
	case "ping":
		return MessageTypePing
	default:
		return ""
	}
}

// handleWebSocket upgrades an HTTP request to a WebSocket session and
// registers the client with the hub.
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("WebSocket upgrade failed")
		return
	}
	client := &wsClient{
		hub:  s.hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	s.hub.register <- client

	go client.writePump()
	go client.readPump()
}

// readPump drains inbound messages until the socket closes, handling
// client pings in-band. Pong responses keep the read deadline alive.
func (c *wsClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
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
		var msg Message
		if json.Unmarshal(message, &msg) == nil && msg.Type == MessageTypePing {
			pong := &Message{Type: MessageTypePong, Timestamp: time.Now().Unix()}
			if data, err := json.Marshal(pong); err == nil {
				select {
				case c.send <- data:
				default:
					log.Warn().Msg("ui: WebSocket client send buffer full")
					return
				}
			}
		}
	}
}

// writePump flushes pending frames to the socket and emits a protocol
// ping every 30s to keep proxies and idle timeouts honest.
func (c *wsClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)
			// Batch queued frames into one text message, separated
			// by newlines, so a burst of events ships in one TCP
			// write like the embedded dashboard did.
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.send)
			}
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
