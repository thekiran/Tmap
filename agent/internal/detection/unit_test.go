package detection

import "testing"

func TestNormalizeModel(t *testing.T) {
	cases := map[string]string{
		"  TP-LINK_VR400  ":     "TP-LINK VR400",
		"Archer VR400 v3":       "Archer VR400 v3",
		"Huawei   HG8245H":      "Huawei HG8245H",
		"ZXHN_H168A":            "ZXHN H168A",
	}
	for in, want := range cases {
		if got := NormalizeModel(in); got != want {
			t.Errorf("NormalizeModel(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestComputeConfidence checks the intended ordering properties rather than
// exact values: a dominant, corroborated, fingerprinted verdict should be more
// confident than a weak, contested one.
func TestComputeConfidence(t *testing.T) {
	strong := computeConfidence(map[string]float64{"VDSL": 0.9, "ADSL": 0.1}, 4, true)
	weak := computeConfidence(map[string]float64{"VDSL": 0.4, "Fiber": 0.38}, 1, false)

	if !(strong > weak) {
		t.Errorf("expected strong (%v) > weak (%v)", strong, weak)
	}
	if strong <= 0 || strong > 1 {
		t.Errorf("strong confidence %v out of range", strong)
	}
	if got := computeConfidence(nil, 0, false); got != 0 {
		t.Errorf("empty scores confidence = %v, want 0", got)
	}
}
