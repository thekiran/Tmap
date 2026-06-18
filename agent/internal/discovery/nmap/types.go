// Package nmap integrates the optional external Nmap scanner for LAN service
// discovery. It only runs safe TCP connect profiles, requests XML output, and
// maps parsed hosts into discovery.ScannedHost values.
package nmap

import "github.com/thekiran/iad/pkg/models"

// Host is a parsed Nmap host result, reduced to the fields the topology mapper
// uses. Parser output includes only up hosts and open ports.
type Host struct {
	IP       string
	MAC      string
	Vendor   string
	Hostname string
	Ports    []Port
}

// Port is a parsed open port.
type Port struct {
	ID       int
	Protocol string
	Service  string
	Product  string
	Version  string
}

// Detection describes an installed Nmap binary.
type Detection struct {
	Found   bool
	Path    string
	Version string
}

// MergeDevice is the minimal device shape used by MergeServices. It matches the
// report model without making this helper responsible for full discovery state.
type MergeDevice struct {
	IP       string
	Services []models.Service
}
