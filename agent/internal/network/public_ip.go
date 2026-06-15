package network

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// ipServices are plain-text "what is my IP" endpoints, tried in order. They are
// only ever contacted when the caller has opted into online probes.
var ipServices = []string{
	"https://api.ipify.org",
	"https://ifconfig.me/ip",
	"https://icanhazip.com",
}

// PublicIP fetches the host's public IP from an external echo service. This is
// an outbound call by nature, so callers must gate it behind ScanInput.Online.
func PublicIP(ctx context.Context) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	var lastErr error
	for _, url := range ipServices {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			lastErr = err
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		ip := strings.TrimSpace(string(body))
		if net.ParseIP(ip) != nil {
			return ip, nil
		}
	}
	if lastErr == nil {
		lastErr = io.EOF
	}
	return "", lastErr
}

// IsCGNAT reports whether ip falls in the carrier-grade NAT range
// 100.64.0.0/10 (RFC 6598). A public-facing CGNAT address is a strong hint of
// mobile/FWA/satellite access, where the ISP shares one public IP across many
// subscribers.
func IsCGNAT(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	_, cgnat, _ := net.ParseCIDR("100.64.0.0/10")
	return cgnat.Contains(parsed)
}
