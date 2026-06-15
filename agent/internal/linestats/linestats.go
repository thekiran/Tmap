// Package linestats parses fine-grained physical-layer line properties from
// authorized CPE telemetry (TR-064 / SNMP / UPnP-IGD / vendor APIs) into a
// normalized models.LineProfile.
//
// This is the deepest evidence the system can obtain. Where the rest of the
// engine *infers* an access medium from behaviour, this package *reads* what the
// line negotiated: the VDSL2 profile (8a/12a/17a/35b and vectoring), DSL noise
// margin / attenuation / attainable & sync rates, DOCSIS version + OFDM/OFDMA +
// channels + power/MER, and PON type + optical Rx/Tx power.
//
// Every function here is pure (no I/O), so it is fully unit-testable from canned
// CPE key/value maps and free text — exactly the shape probes already emit.
package linestats

import (
	"regexp"
	"strings"

	"github.com/thekiran/iad/pkg/models"
)

// vdsl2Profile describes a VDSL2 band-plan profile. The band is the nominal
// upper frequency the profile uses; higher band ⇒ higher rate potential but more
// sensitivity to loop length and copper quality (35b is the highest, and in
// practice always paired with (super) vectoring).
type vdsl2Profile struct {
	bandMHz        float64
	superVectoring bool
}

// The recognised VDSL2 profiles (ITU-T G.993.2 Table). 35b is "Super Vectoring".
var vdsl2Profiles = map[string]vdsl2Profile{
	"8a":  {8.5, false},
	"8b":  {8.5, false},
	"8c":  {8.5, false},
	"8d":  {8.5, false},
	"12a": {12.0, false},
	"12b": {12.0, false},
	"17a": {17.664, false},
	"30a": {30.0, false},
	"35b": {35.328, true},
}

// profileOrder lets callers/tests reason about "stronger" profiles.
var profileOrder = map[string]int{
	"8a": 1, "8b": 2, "8c": 3, "8d": 4,
	"12a": 5, "12b": 6, "17a": 7, "30a": 8, "35b": 9,
}

// ProfileRank returns a comparable rank for a VDSL2 profile (0 if unknown), so
// 8a < 12a < 17a < 35b.
func ProfileRank(profile string) int { return profileOrder[strings.ToLower(profile)] }

var (
	// profile token only after the word "profile" (strong, unambiguous).
	reProfileLabeled = regexp.MustCompile(`(?i)profile[^a-z0-9]{0,6}(8a|8b|8c|8d|12a|12b|17a|30a|35b)\b`)
	// bare profile token (only trusted when VDSL context is otherwise present).
	reProfileBare = regexp.MustCompile(`(?i)\b(8a|8b|8c|8d|12a|12b|17a|30a|35b)\b`)
	reFloat       = regexp.MustCompile(`-?\d+(?:\.\d+)?`)
	reIntOnly     = regexp.MustCompile(`^-?\d+$`)
)

// Parse reads a normalized LineProfile from CPE key/value pairs and free text.
// kv holds structured CPE fields (e.g. TR-064 "NewDownstreamNoiseMargin" or SNMP
// labels); text is any free-form CPE/device text. Either may be empty. It returns
// nil when nothing physical-layer-relevant is present.
func Parse(kv map[string]string, text, source string) *models.LineProfile {
	norm := normalizeKV(kv)
	// Fold kv content into the search text so medium/technology detectors see both.
	hay := strings.ToLower(strings.TrimSpace(text + " " + kvText(kv)))
	if hay == "" && len(norm) == 0 {
		return nil
	}

	switch medium := detectMedium(hay, norm); medium {
	case "dsl":
		return parseDSL(norm, hay, source)
	case "docsis":
		return parseDOCSIS(norm, hay, source)
	case "pon":
		return parsePON(norm, hay, source)
	default:
		return nil
	}
}

// reONT matches a standalone ONT/ONU token (not "front", "control", ...).
var reONT = regexp.MustCompile(`(?i)\b(ont|onu)\b`)

// detectMedium decides which physical medium the evidence describes. Order
// matters: explicit PON/DOCSIS markers win over generic DSL keywords.
func detectMedium(hay string, norm map[string]string) string {
	switch {
	case containsAny(hay, "xgs-pon", "xgspon", "xg-pon", "xgpon", "gpon", "epon", "10g-epon",
		"gigabit-capable passive optical") || reONT.MatchString(hay):
		return "pon"
	case containsAny(hay, "docsis", "eurodocsis", "cable modem", "cablemodem", "ofdma", "cmts", " mer "):
		return "docsis"
	case containsAny(hay, "vdsl", "adsl", "g.fast", "gfast", "g.993", "g993", "g.992", "g992",
		"g.9701", "g9701", "shdsl", " ptm", " atm", "noisemargin", "noise margin", "snr margin",
		"lineattenuation", "line attenuation", "dsl"):
		return "dsl"
	case hasAnyKey(norm, "newupstreamnoisemargin", "newdownstreamnoisemargin", "newmodulationtype"):
		return "dsl"
	default:
		return ""
	}
}

// ---------------------------------------------------------------------------
// DSL
// ---------------------------------------------------------------------------

func parseDSL(norm map[string]string, hay, source string) *models.LineProfile {
	d := &models.DSLProfile{}

	d.Standard = detectDSLStandard(hay)
	d.Profile, d.ProfileBandMHz, d.Vectoring = detectVDSL2Profile(hay)
	d.Mode = detectDSLMode(hay, d.Standard, d.Profile)
	if strings.Contains(hay, "vector") || d.Standard == "G.993.5" {
		d.Vectoring = true
	}
	if a := detectAnnex(hay); a != "" {
		d.Annex = a
	}

	// dB-valued line stats (TR-064 / ADSL-MIB encode these in 0.1 dB).
	d.SNRMarginDownDB = lookupTenthsDB(norm, "newdownstreamnoisemargin", "downstreamnoisemargin", "downstreamsnrmargin", "dssnrmargin")
	d.SNRMarginUpDB = lookupTenthsDB(norm, "newupstreamnoisemargin", "upstreamnoisemargin", "upstreamsnrmargin", "ussnrmargin")
	d.AttenuationDownDB = lookupTenthsDB(norm, "newdownstreamattenuation", "downstreamattenuation", "dsattenuation", "lineattenuationdown")
	d.AttenuationUpDB = lookupTenthsDB(norm, "newupstreamattenuation", "upstreamattenuation", "usattenuation", "lineattenuationup")

	// rates in kbit/s.
	d.SyncDownKbps = lookupKbps(norm, "newdownstreamcurrrate", "downstreamcurrrate", "currentdownstreamrate", "downstreamsyncrate")
	d.SyncUpKbps = lookupKbps(norm, "newupstreamcurrrate", "upstreamcurrrate", "currentupstreamrate", "upstreamsyncrate")
	d.AttainableDownKbps = lookupKbps(norm, "newdownstreammaxrate", "downstreammaxrate", "attainabledownstreamrate", "maxdownstreamrate")
	d.AttainableUpKbps = lookupKbps(norm, "newupstreammaxrate", "upstreammaxrate", "attainableupstreamrate", "maxupstreamrate")

	if p := strings.ToLower(lookupKV(norm, "newdatapath", "datapath", "path")); p != "" {
		if strings.Contains(p, "interleav") {
			d.Path = "interleaved"
		} else if strings.Contains(p, "fast") {
			d.Path = "fast"
		}
	}
	d.InterleaveDepth = lookupKV(norm, "newinterleavedepth", "interleavedepth")

	lp := &models.LineProfile{Medium: "dsl", DSL: d, Source: source}
	lp.Technology = dslTechnology(d)
	lp.Subtype = dslSubtype(d)
	lp.Confidence = dslConfidence(d)
	lp.Notes = dslNotes(d)
	if lp.Technology == "" && !hasDSLStats(d) {
		return nil
	}
	return lp
}

func detectDSLStandard(hay string) string {
	switch {
	case containsAny(hay, "g.9701", "g9701", "g.fast", "gfast"):
		return "G.9701"
	case containsAny(hay, "g.993.5", "g9935"):
		return "G.993.5"
	case containsAny(hay, "g.993.2", "g9932", "vdsl2"):
		return "G.993.2"
	case containsAny(hay, "g.992.5", "g9925", "adsl2+", "adsl2plus"):
		return "G.992.5"
	case containsAny(hay, "g.992.3", "g9923"):
		return "G.992.3"
	case containsAny(hay, "g.992.1", "g9921"):
		return "G.992.1"
	case containsAny(hay, "g.991.2", "g9912", "shdsl"):
		return "G.991.2"
	}
	return ""
}

// detectVDSL2Profile finds the band-plan profile token. A label-qualified token
// ("Profile 35b") is always trusted; a bare token is trusted only when VDSL
// context is present so we never read "8a" out of an unrelated string.
func detectVDSL2Profile(hay string) (profile string, bandMHz float64, vectoring bool) {
	if m := reProfileLabeled.FindStringSubmatch(hay); m != nil {
		profile = strings.ToLower(m[1])
	} else if containsAny(hay, "vdsl", "g.993", "g993") {
		if m := reProfileBare.FindStringSubmatch(hay); m != nil {
			profile = strings.ToLower(m[1])
		}
	}
	if profile == "" {
		// "super vectoring" strongly implies 35b but we do not fabricate a token.
		return "", 0, strings.Contains(hay, "super vectoring") || strings.Contains(hay, "supervectoring")
	}
	p := vdsl2Profiles[profile]
	return profile, p.bandMHz, p.superVectoring
}

func detectDSLMode(hay, standard, profile string) string {
	switch standard {
	case "G.9701":
		return "G.fast"
	case "G.993.5", "G.993.2":
		return "VDSL2"
	case "G.992.5":
		return "ADSL2+"
	case "G.992.3":
		return "ADSL2"
	case "G.992.1":
		return "ADSL"
	case "G.991.2":
		return "SHDSL"
	}
	if profile != "" { // any 8a..35b token is a VDSL2 profile
		return "VDSL2"
	}
	switch {
	case containsAny(hay, "g.fast", "gfast"):
		return "G.fast"
	case containsAny(hay, "vdsl2"):
		return "VDSL2"
	case containsAny(hay, "adsl2+", "adsl2plus"):
		return "ADSL2+"
	case containsAny(hay, "adsl2"):
		return "ADSL2"
	case containsAny(hay, "shdsl"):
		return "SHDSL"
	case containsAny(hay, "sdsl"):
		return "SDSL"
	case containsAny(hay, "vdsl"):
		return "VDSL"
	case containsAny(hay, "adsl"):
		return "ADSL"
	}
	return ""
}

func detectAnnex(hay string) string {
	switch {
	case containsAny(hay, "annex b", "annexb"):
		return "Annex B"
	case containsAny(hay, "annex a", "annexa"):
		return "Annex A"
	case containsAny(hay, "annex j", "annexj"):
		return "Annex J"
	case containsAny(hay, "annex m", "annexm"):
		return "Annex M"
	}
	return ""
}

func dslTechnology(d *models.DSLProfile) string {
	if d.Mode == "VDSL2" && d.Vectoring {
		return "VDSL2 (vectoring)"
	}
	return d.Mode
}

func dslSubtype(d *models.DSLProfile) string {
	if d.Mode == "" {
		return ""
	}
	if d.Mode == "VDSL2" && d.Profile != "" {
		s := "VDSL2 Profile " + d.Profile
		if d.Profile == "35b" {
			return s + " (Super Vectoring)"
		}
		if d.Vectoring {
			return s + " (Vectoring)"
		}
		return s
	}
	return d.Mode
}

func dslNotes(d *models.DSLProfile) []string {
	var notes []string
	if d.Profile == "35b" {
		notes = append(notes, "VDSL2 profile 35b uses ~35 MHz (Super Vectoring): highest rate potential, most sensitive to loop length/copper quality.")
	} else if d.Profile != "" {
		notes = append(notes, "VDSL2 profile "+d.Profile+" denotes the band plan (a working mode), not the delivered speed; real rate depends on loop length, copper quality and SNR.")
	}
	if d.Mode == "VDSL2" && d.Vectoring && d.Profile != "35b" {
		notes = append(notes, "Vectoring (G.993.5) is active: crosstalk cancellation raises attainable rate.")
	}
	return notes
}

func dslConfidence(d *models.DSLProfile) float64 {
	switch {
	case d.Mode != "" && hasDSLStats(d):
		return 0.92
	case d.Mode != "" && d.Profile != "":
		return 0.88
	case d.Mode != "":
		return 0.75
	case hasDSLStats(d):
		return 0.6
	}
	return 0
}

func hasDSLStats(d *models.DSLProfile) bool {
	return d.SNRMarginDownDB != 0 || d.SNRMarginUpDB != 0 || d.AttenuationDownDB != 0 ||
		d.AttenuationUpDB != 0 || d.SyncDownKbps != 0 || d.SyncUpKbps != 0 ||
		d.AttainableDownKbps != 0 || d.AttainableUpKbps != 0
}

// ---------------------------------------------------------------------------
// DOCSIS
// ---------------------------------------------------------------------------

func parseDOCSIS(norm map[string]string, hay, source string) *models.LineProfile {
	c := &models.DOCSISProfile{}
	c.OFDM = strings.Contains(hay, "ofdm")
	c.OFDMA = strings.Contains(hay, "ofdma")
	c.Version = detectDOCSISVersion(hay, c)
	c.DownstreamChannels = int(lookupNumber(norm, "downstreamchannels", "numberofdownstreamchannels", "dschannels", "bondeddownstreamchannels"))
	c.UpstreamChannels = int(lookupNumber(norm, "upstreamchannels", "numberofupstreamchannels", "uschannels", "bondedupstreamchannels"))
	c.DownstreamPowerDBmV = lookupNumber(norm, "downstreampower", "dspower", "downstreampowerlevel", "rxpower")
	c.UpstreamPowerDBmV = lookupNumber(norm, "upstreampower", "uspower", "upstreampowerlevel", "txpower")
	c.SNRMERdB = lookupNumber(norm, "snr", "mer", "downstreamsnr", "downstreammer", "rxmer")

	lp := &models.LineProfile{Medium: "docsis", DOCSIS: c, Source: source}
	lp.Technology = "DOCSIS"
	if c.Version != "" {
		lp.Technology = "DOCSIS " + c.Version
	}
	lp.Subtype = lp.Technology
	if c.OFDM || c.OFDMA {
		lp.Subtype = lp.Technology + " (OFDM/OFDMA)"
	}
	lp.Confidence = docsisConfidence(c)
	return lp
}

func detectDOCSISVersion(hay string, c *models.DOCSISProfile) string {
	switch {
	case containsAny(hay, "docsis 4", "docsis4", "docsis 4.0", "full duplex", "full-duplex", "fdx", "extended spectrum"):
		return "4.0"
	case containsAny(hay, "docsis 3.1", "docsis3.1", "d3.1"):
		return "3.1"
	case containsAny(hay, "docsis 3.0", "docsis3.0", "d3.0"):
		return "3.0"
	case containsAny(hay, "docsis 2.0", "docsis2.0"):
		return "2.0"
	}
	// Infer from PHY: OFDM/OFDMA only exist in 3.1+.
	if c.OFDM || c.OFDMA {
		return "3.1"
	}
	return ""
}

func docsisConfidence(c *models.DOCSISProfile) float64 {
	switch {
	case c.Version != "" && (c.DownstreamChannels > 0 || c.SNRMERdB != 0 || c.DownstreamPowerDBmV != 0):
		return 0.9
	case c.Version != "":
		return 0.8
	default:
		return 0.6
	}
}

// ---------------------------------------------------------------------------
// PON
// ---------------------------------------------------------------------------

func parsePON(norm map[string]string, hay, source string) *models.LineProfile {
	p := &models.PONProfile{}
	p.Type = detectPONType(hay)
	p.ONTModel = lookupKV(norm, "ontmodel", "onumodel", "opticalmodulemodel", "model")
	p.RxPowerDBm = lookupNumber(norm, "rxpower", "opticalrxpower", "rxopticalpower", "opticalsignallevel", "rxopticalsignal")
	p.TxPowerDBm = lookupNumber(norm, "txpower", "opticaltxpower", "txopticalpower", "txopticalsignal")

	lp := &models.LineProfile{Medium: "pon", PON: p, Source: source}
	tech := p.Type
	if tech == "" {
		tech = "PON"
	}
	lp.Technology = tech
	lp.Subtype = tech
	lp.Confidence = ponConfidence(p)
	if p.Type == "" && p.RxPowerDBm == 0 && p.TxPowerDBm == 0 && p.ONTModel == "" {
		return nil
	}
	return lp
}

func detectPONType(hay string) string {
	switch {
	case containsAny(hay, "xgs-pon", "xgspon"):
		return "XGS-PON"
	case containsAny(hay, "xg-pon", "xgpon"):
		return "XG-PON"
	case containsAny(hay, "10g-epon", "10gepon"):
		return "10G-EPON"
	case containsAny(hay, "gpon"):
		return "GPON"
	case containsAny(hay, "epon"):
		return "EPON"
	}
	return ""
}

func ponConfidence(p *models.PONProfile) float64 {
	switch {
	case p.Type != "" && (p.RxPowerDBm != 0 || p.TxPowerDBm != 0):
		return 0.92
	case p.Type != "":
		return 0.85
	case p.RxPowerDBm != 0 || p.TxPowerDBm != 0:
		return 0.7
	default:
		return 0.5
	}
}

// ---------------------------------------------------------------------------
// Access-type mapping & summary (used by the detection engine)
// ---------------------------------------------------------------------------

// AccessHints returns the scoreable access-type keys (models.Type* constants)
// implied by a line profile, strongest/most-specific first. These are fed to the
// engine as strong physical-layer evidence so a verdict can commit out of
// Unknown. Returns nil for a nil/empty profile.
func AccessHints(lp *models.LineProfile) []string {
	if lp == nil {
		return nil
	}
	switch {
	case lp.DSL != nil:
		switch lp.DSL.Mode {
		case "VDSL2":
			return []string{models.TypeVDSL2, models.TypeVDSL, models.TypeDSL}
		case "VDSL":
			return []string{models.TypeVDSL, models.TypeDSL}
		case "ADSL2+":
			return []string{models.TypeADSL2, models.TypeADSL, models.TypeDSL}
		case "ADSL2":
			return []string{models.TypeADSL, models.TypeDSL}
		case "ADSL":
			return []string{models.TypeADSL, models.TypeDSL}
		case "G.fast":
			return []string{models.TypeGfast, models.TypeDSL}
		case "SHDSL":
			return []string{models.TypeSHDSL, models.TypeDSL}
		case "SDSL":
			return []string{models.TypeSDSL, models.TypeDSL}
		default:
			return []string{models.TypeDSL}
		}
	case lp.DOCSIS != nil:
		return []string{models.TypeDOCSIS, models.TypeCable}
	case lp.PON != nil:
		switch lp.PON.Type {
		case "XGS-PON":
			return []string{models.TypeXGSPON, models.TypeFTTH, models.TypeFiber}
		case "XG-PON":
			return []string{models.TypeXGPON, models.TypeFTTH, models.TypeFiber}
		case "EPON":
			return []string{models.TypeEPON, models.TypeFTTH, models.TypeFiber}
		case "10G-EPON":
			return []string{models.TypeTenGEPON, models.TypeFTTH, models.TypeFiber}
		case "GPON":
			return []string{models.TypeGPON, models.TypeFTTH, models.TypeFiber}
		default:
			return []string{models.TypeFTTH, models.TypeFiber}
		}
	}
	return nil
}

// Summary returns human-readable lines describing the line profile, suitable for
// an explanation block. Empty for a nil profile.
func Summary(lp *models.LineProfile) []string {
	if lp == nil {
		return nil
	}
	var lines []string
	switch {
	case lp.DSL != nil:
		d := lp.DSL
		head := "DSL line read from the modem"
		if lp.Subtype != "" {
			head += ": " + lp.Subtype
		} else if lp.Technology != "" {
			head += ": " + lp.Technology
		}
		if d.Standard != "" {
			head += " [" + d.Standard + "]"
		}
		lines = append(lines, head+".")
		if d.SyncDownKbps > 0 || d.SyncUpKbps > 0 {
			lines = append(lines, "Sync rate "+mbps(d.SyncDownKbps)+" / "+mbps(d.SyncUpKbps)+
				" Mbps (attainable "+mbps(d.AttainableDownKbps)+" / "+mbps(d.AttainableUpKbps)+").")
		}
		if d.SNRMarginDownDB != 0 || d.AttenuationDownDB != 0 {
			lines = append(lines, "SNR margin "+oneDP(d.SNRMarginDownDB)+" dB, attenuation "+oneDP(d.AttenuationDownDB)+" dB (downstream).")
		}
		lines = append(lines, lp.Notes...)
	case lp.DOCSIS != nil:
		c := lp.DOCSIS
		head := "Cable line read from the modem: " + lp.Subtype
		lines = append(lines, head+".")
		if c.DownstreamChannels > 0 || c.UpstreamChannels > 0 {
			lines = append(lines, "Channels "+itoa(c.DownstreamChannels)+" downstream / "+itoa(c.UpstreamChannels)+" upstream.")
		}
		if c.SNRMERdB != 0 || c.DownstreamPowerDBmV != 0 {
			lines = append(lines, "Downstream power "+oneDP(c.DownstreamPowerDBmV)+" dBmV, MER/SNR "+oneDP(c.SNRMERdB)+" dB.")
		}
	case lp.PON != nil:
		p := lp.PON
		head := "Optical line read from the ONT: " + lp.Subtype
		if p.ONTModel != "" {
			head += " (ONT " + p.ONTModel + ")"
		}
		lines = append(lines, head+".")
		if p.RxPowerDBm != 0 || p.TxPowerDBm != 0 {
			lines = append(lines, "Optical Rx "+oneDP(p.RxPowerDBm)+" dBm, Tx "+oneDP(p.TxPowerDBm)+" dBm.")
		}
	}
	return lines
}
