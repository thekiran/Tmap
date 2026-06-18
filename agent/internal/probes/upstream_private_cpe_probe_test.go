package probes

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

func TestUpstreamPrivateCPEProbesOnlyPrivateCandidates(t *testing.T) {
	probed := map[string]bool{}
	p := UpstreamPrivateCPEProbe{funcs: upstreamPrivateCPEFuncs{
		gateway: func() (net.IP, error) { return net.ParseIP("192.168.31.1"), nil },
		traceroute: func(ctx context.Context, host string, maxHops int) ([]string, error) {
			return []string{"192.168.31.1", "192.168.1.1", "8.8.8.8"}, nil
		},
		checkTCP: func(ctx context.Context, ip, port string) bool {
			probed[ip] = true
			return false
		},
		fetch: func(ctx context.Context, method, endpoint string) (httpFingerprintV2Result, error) {
			probed[endpointIP(endpoint)] = true
			return httpFingerprintV2Result{}, errFakeHTTP
		},
		tlsInfo: func(ctx context.Context, ip, port string) (string, []string, string) {
			probed[ip] = true
			return "", nil, ""
		},
	}}

	res, err := p.Run(context.Background(), models.ScanInput{Mode: models.ModeDeep, Online: true})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if probed["8.8.8.8"] {
		t.Fatalf("public hop was probed: %v", probed)
	}
	if !probed["192.168.31.1"] || !probed["192.168.1.1"] {
		t.Fatalf("private candidates were not probed: %v", probed)
	}
	if chain := res.Evidence["gateway_chain"].([]string); len(chain) != 2 {
		t.Fatalf("gateway_chain = %v, want two private hops", chain)
	}
}

func TestFailedUpstreamCPEProbeDoesNotClassifyAccess(t *testing.T) {
	p := UpstreamPrivateCPEProbe{funcs: upstreamPrivateCPEFuncs{
		gateway: func() (net.IP, error) { return net.ParseIP("192.168.31.1"), nil },
		traceroute: func(ctx context.Context, host string, maxHops int) ([]string, error) {
			return []string{"192.168.31.1", "192.168.1.1", "95.15.180.1"}, nil
		},
		checkTCP: func(ctx context.Context, ip, port string) bool { return false },
		fetch: func(ctx context.Context, method, endpoint string) (httpFingerprintV2Result, error) {
			return httpFingerprintV2Result{}, errFakeHTTP
		},
		tlsInfo: func(ctx context.Context, ip, port string) (string, []string, string) {
			return "", nil, ""
		},
	}}

	res, err := p.Run(context.Background(), models.ScanInput{Mode: models.ModeDeep, Online: true})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Hints) != 0 {
		t.Fatalf("failed upstream CPE probe must not hint access type: %v", res.Hints)
	}
	devices := res.Evidence["gateway_devices"].([]models.GatewayDevice)
	if len(devices) != 2 {
		t.Fatalf("devices len = %d, want 2", len(devices))
	}
	upstream := devices[1]
	if upstream.ReachableState != models.ReachableUnknown && upstream.ReachableState != models.ReachableFalse {
		t.Fatalf("reachable_state = %q, want unknown or false", upstream.ReachableState)
	}
	if len(upstream.FailedAttempts) == 0 {
		t.Fatalf("failed attempts not recorded: %#v", upstream)
	}
	if upstream.AccessConfidence != 0 || len(upstream.AccessEvidence) != 0 || len(upstream.AccessHints) != 0 {
		t.Fatalf("failed probe produced access evidence: %#v", upstream)
	}
	for _, a := range upstream.FailedAttempts {
		if strings.Contains(a.Target, "95.15.") || strings.Contains(a.URL, "95.15.") {
			t.Fatalf("public target appeared in failed attempts: %#v", upstream.FailedAttempts)
		}
	}
}
