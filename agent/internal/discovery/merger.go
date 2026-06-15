package discovery

import (
	"sort"
	"strconv"
	"strings"

	"github.com/thekiran/iad/internal/model"
	"github.com/thekiran/iad/internal/safety"
)

func MergeDevices(results []model.ProbeResult) []model.Device {
	byKey := map[string]*model.Device{}
	ipToKey := map[string]string{}
	macToKey := map[string]string{}

	for _, result := range results {
		for _, d := range result.Devices {
			if hasPublicOnlyIPs(d) {
				continue
			}
			key := mergeKey(d, ipToKey, macToKey)
			if key == "" {
				continue
			}
			dst := byKey[key]
			if dst == nil {
				cp := d
				cp.IPAddresses = uniqueSorted(cp.IPAddresses)
				cp.MACAddresses = uniqueSorted(normalizeMACs(cp.MACAddresses))
				cp.Hostnames = uniqueSorted(cp.Hostnames)
				cp.Evidence = mergeEvidence(nil, result.Evidence, d.Evidence)
				cp.Confidence = topologyDeviceConfidence(cp)
				byKey[key] = &cp
				dst = &cp
			} else {
				mergeIntoDevice(dst, d, result.Evidence)
			}
			for _, ip := range dst.IPAddresses {
				ipToKey[ip] = key
			}
			for _, mac := range dst.MACAddresses {
				macToKey[mac] = key
			}
		}
	}

	out := make([]model.Device, 0, len(byKey))
	for _, d := range byKey {
		*d = ClassifyDevice(*d, d.Evidence)
		out = append(out, *d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func mergeIntoDevice(dst *model.Device, src model.Device, resultEvidence []model.Evidence) {
	dst.IPAddresses = uniqueSorted(append(dst.IPAddresses, src.IPAddresses...))
	dst.MACAddresses = uniqueSorted(append(dst.MACAddresses, normalizeMACs(src.MACAddresses)...))
	dst.Hostnames = uniqueSorted(append(dst.Hostnames, src.Hostnames...))
	dst.Roles = uniqueRoles(append(dst.Roles, src.Roles...))
	dst.OpenPorts = mergePorts(dst.OpenPorts, src.OpenPorts)
	dst.Services = mergeServices(dst.Services, src.Services)
	dst.Evidence = mergeEvidence(dst.Evidence, resultEvidence, src.Evidence)
	if dst.Vendor == "" {
		dst.Vendor = src.Vendor
	}
	if dst.Manufacturer == "" {
		dst.Manufacturer = src.Manufacturer
	}
	if dst.Model == "" {
		dst.Model = src.Model
	}
	if dst.SerialNumber == "" {
		dst.SerialNumber = src.SerialNumber
	}
	if dst.OSGuess == "" {
		dst.OSGuess = src.OSGuess
	}
	if src.DeviceType != "" && (dst.DeviceType == "" || dst.DeviceType == model.DeviceTypeUnknown || src.Confidence > dst.Confidence) {
		dst.DeviceType = src.DeviceType
	}
	if src.LastSeen.After(dst.LastSeen) {
		dst.LastSeen = src.LastSeen
	}
	dst.Inferred = dst.Inferred || src.Inferred
	dst.Confidence = topologyDeviceConfidence(*dst)
}

func mergeKey(d model.Device, ipToKey, macToKey map[string]string) string {
	macs := normalizeMACs(d.MACAddresses)
	for _, mac := range macs {
		if key := macToKey[mac]; key != "" {
			return key
		}
	}
	for _, ip := range d.IPAddresses {
		if key := ipToKey[ip]; key != "" {
			return key
		}
	}
	if len(macs) > 0 {
		return "mac_" + macs[0]
	}
	if len(d.IPAddresses) > 0 {
		return "ip_" + d.IPAddresses[0]
	}
	if d.ID != "" {
		return d.ID
	}
	return ""
}

func hasPublicOnlyIPs(d model.Device) bool {
	if len(d.IPAddresses) == 0 {
		return false
	}
	for _, ip := range d.IPAddresses {
		if safety.IsPrivateIPString(ip) {
			return false
		}
	}
	return d.DeviceType != model.DeviceTypeISPHop
}

func normalizeMACs(in []string) []string {
	out := make([]string, 0, len(in))
	for _, mac := range in {
		mac = strings.ToLower(strings.TrimSpace(mac))
		if mac != "" {
			out = append(out, mac)
		}
	}
	return out
}

func mergeEvidence(existing []model.Evidence, groups ...[]model.Evidence) []model.Evidence {
	out := append([]model.Evidence{}, existing...)
	seen := map[string]bool{}
	for _, ev := range out {
		seen[evidenceKey(ev)] = true
	}
	for _, group := range groups {
		for _, ev := range group {
			key := evidenceKey(ev)
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, ev)
		}
	}
	return out
}

func evidenceKey(ev model.Evidence) string {
	return ev.Source + "|" + ev.Target + "|" + ev.Reason
}

func uniqueSorted(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func uniqueRoles(in []model.DeviceRole) []model.DeviceRole {
	seen := map[model.DeviceRole]bool{}
	var out []model.DeviceRole
	for _, v := range in {
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func mergePorts(a, b []model.PortInfo) []model.PortInfo {
	seen := map[string]bool{}
	var out []model.PortInfo
	for _, p := range append(a, b...) {
		key := p.Protocol + ":" + strconv.Itoa(p.Port)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Port < out[j].Port })
	return out
}

func mergeServices(a, b []model.ServiceInfo) []model.ServiceInfo {
	seen := map[string]bool{}
	var out []model.ServiceInfo
	for _, s := range append(a, b...) {
		key := s.Name + ":" + strconv.Itoa(s.Port) + ":" + s.Protocol
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, s)
	}
	return out
}
