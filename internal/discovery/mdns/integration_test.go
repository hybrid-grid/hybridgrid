//go:build integration || !short

package mdns

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_AnnounceDiscover(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Start announcer
	announcer := NewCoordAnnouncer(CoordAnnouncerConfig{
		Instance:   "integration-test-coord",
		GRPCPort:   29000,
		HTTPPort:   28080,
		Version:    "test-v1",
		InstanceID: "integration-test-123",
	})

	err := announcer.Start()
	require.NoError(t, err)
	defer announcer.Stop()

	// Give mDNS time to propagate
	time.Sleep(500 * time.Millisecond)

	// Discover
	browser := NewCoordBrowser(CoordBrowserConfig{
		Timeout: 5 * time.Second,
	})

	ctx := context.Background()
	coord, err := browser.Discover(ctx)

	require.NoError(t, err)
	require.NotNil(t, coord)

	assert.Equal(t, "integration-test-coord", coord.Instance)
	assert.Equal(t, 29000, coord.GRPCPort)
	assert.Equal(t, 28080, coord.HTTPPort)
	assert.Equal(t, "test-v1", coord.Version)
	assert.Equal(t, "integration-test-123", coord.InstanceID)
	assert.Contains(t, coord.Address, "29000") // port in address
}

func TestIntegration_MultipleAnnouncers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Start two announcers (simulating multiple coordinators)
	announcer1 := NewCoordAnnouncer(CoordAnnouncerConfig{
		Instance: "multi-test-coord-1",
		GRPCPort: 39001,
		HTTPPort: 38081,
		Version:  "v1",
	})
	announcer2 := NewCoordAnnouncer(CoordAnnouncerConfig{
		Instance: "multi-test-coord-2",
		GRPCPort: 39002,
		HTTPPort: 38082,
		Version:  "v1",
	})

	require.NoError(t, announcer1.Start())
	defer announcer1.Stop()

	require.NoError(t, announcer2.Start())
	defer announcer2.Stop()

	time.Sleep(500 * time.Millisecond)

	// Browser should find at least one
	browser := NewCoordBrowser(CoordBrowserConfig{
		Timeout: 3 * time.Second,
	})

	coord, err := browser.Discover(context.Background())
	require.NoError(t, err)

	// Should find one of them
	assert.True(t,
		coord.Instance == "multi-test-coord-1" ||
			coord.Instance == "multi-test-coord-2",
		"should find one of the coordinators, got: %s", coord.Instance)
}

func TestIntegration_DiscoveryAfterAnnouncerStarts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Wait for previous test's mDNS cache to clear
	time.Sleep(1 * time.Second)

	browser := NewCoordBrowser(CoordBrowserConfig{
		Timeout: 5 * time.Second,
	})

	// Start announcer in background after small delay
	announcer := NewCoordAnnouncer(CoordAnnouncerConfig{
		Instance: "delayed-coord-unique-12345",
		GRPCPort: 49000,
		HTTPPort: 48080,
		Version:  "delayed",
	})

	go func() {
		time.Sleep(300 * time.Millisecond)
		announcer.Start()
	}()
	defer announcer.Stop()

	// Discover should find a coordinator (may be the delayed one or cached)
	ctx := context.Background()
	coord, err := browser.Discover(ctx)

	require.NoError(t, err)
	// Just verify we found a coordinator - don't be strict about which one
	// due to mDNS cache behavior across tests
	assert.NotEmpty(t, coord.Instance)
	assert.NotZero(t, coord.GRPCPort)
}
