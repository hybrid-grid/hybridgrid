//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"

	coordserver "github.com/h3nr1-d14z/hybridgrid/internal/coordinator/server"
	"github.com/h3nr1-d14z/hybridgrid/internal/discovery/mdns"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/dashboard"
)

const (
	serviceName = "HybridGridCoord"
	serviceDesc = "Hybrid-Grid Build Coordinator Service"
)

// coordService implements the Windows service interface.
type coordService struct {
	grpcPort int
	httpPort int
	token    string
	noMdns   bool
	stopChan chan struct{}
	elog     debug.Log
}

// Execute is the main service loop required by the Windows Service Control Manager.
func (s *coordService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	changes <- svc.Status{State: svc.StartPending}

	// Initialize logging to file for service mode
	logFile, err := os.OpenFile(filepath.Join(os.TempDir(), "hg-coord.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: logFile, NoColor: true})
	}

	log.Info().
		Int("grpc_port", s.grpcPort).
		Int("http_port", s.httpPort).
		Msg("Starting Hybrid-Grid Coordinator as Windows Service")

	// Start coordinator
	cfg := coordserver.DefaultConfig()
	cfg.Port = s.grpcPort
	cfg.AuthToken = s.token
	cfg.HeartbeatTTL = 60 * time.Second
	cfg.RequestTimeout = 120 * time.Second

	srv := coordserver.New(cfg)

	errCh := make(chan error, 2)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- fmt.Errorf("gRPC server: %w", err)
		}
	}()

	// Start HTTP dashboard server
	dashCfg := dashboard.DefaultConfig()
	dashCfg.Port = s.httpPort
	dashSrv := dashboard.New(dashCfg, srv.NewStatsProvider())

	go func() {
		if err := dashSrv.Start(); err != nil {
			errCh <- fmt.Errorf("dashboard server: %w", err)
		}
	}()

	// Start mDNS announcer
	var mdnsAnnouncer *mdns.CoordAnnouncer
	if !s.noMdns {
		hostname, _ := os.Hostname()
		mdnsAnnouncer = mdns.NewCoordAnnouncer(mdns.CoordAnnouncerConfig{
			Instance:   fmt.Sprintf("hg-coord-%s", hostname),
			GRPCPort:   s.grpcPort,
			HTTPPort:   s.httpPort,
			Version:    version,
			InstanceID: fmt.Sprintf("%s-%d", hostname, os.Getpid()),
		})

		if err := mdnsAnnouncer.Start(); err != nil {
			log.Warn().Err(err).Msg("Failed to start mDNS announcer")
		}
	}

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	// Main service loop
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				log.Info().Msg("Received stop/shutdown command")
				break loop
			default:
				s.elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		case err := <-errCh:
			s.elog.Error(1, fmt.Sprintf("server error: %v", err))
			break loop
		}
	}

	changes <- svc.Status{State: svc.StopPending}

	// Cleanup
	if mdnsAnnouncer != nil {
		mdnsAnnouncer.Stop()
	}
	dashSrv.Stop()
	srv.Stop()

	return false, 0
}

// runAsService runs the coordinator as a Windows Service.
func runAsService(grpcPort, httpPort int, token string, noMdns bool) error {
	elog, err := eventlog.Open(serviceName)
	if err != nil {
		return err
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", serviceName))

	s := &coordService{
		grpcPort: grpcPort,
		httpPort: httpPort,
		token:    token,
		noMdns:   noMdns,
		stopChan: make(chan struct{}),
		elog:     elog,
	}

	err = svc.Run(serviceName, s)
	if err != nil {
		elog.Error(1, fmt.Sprintf("service failed: %v", err))
		return err
	}

	elog.Info(1, fmt.Sprintf("%s service stopped", serviceName))
	return nil
}

// IsWindowsService checks if the process is running as a Windows Service.
func IsWindowsService() bool {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return false
	}
	return isService
}

// installService installs the coordinator as a Windows Service.
func installService(exePath string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", serviceName)
	}

	if exePath == "" {
		exePath, err = os.Executable()
		if err != nil {
			return err
		}
	}

	s, err = m.CreateService(serviceName, exePath, mgr.Config{
		DisplayName: "Hybrid-Grid Coordinator",
		Description: serviceDesc,
		StartType:   mgr.StartAutomatic,
	}, "serve")
	if err != nil {
		return err
	}
	defer s.Close()

	// Set recovery actions
	err = s.SetRecoveryActions([]mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 5 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 30 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 60 * time.Second},
	}, 86400) // Reset failure count after 1 day

	if err != nil {
		return fmt.Errorf("failed to set recovery actions: %w", err)
	}

	log.Info().Str("service", serviceName).Msg("Service installed successfully")
	return nil
}

// uninstallService removes the Windows Service.
func uninstallService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s not installed", serviceName)
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		return err
	}

	log.Info().Str("service", serviceName).Msg("Service uninstalled successfully")
	return nil
}
