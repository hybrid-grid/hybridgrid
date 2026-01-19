package mdns

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCoordAnnouncer(t *testing.T) {
	cfg := CoordAnnouncerConfig{
		Instance:   "test-coord",
		GRPCPort:   9000,
		HTTPPort:   8080,
		Version:    "v1.0.0",
		InstanceID: "test-123",
	}

	announcer := NewCoordAnnouncer(cfg)

	assert.NotNil(t, announcer)
	assert.Equal(t, cfg.Instance, announcer.cfg.Instance)
	assert.Equal(t, cfg.GRPCPort, announcer.cfg.GRPCPort)
	assert.Equal(t, cfg.HTTPPort, announcer.cfg.HTTPPort)
	assert.Equal(t, cfg.Version, announcer.cfg.Version)
	assert.Equal(t, cfg.InstanceID, announcer.cfg.InstanceID)
}

func TestCoordAnnouncer_BuildTXTRecords(t *testing.T) {
	announcer := NewCoordAnnouncer(CoordAnnouncerConfig{
		Instance:   "test",
		GRPCPort:   9000,
		HTTPPort:   8080,
		Version:    "v1.0.0",
		InstanceID: "abc123",
	})

	txt := announcer.buildTXTRecords()

	assert.Contains(t, txt, "grpc_port=9000")
	assert.Contains(t, txt, "http_port=8080")
	assert.Contains(t, txt, "version=v1.0.0")
	assert.Contains(t, txt, "instance_id=abc123")
}

func TestCoordAnnouncer_BuildTXTRecords_Minimal(t *testing.T) {
	announcer := NewCoordAnnouncer(CoordAnnouncerConfig{
		Instance: "test",
		GRPCPort: 9000,
		HTTPPort: 8080,
		// no version or instance_id
	})

	txt := announcer.buildTXTRecords()

	assert.Contains(t, txt, "grpc_port=9000")
	assert.Contains(t, txt, "http_port=8080")
	assert.Len(t, txt, 2) // only port fields
}

func TestCoordAnnouncer_StartStop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	announcer := NewCoordAnnouncer(CoordAnnouncerConfig{
		Instance: "test-coord-mdns",
		GRPCPort: 19000, // use high port to avoid conflicts
		HTTPPort: 18080,
		Version:  "test",
	})

	// Start
	err := announcer.Start()
	require.NoError(t, err)

	// Double start should error
	err = announcer.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already started")

	// Give network time to register
	time.Sleep(100 * time.Millisecond)

	// Stop
	announcer.Stop()

	// Double stop should be safe (no panic)
	announcer.Stop()
}

func TestCoordAnnouncer_StopWithoutStart(t *testing.T) {
	announcer := NewCoordAnnouncer(CoordAnnouncerConfig{
		Instance: "test",
		GRPCPort: 9000,
		HTTPPort: 8080,
	})

	// Stop without start should be safe
	announcer.Stop()
}

func TestCoordAnnouncer_ConcurrentStartStop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	announcer := NewCoordAnnouncer(CoordAnnouncerConfig{
		Instance: "concurrent-test-coord",
		GRPCPort: 29001,
		HTTPPort: 28081,
		Version:  "concurrent-test",
	})

	var wg sync.WaitGroup

	// Try concurrent starts - only one should succeed
	startErrors := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := announcer.Start()
			startErrors <- err
		}()
	}

	wg.Wait()
	close(startErrors)

	// Count successes and failures
	successCount := 0
	for err := range startErrors {
		if err == nil {
			successCount++
		}
	}

	// Exactly one should succeed
	assert.Equal(t, 1, successCount, "exactly one concurrent Start should succeed")

	// Concurrent stops should be safe
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			announcer.Stop()
		}()
	}

	wg.Wait()
}

func TestCoordAnnouncer_RestartAfterStop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	announcer := NewCoordAnnouncer(CoordAnnouncerConfig{
		Instance: "restart-test-coord",
		GRPCPort: 29002,
		HTTPPort: 28082,
		Version:  "restart-test",
	})

	// First start
	err := announcer.Start()
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	// Stop
	announcer.Stop()

	time.Sleep(50 * time.Millisecond)

	// Should be able to restart
	err = announcer.Start()
	require.NoError(t, err)

	// Wait for goroutines to settle before stopping (avoid race with zeroconf internals)
	time.Sleep(100 * time.Millisecond)

	announcer.Stop()
}
