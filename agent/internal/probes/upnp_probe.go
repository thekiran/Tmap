package probes

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

// UPnPProbe discovers the gateway device over SSDP and reads its UPnP device
// description for manufacturer/model. On a client OS this is the single best
// source of modem evidence, since the modem model is what the fingerprint and
// rule engines key on. It is a LAN-only multicast exchange — no external
// service is contacted — so it runs even in offline mode.
type UPnPProbe struct{}

func (UPnPProbe) Name() string { return "upnp_probe" }

const ssdpAddr = "239.255.255.250:1900"

func (p UPnPProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())

	locations, err := ssdpSearch(ctx, 2*time.Second)
	if err != nil {
		return res, err
	}
	if len(locations) == 0 {
		res.Evidence["found"] = false
		res.Confidence = 0
		return res, nil
	}

	dev, loc, err := firstGatewayDevice(ctx, locations)
	if err != nil || dev == nil {
		res.Evidence["found"] = false
		res.Confidence = 0
		return res, nil
	}

	res.Evidence["found"] = true
	res.Evidence["location"] = loc
	res.Evidence["manufacturer"] = dev.Manufacturer
	res.Evidence["model"] = strings.TrimSpace(dev.ModelName + " " + dev.ModelNumber)
	res.Evidence["device_type"] = dev.DeviceType
	res.Evidence["friendly_name"] = dev.FriendlyName

	// router_model is the canonical "vendor + model" string the matcher/rules
	// key on; router_text adds the looser fields so keyword rules can also hit
	// model descriptions and friendly names.
	res.Evidence["router_model"] = strings.TrimSpace(dev.Manufacturer + " " + dev.ModelName + " " + dev.ModelNumber)
	res.Evidence["router_text"] = strings.Join([]string{
		dev.Manufacturer, dev.ModelName, dev.ModelNumber, dev.ModelDescription, dev.FriendlyName, dev.DeviceType,
	}, " ")
	res.Confidence = 0.7
	return res, nil
}

// ssdpSearch sends an M-SEARCH and collects the LOCATION URLs from the
// responses received within the wait window.
func ssdpSearch(ctx context.Context, wait time.Duration) ([]string, error) {
	raddr, err := net.ResolveUDPAddr("udp4", ssdpAddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	msearch := "M-SEARCH * HTTP/1.1\r\n" +
		"HOST: " + ssdpAddr + "\r\n" +
		"MAN: \"ssdp:discover\"\r\n" +
		"MX: 2\r\n" +
		"ST: upnp:rootdevice\r\n\r\n"
	if _, err := conn.WriteToUDP([]byte(msearch), raddr); err != nil {
		return nil, err
	}

	deadline := time.Now().Add(wait)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	conn.SetReadDeadline(deadline)

	seen := map[string]bool{}
	var locations []string
	buf := make([]byte, 2048)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			break // deadline reached or socket closed
		}
		if loc := parseLocation(string(buf[:n])); loc != "" && !seen[loc] {
			seen[loc] = true
			locations = append(locations, loc)
		}
	}
	return locations, nil
}

// parseLocation extracts the LOCATION header value from an SSDP response.
func parseLocation(resp string) string {
	for _, line := range strings.Split(resp, "\r\n") {
		if len(line) >= 9 && strings.EqualFold(line[:9], "LOCATION:") {
			return strings.TrimSpace(line[9:])
		}
	}
	return ""
}

// upnpDevice mirrors the subset of the UPnP device-description XML we use.
type upnpDevice struct {
	DeviceType       string `xml:"deviceType"`
	FriendlyName     string `xml:"friendlyName"`
	Manufacturer     string `xml:"manufacturer"`
	ModelName        string `xml:"modelName"`
	ModelNumber      string `xml:"modelNumber"`
	ModelDescription string `xml:"modelDescription"`
}

type upnpRoot struct {
	Device upnpDevice `xml:"device"`
}

// firstGatewayDevice fetches each description and returns the first device,
// preferring an InternetGatewayDevice when several are advertised.
func firstGatewayDevice(ctx context.Context, locations []string) (*upnpDevice, string, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	var fallback *upnpDevice
	var fallbackLoc string
	var lastErr error
	for _, loc := range locations {
		dev, err := fetchDescription(ctx, client, loc)
		if err != nil {
			lastErr = err
			continue
		}
		if strings.Contains(strings.ToLower(dev.DeviceType), "internetgatewaydevice") {
			return dev, loc, nil
		}
		if fallback == nil {
			fallback = dev
			fallbackLoc = loc
		}
	}
	if fallback != nil {
		return fallback, fallbackLoc, nil
	}
	return nil, "", lastErr
}

func fetchDescription(ctx context.Context, client *http.Client, loc string) (*upnpDevice, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, loc, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil, err
	}
	var root upnpRoot
	if err := xml.Unmarshal(body, &root); err != nil {
		return nil, fmt.Errorf("parse upnp description: %w", err)
	}
	return &root.Device, nil
}
