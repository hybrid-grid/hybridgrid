package ui

import (
	"encoding/json"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

func (s *Server) handleWorkerJoin(w http.ResponseWriter, r *http.Request) {
	address := s.config.CoordinatorAddress
	if s.config.WorkerConnectAddress != "" {
		address = s.config.WorkerConnectAddress
		// Compose publishes gRPC on the same host as this dashboard.
		if host, port, err := net.SplitHostPort(address); err == nil && host == "" {
			u := url.URL{Host: r.Host}
			h := u.Hostname()
			ip := net.ParseIP(h)
			address = ""
			if h != "" && !strings.EqualFold(h, "localhost") && !ip.IsLoopback() && !ip.IsUnspecified() {
				address = net.JoinHostPort(h, port)
			}
		}
	}
	addresses := workerJoinAddresses(address, net.InterfaceAddrs)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"addresses": addresses})
}

// Expand loopback only when the dashboard connects to its own machine.
func workerJoinAddresses(address string, interfaces func() ([]net.Addr, error)) []string {
	result := []string{}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return result
	}
	n, err := strconv.ParseUint(port, 10, 16)
	if err != nil || n == 0 {
		return result
	}
	ip := net.ParseIP(host)
	if host != "" && !strings.EqualFold(host, "localhost") && !ip.IsLoopback() && !ip.IsUnspecified() {
		return []string{address}
	}
	addrs, err := interfaces()
	if err != nil {
		return result
	}
	seen := make(map[netip.Addr]bool)
	var ips []netip.Addr
	for _, addr := range addrs {
		prefix, err := netip.ParsePrefix(addr.String())
		if err != nil {
			continue
		}
		ip := prefix.Addr().Unmap()
		if ip.IsGlobalUnicast() && !seen[ip] {
			seen[ip] = true
			ips = append(ips, ip)
		}
	}
	sort.Slice(ips, func(i, j int) bool {
		if ips[i].Is4() != ips[j].Is4() {
			return ips[i].Is4()
		}
		if ips[i].IsPrivate() != ips[j].IsPrivate() {
			return ips[i].IsPrivate()
		}
		return ips[i].Less(ips[j])
	})
	for _, ip := range ips {
		result = append(result, net.JoinHostPort(ip.String(), port))
	}
	return result
}
