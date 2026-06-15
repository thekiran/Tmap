package models

// LineProfile is the normalized, fine-grained physical-layer line profile parsed
// from authorized CPE telemetry (TR-064 / SNMP / UPnP-IGD / vendor API). It is
// the deepest evidence the system can obtain: it does not infer the medium from
// behaviour, it reads what the line actually negotiated.
//
// It is Physical-tier evidence and is only populated when the CPE genuinely
// exposed it. A nil *LineProfile means "we could not read the line", never
// "the line is absent".
type LineProfile struct {
	// Medium is the broad physical medium the profile describes:
	// "dsl" | "docsis" | "pon" | "ethernet".
	Medium string `json:"medium,omitempty"`
	// Technology is the negotiated technology, e.g. "VDSL2", "ADSL2+", "G.fast",
	// "DOCSIS 3.1", "GPON", "XGS-PON".
	Technology string `json:"technology,omitempty"`
	// Subtype is a human-facing precise descriptor, e.g.
	// "VDSL2 Profile 35b (Super Vectoring)" or "DOCSIS 3.1 (OFDM/OFDMA)".
	Subtype string `json:"subtype,omitempty"`

	DSL    *DSLProfile    `json:"dsl,omitempty"`
	DOCSIS *DOCSISProfile `json:"docsis,omitempty"`
	PON    *PONProfile    `json:"pon,omitempty"`

	// Source identifies where the profile was read from (e.g. "tr064_probe").
	Source string `json:"source,omitempty"`
	// Confidence in the *line reading itself* (not the whole scan). High when the
	// CPE returned a recognized technology + at least one corroborating stat.
	Confidence float64 `json:"confidence,omitempty"`
	// Notes carry plain-language remarks (e.g. "35b implies super vectoring").
	Notes []string `json:"notes,omitempty"`
}

// DSLProfile holds DSL physical-layer line parameters. Rates are in kbit/s and
// dB-valued fields are in dB (already converted from any 0.1 dB CPE encoding).
type DSLProfile struct {
	Mode     string `json:"mode,omitempty"`     // ADSL | ADSL2 | ADSL2+ | VDSL2 | G.fast | SDSL | SHDSL
	Standard string `json:"standard,omitempty"` // ITU standard, e.g. G.993.2, G.993.5, G.9701
	Annex    string `json:"annex,omitempty"`    // e.g. Annex A / Annex B

	// Profile is the VDSL2 profile token: 8a/8b/8c/8d/12a/12b/17a/30a/35b.
	Profile        string  `json:"profile,omitempty"`
	ProfileBandMHz float64 `json:"profile_band_mhz,omitempty"` // nominal band of the profile
	Vectoring      bool    `json:"vectoring,omitempty"`        // G.993.5 vectoring / super vectoring

	SNRMarginDownDB   float64 `json:"snr_margin_down_db,omitempty"`
	SNRMarginUpDB     float64 `json:"snr_margin_up_db,omitempty"`
	AttenuationDownDB float64 `json:"attenuation_down_db,omitempty"`
	AttenuationUpDB   float64 `json:"attenuation_up_db,omitempty"`

	AttainableDownKbps int64  `json:"attainable_down_kbps,omitempty"`
	AttainableUpKbps   int64  `json:"attainable_up_kbps,omitempty"`
	SyncDownKbps       int64  `json:"sync_down_kbps,omitempty"`
	SyncUpKbps         int64  `json:"sync_up_kbps,omitempty"`
	Path               string `json:"path,omitempty"` // interleaved | fast
	InterleaveDepth    string `json:"interleave_depth,omitempty"`
}

// DOCSISProfile holds cable-modem physical-layer parameters.
type DOCSISProfile struct {
	Version             string  `json:"version,omitempty"` // 2.0 | 3.0 | 3.1 | 4.0
	OFDM                bool    `json:"ofdm,omitempty"`    // downstream OFDM (3.1+)
	OFDMA               bool    `json:"ofdma,omitempty"`   // upstream OFDMA (3.1+)
	DownstreamChannels  int     `json:"downstream_channels,omitempty"`
	UpstreamChannels    int     `json:"upstream_channels,omitempty"`
	DownstreamPowerDBmV float64 `json:"downstream_power_dbmv,omitempty"`
	UpstreamPowerDBmV   float64 `json:"upstream_power_dbmv,omitempty"`
	SNRMERdB            float64 `json:"snr_mer_db,omitempty"`
}

// PONProfile holds passive-optical-network / ONT parameters.
type PONProfile struct {
	Type       string  `json:"type,omitempty"` // GPON | EPON | XG-PON | XGS-PON | 10G-EPON
	ONTModel   string  `json:"ont_model,omitempty"`
	RxPowerDBm float64 `json:"rx_power_dbm,omitempty"`
	TxPowerDBm float64 `json:"tx_power_dbm,omitempty"`
}
