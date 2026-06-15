package models

// Access type identifiers. The detection engine, rules and fingerprints all
// score against these strings. They mix specific technologies (VDSL2, GPON)
// with broader buckets (DSL, Fiber) on purpose: the strength of the evidence
// determines how specific a verdict we can honestly make.
const (
	// DSL family
	TypeDSL   = "DSL"
	TypeADSL  = "ADSL"
	TypeADSL2 = "ADSL2+"
	TypeVDSL  = "VDSL"
	TypeVDSL2 = "VDSL2"
	TypeGfast = "G.fast"
	TypeSDSL  = "SDSL"
	TypeSHDSL = "SHDSL"

	// Fiber family
	TypeFiber         = "Fiber"
	TypeFTTH          = "FTTH"
	TypeFTTB          = "FTTB"
	TypeFTTC          = "FTTC"
	TypeFTTN          = "FTTN"
	TypeFTTP          = "FTTP"
	TypeGPON          = "GPON"
	TypeEPON          = "EPON"
	TypeXGPON         = "XG-PON"
	TypeXGSPON        = "XGS-PON"
	TypeTenGEPON      = "10G-EPON"
	TypeActiveEthernet = "Active Ethernet"
	TypeEthernetWAN   = "Ethernet WAN"

	// Cable family
	TypeCable      = "Cable"
	TypeDOCSIS     = "DOCSIS"
	TypeEuroDOCSIS = "EuroDOCSIS"
	TypeHFC        = "HFC"

	// Fixed wireless
	TypeFixedWireless = "Fixed Wireless"
	TypeWISP          = "WISP"
	TypeMicrowave     = "Microwave"
	TypePTPWireless   = "Point-to-Point Wireless"
	TypePMPWireless   = "Point-to-Multipoint Wireless"

	// Mobile / fixed-wireless-access
	TypeMobile = "Mobile"
	TypeCellular = "Cellular"
	TypeFWA    = "FWA"
	TypeFWA5G  = "5G FWA"
	TypeLTE    = "LTE"
	TypeNR5G   = "5G"
	TypeWWAN   = "WWAN"

	// Satellite
	TypeSatellite = "Satellite"
	TypeLEOSatellite = "LEO Satellite"
	TypeGEOSatellite = "GEO Satellite"
	TypeMEOSatellite = "MEO Satellite"
	TypeVSAT = "VSAT"

	// Enterprise
	TypeEnterprise = "Enterprise"
	TypeEthernet = "Ethernet"
	TypeMetroEthernet = "Metro Ethernet"
	TypeDIA = "DIA"
	TypeMPLS = "MPLS"
	TypeLeasedLine = "Leased Line"
	TypePublicWiFi = "Public Wi-Fi"
	TypeVPNOverlay = "VPN / Overlay Network"
)

// Access categories (the coarse grouping shown next to the primary verdict).
const (
	CatDSL        = "DSL"
	CatFiber      = "Fiber"
	CatCable      = "Cable"
	CatWireless   = "Fixed Wireless"
	CatMobile     = "Mobile"
	CatSatellite  = "Satellite"
	CatEnterprise = "Enterprise"
	CatUnknown    = "Unknown"
)

// categoryByType maps every scoreable type key to its category. Category-level
// keys (e.g. "DSL", "Fiber") map to themselves so the classifier can resolve a
// category even when only a coarse verdict was reached.
var categoryByType = map[string]string{
	TypeDSL: CatDSL, TypeADSL: CatDSL, TypeADSL2: CatDSL, TypeVDSL: CatDSL, TypeVDSL2: CatDSL, TypeGfast: CatDSL, TypeSDSL: CatDSL, TypeSHDSL: CatDSL,
	TypeFiber: CatFiber, TypeFTTH: CatFiber, TypeFTTB: CatFiber, TypeFTTC: CatFiber, TypeFTTN: CatFiber, TypeFTTP: CatFiber, TypeGPON: CatFiber, TypeEPON: CatFiber, TypeXGPON: CatFiber, TypeXGSPON: CatFiber, TypeTenGEPON: CatFiber, TypeActiveEthernet: CatFiber, TypeEthernetWAN: CatFiber,
	TypeCable: CatCable, TypeDOCSIS: CatCable, TypeEuroDOCSIS: CatCable, TypeHFC: CatCable,
	// Cellular FWA (4G/5G home internet) is delivered over the mobile radio
	// network, so it shares the Mobile category. Only genuinely non-cellular fixed
	// wireless (WISP, microwave, PtP/PtMP links) stays in Fixed Wireless. This
	// keeps "Mobile vs FWA" a subtype question, not a cross-medium ambiguity.
	TypeFixedWireless: CatWireless, TypeWISP: CatWireless, TypeMicrowave: CatWireless, TypePTPWireless: CatWireless, TypePMPWireless: CatWireless,
	TypeMobile: CatMobile, TypeCellular: CatMobile, TypeLTE: CatMobile, TypeNR5G: CatMobile, TypeWWAN: CatMobile, TypeFWA: CatMobile, TypeFWA5G: CatMobile,
	TypeSatellite: CatSatellite, TypeLEOSatellite: CatSatellite, TypeGEOSatellite: CatSatellite, TypeMEOSatellite: CatSatellite, TypeVSAT: CatSatellite,
	TypeEnterprise: CatEnterprise, TypeEthernet: CatEnterprise, TypeMetroEthernet: CatEnterprise, TypeDIA: CatEnterprise, TypeMPLS: CatEnterprise, TypeLeasedLine: CatEnterprise, TypePublicWiFi: CatEnterprise, TypeVPNOverlay: CatEnterprise,
}

// CategoryFor returns the access category for a type key, or CatUnknown.
func CategoryFor(t string) string {
	if c, ok := categoryByType[t]; ok {
		return c
	}
	return CatUnknown
}
