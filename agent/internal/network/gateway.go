package network

import (
	"net"

	"github.com/jackpal/gateway"
)

// Gateway returns the host's default gateway IP. jackpal/gateway is pure Go and
// reads the routing table on Windows, Linux and macOS, so we avoid per-OS code
// here.
func Gateway() (net.IP, error) {
	return gateway.DiscoverGateway()
}
