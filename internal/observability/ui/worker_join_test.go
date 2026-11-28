package ui

import (
	"net"
	"slices"
	"testing"
)

func TestWorkerJoinAddresses(t *testing.T) {
	interfaces := func() ([]net.Addr, error) {
		var result []net.Addr
		for _, cidr := range []string{"fd00::10/64", "127.0.0.1/8", "192.168.1.9/24", "fe80::1/64"} {
			ip, network, err := net.ParseCIDR(cidr)
			if err != nil {
				t.Fatal(err)
			}
			network.IP = ip
			result = append(result, network)
		}
		return result, nil
	}
	for addr, want := range map[string][]string{
		"localhost:9100":     {"192.168.1.9:9100", "[fd00::10]:9100"},
		"[::1]:9100":         {"192.168.1.9:9100", "[fd00::10]:9100"},
		"coord.example:9101": {"coord.example:9101"},
		"invalid":            {},
	} {
		if got := workerJoinAddresses(addr, interfaces); !slices.Equal(got, want) {
			t.Errorf("%s: got %v, want %v", addr, got, want)
		}
	}
}
