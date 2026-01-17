package mdns

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/rs/zerolog/log"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// DiscoveredWorker represents a worker found via mDNS.
type DiscoveredWorker struct {
	ID           string
	Address      string // host:port
	Capabilities *pb.WorkerCapabilities
	DiscoveredAt time.Time
	Source       string // "mdns"
}

// WorkerCallback is called when a worker is discovered or lost.
type WorkerCallback func(worker *DiscoveredWorker, event string)

// Browser discovers workers via mDNS.
type Browser struct {
	mu           sync.RWMutex
	workers      map[string]*DiscoveredWorker
	callback     WorkerCallback
	resolver     *zeroconf.Resolver
	browseCtx    context.Context
	browseCancel context.CancelFunc
	running      bool
	ttl          time.Duration
}

// BrowserConfig holds browser configuration.
type BrowserConfig struct {
	TTL time.Duration // How long to keep workers without re-discovery
}

// DefaultBrowserConfig returns sensible defaults.
func DefaultBrowserConfig() BrowserConfig {
	return BrowserConfig{
		TTL: 60 * time.Second,
	}
}

// NewBrowser creates a new mDNS browser.
func NewBrowser(cfg BrowserConfig, callback WorkerCallback) *Browser {
	if cfg.TTL == 0 {
		cfg.TTL = 60 * time.Second
	}
	return &Browser{
		workers:  make(map[string]*DiscoveredWorker),
		callback: callback,
		ttl:      cfg.TTL,
	}
}

// Start begins browsing for workers via mDNS.
func (b *Browser) Start() error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return fmt.Errorf("browser already running")
	}
	b.running = true
	b.mu.Unlock()

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return fmt.Errorf("failed to create resolver: %w", err)
	}
	b.resolver = resolver

	b.browseCtx, b.browseCancel = context.WithCancel(context.Background())

	// Start browsing in background
	go b.browse()

	// Start TTL cleanup goroutine
	go b.cleanupLoop()

	log.Info().
		Str("service", ServiceType).
		Dur("ttl", b.ttl).
		Msg("mDNS browser started")

	return nil
}

// Stop stops browsing for workers.
func (b *Browser) Stop() {
	b.mu.Lock()
	if !b.running {
		b.mu.Unlock()
		return
	}
	b.running = false
	b.mu.Unlock()

	if b.browseCancel != nil {
		b.browseCancel()
	}

	log.Info().Msg("mDNS browser stopped")
}

// browse continuously listens for mDNS announcements.
func (b *Browser) browse() {
	entries := make(chan *zeroconf.ServiceEntry, 100)

	// Start browsing in a single goroutine
	go func() {
		defer func() {
			// Recover from panic when channel is closed
			if r := recover(); r != nil {
				log.Debug().Interface("panic", r).Msg("mDNS browse goroutine recovered")
			}
		}()

		for {
			select {
			case <-b.browseCtx.Done():
				return
			default:
				// Browse with a timeout, then restart
				browseCtx, cancel := context.WithTimeout(b.browseCtx, 10*time.Second)
				err := b.resolver.Browse(browseCtx, ServiceType, Domain, entries)
				cancel()
				if err != nil && b.browseCtx.Err() == nil {
					log.Error().Err(err).Msg("mDNS browse error")
					time.Sleep(5 * time.Second)
				}
			}
		}
	}()

	// Process discovered services
	for {
		select {
		case <-b.browseCtx.Done():
			return
		case entry, ok := <-entries:
			if !ok {
				return
			}
			if entry == nil {
				continue
			}
			b.handleDiscovery(entry)
		}
	}
}

// handleDiscovery processes a discovered service entry.
func (b *Browser) handleDiscovery(entry *zeroconf.ServiceEntry) {
	// Parse TXT records
	txtMap := ParseTXTRecords(entry.Text)

	// Get worker ID
	workerID := txtMap["id"]
	if workerID == "" {
		workerID = entry.Instance
	}

	// Build address (prefer IPv4)
	var addr string
	for _, ip := range entry.AddrIPv4 {
		addr = net.JoinHostPort(ip.String(), strconv.Itoa(entry.Port))
		break
	}
	if addr == "" {
		for _, ip := range entry.AddrIPv6 {
			addr = net.JoinHostPort(ip.String(), strconv.Itoa(entry.Port))
			break
		}
	}
	if addr == "" {
		addr = net.JoinHostPort(entry.HostName, strconv.Itoa(entry.Port))
	}

	// Build capabilities from TXT records
	caps := parseCapsFromTXT(txtMap, entry)

	worker := &DiscoveredWorker{
		ID:           workerID,
		Address:      addr,
		Capabilities: caps,
		DiscoveredAt: time.Now(),
		Source:       "mdns",
	}

	b.mu.Lock()
	existing, exists := b.workers[workerID]
	b.workers[workerID] = worker
	b.mu.Unlock()

	if !exists {
		log.Info().
			Str("worker_id", workerID).
			Str("address", addr).
			Str("arch", caps.NativeArch.String()).
			Msg("Discovered worker via mDNS")

		if b.callback != nil {
			b.callback(worker, "found")
		}
	} else {
		// Update existing worker's discovery time
		existing.DiscoveredAt = time.Now()
	}
}

// parseCapsFromTXT builds WorkerCapabilities from TXT records.
func parseCapsFromTXT(txt map[string]string, entry *zeroconf.ServiceEntry) *pb.WorkerCapabilities {
	caps := &pb.WorkerCapabilities{
		WorkerId: txt["id"],
		Hostname: txt["host"],
		Os:       txt["os"],
		Version:  txt["version"],
	}

	if caps.Hostname == "" && entry != nil {
		caps.Hostname = strings.TrimSuffix(entry.HostName, ".")
	}

	// Parse CPU cores
	if cpu, err := strconv.Atoi(txt["cpu"]); err == nil {
		caps.CpuCores = int32(cpu)
	}

	// Parse RAM (stored as "XG")
	if ram := txt["ram"]; ram != "" {
		ram = strings.TrimSuffix(ram, "G")
		if gb, err := strconv.ParseInt(ram, 10, 64); err == nil {
			caps.MemoryBytes = gb * 1024 * 1024 * 1024
		}
	}

	// Parse architecture
	if arch := txt["arch"]; arch != "" {
		switch arch {
		case "ARCH_X86_64":
			caps.NativeArch = pb.Architecture_ARCH_X86_64
		case "ARCH_ARM64":
			caps.NativeArch = pb.Architecture_ARCH_ARM64
		case "ARCH_ARMV7":
			caps.NativeArch = pb.Architecture_ARCH_ARMV7
		}
	}

	// Parse Docker availability
	if docker := txt["docker"]; docker == "true" {
		caps.DockerAvailable = true
	}

	// Parse Docker images
	if images := txt["images"]; images != "" {
		caps.DockerImages = strings.Split(images, ",")
	}

	// Parse max parallel tasks
	if maxp, err := strconv.Atoi(txt["max_parallel"]); err == nil {
		caps.MaxParallelTasks = int32(maxp)
	}

	return caps
}

// cleanupLoop removes stale workers that haven't been re-discovered.
func (b *Browser) cleanupLoop() {
	ticker := time.NewTicker(b.ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-b.browseCtx.Done():
			return
		case <-ticker.C:
			b.cleanup()
		}
	}
}

// cleanup removes workers that haven't been seen recently.
func (b *Browser) cleanup() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	for id, worker := range b.workers {
		if now.Sub(worker.DiscoveredAt) > b.ttl {
			delete(b.workers, id)

			log.Info().
				Str("worker_id", id).
				Dur("age", now.Sub(worker.DiscoveredAt)).
				Msg("Worker removed (TTL expired)")

			if b.callback != nil {
				b.callback(worker, "lost")
			}
		}
	}
}

// List returns all currently known workers.
func (b *Browser) List() []*DiscoveredWorker {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]*DiscoveredWorker, 0, len(b.workers))
	for _, w := range b.workers {
		result = append(result, w)
	}
	return result
}

// Get returns a specific worker by ID.
func (b *Browser) Get(id string) (*DiscoveredWorker, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	w, ok := b.workers[id]
	return w, ok
}

// Count returns the number of known workers.
func (b *Browser) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.workers)
}
