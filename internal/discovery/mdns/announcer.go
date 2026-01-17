package mdns

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/grandcat/zeroconf"
	"github.com/rs/zerolog/log"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

const (
	ServiceType      = "_hybridgrid._tcp"
	CoordServiceType = "_hybridgrid-coord._tcp"
	Domain           = "local."
)

// Announcer advertises a worker via mDNS.
type Announcer struct {
	mu       sync.Mutex
	server   *zeroconf.Server
	instance string
	port     int
}

// AnnouncerConfig holds announcer configuration.
type AnnouncerConfig struct {
	Instance string // e.g., "worker-hostname-1234"
	Port     int
}

// NewAnnouncer creates a new mDNS announcer.
func NewAnnouncer(cfg AnnouncerConfig) *Announcer {
	return &Announcer{
		instance: cfg.Instance,
		port:     cfg.Port,
	}
}

// Start begins advertising the worker service via mDNS.
func (a *Announcer) Start(caps *pb.WorkerCapabilities) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.server != nil {
		return fmt.Errorf("announcer already started")
	}

	// Build TXT records from capabilities
	txt := buildTXTRecords(caps)

	log.Debug().
		Str("instance", a.instance).
		Int("port", a.port).
		Strs("txt", txt).
		Msg("Starting mDNS announcer")

	server, err := zeroconf.Register(
		a.instance,
		ServiceType,
		Domain,
		a.port,
		txt,
		nil, // Use all interfaces
	)
	if err != nil {
		return fmt.Errorf("failed to register mDNS service: %w", err)
	}

	a.server = server

	log.Info().
		Str("instance", a.instance).
		Str("service", ServiceType).
		Int("port", a.port).
		Msg("mDNS announcer started")

	return nil
}

// Stop stops advertising the worker service.
func (a *Announcer) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.server != nil {
		a.server.Shutdown()
		a.server = nil
		log.Info().Str("instance", a.instance).Msg("mDNS announcer stopped")
	}
}

// buildTXTRecords creates TXT records from worker capabilities.
func buildTXTRecords(caps *pb.WorkerCapabilities) []string {
	var txt []string

	if caps == nil {
		return txt
	}

	// Core info
	if caps.WorkerId != "" {
		txt = append(txt, "id="+caps.WorkerId)
	}
	if caps.Hostname != "" {
		txt = append(txt, "host="+caps.Hostname)
	}

	// Hardware
	txt = append(txt, "cpu="+strconv.Itoa(int(caps.CpuCores)))
	txt = append(txt, "ram="+strconv.FormatInt(caps.MemoryBytes/(1024*1024*1024), 10)+"G")
	txt = append(txt, "arch="+caps.NativeArch.String())

	// Docker
	txt = append(txt, "docker="+strconv.FormatBool(caps.DockerAvailable))
	if len(caps.DockerImages) > 0 {
		// Limit to first 5 images to avoid TXT record size limits
		images := caps.DockerImages
		if len(images) > 5 {
			images = images[:5]
		}
		txt = append(txt, "images="+strings.Join(images, ","))
	}

	// Capacity
	txt = append(txt, "max_parallel="+strconv.Itoa(int(caps.MaxParallelTasks)))

	// Version
	if caps.Version != "" {
		txt = append(txt, "version="+caps.Version)
	}

	// OS
	if caps.Os != "" {
		txt = append(txt, "os="+caps.Os)
	}

	return txt
}

// ParseTXTRecords parses TXT records back into a map.
func ParseTXTRecords(txt []string) map[string]string {
	result := make(map[string]string)
	for _, record := range txt {
		parts := strings.SplitN(record, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

// CoordAnnouncerConfig holds coordinator announcer configuration.
type CoordAnnouncerConfig struct {
	Instance   string // e.g., "coord-hostname"
	GRPCPort   int
	HTTPPort   int
	Version    string
	InstanceID string // unique ID for this coordinator instance
}

// CoordAnnouncer advertises a coordinator via mDNS.
type CoordAnnouncer struct {
	mu     sync.Mutex
	server *zeroconf.Server
	cfg    CoordAnnouncerConfig
}

// NewCoordAnnouncer creates a new coordinator mDNS announcer.
func NewCoordAnnouncer(cfg CoordAnnouncerConfig) *CoordAnnouncer {
	return &CoordAnnouncer{cfg: cfg}
}

// Start begins advertising the coordinator service via mDNS.
func (a *CoordAnnouncer) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.server != nil {
		return fmt.Errorf("coordinator announcer already started")
	}

	txt := a.buildTXTRecords()

	log.Debug().
		Str("instance", a.cfg.Instance).
		Int("grpc_port", a.cfg.GRPCPort).
		Int("http_port", a.cfg.HTTPPort).
		Strs("txt", txt).
		Msg("Starting coordinator mDNS announcer")

	server, err := zeroconf.Register(
		a.cfg.Instance,
		CoordServiceType,
		Domain,
		a.cfg.GRPCPort,
		txt,
		nil, // all interfaces
	)
	if err != nil {
		return fmt.Errorf("failed to register coordinator mDNS: %w", err)
	}

	a.server = server

	log.Info().
		Str("instance", a.cfg.Instance).
		Str("service", CoordServiceType).
		Int("grpc_port", a.cfg.GRPCPort).
		Msg("Coordinator mDNS announcer started")

	return nil
}

// buildTXTRecords creates TXT records for coordinator.
func (a *CoordAnnouncer) buildTXTRecords() []string {
	txt := []string{
		"grpc_port=" + strconv.Itoa(a.cfg.GRPCPort),
		"http_port=" + strconv.Itoa(a.cfg.HTTPPort),
	}
	if a.cfg.Version != "" {
		txt = append(txt, "version="+a.cfg.Version)
	}
	if a.cfg.InstanceID != "" {
		txt = append(txt, "instance_id="+a.cfg.InstanceID)
	}
	return txt
}

// Stop stops advertising the coordinator service.
func (a *CoordAnnouncer) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.server != nil {
		a.server.Shutdown()
		a.server = nil
		log.Info().Str("instance", a.cfg.Instance).Msg("Coordinator mDNS announcer stopped")
	}
}
