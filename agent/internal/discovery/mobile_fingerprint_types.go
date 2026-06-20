package discovery

import (
	"time"

	"github.com/thekiran/iad/pkg/models"
)

type MobileFingerprint = models.MobileFingerprint
type MobileEvidenceItem = models.MobileEvidenceItem
type MobileConflictItem = models.MobileConflictItem

type MobileFingerprintInput struct {
	DeviceID       string
	MACAddresses   []string
	OUIVendors     []string
	VendorHints    []MobileObservedValue
	Hostnames      []MobileObservedValue
	DHCPHostnames  []MobileObservedValue
	MDNSRecords    []MobileMDNSRecord
	DNSQueries     []MobileObservedValue
	Services       []MobileServiceObservation
	UserAgents     []MobileObservedValue
	PassivePackets []MobilePassivePacket
	Timestamp      time.Time
}

type MobileObservedValue struct {
	Value     string
	Source    string
	Timestamp time.Time
}

type MobileMDNSRecord struct {
	Name      string
	Service   string
	Target    string
	Text      []string
	Source    string
	Timestamp time.Time
}

type MobileServiceObservation struct {
	Port      int
	Protocol  string
	Name      string
	Product   string
	Source    string
	Timestamp time.Time
}

type MobilePassivePacket struct {
	Protocol        string
	SourcePort      int
	DestinationPort int
	Value           string
	Source          string
	Timestamp       time.Time
}
