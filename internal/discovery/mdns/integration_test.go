//go:build integration || !short

package mdns

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCoordServiceType = "_hybridgrid-coord-test._tcp"

func TestIntegration_AnnounceDiscover(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	announcer := NewCoordAnnouncer(CoordAnnouncerConfig{
		Instance:    "integration-test-coord",
		GRPCPort:    29000,
		HTTPPort:    28080,
		Version:     "test-v1",
		InstanceID:  "integration-test-123",
		ServiceName: testCoordServiceType,
	})

	err := announcer.Start()
	require.NoError(t, err)
	defer announcer.Stop()

	time.Sleep(500 * time.Millisecond)

	browser := NewCoordBrowser(CoordBrowserConfig{
		Timeout:     5 * time.Second,
		ServiceName: testCoordServiceType,
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
	assert.Contains(t, coord.Address, "29000")
}

func TestIntegration_MultipleAnnouncers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	announcer1 := NewCoordAnnouncer(CoordAnnouncerConfig{
		Instance:    "multi-test-coord-1",
		GRPCPort:    39001,
		HTTPPort:    38081,
		Version:     "v1",
		ServiceName: testCoordServiceType,
	})
	announcer2 := NewCoordAnnouncer(CoordAnnouncerConfig{
		Instance:    "multi-test-coord-2",
		GRPCPort:    39002,
		HTTPPort:    38082,
		Version:     "v1",
		ServiceName: testCoordServiceType,
	})

	require.NoError(t, announcer1.Start())
	defer announcer1.Stop()

	require.NoError(t, announcer2.Start())
	defer announcer2.Stop()

	time.Sleep(500 * time.Millisecond)

	browser := NewCoordBrowser(CoordBrowserConfig{
		Timeout:     3 * time.Second,
		ServiceName: testCoordServiceType,
	})

	coord, err := browser.Discover(context.Background())
	require.NoError(t, err)

	assert.True(t,
		coord.Instance == "multi-test-coord-1" ||
			coord.Instance == "multi-test-coord-2",
		"should find one of the coordinators, got: %s", coord.Instance)
}

func TestIntegration_DiscoveryAfterAnnouncerStarts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	time.Sleep(1 * time.Second)

	browser := NewCoordBrowser(CoordBrowserConfig{
		Timeout:     5 * time.Second,
		ServiceName: testCoordServiceType,
	})

	announcer := NewCoordAnnouncer(CoordAnnouncerConfig{
		Instance:    "delayed-coord-unique-12345",
		GRPCPort:    49000,
		HTTPPort:    48080,
		Version:     "delayed",
		ServiceName: testCoordServiceType,
	})

	go func() {
		time.Sleep(300 * time.Millisecond)
		announcer.Start()
	}()
	defer announcer.Stop()

	ctx := context.Background()
	coord, err := browser.Discover(ctx)

	require.NoError(t, err)
	assert.NotEmpty(t, coord.Instance)
	assert.NotZero(t, coord.GRPCPort)
}
