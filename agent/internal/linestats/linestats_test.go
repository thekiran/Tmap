package linestats

import (
	"strings"
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

func TestParseVDSL2_35b_SuperVectoring(t *testing.T) {
	// Realistic TR-064 WANDSLInterfaceConfig GetInfo (dB fields in 0.1 dB units).
	kv := map[string]string{
		"NewModulationType":        "VDSL2",
		"NewStatus":                "Up",
		"NewUpstreamNoiseMargin":   "62",  // 6.2 dB
		"NewDownstreamNoiseMargin": "61",  // 6.1 dB
		"NewUpstreamAttenuation":   "85",  // 8.5 dB
		"NewDownstreamAttenuation": "115", // 11.5 dB
		"NewUpstreamCurrRate":      "48000",
		"NewDownstreamCurrRate":    "294000",
		"NewUpstreamMaxRate":       "52000",
		"NewDownstreamMaxRate":     "310000",
		"NewDataPath":              "Interleaved",
		"X_AVM-DE_VDSLProfile":     "Profile 35b",
	}
	lp := Parse(kv, "AVM FRITZ!Box 7590 VDSL2 G.993.5 Super Vectoring", "tr064_probe")
	if lp == nil {
		t.Fatal("expected a line profile, got nil")
	}
	if lp.Medium != "dsl" || lp.DSL == nil {
		t.Fatalf("medium=%q dsl=%v, want dsl", lp.Medium, lp.DSL)
	}
	d := lp.DSL
	if d.Mode != "VDSL2" {
		t.Errorf("mode = %q, want VDSL2", d.Mode)
	}
	if d.Profile != "35b" {
		t.Errorf("profile = %q, want 35b", d.Profile)
	}
	if d.ProfileBandMHz < 35 || d.ProfileBandMHz > 36 {
		t.Errorf("band = %v MHz, want ~35.3", d.ProfileBandMHz)
	}
	if !d.Vectoring {
		t.Error("vectoring should be true for 35b / G.993.5")
	}
	if d.SNRMarginDownDB != 6.1 {
		t.Errorf("snr margin down = %v, want 6.1 (tenths dB conversion)", d.SNRMarginDownDB)
	}
	if d.AttenuationDownDB != 11.5 {
		t.Errorf("attenuation down = %v, want 11.5", d.AttenuationDownDB)
	}
	if d.SyncDownKbps != 294000 || d.SyncUpKbps != 48000 {
		t.Errorf("sync = %d/%d, want 294000/48000", d.SyncDownKbps, d.SyncUpKbps)
	}
	if d.AttainableDownKbps != 310000 {
		t.Errorf("attainable down = %d, want 310000", d.AttainableDownKbps)
	}
	if d.Path != "interleaved" {
		t.Errorf("path = %q, want interleaved", d.Path)
	}
	if !strings.Contains(lp.Subtype, "35b") || !strings.Contains(lp.Subtype, "Super Vectoring") {
		t.Errorf("subtype = %q, want it to mention 35b / Super Vectoring", lp.Subtype)
	}
	if got := AccessHints(lp); !equalStrings(got, []string{models.TypeVDSL2, models.TypeVDSL, models.TypeDSL}) {
		t.Errorf("access hints = %v, want [VDSL2 VDSL DSL]", got)
	}
	if lp.Confidence < 0.9 {
		t.Errorf("confidence = %v, want >= 0.9 with mode + stats", lp.Confidence)
	}
}

func TestParseVDSL2_17a(t *testing.T) {
	lp := Parse(map[string]string{"NewModulationType": "VDSL2"}, "VDSL2 Profile 17a", "tr064_probe")
	if lp == nil || lp.DSL == nil {
		t.Fatal("expected dsl line profile")
	}
	if lp.DSL.Profile != "17a" {
		t.Errorf("profile = %q, want 17a", lp.DSL.Profile)
	}
	if lp.DSL.Vectoring {
		t.Error("17a alone should not be flagged as vectoring")
	}
	if ProfileRank("17a") <= ProfileRank("12a") || ProfileRank("35b") <= ProfileRank("17a") {
		t.Error("profile ordering broken: want 12a < 17a < 35b")
	}
}

func TestParseADSL2Plus(t *testing.T) {
	lp := Parse(map[string]string{"NewModulationType": "ADSL2+"}, "G.992.5 ADSL2+", "tr064_probe")
	if lp == nil || lp.DSL == nil || lp.DSL.Mode != "ADSL2+" {
		t.Fatalf("want ADSL2+, got %+v", lp)
	}
	if got := AccessHints(lp); !equalStrings(got, []string{models.TypeADSL2, models.TypeADSL, models.TypeDSL}) {
		t.Errorf("access hints = %v, want [ADSL2+ ADSL DSL]", got)
	}
}

func TestParseGfast(t *testing.T) {
	lp := Parse(nil, "Sercomm G.fast modem G.9701", "tr064_probe")
	if lp == nil || lp.DSL == nil || lp.DSL.Mode != "G.fast" {
		t.Fatalf("want G.fast, got %+v", lp)
	}
	if got := AccessHints(lp); !equalStrings(got, []string{models.TypeGfast, models.TypeDSL}) {
		t.Errorf("access hints = %v, want [G.fast DSL]", got)
	}
}

func TestParseDOCSIS31(t *testing.T) {
	kv := map[string]string{
		"DownstreamChannels": "32",
		"UpstreamChannels":   "8",
		"DownstreamPower":    "3.5",
		"SNR":                "38.5",
	}
	lp := Parse(kv, "Technicolor CGA4233 DOCSIS 3.1 Cable Modem OFDM OFDMA", "tr064_probe")
	if lp == nil || lp.DOCSIS == nil {
		t.Fatalf("want docsis profile, got %+v", lp)
	}
	c := lp.DOCSIS
	if c.Version != "3.1" {
		t.Errorf("version = %q, want 3.1", c.Version)
	}
	if !c.OFDM || !c.OFDMA {
		t.Errorf("ofdm/ofdma = %v/%v, want true/true", c.OFDM, c.OFDMA)
	}
	if c.DownstreamChannels != 32 || c.UpstreamChannels != 8 {
		t.Errorf("channels = %d/%d, want 32/8", c.DownstreamChannels, c.UpstreamChannels)
	}
	if c.SNRMERdB != 38.5 {
		t.Errorf("mer = %v, want 38.5", c.SNRMERdB)
	}
	if got := AccessHints(lp); !equalStrings(got, []string{models.TypeDOCSIS, models.TypeCable}) {
		t.Errorf("access hints = %v, want [DOCSIS Cable]", got)
	}
}

func TestParseDOCSISVersionInferredFromOFDM(t *testing.T) {
	// No explicit version, but OFDM/OFDMA only exist in 3.1+.
	lp := Parse(nil, "cable modem OFDMA", "tr064_probe")
	if lp == nil || lp.DOCSIS == nil || lp.DOCSIS.Version != "3.1" {
		t.Fatalf("want inferred DOCSIS 3.1, got %+v", lp)
	}
}

func TestParseGPON_Optical(t *testing.T) {
	kv := map[string]string{
		"RxPower":  "-19.2",
		"TxPower":  "2.1",
		"ONTModel": "HG8245H",
	}
	lp := Parse(kv, "Huawei HG8245H GPON ONT", "tr064_probe")
	if lp == nil || lp.PON == nil {
		t.Fatalf("want pon profile, got %+v", lp)
	}
	if lp.PON.Type != "GPON" {
		t.Errorf("type = %q, want GPON", lp.PON.Type)
	}
	if lp.PON.RxPowerDBm != -19.2 || lp.PON.TxPowerDBm != 2.1 {
		t.Errorf("optical = %v/%v, want -19.2/2.1", lp.PON.RxPowerDBm, lp.PON.TxPowerDBm)
	}
	if got := AccessHints(lp); !equalStrings(got, []string{models.TypeGPON, models.TypeFTTH, models.TypeFiber}) {
		t.Errorf("access hints = %v, want [GPON FTTH Fiber]", got)
	}
}

func TestParseXGSPON(t *testing.T) {
	lp := Parse(nil, "Nokia XGS-PON ONT 10 Gbps", "tr064_probe")
	if lp == nil || lp.PON == nil || lp.PON.Type != "XGS-PON" {
		t.Fatalf("want XGS-PON, got %+v", lp)
	}
	if got := AccessHints(lp); got[0] != models.TypeXGSPON {
		t.Errorf("primary access hint = %q, want XGS-PON", got[0])
	}
}

func TestParseReturnsNilForNoCPEEvidence(t *testing.T) {
	if lp := Parse(nil, "", ""); lp != nil {
		t.Errorf("want nil for empty input, got %+v", lp)
	}
	if lp := Parse(nil, "the quick brown fox visited a website", ""); lp != nil {
		t.Errorf("want nil for non-CPE text, got %+v", lp)
	}
}

func TestProfileNotReadFromUnrelatedText(t *testing.T) {
	// "8a" appears but there is no VDSL/profile context, so it must not be read.
	lp := Parse(nil, "ticket #8a closed at building 12a", "")
	if lp != nil {
		t.Errorf("must not infer a DSL profile from unrelated text, got %+v", lp)
	}
}

func TestTenthsVsDecimalDB(t *testing.T) {
	// Integer (TR-064 tenths) vs explicit decimal (web UI) must both work.
	tenths := Parse(map[string]string{"NewModulationType": "VDSL2", "NewDownstreamNoiseMargin": "61"}, "", "")
	if tenths.DSL.SNRMarginDownDB != 6.1 {
		t.Errorf("tenths: got %v, want 6.1", tenths.DSL.SNRMarginDownDB)
	}
	decimal := Parse(map[string]string{"NewModulationType": "VDSL2", "NewDownstreamNoiseMargin": "6.3 dB"}, "", "")
	if decimal.DSL.SNRMarginDownDB != 6.3 {
		t.Errorf("decimal: got %v, want 6.3", decimal.DSL.SNRMarginDownDB)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
