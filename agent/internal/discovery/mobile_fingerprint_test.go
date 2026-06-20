package discovery

import (
	"strings"
	"testing"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

func TestMobileFingerprintRequestedCases(t *testing.T) {
	now := time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	engine := NewMobileDeviceFingerprintEngine(func() time.Time { return now })

	tests := []struct {
		name  string
		in    MobileFingerprintInput
		check func(t *testing.T, fp models.MobileFingerprint)
	}{
		{
			name: "iPhone hostname classification",
			in: input(now,
				withHostname("KIRAN-iPhone"),
			),
			check: wantClassification(models.MobileClassificationPossibleIOS),
		},
		{
			name: "iPad hostname classification",
			in: input(now,
				withHostname("Living-Room-iPad"),
			),
			check: wantClassification(models.MobileClassificationPossibleIPadOS),
		},
		{
			name: "Android hostname classification",
			in: input(now,
				withHostname("Galaxy-S23"),
			),
			check: wantClassification(models.MobileClassificationPossibleAndroid),
		},
		{
			name: "Apple OUI scoring",
			in: input(now,
				withMAC("a4:83:e7:12:34:56"),
			),
			check: func(t *testing.T, fp models.MobileFingerprint) {
				t.Helper()
				if fp.IOSScore != 35 || fp.IPadScore != 35 {
					t.Fatalf("Apple OUI scores = ios %d ipad %d, want 35/35", fp.IOSScore, fp.IPadScore)
				}
				if strings.Contains(fp.Classification, "confirmed") || strings.Contains(fp.Classification, "probable") {
					t.Fatalf("Apple OUI alone classified too strongly: %s", fp.Classification)
				}
			},
		},
		{
			name: "Android vendor OUI scoring",
			in: input(now,
				withMAC("ec:1f:72:12:34:56"),
			),
			check: func(t *testing.T, fp models.MobileFingerprint) {
				t.Helper()
				if fp.AndroidScore != 35 {
					t.Fatalf("Android OUI score = %d, want 35", fp.AndroidScore)
				}
				if strings.Contains(fp.Classification, "confirmed") || strings.Contains(fp.Classification, "probable") {
					t.Fatalf("Android OUI alone classified too strongly: %s", fp.Classification)
				}
			},
		},
		{
			name: "randomized MAC confidence downgrade",
			in: input(now,
				withMAC("02:00:00:12:34:56"),
				withOUIVendor("Apple"),
			),
			check: func(t *testing.T, fp models.MobileFingerprint) {
				t.Helper()
				if fp.IOSScore >= 35 {
					t.Fatalf("randomized Apple MAC score = %d, want downgraded below clear OUI score", fp.IOSScore)
				}
				if !containsWarning(fp.Warnings, "randomization") {
					t.Fatalf("warnings = %#v, want MAC randomization warning", fp.Warnings)
				}
			},
		},
		{
			name: "UDP 5353 alone not classified as iPhone",
			in: input(now,
				withService(5353, "udp", "mdns"),
			),
			check: func(t *testing.T, fp models.MobileFingerprint) {
				t.Helper()
				if fp.Classification != models.MobileClassificationUnknownDevice {
					t.Fatalf("classification = %s, want unknown_device", fp.Classification)
				}
				if fp.IOSScore != 0 || fp.AndroidScore != 0 {
					t.Fatalf("UDP 5353 scored OS evidence: ios=%d android=%d", fp.IOSScore, fp.AndroidScore)
				}
			},
		},
		{
			name: "TCP 443 alone not classified",
			in: input(now,
				withService(443, "tcp", "https"),
			),
			check: wantClassification(models.MobileClassificationUnknownDevice),
		},
		{
			name: "Apple DNS hints alone only possible",
			in: input(now,
				withDNS("captive.apple.com"),
				withDNS("p42-escrowproxy.icloud.com"),
			),
			check: func(t *testing.T, fp models.MobileFingerprint) {
				t.Helper()
				if fp.Classification != models.MobileClassificationPossibleIOS {
					t.Fatalf("classification = %s, want possible_ios", fp.Classification)
				}
				if strings.Contains(fp.Classification, "confirmed") {
					t.Fatalf("DNS alone must not confirm: %s", fp.Classification)
				}
			},
		},
		{
			name: "Android DNS hints alone only possible",
			in: input(now,
				withDNS("android.clients.google.com"),
				withDNS("mtalk.google.com"),
			),
			check: wantClassification(models.MobileClassificationPossibleAndroid),
		},
		{
			name: "conflicting Apple Android evidence",
			in: input(now,
				withHostname("KIRAN-iPhone"),
				withOUIVendor("Samsung"),
			),
			check: func(t *testing.T, fp models.MobileFingerprint) {
				t.Helper()
				if fp.Classification != models.MobileClassificationConflict {
					t.Fatalf("classification = %s, want conflict", fp.Classification)
				}
				if len(fp.Conflicts) == 0 {
					t.Fatal("expected conflict items")
				}
			},
		},
		{
			name: "evidence explanation output",
			in: input(now,
				withHostname("Pixel-8"),
			),
			check: func(t *testing.T, fp models.MobileFingerprint) {
				t.Helper()
				if len(fp.Evidence) == 0 || fp.Evidence[0].Explanation == "" {
					t.Fatalf("evidence = %#v, want explanations", fp.Evidence)
				}
				if fp.WhyThisClassification == "" || fp.WhyNotCertain == "" {
					t.Fatalf("missing explanation fields: %#v", fp)
				}
			},
		},
		{
			name:  "unknown device fallback",
			in:    input(now),
			check: wantClassification(models.MobileClassificationUnknownDevice),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, engine.Fingerprint(tt.in))
		})
	}
}

type inputOption func(*MobileFingerprintInput)

func input(now time.Time, opts ...inputOption) MobileFingerprintInput {
	in := MobileFingerprintInput{DeviceID: "dev-192.168.1.20", Timestamp: now}
	for _, opt := range opts {
		opt(&in)
	}
	return in
}

func withHostname(value string) inputOption {
	return func(in *MobileFingerprintInput) {
		in.Hostnames = append(in.Hostnames, MobileObservedValue{Value: value, Source: "test_hostname", Timestamp: in.Timestamp})
	}
}

func withMAC(value string) inputOption {
	return func(in *MobileFingerprintInput) {
		in.MACAddresses = append(in.MACAddresses, value)
	}
}

func withOUIVendor(value string) inputOption {
	return func(in *MobileFingerprintInput) {
		in.OUIVendors = append(in.OUIVendors, value)
	}
}

func withDNS(value string) inputOption {
	return func(in *MobileFingerprintInput) {
		in.DNSQueries = append(in.DNSQueries, MobileObservedValue{Value: value, Source: "passive_dns_test", Timestamp: in.Timestamp})
	}
}

func withService(port int, protocol, name string) inputOption {
	return func(in *MobileFingerprintInput) {
		in.Services = append(in.Services, MobileServiceObservation{Port: port, Protocol: protocol, Name: name, Source: "service_test", Timestamp: in.Timestamp})
	}
}

func wantClassification(want string) func(*testing.T, models.MobileFingerprint) {
	return func(t *testing.T, fp models.MobileFingerprint) {
		t.Helper()
		if fp.Classification != want {
			t.Fatalf("classification = %s, want %s", fp.Classification, want)
		}
	}
}

func containsWarning(warnings []string, needle string) bool {
	for _, warning := range warnings {
		if strings.Contains(strings.ToLower(warning), strings.ToLower(needle)) {
			return true
		}
	}
	return false
}
