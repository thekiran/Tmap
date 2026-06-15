//go:build !windows

package network

import (
	"context"
	"os"
	"strings"
)

// DNSServers returns the configured DNS servers by reading /etc/resolv.conf.
// The ctx argument is unused here but keeps the signature identical to the
// Windows implementation.
func DNSServers(ctx context.Context) ([]string, error) {
	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}
	var servers []string
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "nameserver" {
			servers = append(servers, fields[1])
		}
	}
	return servers, nil
}
