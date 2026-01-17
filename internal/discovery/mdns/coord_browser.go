package mdns

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/rs/zerolog/log"
)

// DiscoveredCoordinator represents a coordinator found via mDNS.
type DiscoveredCoordinator struct {
	Instance   string
	Address    string // host:grpc_port
	GRPCPort   int
	HTTPPort   int
	Version    string
	InstanceID string
}

// CoordBrowserConfig holds coordinator browser configuration.
type CoordBrowserConfig struct {
	Timeout time.Duration // discovery timeout
}

// DefaultCoordBrowserConfig returns sensible defaults.
func DefaultCoordBrowserConfig() CoordBrowserConfig {
	return CoordBrowserConfig{
		Timeout: 10 * time.Second,
	}
}

// CoordBrowser discovers coordinators via mDNS.
type CoordBrowser struct {
	timeout time.Duration
}

// NewCoordBrowser creates a new coordinator mDNS browser.
func NewCoordBrowser(cfg CoordBrowserConfig) *CoordBrowser {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &CoordBrowser{
		timeout: cfg.Timeout,
	}
}

// Discover searches for a coordinator on the local network.
// Returns the first coordinator found or error if timeout expires.
func (b *CoordBrowser) Discover(ctx context.Context) (*DiscoveredCoordinator, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry, 10)
	result := make(chan *DiscoveredCoordinator, 1)
	errCh := make(chan error, 1)

	// Create timeout context
	discoverCtx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	log.Debug().
		Str("service", CoordServiceType).
		Dur("timeout", b.timeout).
		Msg("Starting coordinator discovery")

	// Start browsing
	go func() {
		err := resolver.Browse(discoverCtx, CoordServiceType, Domain, entries)
		if err != nil {
			select {
			case errCh <- fmt.Errorf("browse failed: %w", err):
			default:
			}
		}
	}()

	// Process entries
	go func() {
		for entry := range entries {
			if entry == nil {
				continue
			}
			coord := b.parseEntry(entry)
			if coord != nil {
				select {
				case result <- coord:
				default:
				}
				return
			}
		}
	}()

	// Wait for result, error, or timeout
	select {
	case coord := <-result:
		log.Info().
			Str("instance", coord.Instance).
			Str("address", coord.Address).
			Msg("Discovered coordinator via mDNS")
		return coord, nil
	case err := <-errCh:
		return nil, err
	case <-discoverCtx.Done():
		return nil, fmt.Errorf("coordinator discovery timeout after %v", b.timeout)
	}
}

// parseEntry converts a zeroconf entry to DiscoveredCoordinator.
func (b *CoordBrowser) parseEntry(entry *zeroconf.ServiceEntry) *DiscoveredCoordinator {
	// Parse TXT records
	txt := ParseTXTRecords(entry.Text)

	// Get gRPC port from TXT or use entry port
	grpcPort := entry.Port
	if p, err := strconv.Atoi(txt["grpc_port"]); err == nil {
		grpcPort = p
	}

	httpPort := 0
	if p, err := strconv.Atoi(txt["http_port"]); err == nil {
		httpPort = p
	}

	// Build address (prefer IPv4)
	var host string
	for _, ip := range entry.AddrIPv4 {
		host = ip.String()
		break
	}
	if host == "" {
		for _, ip := range entry.AddrIPv6 {
			host = ip.String()
			break
		}
	}
	if host == "" {
		host = entry.HostName
	}

	addr := net.JoinHostPort(host, strconv.Itoa(grpcPort))

	return &DiscoveredCoordinator{
		Instance:   entry.Instance,
		Address:    addr,
		GRPCPort:   grpcPort,
		HTTPPort:   httpPort,
		Version:    txt["version"],
		InstanceID: txt["instance_id"],
	}
}

// DiscoverWithFallback tries mDNS discovery, falls back to provided address.
func (b *CoordBrowser) DiscoverWithFallback(ctx context.Context, fallback string) (string, error) {
	coord, err := b.Discover(ctx)
	if err == nil {
		return coord.Address, nil
	}

	log.Warn().
		Err(err).
		Str("fallback", fallback).
		Msg("mDNS discovery failed, using fallback")

	if fallback != "" {
		return fallback, nil
	}

	return "", fmt.Errorf("no coordinator found: mDNS failed (%v) and no fallback provided", err)
}
