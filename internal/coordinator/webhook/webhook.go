// Package webhook delivers coordinator task events to HTTP webhook
// endpoints: the lightest plugin tier (see docs/plugins.md). Delivery
// is fire-and-forget by design — NotifyTaskStarted/Completed never
// block the dispatch or completion paths; events queue in a bounded
// buffer and drop (with a log line) under sustained endpoint failure.
package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/server"
)

// Config configures a Notifier.
type Config struct {
	// Endpoints are the URLs every event is POSTed to.
	Endpoints []string

	// Timeout bounds one POST. Default 3s.
	Timeout time.Duration

	// Retries is the number of additional attempts per event per
	// endpoint after a failure. Default 2.
	Retries int

	// Buffer is the event queue depth. Default 256. Overflow drops.
	Buffer int

	// Client overrides the HTTP client (tests).
	Client *http.Client
}

// event is the wire form POSTed to the endpoints.
type event struct {
	Type string `json:"type"` // task_started | task_completed
	server.TaskEvent
}

// Notifier implements server.EventNotifier over HTTP webhooks.
type Notifier struct {
	cfg    Config
	client *http.Client
	ch     chan event
	done   chan struct{}
	wg     sync.WaitGroup
}

// New starts a Notifier and its delivery pump.
func New(cfg Config) *Notifier {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 3 * time.Second
	}
	if cfg.Retries < 0 {
		cfg.Retries = 0
	}
	if cfg.Retries == 0 {
		cfg.Retries = 2
	}
	if cfg.Buffer <= 0 {
		cfg.Buffer = 256
	}
	client := cfg.Client
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
	}
	n := &Notifier{
		cfg:    cfg,
		client: client,
		ch:     make(chan event, cfg.Buffer),
		done:   make(chan struct{}),
	}
	n.wg.Add(1)
	go n.pump()
	return n
}

// NotifyTaskStarted enqueues a start event; never blocks.
func (n *Notifier) NotifyTaskStarted(e *server.TaskEvent) {
	n.enqueue(event{Type: "task_started", TaskEvent: *clone(e)})
}

// NotifyTaskCompleted enqueues a completion event; never blocks.
func (n *Notifier) NotifyTaskCompleted(e *server.TaskEvent) {
	n.enqueue(event{Type: "task_completed", TaskEvent: *clone(e)})
}

// Close stops the pump. Queued events are dropped: webhooks are
// best-effort by contract.
func (n *Notifier) Close() {
	close(n.done)
	n.wg.Wait()
}

func (n *Notifier) enqueue(ev event) {
	select {
	case n.ch <- ev:
	default:
		log.Warn().Str("type", ev.Type).Str("task_id", ev.ID).
			Msg("webhook queue full; event dropped")
	}
}

// pump delivers events sequentially with retries. A slow endpoint
// delays later events (bounded by Buffer) but never the coordinator.
func (n *Notifier) pump() {
	defer n.wg.Done()
	for {
		select {
		case ev := <-n.ch:
			n.deliver(ev)
		case <-n.done:
			return
		}
	}
}

func (n *Notifier) deliver(ev event) {
	body, err := json.Marshal(ev)
	if err != nil {
		log.Error().Err(err).Msg("webhook: encode event failed")
		return
	}
	for _, url := range n.cfg.Endpoints {
		n.post(url, body, ev)
	}
}

func (n *Notifier) post(url string, body []byte, ev event) {
	var lastErr error
	for attempt := 0; attempt <= n.cfg.Retries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 250 * time.Millisecond)
		}
		resp, err := n.client.Post(url, "application/json", bytes.NewReader(body))
		if err != nil {
			lastErr = err
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return
		}
		lastErr = fmt.Errorf("status %d", resp.StatusCode)
	}
	log.Warn().Err(lastErr).Str("url", url).Str("type", ev.Type).Str("task_id", ev.ID).
		Msg("webhook delivery failed after retries; event dropped")
}

// clone copies a TaskEvent so queued events are immune to later
// mutation by the caller.
func clone(e *server.TaskEvent) *server.TaskEvent {
	copied := *e
	return &copied
}
