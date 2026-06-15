package topology

import (
	"fmt"
	"sort"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

// Evidence kinds (the observable facts the builder reasons over).
const (
	EvidenceInterface    = "interface"
	EvidenceGatewayRoute = "gateway_route"
	EvidenceARPTable     = "arp_table"
	EvidenceReverseDNS   = "reverse_dns"
	EvidenceTCPConnect   = "tcp_connect"
	EvidenceICMPEcho     = "icmp_echo"
	EvidenceNmap         = "nmap"
	EvidenceLLDP         = "lldp"
	EvidenceCDP          = "cdp"
	EvidenceSNMPBridge   = "snmp_bridge"
)

// EvidenceStore mints stable, deterministic Evidence IDs and stores the records.
// IDs are "ev-<n>" in creation order so reports are reproducible across runs of
// the same input. The store is not safe for concurrent use; build evidence on a
// single goroutine (discovery fans out, then hands results to the builder).
type EvidenceStore struct {
	now     func() time.Time
	records []models.Evidence
	seq     int
}

// NewEvidenceStore returns an empty store. If now is nil, time.Now is used.
func NewEvidenceStore(now func() time.Time) *EvidenceStore {
	if now == nil {
		now = time.Now
	}
	return &EvidenceStore{now: now}
}

// Add records a piece of evidence and returns its generated ID.
func (s *EvidenceStore) Add(kind, source, summary string, data map[string]string) string {
	s.seq++
	id := fmt.Sprintf("ev-%d", s.seq)
	s.records = append(s.records, models.Evidence{
		ID:        id,
		Kind:      kind,
		Source:    source,
		Summary:   summary,
		Data:      data,
		Timestamp: s.now().UTC(),
	})
	return id
}

// Records returns the evidence in creation order.
func (s *EvidenceStore) Records() []models.Evidence {
	out := make([]models.Evidence, len(s.records))
	copy(out, s.records)
	return out
}

// dedupSorted returns the unique, sorted set of evidence IDs.
func dedupSorted(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
