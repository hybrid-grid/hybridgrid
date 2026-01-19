//go:build !windows

package main

// IsWindowsService returns false on non-Windows platforms.
func IsWindowsService() bool {
	return false
}

// runAsService is a no-op on non-Windows platforms.
func runAsService(grpcPort, httpPort int, token string, noMdns bool) error {
	return nil
}

// installService is a no-op on non-Windows platforms.
func installService(exePath string) error {
	return nil
}

// uninstallService is a no-op on non-Windows platforms.
func uninstallService() error {
	return nil
}
