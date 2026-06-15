// Package nmap integrates the external Nmap scanner safely: it runs Nmap via
// os/exec with a context timeout, requests XML output only, and parses that XML
// with encoding/xml. It never parses Nmap's human terminal output, never enables
// stealth/evasion options, and degrades gracefully when Nmap is not installed.
package nmap

import (
	"encoding/xml"
	"strconv"
	"strings"
)

// Host is a parsed Nmap host result, reduced to the fields the topology mapper
// uses. Only "up" hosts and "open" ports are surfaced by Parse.
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

// --- XML schema (subset of Nmap's DTD) -------------------------------------

type xmlRun struct {
	XMLName xml.Name  `xml:"nmaprun"`
	Hosts   []xmlHost `xml:"host"`
}

type xmlHost struct {
	Status    xmlStatus     `xml:"status"`
	Addresses []xmlAddress  `xml:"address"`
	Hostnames []xmlHostname `xml:"hostnames>hostname"`
	Ports     []xmlPort     `xml:"ports>port"`
}

type xmlStatus struct {
	State string `xml:"state,attr"`
}

type xmlAddress struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
	Vendor   string `xml:"vendor,attr"`
}

type xmlHostname struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

type xmlPort struct {
	Protocol string     `xml:"protocol,attr"`
	PortID   string     `xml:"portid,attr"`
	State    xmlState   `xml:"state"`
	Service  xmlService `xml:"service"`
}

type xmlState struct {
	State string `xml:"state,attr"`
}

type xmlService struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr"`
	Version string `xml:"version,attr"`
}

// Parse converts Nmap XML (`-oX -`) into Hosts. It is pure and deterministic:
// the same XML always yields the same result. Hosts that are not "up" and ports
// that are not "open" are dropped. A malformed document returns an error.
func Parse(data []byte) ([]Host, error) {
	var run xmlRun
	if err := xml.Unmarshal(data, &run); err != nil {
		return nil, err
	}
	var hosts []Host
	for _, h := range run.Hosts {
		if h.Status.State != "" && h.Status.State != "up" {
			continue
		}
		host := Host{}
		for _, a := range h.Addresses {
			switch a.AddrType {
			case "ipv4", "ipv6":
				if host.IP == "" {
					host.IP = a.Addr
				}
			case "mac":
				host.MAC = strings.ToLower(strings.ReplaceAll(a.Addr, "-", ":"))
				host.Vendor = a.Vendor
			}
		}
		if host.IP == "" {
			continue
		}
		for _, hn := range h.Hostnames {
			if hn.Name != "" {
				host.Hostname = hn.Name
				break
			}
		}
		for _, p := range h.Ports {
			if p.State.State != "open" {
				continue
			}
			id, err := strconv.Atoi(p.PortID)
			if err != nil {
				continue
			}
			host.Ports = append(host.Ports, Port{
				ID:       id,
				Protocol: p.Protocol,
				Service:  p.Service.Name,
				Product:  p.Service.Product,
				Version:  p.Service.Version,
			})
		}
		hosts = append(hosts, host)
	}
	return hosts, nil
}
