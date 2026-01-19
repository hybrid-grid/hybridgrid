//go:build windows

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/grpc/client"
	workerserver "github.com/h3nr1-d14z/hybridgrid/internal/worker/server"
)

const (
	serviceName = "HybridGridWorker"
	serviceDesc = "Hybrid-Grid Build Worker Service"
)

// workerService implements the Windows service interface.
type workerService struct {
	coordinator   string
	port          int
	httpPort      int
	token         string
	maxParallel   int
	advertiseAddr string
	stopChan      chan struct{}
	elog          debug.Log
}

// Execute is the main service loop required by the Windows Service Control Manager.
func (s *workerService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	changes <- svc.Status{State: svc.StartPending}

	// Initialize logging to file for service mode
	logFile, err := os.OpenFile(filepath.Join(os.TempDir(), "hg-worker.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: logFile, NoColor: true})
	}

	if s.maxParallel == 0 {
		s.maxParallel = runtime.NumCPU()
	}

	hostname, _ := os.Hostname()
	log.Info().
		Str("coordinator", s.coordinator).
		Str("hostname", hostname).
		Int("port", s.port).
		Int("max_parallel", s.maxParallel).
		Msg("Starting Hybrid-Grid Worker as Windows Service")

	// Create worker server
	cfg := workerserver.DefaultConfig()
	cfg.Port = s.port
	cfg.MaxConcurrent = s.maxParallel

	srv := workerserver.New(cfg)
	caps := srv.Capabilities()
	caps.WorkerId = fmt.Sprintf("worker-%s", hostname)
	caps.MaxParallelTasks = int32(s.maxParallel)
	caps.Version = version

	// Connect to coordinator
	cli, err := client.New(client.Config{
		Address:   s.coordinator,
		AuthToken: s.token,
		Timeout:   30 * time.Second,
		Insecure:  true,
	})
	if err != nil {
		s.elog.Error(1, fmt.Sprintf("failed to connect to coordinator: %v", err))
		return false, 1
	}

	// Determine worker address to advertise
	workerAddr := s.advertiseAddr
	if workerAddr == "" {
		workerAddr = fmt.Sprintf("%s:%d", hostname, s.port)
	}

	// Register with coordinator
	regReq := &pb.HandshakeRequest{
		Capabilities:  caps,
		AuthToken:     s.token,
		WorkerAddress: workerAddr,
	}
	resp, err := cli.Handshake(context.Background(), regReq)
	if err != nil {
		s.elog.Error(1, fmt.Sprintf("handshake failed: %v", err))
		cli.Close()
		return false, 1
	}

	if !resp.Accepted {
		s.elog.Error(1, fmt.Sprintf("worker registration rejected: %s", resp.Message))
		cli.Close()
		return false, 1
	}

	log.Info().
		Str("worker_id", resp.AssignedWorkerId).
		Int32("heartbeat_interval", resp.HeartbeatIntervalSeconds).
		Msg("Worker registered successfully")

	errCh := make(chan error, 2)
	stopHeartbeat := make(chan struct{})

	// Start worker gRPC server
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- fmt.Errorf("gRPC server: %w", err)
		}
	}()

	// Start metrics HTTP server
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.httpPort),
		Handler: metricsMux,
	}
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("metrics server: %w", err)
		}
	}()

	// Start heartbeat loop
	heartbeatInterval := time.Duration(resp.HeartbeatIntervalSeconds) * time.Second
	go func() {
		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				hResp, err := cli.Handshake(context.Background(), regReq)
				if err != nil {
					log.Warn().Err(err).Msg("Heartbeat failed")
				} else {
					log.Debug().Bool("accepted", hResp.Accepted).Msg("Heartbeat sent")
				}
			case <-stopHeartbeat:
				return
			}
		}
	}()

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
	close(stopHeartbeat)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	metricsServer.Shutdown(ctx)
	srv.Stop()
	cli.Close()

	return false, 0
}

// runAsService runs the worker as a Windows Service.
func runAsService(coordinator string, port, httpPort int, token string, maxParallel int, advertiseAddr string) error {
	elog, err := eventlog.Open(serviceName)
	if err != nil {
		return err
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", serviceName))

	s := &workerService{
		coordinator:   coordinator,
		port:          port,
		httpPort:      httpPort,
		token:         token,
		maxParallel:   maxParallel,
		advertiseAddr: advertiseAddr,
		stopChan:      make(chan struct{}),
		elog:          elog,
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

// installService installs the worker as a Windows Service.
func installService(exePath, coordinator string) error {
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

	// Service arguments: serve --coordinator=<addr>
	args := []string{"serve"}
	if coordinator != "" {
		args = append(args, "--coordinator="+coordinator)
	}

	s, err = m.CreateService(serviceName, exePath, mgr.Config{
		DisplayName: "Hybrid-Grid Worker",
		Description: serviceDesc,
		StartType:   mgr.StartAutomatic,
	}, args...)
	if err != nil {
		return err
	}
	defer s.Close()

	// Set recovery actions
	err = s.SetRecoveryActions([]mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: 5 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 30 * time.Second},
		{Type: mgr.ServiceRestart, Delay: 60 * time.Second},
	}, 86400)

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
