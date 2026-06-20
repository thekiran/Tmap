package wireless

import (
	"context"
	"testing"
)

func TestDefaultCapabilityDetectorReportsUnsupportedCleanly(t *testing.T) {
	cap := (DefaultCapabilityDetector{}).Detect(context.Background(), "Wi-Fi")
	if cap.Status != CapabilityUnsupported {
		t.Fatalf("status = %s, want unsupported", cap.Status)
	}
	if cap.Interface != "Wi-Fi" || cap.Reason == "" {
		t.Fatalf("capability should include interface and reason: %#v", cap)
	}
}

func TestMetadataParserAssociationIsProven(t *testing.T) {
	obs := (MetadataParser{}).Parse(RadioFrameMetadata{
		Interface:  "wlan0",
		FrameType:  "association",
		SSID:       "lab",
		BSSID:      "AA-BB-CC-00-00-01",
		StationMAC: "AA-BB-CC-00-00-99",
		Frequency:  5180,
		PHY:        "802.11ax",
		Security:   "WPA3",
	})
	if obs.Relationship != "proven" {
		t.Fatalf("relationship = %q, want proven", obs.Relationship)
	}
	if obs.Confidence < 0.89 {
		t.Fatalf("confidence = %.2f, want high", obs.Confidence)
	}
	if obs.Band != "5GHz" || obs.Source != "wireless_association" {
		t.Fatalf("unexpected parsed metadata: %#v", obs)
	}
}

func TestMetadataParserCorrelationIsWeakInferred(t *testing.T) {
	obs := (MetadataParser{}).Parse(RadioFrameMetadata{
		Interface: "wlan0",
		FrameType: "beacon",
		BSSID:     "aa:bb:cc:00:00:01",
		Channel:   6,
		RSSI:      -55,
	})
	if obs.Relationship != "weak-inferred" {
		t.Fatalf("relationship = %q, want weak-inferred", obs.Relationship)
	}
	if obs.Confidence >= 0.5 {
		t.Fatalf("confidence = %.2f, want weak", obs.Confidence)
	}
}
