package nmap

import (
	"context"

	"github.com/thekiran/iad/internal/discovery"
	"github.com/thekiran/iad/pkg/models"
)

// ServiceScanner adapts Runner to discovery.ServiceScanner.
type ServiceScanner struct {
	Runner  Runner
	Profile string
}

func (s ServiceScanner) Available() bool {
	return s.Runner.Available()
}

func (s ServiceScanner) Scan(ctx context.Context, scope models.ScanScope) ([]discovery.ScannedHost, error) {
	hosts, err := s.Runner.Scan(ctx, scope.CIDR, s.Profile)
	if err != nil {
		return nil, err
	}
	return MapHosts(hosts), nil
}

// MapHosts maps parsed Nmap hosts into discovery scanner results.
func MapHosts(hosts []Host) []discovery.ScannedHost {
	out := make([]discovery.ScannedHost, 0, len(hosts))
	for _, host := range hosts {
		scanned := discovery.ScannedHost{IP: host.IP, MAC: host.MAC, Hostname: host.Hostname}
		for _, port := range host.Ports {
			scanned.Services = append(scanned.Services, models.Service{
				Port:     port.ID,
				Protocol: port.Protocol,
				State:    "open",
				Name:     port.Service,
				Product:  port.Product,
			})
		}
		out = append(out, scanned)
	}
	return out
}
