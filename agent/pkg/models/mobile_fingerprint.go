package models

const (
	MobileClassificationConfirmedIOS     = "confirmed_ios"
	MobileClassificationProbableIOS      = "probable_ios"
	MobileClassificationPossibleIOS      = "possible_ios"
	MobileClassificationConfirmedIPadOS  = "confirmed_ipados"
	MobileClassificationProbableIPadOS   = "probable_ipados"
	MobileClassificationPossibleIPadOS   = "possible_ipados"
	MobileClassificationConfirmedAndroid = "confirmed_android"
	MobileClassificationProbableAndroid  = "probable_android"
	MobileClassificationPossibleAndroid  = "possible_android"
	MobileClassificationConflict         = "conflicting_mobile_os_evidence"
	MobileClassificationUnknownMobile    = "unknown_mobile"
	MobileClassificationUnknownDevice    = "unknown_device"
	MobileDeviceTypeHintPhone            = "phone"
	MobileDeviceTypeHintTablet           = "tablet"
	MobileDeviceTypeHintComputer         = "computer"
	MobileDeviceTypeHintIoT              = "iot"
	MobileDeviceTypeHintRouter           = "router"
	MobileDeviceTypeHintUnknown          = "unknown"
	MobileOSHintIOS                      = "ios"
	MobileOSHintIPadOS                   = "ipados"
	MobileOSHintAndroid                  = "android"
	MobileOSHintUnknown                  = "unknown"
	MobileEvidenceTypeMACOUI             = "mac_oui"
	MobileEvidenceTypeHostname           = "hostname"
	MobileEvidenceTypeMDNS               = "mdns"
	MobileEvidenceTypeDNSQuery           = "dns_query"
	MobileEvidenceTypeDHCPHostname       = "dhcp_hostname"
	MobileEvidenceTypeServicePort        = "service_port"
	MobileEvidenceTypeUserAgent          = "user_agent"
	MobileEvidenceTypePassivePacket      = "passive_packet"
	MobileEvidenceTypeVendorHint         = "vendor_hint"
	MobileEvidenceTypeServiceName        = "service_name"
	MobileEvidenceStrengthStrong         = "strong"
	MobileEvidenceStrengthMedium         = "medium"
	MobileEvidenceStrengthWeak           = "weak"
	MobileConflictSeverityInfo           = "info"
	MobileConflictSeverityWarning        = "warning"
)

type MobileFingerprint struct {
	Classification        string               `json:"classification"`
	IOSScore              int                  `json:"iosScore"`
	AndroidScore          int                  `json:"androidScore"`
	IPadScore             int                  `json:"ipadScore"`
	Confidence            float64              `json:"confidence"`
	Evidence              []MobileEvidenceItem `json:"evidence,omitempty"`
	Conflicts             []MobileConflictItem `json:"conflicts,omitempty"`
	Warnings              []string             `json:"warnings,omitempty"`
	LastUpdatedAt         string               `json:"lastUpdatedAt,omitempty"`
	WhyThisClassification string               `json:"whyThisClassification,omitempty"`
	WhyNotCertain         string               `json:"whyNotCertain,omitempty"`
}

type MobileEvidenceItem struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	Value            string `json:"value"`
	OSHint           string `json:"osHint"`
	ConfidenceImpact int    `json:"confidenceImpact"`
	Strength         string `json:"strength"`
	Source           string `json:"source"`
	Timestamp        string `json:"timestamp"`
	Explanation      string `json:"explanation"`
}

type MobileConflictItem struct {
	Reason             string   `json:"reason"`
	IOSEvidenceIDs     []string `json:"iosEvidenceIds,omitempty"`
	AndroidEvidenceIDs []string `json:"androidEvidenceIds,omitempty"`
	Severity           string   `json:"severity"`
	ResolutionHint     string   `json:"resolutionHint"`
}
