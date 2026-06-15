//go:build windows

package network

import (
	"context"

	"github.com/thekiran/iad/internal/system"
)

// DNSServers returns the configured IPv4 DNS server addresses. We shell out to
// PowerShell's Get-DnsClientServerAddress instead of parsing `ipconfig /all`
// because its output is locale-independent (the host may be non-English) and
// yields a clean list of IPs.
func DNSServers(ctx context.Context) ([]string, error) {
	out, err := system.Run(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command",
		"Get-DnsClientServerAddress -AddressFamily IPv4 | Select-Object -ExpandProperty ServerAddresses")
	servers := extractIPv4(out)
	if len(servers) == 0 && err != nil {
		return nil, err
	}
	return servers, nil
}
