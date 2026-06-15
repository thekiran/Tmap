package detection

import "testing"

func TestDetectGatewayChain_DoubleNAT(t *testing.T) {
	hops := []string{"192.168.31.1", "192.168.1.1", "95.15.180.1", "81.212.73.143"}
	res := detectGatewayChain(hops)
	if !res.DoubleNATPossible {
		t.Error("two leading private gateways must flag double NAT")
	}
	if res.PrivateHopCount != 2 {
		t.Errorf("private hop count = %d, want 2", res.PrivateHopCount)
	}
	if len(res.Chain) != 2 || res.Chain[0] != "192.168.31.1" || res.Chain[1] != "192.168.1.1" {
		t.Errorf("chain = %v, want [192.168.31.1 192.168.1.1]", res.Chain)
	}
}

func TestDetectGatewayChain_SingleGateway(t *testing.T) {
	hops := []string{"192.168.1.1", "100.64.0.1", "1.2.3.4"}
	res := detectGatewayChain(hops)
	if res.DoubleNATPossible {
		t.Error("a single private gateway must not flag double NAT")
	}
	if res.PrivateHopCount != 1 {
		t.Errorf("private hop count = %d, want 1", res.PrivateHopCount)
	}
}

func TestDetectGatewayChain_ToleratesTimeout(t *testing.T) {
	hops := []string{"192.168.0.1", "*", "192.168.1.1", "8.8.8.8"}
	res := detectGatewayChain(hops)
	if res.PrivateHopCount != 2 {
		t.Errorf("private hop count = %d, want 2 (timeout tolerated)", res.PrivateHopCount)
	}
}
