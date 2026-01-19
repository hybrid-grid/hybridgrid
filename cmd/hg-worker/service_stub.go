//go:build !windows

package main

// IsWindowsService returns false on non-Windows platforms.
func IsWindowsService() bool {
	return false
}

// runAsService is a no-op on non-Windows platforms.
//
//nolint:unused
func runAsService(coordinator string, port, httpPort int, token string, maxParallel int, advertiseAddr string) error {
	return nil
}

// installService is a no-op on non-Windows platforms.
//
//nolint:unused
func installService(exePath, coordinator string) error {
	return nil
}

// uninstallService is a no-op on non-Windows platforms.
//
//nolint:unused
func uninstallService() error {
	return nil
}
