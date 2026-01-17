package mdns

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCoordBrowserConfig(t *testing.T) {
	cfg := DefaultCoordBrowserConfig()

	assert.Equal(t, 10*time.Second, cfg.Timeout)
}

func TestNewCoordBrowser(t *testing.T) {
	browser := NewCoordBrowser(CoordBrowserConfig{
		Timeout: 5 * time.Second,
	})

	assert.NotNil(t, browser)
	assert.Equal(t, 5*time.Second, browser.timeout)
}

func TestNewCoordBrowser_DefaultTimeout(t *testing.T) {
	browser := NewCoordBrowser(CoordBrowserConfig{})

	assert.Equal(t, 10*time.Second, browser.timeout)
}

func TestCoordBrowser_DiscoverTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	browser := NewCoordBrowser(CoordBrowserConfig{
		Timeout: 500 * time.Millisecond, // short timeout
	})

	ctx := context.Background()
	_, err := browser.Discover(ctx)

	// Should timeout since no coordinator running
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestCoordBrowser_DiscoverWithFallback_UsesFallback(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	browser := NewCoordBrowser(CoordBrowserConfig{
		Timeout: 100 * time.Millisecond,
	})

	ctx := context.Background()
	addr, err := browser.DiscoverWithFallback(ctx, "fallback:9000")

	require.NoError(t, err)
	assert.Equal(t, "fallback:9000", addr)
}

func TestCoordBrowser_DiscoverWithFallback_NoFallback(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	browser := NewCoordBrowser(CoordBrowserConfig{
		Timeout: 100 * time.Millisecond,
	})

	ctx := context.Background()
	_, err := browser.DiscoverWithFallback(ctx, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no coordinator found")
}

func TestCoordBrowser_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	browser := NewCoordBrowser(CoordBrowserConfig{
		Timeout: 30 * time.Second, // long timeout
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err := browser.Discover(ctx)
	elapsed := time.Since(start)

	// Should return quickly due to cancellation (< 1 second)
	assert.Error(t, err)
	assert.Less(t, elapsed, 2*time.Second)
}

func TestCoordBrowser_ParseEntry_IPv4(t *testing.T) {
	browser := NewCoordBrowser(CoordBrowserConfig{})

	entry := zeroconf.NewServiceEntry("test-coord", CoordServiceType, Domain)
	entry.Port = 9000
	entry.AddrIPv4 = []net.IP{net.ParseIP("192.168.1.100")}
	entry.Text = []string{"grpc_port=9001", "http_port=8080", "version=v1", "instance_id=abc"}

	coord := browser.parseEntry(entry)

	assert.Equal(t, "test-coord", coord.Instance)
	assert.Equal(t, "192.168.1.100:9001", coord.Address)
	assert.Equal(t, 9001, coord.GRPCPort)
	assert.Equal(t, 8080, coord.HTTPPort)
	assert.Equal(t, "v1", coord.Version)
	assert.Equal(t, "abc", coord.InstanceID)
}

func TestCoordBrowser_ParseEntry_IPv6Fallback(t *testing.T) {
	browser := NewCoordBrowser(CoordBrowserConfig{})

	entry := zeroconf.NewServiceEntry("test-coord", CoordServiceType, Domain)
	entry.Port = 9000
	entry.AddrIPv4 = nil // no IPv4
	entry.AddrIPv6 = []net.IP{net.ParseIP("::1")}
	entry.Text = []string{"grpc_port=9000"}

	coord := browser.parseEntry(entry)

	assert.Equal(t, "[::1]:9000", coord.Address)
}

func TestCoordBrowser_ParseEntry_HostnameFallback(t *testing.T) {
	browser := NewCoordBrowser(CoordBrowserConfig{})

	entry := zeroconf.NewServiceEntry("test-coord", CoordServiceType, Domain)
	entry.Port = 9000
	entry.AddrIPv4 = nil
	entry.AddrIPv6 = nil
	entry.HostName = "coordinator.local"
	entry.Text = []string{"grpc_port=9000"}

	coord := browser.parseEntry(entry)

	assert.Equal(t, "coordinator.local:9000", coord.Address)
}

func TestCoordBrowser_ParseEntry_NoTXTRecords(t *testing.T) {
	browser := NewCoordBrowser(CoordBrowserConfig{})

	entry := zeroconf.NewServiceEntry("test-coord", CoordServiceType, Domain)
	entry.Port = 9000
	entry.AddrIPv4 = []net.IP{net.ParseIP("10.0.0.1")}
	entry.Text = nil // no TXT records

	coord := browser.parseEntry(entry)

	// Should use entry port as fallback
	assert.Equal(t, 9000, coord.GRPCPort)
	assert.Equal(t, 0, coord.HTTPPort)
	assert.Equal(t, "", coord.Version)
}

func TestCoordBrowser_ParseEntry_InvalidPortInTXT(t *testing.T) {
	browser := NewCoordBrowser(CoordBrowserConfig{})

	entry := zeroconf.NewServiceEntry("test-coord", CoordServiceType, Domain)
	entry.Port = 9000
	entry.AddrIPv4 = []net.IP{net.ParseIP("10.0.0.1")}
	entry.Text = []string{"grpc_port=invalid", "http_port=notanumber"}

	coord := browser.parseEntry(entry)

	// Should fall back to entry port
	assert.Equal(t, 9000, coord.GRPCPort)
	assert.Equal(t, 0, coord.HTTPPort)
}
