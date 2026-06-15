package probes

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

// UPnPIGDProbe reads the InternetGatewayDevice WANCommonInterfaceConfig service
// when a LAN gateway advertises it. Generic IGD/NAT presence is device context
// only; classification hints are emitted only for explicit WAN access values.
type UPnPIGDProbe struct {
	funcs upnpIGDProbeFuncs
}

type UPnPIGDDeepProbe struct {
	funcs upnpIGDProbeFuncs
}

type upnpIGDProbeFuncs struct {
	ssdpSearch       func(context.Context, time.Duration) ([]string, error)
	fetchDescription func(context.Context, string) (*upnpIGDDescription, error)
	soapAction       func(context.Context, string, string, string) (map[string]string, error)
}

type upnpIGDDescription struct {
	URLBase string        `xml:"URLBase"`
	Device  upnpIGDDevice `xml:"device"`
}

type upnpIGDDevice struct {
	DeviceType       string           `xml:"deviceType"`
	FriendlyName     string           `xml:"friendlyName"`
	Manufacturer     string           `xml:"manufacturer"`
	ModelName        string           `xml:"modelName"`
	ModelNumber      string           `xml:"modelNumber"`
	ModelDescription string           `xml:"modelDescription"`
	Services         []upnpIGDService `xml:"serviceList>service"`
	Devices          []upnpIGDDevice  `xml:"deviceList>device"`
}

type upnpIGDService struct {
	ServiceType string `xml:"serviceType"`
	ServiceID   string `xml:"serviceId"`
	ControlURL  string `xml:"controlURL"`
	SCPDURL     string `xml:"SCPDURL"`
}

func (UPnPIGDProbe) Name() string { return "upnp_igd_probe" }

func (p UPnPIGDProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	return runUPnPIGD(ctx, p.Name(), p.withDefaults())
}

func (UPnPIGDDeepProbe) Name() string { return "upnp_igd_deep_probe" }

func (p UPnPIGDDeepProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	if in.Mode != models.ModeDeep {
		res := newResult(p.Name())
		res.Status = models.StatusSkipped
		return res, nil
	}
	return runUPnPIGD(ctx, p.Name(), p.withDefaults())
}

func runUPnPIGD(ctx context.Context, name string, f upnpIGDProbeFuncs) (*models.ProbeResult, error) {
	res := newResult(name)

	locations, err := f.ssdpSearch(ctx, 2*time.Second)
	if err != nil {
		return res, err
	}
	if len(locations) == 0 {
		res.Evidence["igd_wan_common_found"] = false
		res.Confidence = 0
		return res, nil
	}

	var services []string
	for _, loc := range locations {
		desc, err := f.fetchDescription(ctx, loc)
		if err != nil || desc == nil {
			continue
		}
		if !containsIGD(desc.Device) {
			continue
		}
		for _, svc := range allServices(desc.Device) {
			services = appendUnique(services, svc.ServiceType)
			if !strings.Contains(strings.ToLower(svc.ServiceType), "wancommoninterfaceconfig") {
				continue
			}
			res.Evidence["igd_wan_common_found"] = true
			res.Evidence["location"] = loc
			res.Evidence["cpe_services"] = services

			controlURL := resolveDeviceURL(loc, desc.URLBase, svc.ControlURL)
			values, err := f.soapAction(ctx, controlURL, svc.ServiceType, "GetCommonLinkProperties")
			if err != nil {
				res.Evidence["device_confidence"] = 0.45
				res.Evidence["access_confidence"] = 0.0
				res.Confidence = 0.45
				return res, nil
			}
			applyWANCommonEvidence(res, name, loc, values)
			return res, nil
		}
	}

	if len(services) > 0 {
		res.Evidence["cpe_services"] = services
		res.Evidence["device_confidence"] = 0.35
		res.Confidence = 0.35
	}
	if _, ok := res.Evidence["igd_wan_common_found"]; !ok {
		res.Evidence["igd_wan_common_found"] = false
	}
	return res, nil
}

func applyWANCommonEvidence(res *models.ProbeResult, source, loc string, values map[string]string) {
	accessType := firstMapValue(values, "NewWANAccessType", "WANAccessType")
	physical := firstMapValue(values, "NewPhysicalLinkStatus", "PhysicalLinkStatus")
	up := parseBitrate(firstMapValue(values, "NewLayer1UpstreamMaxBitRate", "Layer1UpstreamMaxBitRate"))
	down := parseBitrate(firstMapValue(values, "NewLayer1DownstreamMaxBitRate", "Layer1DownstreamMaxBitRate"))
	provider := firstMapValue(values, "NewWANAccessProvider", "WANAccessProvider")

	if accessType != "" {
		res.Evidence["wan_access_type"] = accessType
	}
	if provider != "" {
		res.Evidence["wan_access_provider"] = provider
	}
	if physical != "" {
		res.Evidence["physical_link_status"] = physical
	}
	if up > 0 {
		res.Evidence["layer1_upstream_bps"] = up
	}
	if down > 0 {
		res.Evidence["layer1_downstream_bps"] = down
	}

	detail := strings.Join([]string{accessType, physical, provider}, " ")
	hints := wanHintsFromText(detail)
	signals := []models.WANSignal{}
	if strings.TrimSpace(detail) != "" {
		signals = append(signals, wanSignal(source, locationHost(loc), "wan_common_interface", detail, "UPnP WANCommonInterfaceConfig", hints))
	}
	accessConf := accessConfidenceFromHints(hints, true)
	res.Evidence["wan_signals"] = signals
	res.Evidence["device_confidence"] = 0.55
	res.Evidence["access_confidence"] = accessConf
	res.Evidence["strong_access_evidence"] = len(hints) > 0
	res.Hints = hints
	if accessConf > 0 {
		res.Confidence = accessConf
	} else {
		res.Confidence = 0.55
	}
}

func (p UPnPIGDProbe) withDefaults() upnpIGDProbeFuncs {
	f := p.funcs
	if f.ssdpSearch == nil {
		f.ssdpSearch = ssdpSearch
	}
	if f.fetchDescription == nil {
		f.fetchDescription = fetchUPnPIGDDescription
	}
	if f.soapAction == nil {
		f.soapAction = soapUPnPAction
	}
	return f
}

func (p UPnPIGDDeepProbe) withDefaults() upnpIGDProbeFuncs {
	return UPnPIGDProbe{funcs: p.funcs}.withDefaults()
}

func fetchUPnPIGDDescription(ctx context.Context, loc string) (*upnpIGDDescription, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, loc, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("description status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 128*1024))
	if err != nil {
		return nil, err
	}
	var desc upnpIGDDescription
	if err := xml.Unmarshal(body, &desc); err != nil {
		return nil, fmt.Errorf("parse upnp igd description: %w", err)
	}
	return &desc, nil
}

func soapUPnPAction(ctx context.Context, controlURL, serviceType, action string) (map[string]string, error) {
	body := fmt.Sprintf(`<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
  <s:Body>
    <u:%s xmlns:u="%s"></u:%s>
  </s:Body>
</s:Envelope>`, action, serviceType, action)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, controlURL, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("SOAPAction", `"`+serviceType+"#"+action+`"`)
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("soap status %d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil, err
	}
	return parseSOAPValues(respBody), nil
}

func containsIGD(d upnpIGDDevice) bool {
	if strings.Contains(strings.ToLower(d.DeviceType), "internetgatewaydevice") {
		return true
	}
	for _, child := range d.Devices {
		if containsIGD(child) {
			return true
		}
	}
	return false
}

func allServices(d upnpIGDDevice) []upnpIGDService {
	out := append([]upnpIGDService{}, d.Services...)
	for _, child := range d.Devices {
		out = append(out, allServices(child)...)
	}
	return out
}

func resolveDeviceURL(location, urlBase, path string) string {
	if path == "" {
		return ""
	}
	if u, err := url.Parse(path); err == nil && u.IsAbs() {
		return path
	}
	base := strings.TrimSpace(urlBase)
	if base == "" {
		base = location
	}
	u, err := url.Parse(base)
	if err != nil {
		return path
	}
	ref, err := url.Parse(path)
	if err != nil {
		return path
	}
	return u.ResolveReference(ref).String()
}

func parseSOAPValues(body []byte) map[string]string {
	dec := xml.NewDecoder(bytes.NewReader(body))
	values := map[string]string{}
	var stack []string
	var chars []string
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			stack = append(stack, t.Name.Local)
			chars = append(chars, "")
		case xml.CharData:
			if len(chars) > 0 {
				chars[len(chars)-1] += string(t)
			}
		case xml.EndElement:
			if len(stack) == 0 {
				continue
			}
			name := stack[len(stack)-1]
			text := strings.TrimSpace(chars[len(chars)-1])
			stack = stack[:len(stack)-1]
			chars = chars[:len(chars)-1]
			if text != "" {
				values[name] = text
			}
		}
	}
	return values
}

func firstMapValue(m map[string]string, keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(m[k]); v != "" {
			return v
		}
	}
	return ""
}

func locationHost(loc string) string {
	u, err := url.Parse(loc)
	if err != nil {
		return ""
	}
	return u.Hostname()
}
