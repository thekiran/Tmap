// Package scoring turns evidence into per-access-type scores. Rules live in YAML
// (doc §5.3) so new modem models or ISP patterns can be added without
// recompiling; this file holds only the few code-side weights.
package scoring

// Evidence-strength weights, in the same "points" unit the YAML rules use. A
// fingerprint match (a known modem model) is the strongest single signal; a
// probe hint (e.g. CGNAT → Mobile) is moderate; a device's secondary supported
// technologies get a smaller nudge so a specific verdict can edge out the
// coarse category.
const (
	WeightFingerprintHint = 40.0
	WeightFingerprintSupp = 15.0
	WeightProbeHint       = 25.0
	WeightStrongAccessHint = 45.0
)

// Evidence tiers, used by the decision layer to judge whether the verdict rests
// on anything that actually proves the physical access type.
//
//	strong  — modem fingerprint, WAN interface (ptm0/atm0/gpon/ont/lte), DSL line
//	          stats, ONT/GPON/EPON, DOCSIS/cable-modem.
//	medium  — UPnP device type, HTTP banner, gateway MAC vendor, interface hints,
//	          ISP-specific patterns, CGNAT.
//	weak    — PTR, ASN org, public IP, latency, DNS servers, local adapter name.
const (
	EvidenceWeak   = "weak"
	EvidenceMedium = "medium"
	EvidenceStrong = "strong"
)

// Latency-band supportive weights. Latency is weak corroboration only and must
// never decide a verdict on its own, so these are deliberately tiny next to the
// fingerprint/rule weights above. The bands follow the brief's §5 guidance.
const (
	LatLowFiber   = 4.0  // <= 8 ms  → compatible with fiber
	LatLowVDSL    = 3.0  // <= 8 ms  → or short-loop VDSL
	LatMidGeneric = 2.0  // 8-25 ms  → generic fixed broadband (no type preference)
	LatDSLFWA     = 2.0  // 25-80 ms → DSL or FWA possible
	LatMobile     = 4.0  // 80-200 ms → mobile/FWA more likely
	LatSatellite  = 15.0 // >= 500 ms → satellite plausible (still not decisive)
)
