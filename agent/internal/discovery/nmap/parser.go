package nmap

import (
	"encoding/xml"
	"strconv"
	"strings"
)

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

// Parse converts Nmap XML (-oX -) into Hosts. Hosts that are not up and ports
// that are not open are dropped.
func Parse(data []byte) ([]Host, error) {
	var run xmlRun
	if err := xml.Unmarshal(data, &run); err != nil {
		return nil, err
	}
	hosts := make([]Host, 0, len(run.Hosts))
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
				host.MAC = normalizeMAC(a.Addr)
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

func normalizeMAC(mac string) string {
	return strings.ToLower(strings.ReplaceAll(mac, "-", ":"))
}
