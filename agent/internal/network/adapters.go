// Package network gathers host-side network facts (adapters, default gateway,
// configured DNS servers, public IP). It is the cross-platform foundation the
// probes build on; OS-specific bits live in the *_windows.go / *_unix.go files.
package network

import "net"

// Adapter is a non-loopback network interface and its addresses.
type Adapter struct {
	Name  string   `json:"name"`
	MAC   string   `json:"mac"`
	Up    bool     `json:"up"`
	Addrs []string `json:"addrs"`
}

// Adapters returns the host's non-loopback interfaces. Uses only the standard
// library, so it is identical on Windows, Linux and macOS.
func Adapters() ([]Adapter, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var out []Adapter
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		a := Adapter{
			Name: ifc.Name,
			MAC:  ifc.HardwareAddr.String(),
			Up:   ifc.Flags&net.FlagUp != 0,
		}
		addrs, _ := ifc.Addrs()
		for _, ad := range addrs {
			a.Addrs = append(a.Addrs, ad.String())
		}
		out = append(out, a)
	}
	return out, nil
}
