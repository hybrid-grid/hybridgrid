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
	clients          map[*Client]bool
	broadcast        chan []byte
	register         chan *Client
	unregister       chan *Client
	done             chan struct{}
	mu               sync.RWMutex
	recentEvents     [][]byte // Store recent events for new clients
	eventsMu         sync.RWMutex
	maxEvents        int
	tasks            map[string]*TaskInfo
	taskOrder        []string
	buildTasks       map[string][]string
	buildOrder       []string
	buildTruncated   map[string]bool
	maxBuilds        int
	maxTasksTotal    int
	maxTasksPerBuild int
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		clients:          make(map[*Client]bool),
		broadcast:        make(chan []byte, 256),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		done:             make(chan struct{}),
		recentEvents:     make([][]byte, 0, 100),
		maxEvents:        100, // Keep last 100 events
		tasks:            make(map[string]*TaskInfo),
		buildTasks:       make(map[string][]string),
		buildTruncated:   make(map[string]bool),
		maxBuilds:        100,
		maxTasksTotal:    20000,
		maxTasksPerBuild: 5000,
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

			// Send recent events to new client
			h.eventsMu.RLock()
			for _, event := range h.recentEvents {
				select {
				case client.send <- event:
				default:
					// Skip if client buffer full
				}
			}
			h.eventsMu.RUnlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Debug().Int("clients", len(h.clients)).Msg("WebSocket client disconnected")

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client send buffer full, disconnect
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

// GetRecentEvents returns recent events as parsed JSON.
func (h *Hub) GetRecentEvents() []json.RawMessage {
	h.eventsMu.RLock()
	defer h.eventsMu.RUnlock()

	events := make([]json.RawMessage, len(h.recentEvents))
	for i, data := range h.recentEvents {
		events[i] = json.RawMessage(data)
	}
	return events
}

// GetTasks returns the latest state for each task, newest first.
func (h *Hub) GetTasks() []*TaskInfo {
	h.eventsMu.RLock()
	defer h.eventsMu.RUnlock()

	tasks := make([]*TaskInfo, 0, len(h.taskOrder))
	for i := len(h.taskOrder) - 1; i >= 0; i-- {
		task, ok := h.tasks[h.taskOrder[i]]
		if !ok {
			continue
		}
		copy := *task
		tasks = append(tasks, &copy)
	}
	return tasks
}

// GetBuilds returns derived logical build aggregates, newest first.
func (h *Hub) GetBuilds() []*BuildInfo {
	h.eventsMu.RLock()
	defer h.eventsMu.RUnlock()
	builds := make([]*BuildInfo, 0, len(h.buildOrder))
	for i := len(h.buildOrder) - 1; i >= 0; i-- {
		builds = append(builds, h.buildInfo(h.buildOrder[i]))
	}
	return builds
}

// GetBuildDetail returns a logical build and its tasks, newest first.
func (h *Hub) GetBuildDetail(buildID string) (*BuildInfo, []*TaskInfo, bool) {
	h.eventsMu.RLock()
	defer h.eventsMu.RUnlock()
	if _, ok := h.buildTasks[buildID]; !ok {
		return nil, nil, false
	}
	tasks := make([]*TaskInfo, 0, len(h.buildTasks[buildID]))
	for i := len(h.buildTasks[buildID]) - 1; i >= 0; i-- {
		task, ok := h.tasks[h.buildTasks[buildID][i]]
		if !ok {
			continue
		}
		copy := *task
		tasks = append(tasks, &copy)
	}
	return h.buildInfo(buildID), tasks, true
}

// buildInfo derives a logical build aggregate. The caller must hold eventsMu.
func (h *Hub) buildInfo(buildID string) *BuildInfo {
	build := &BuildInfo{ID: buildID, Status: "completed", Truncated: h.buildTruncated[buildID]}
	for _, task := range h.tasks {
		if task.BuildID != buildID {
			continue
		}
		if build.BuildType == "" {
			build.BuildType = task.BuildType
		}
		build.TotalTasks++
		switch task.Status {
		case "running":
			build.RunningTasks++
		case "failed":
			build.FailedTasks++
		default:
			build.CompletedTasks++
		}
		if task.FromCache {
			build.FromCacheCount++
		}
		if task.StartedAt != 0 && (build.FirstTaskAt == 0 || task.StartedAt < build.FirstTaskAt) {
			build.FirstTaskAt = task.StartedAt
		}
		lastAt := task.CompletedAt
		if lastAt == 0 {
			lastAt = task.StartedAt
		}
		if lastAt > build.LastTaskAt {
			build.LastTaskAt = lastAt
		}
	}
	if build.RunningTasks > 0 {
		build.Status = "running"
	} else if build.FailedTasks > 0 {
		build.Status = "failed"
	}
	return build
}

// storeTask records the latest task state and updates build membership.
// The caller must hold eventsMu for writing.
func (h *Hub) storeTask(task *TaskInfo) {
	if task == nil {
		return
	}

	updated := *task
	previous, exists := h.tasks[task.ID]
	if !exists {
		h.taskOrder = append(h.taskOrder, task.ID)
	} else if previous.BuildID != task.BuildID {
		h.removeTaskFromBuild(previous.BuildID, task.ID)
	}
	h.tasks[task.ID] = &updated

	if task.BuildID != "" {
		h.ensureBuild(task.BuildID)
		if ids := h.buildTasks[task.BuildID]; len(ids) < h.maxTasksPerBuild && !containsString(ids, task.ID) {
			h.buildTasks[task.BuildID] = append(ids, task.ID)
		}
		if len(h.buildTasks[task.BuildID]) >= h.maxTasksPerBuild {
			h.buildTruncated[task.BuildID] = true
		}
	}

	for len(h.tasks) > h.maxTasksTotal {
		h.evictOldestTask()
	}
}

func (h *Hub) ensureBuild(buildID string) {
	if _, exists := h.buildTasks[buildID]; exists {
		return
	}
	if h.maxBuilds <= 0 {
		return
	}
	for len(h.buildOrder) >= h.maxBuilds {
		h.evictOldestBuild()
	}
	h.buildTasks[buildID] = make([]string, 0)
	h.buildOrder = append(h.buildOrder, buildID)
}

func (h *Hub) evictOldestBuild() {
	if len(h.buildOrder) == 0 {
		return
	}
	buildID := h.buildOrder[0]
	h.buildOrder = h.buildOrder[1:]
	delete(h.buildTasks, buildID)
	delete(h.buildTruncated, buildID)
	for id, task := range h.tasks {
		if task.BuildID == buildID {
			delete(h.tasks, id)
		}
	}
	h.removeMissingTasksFromOrder()
}

func (h *Hub) evictOldestTask() {
	for len(h.taskOrder) > 0 {
		id := h.taskOrder[0]
		h.taskOrder = h.taskOrder[1:]
		if task, ok := h.tasks[id]; ok {
			delete(h.tasks, id)
			h.removeTaskFromBuild(task.BuildID, id)
			return
		}
	}
}

func (h *Hub) removeMissingTasksFromOrder() {
	order := h.taskOrder[:0]
	for _, id := range h.taskOrder {
		if _, ok := h.tasks[id]; ok {
			order = append(order, id)
		}
	}
	h.taskOrder = order
}

func (h *Hub) removeTaskFromBuild(buildID, taskID string) {
	if buildID == "" {
		return
	}
	ids := h.buildTasks[buildID]
	for i, id := range ids {
		if id == taskID {
			h.buildTasks[buildID] = append(ids[:i], ids[i+1:]...)
			return
		}
	}
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(msg *Message) {
	msg.Timestamp = time.Now().Unix()
	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal WebSocket message")
		return
	}

	// Store task events for new clients (not stats updates)
	if msg.Type == MessageTypeTaskStarted || msg.Type == MessageTypeTaskComplete {
		h.eventsMu.Lock()
		h.recentEvents = append(h.recentEvents, data)
		// Keep only last maxEvents
		if len(h.recentEvents) > h.maxEvents {
			h.recentEvents = h.recentEvents[len(h.recentEvents)-h.maxEvents:]
		}
		h.eventsMu.Unlock()
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
	h.eventsMu.Lock()
	h.storeTask(task)
	h.eventsMu.Unlock()
	h.Broadcast(&Message{
		Type: MessageTypeTaskStarted,
		Data: task,
	})
}

// BroadcastTaskCompleted notifies clients of a completed task.
func (h *Hub) BroadcastTaskCompleted(task *TaskInfo) {
	h.eventsMu.Lock()
	h.storeTask(task)
	h.eventsMu.Unlock()
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
				select {
				case c.send <- data:
				default:
					log.Warn().Msg("WebSocket client send buffer full")
					return
				}
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
