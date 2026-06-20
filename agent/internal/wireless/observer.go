package wireless

import (
	"context"
	"runtime"
	"sort"
	"strings"
	"time"
)

type CapabilityStatus string

const (
	CapabilitySupported   CapabilityStatus = "supported"
	CapabilityUnsupported CapabilityStatus = "unsupported"
	CapabilityPermission  CapabilityStatus = "permission_required"
	CapabilityUnknown     CapabilityStatus = "unknown"
)

// WirelessCapabilityDetector reports whether monitor/radiotap-style metadata
// is available. Unsupported is a normal result, especially on Windows adapters.
type WirelessCapabilityDetector interface {
	Detect(ctx context.Context, iface string) WirelessCapability
}

type WirelessCapability struct {
	Interface        string           `json:"interface,omitempty"`
	Status           CapabilityStatus `json:"status"`
	MonitorMode      bool             `json:"monitor_mode"`
	RadiotapMetadata bool             `json:"radiotap_metadata"`
	ChannelMetadata  bool             `json:"channel_metadata"`
	Reason           string           `json:"reason,omitempty"`
	Driver           string           `json:"driver,omitempty"`
	OS               string           `json:"os,omitempty"`
}

// DefaultCapabilityDetector is conservative and never assumes monitor support.
// Platform-specific implementations can replace it when permissions and driver
// capabilities are known.
type DefaultCapabilityDetector struct{}

func (DefaultCapabilityDetector) Detect(_ context.Context, iface string) WirelessCapability {
	return WirelessCapability{
		Interface: iface,
		Status:    CapabilityUnsupported,
		OS:        runtime.GOOS,
		Reason:    "wireless monitor metadata backend is not configured for this platform/adapter",
	}
}

type WirelessPassiveObserver interface {
	Observe(ctx context.Context, iface string, sink MetadataSink) error
}

type ChannelMetadataCollector interface {
	Collect(ctx context.Context, iface string) ([]ChannelMetadata, error)
}

type RadioFrameMetadataParser interface {
	Parse(frame RadioFrameMetadata) WirelessObservation
}

type MetadataSink interface {
	RecordWireless(obs WirelessObservation)
}

type ChannelMetadata struct {
	Interface string `json:"interface,omitempty"`
	Channel   int    `json:"channel,omitempty"`
	Frequency int    `json:"frequency,omitempty"`
	Band      string `json:"band,omitempty"`
	Noise     int    `json:"noise,omitempty"`
}

type RadioFrameMetadata struct {
	Interface      string
	FrameType      string // beacon, probe_response, association, reassociation, disassociation, data
	SSID           string
	BSSID          string
	APMAC          string
	StationMAC     string
	Channel        int
	Frequency      int
	RSSI           int
	Noise          int
	Security       string
	PHY            string
	SupportedRates []string
	Capabilities   []string
	WPS            bool
	MeshHint       bool
	RepeaterHint   bool
	Timestamp      time.Time
}

type WirelessObservation struct {
	Interface        string    `json:"interface,omitempty"`
	Source           string    `json:"source"`
	SSID             string    `json:"ssid,omitempty"`
	BSSID            string    `json:"bssid,omitempty"`
	APMAC            string    `json:"ap_mac,omitempty"`
	StationMAC       string    `json:"station_mac,omitempty"`
	Channel          int       `json:"channel,omitempty"`
	Frequency        int       `json:"frequency,omitempty"`
	Band             string    `json:"band,omitempty"`
	RSSI             int       `json:"rssi,omitempty"`
	Noise            int       `json:"noise,omitempty"`
	Security         string    `json:"security,omitempty"`
	PHY              string    `json:"phy,omitempty"`
	Capabilities     []string  `json:"capabilities,omitempty"`
	WPS              bool      `json:"wps_advertised"`
	MeshHint         bool      `json:"mesh_hint"`
	RepeaterHint     bool      `json:"repeater_hint"`
	Relationship     string    `json:"relationship,omitempty"` // proven, observed, weak-inferred
	Confidence       float64   `json:"confidence"`
	FirstSeen        time.Time `json:"first_seen,omitempty"`
	LastSeen         time.Time `json:"last_seen,omitempty"`
	ObservationCount int       `json:"observation_count"`
}

type MetadataParser struct{}

func (MetadataParser) Parse(frame RadioFrameMetadata) WirelessObservation {
	ts := frame.Timestamp
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	caps := append([]string{}, frame.Capabilities...)
	caps = append(caps, frame.SupportedRates...)
	sort.Strings(caps)
	obs := WirelessObservation{
		Interface:        frame.Interface,
		Source:           sourceForFrame(frame.FrameType),
		SSID:             strings.TrimSpace(frame.SSID),
		BSSID:            normalizeMAC(frame.BSSID),
		APMAC:            normalizeMAC(firstNonEmpty(frame.APMAC, frame.BSSID)),
		StationMAC:       normalizeMAC(frame.StationMAC),
		Channel:          frame.Channel,
		Frequency:        frame.Frequency,
		Band:             bandFor(frame.Frequency, frame.Channel),
		RSSI:             frame.RSSI,
		Noise:            frame.Noise,
		Security:         frame.Security,
		PHY:              frame.PHY,
		Capabilities:     unique(caps),
		WPS:              frame.WPS,
		MeshHint:         frame.MeshHint,
		RepeaterHint:     frame.RepeaterHint,
		Relationship:     relationshipForFrame(frame),
		Confidence:       confidenceForFrame(frame),
		FirstSeen:        ts,
		LastSeen:         ts,
		ObservationCount: 1,
	}
	return obs
}

func sourceForFrame(frameType string) string {
	switch strings.ToLower(frameType) {
	case "beacon":
		return "wireless_beacon"
	case "probe_response":
		return "wireless_probe_response"
	case "association", "reassociation":
		return "wireless_association"
	default:
		return "passive_wifi"
	}
}

func relationshipForFrame(frame RadioFrameMetadata) string {
	switch strings.ToLower(frame.FrameType) {
	case "association", "reassociation":
		if frame.StationMAC != "" && (frame.BSSID != "" || frame.APMAC != "") {
			return "proven"
		}
	case "data":
		if frame.StationMAC != "" && frame.BSSID != "" {
			return "observed"
		}
	}
	if frame.RSSI != 0 && (frame.Channel != 0 || frame.Frequency != 0) {
		return "weak-inferred"
	}
	return "observed"
}

func confidenceForFrame(frame RadioFrameMetadata) float64 {
	switch relationshipForFrame(frame) {
	case "proven":
		return 0.90
	case "observed":
		return 0.65
	case "weak-inferred":
		return 0.35
	default:
		return 0.25
	}
}

func bandFor(freq, channel int) string {
	switch {
	case freq >= 5925 || channel >= 1 && channel <= 233 && freq > 5900:
		return "6GHz"
	case freq >= 5000 || channel > 14:
		return "5GHz"
	case freq >= 2400 || channel >= 1 && channel <= 14:
		return "2.4GHz"
	default:
		return ""
	}
}

func normalizeMAC(raw string) string {
	return strings.ToLower(strings.TrimSpace(strings.ReplaceAll(raw, "-", ":")))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func unique(values []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
