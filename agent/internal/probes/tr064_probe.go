package probes

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/thekiran/iad/internal/network"
	"github.com/thekiran/iad/internal/system"
	"github.com/thekiran/iad/pkg/models"
)

// TR064Probe reads unauthenticated LAN-side CPE metadata from already observed
// private gateways. It never logs in, submits forms, tries credentials, scans a
// subnet, or contacts public/CGNAT addresses.
type TR064Probe struct {
	funcs tr064ProbeFuncs
}

type tr064ProbeFuncs struct {
	gateway          func() (net.IP, error)
	traceroute       func(context.Context, string, int) ([]string, error)
	checkTCP         func(context.Context, string, string) bool
	fetchDescription func(context.Context, string) (*upnpIGDDescription, error)
	soapAction       func(context.Context, string, string, string) (map[string]string, error)
}

func (TR064Probe) Name() string { return "tr064_probe" }

func (p TR064Probe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	f := p.withDefaults()

	candidates := p.candidates(ctx, in, f)
	res.Evidence["gateway_candidates"] = candidates
	if len(candidates) == 0 {
		res.Confidence = 0
		return res, nil
	}

	var cpeServices []string
	var signals []models.WANSignal
	var hints []string
	cpeKV := map[string]string{}
	bestDeviceConf := 0.0
	bestAccessConf := 0.0
	var firstIP string
	var manufacturer string
	var model string
	tr064Found := false

	for _, ip := range candidates {
		if !f.checkTCP(ctx, ip, "49000") {
			continue
		}
		tr064Found = true
		if firstIP == "" {
			firstIP = ip
		}
		bestDeviceConf = maxFloat(bestDeviceConf, 0.40)
		for _, path := range []string{"/tr64desc.xml", "/igddesc.xml"} {
			descURL := "http://" + net.JoinHostPort(ip, "49000") + path
			desc, err := f.fetchDescription(ctx, descURL)
			if err != nil {
				if isAuthRequiredError(err) {
					res.Evidence["auth_required"] = true
					bestDeviceConf = maxFloat(bestDeviceConf, 0.45)
					break
				}
				continue
			}
			if desc == nil {
				continue
			}
			bestDeviceConf = maxFloat(bestDeviceConf, 0.55)
			if manufacturer == "" {
				manufacturer = strings.TrimSpace(desc.Device.Manufacturer)
			}
			if model == "" {
				model = strings.TrimSpace(desc.Device.ModelName + " " + desc.Device.ModelNumber)
			}
			deviceText := strings.Join([]string{
				desc.Device.Manufacturer,
				desc.Device.ModelName,
				desc.Device.ModelNumber,
				desc.Device.ModelDescription,
				desc.Device.FriendlyName,
				desc.Device.DeviceType,
			}, " ")
			for _, h := range wanHintsFromText(deviceText) {
				hints = appendUnique(hints, h)
			}
			for _, svc := range allServices(desc.Device) {
				cpeServices = appendUnique(cpeServices, svc.ServiceType)
				controlURL := resolveDeviceURL(descURL, desc.URLBase, svc.ControlURL)
				values, err := p.queryService(ctx, f, controlURL, svc.ServiceType)
				if err != nil || len(values) == 0 {
					continue
				}
				text := valuesText(values)
				serviceHints := wanHintsFromText(svc.ServiceType + " " + text)
				for _, h := range serviceHints {
					hints = appendUnique(hints, h)
				}
				if len(serviceHints) > 0 {
					signals = append(signals, wanSignal(p.Name(), ip, "tr064_service", text, svc.ServiceType, serviceHints))
				}
				bestAccessConf = maxFloat(bestAccessConf, accessConfidenceFromHints(serviceHints, true))
				p.applyCommonValues(res, values)
				// Preserve the raw physical-layer fields (DSL profile/modulation, noise
				// margin, attenuation, rates, DOCSIS/PON values) for line-stat parsing.
				for k, v := range values {
					if strings.TrimSpace(v) == "" {
						continue
					}
					if _, ok := cpeKV[k]; !ok {
						cpeKV[k] = v
					}
				}
			}
			break
		}
	}

	if tr064Found {
		res.Evidence["tr064_found"] = true
		if firstIP != "" {
			res.Evidence["ip"] = firstIP
		}
		if manufacturer != "" {
			res.Evidence["manufacturer"] = manufacturer
		}
		if strings.TrimSpace(model) != "" {
			res.Evidence["model"] = strings.TrimSpace(model)
		}
		if len(cpeServices) > 0 {
			res.Evidence["cpe_services"] = cpeServices
		}
		for _, h := range wanHintsFromText(manufacturer + " " + model) {
			hints = appendUnique(hints, h)
		}
		bestAccessConf = maxFloat(bestAccessConf, accessConfidenceFromHints(hints, true))
	} else {
		res.Evidence["tr064_found"] = false
	}

	if len(cpeKV) > 0 {
		res.Evidence["cpe_kv"] = cpeKV
		res.Evidence["cpe_text"] = strings.TrimSpace(model + " " + valuesText(cpeKV))
	}

	res.Hints = hints
	res.Evidence["wan_signals"] = signals
	res.Evidence["device_confidence"] = bestDeviceConf
	res.Evidence["access_confidence"] = bestAccessConf
	res.Evidence["strong_access_evidence"] = bestAccessConf > 0
	res.Confidence = bestDeviceConf
	if bestAccessConf > res.Confidence {
		res.Confidence = bestAccessConf
	}
	return res, nil
}

func isAuthRequiredError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "401") || strings.Contains(msg, "403") || strings.Contains(msg, "auth")
}

func (p TR064Probe) withDefaults() tr064ProbeFuncs {
	f := p.funcs
	if f.gateway == nil {
		f.gateway = network.Gateway
	}
	if f.traceroute == nil {
		f.traceroute = system.Traceroute
	}
	if f.checkTCP == nil {
		f.checkTCP = checkTCPPort
	}
	if f.fetchDescription == nil {
		f.fetchDescription = fetchUPnPIGDDescription
	}
	if f.soapAction == nil {
		f.soapAction = soapUPnPAction
	}
	return f
}

func (p TR064Probe) candidates(ctx context.Context, in models.ScanInput, f tr064ProbeFuncs) []string {
	var out []string
	if gw, err := f.gateway(); err == nil && isRFC1918IPv4(gw.String()) {
		out = appendUnique(out, gw.String())
	}
	if in.Mode == models.ModeDeep && in.Online {
		if hops, err := f.traceroute(ctx, traceTarget, 6); err == nil {
			for _, h := range leadingPrivateHops(hops) {
				out = appendUnique(out, h)
			}
		}
	}
	return out
}

func (p TR064Probe) queryService(ctx context.Context, f tr064ProbeFuncs, controlURL, serviceType string) (map[string]string, error) {
	l := strings.ToLower(serviceType)
	switch {
	case strings.Contains(l, "wancommoninterfaceconfig"):
		return f.soapAction(ctx, controlURL, serviceType, "GetCommonLinkProperties")
	case strings.Contains(l, "wandslinterfaceconfig"):
		return f.soapAction(ctx, controlURL, serviceType, "GetInfo")
	case strings.Contains(l, "deviceinfo"):
		return f.soapAction(ctx, controlURL, serviceType, "GetInfo")
	default:
		return nil, nil
	}
}

func (p TR064Probe) applyCommonValues(res *models.ProbeResult, values map[string]string) {
	if v := firstMapValue(values, "NewWANAccessType", "WANAccessType"); v != "" {
		res.Evidence["wan_access_type"] = v
	}
	if v := firstMapValue(values, "NewPhysicalLinkStatus", "PhysicalLinkStatus", "NewStatus", "Status"); v != "" {
		res.Evidence["physical_link_status"] = v
	}
	if v := parseBitrate(firstMapValue(values, "NewLayer1UpstreamMaxBitRate", "Layer1UpstreamMaxBitRate", "NewUpstreamCurrRate", "NewUpstreamMaxRate")); v > 0 {
		res.Evidence["layer1_upstream_bps"] = v
	}
	if v := parseBitrate(firstMapValue(values, "NewLayer1DownstreamMaxBitRate", "Layer1DownstreamMaxBitRate", "NewDownstreamCurrRate", "NewDownstreamMaxRate")); v > 0 {
		res.Evidence["layer1_downstream_bps"] = v
	}
}

func valuesText(values map[string]string) string {
	var parts []string
	for k, v := range values {
		if strings.TrimSpace(v) == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s %s", k, v))
	}
	return strings.Join(parts, " ")
}

func maxFloat(a, b float64) float64 {
	if b > a {
		return b
	}
	return a
}
