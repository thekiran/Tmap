package topology

import (
	"fmt"

	"github.com/thekiran/iad/pkg/models"
)

// validEdgeTypes is the closed set of edge types a report may contain.
var validEdgeTypes = map[string]bool{
	models.EdgeDirectLLDP:     true,
	models.EdgeDirectCDP:      true,
	models.EdgeSNMPBridge:     true,
	models.EdgeRouteHop:       true,
	models.EdgeInferredL2:     true,
	models.EdgeGatewayDefault: true,
}

// ValidateReport checks a ScanReport for internal consistency: known schema, edge
// types, confidence ranges, mandatory reasons, and that every cross-reference
// (edge endpoints, evidence IDs, summary counts) resolves. It returns a list of
// human-readable problems; an empty list means the report is valid. It is pure,
// so it doubles as a test oracle and the `validate` command's engine.
func ValidateReport(r models.ScanReport) []string {
	var problems []string
	add := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }

	if r.SchemaVersion != models.TopologyReportSchema {
		add("schema_version %q is not the expected %q", r.SchemaVersion, models.TopologyReportSchema)
	}

	deviceIDs := map[string]bool{}
	for i, d := range r.Devices {
		if d.ID == "" {
			add("device[%d] has an empty id", i)
			continue
		}
		if deviceIDs[d.ID] {
			add("duplicate device id %q", d.ID)
		}
		deviceIDs[d.ID] = true
		if len(d.Addresses) == 0 {
			add("device %q has no addresses", d.ID)
		}
		if d.Confidence < 0 || d.Confidence > 1 {
			add("device %q confidence %.3f out of range [0,1]", d.ID, d.Confidence)
		}
	}

	evidenceIDs := map[string]bool{}
	for i, e := range r.Evidence {
		if e.ID == "" {
			add("evidence[%d] has an empty id", i)
			continue
		}
		if evidenceIDs[e.ID] {
			add("duplicate evidence id %q", e.ID)
		}
		evidenceIDs[e.ID] = true
	}
	registryIDs := map[string]bool{}
	for i, e := range r.EvidenceRegistry {
		if e.ID == "" {
			add("evidence_registry[%d] has an empty id", i)
			continue
		}
		if registryIDs[e.ID] {
			add("duplicate evidence_registry id %q", e.ID)
		}
		registryIDs[e.ID] = true
		evidenceIDs[e.ID] = true
		if e.Source == "" {
			add("evidence_registry %q has no source", e.ID)
		}
		if e.Kind == "" {
			add("evidence_registry %q has no kind", e.ID)
		}
	}

	// Devices/services reference valid evidence.
	for _, d := range r.Devices {
		checkEvidence(add, "device "+d.ID, d.EvidenceIDs, evidenceIDs)
		for _, s := range d.Services {
			checkEvidence(add, fmt.Sprintf("device %s service %d/%s", d.ID, s.Port, s.Protocol), s.EvidenceIDs, evidenceIDs)
		}
	}

	edgeIDs := map[string]bool{}
	for i, e := range r.Edges {
		if e.ID == "" {
			add("edge[%d] has an empty id", i)
		} else if edgeIDs[e.ID] {
			add("duplicate edge id %q", e.ID)
		}
		edgeIDs[e.ID] = true
		if !validEdgeTypes[e.Type] {
			add("edge %q has unknown type %q", e.ID, e.Type)
		}
		if e.Layer == "" {
			add("edge %q has no layer", e.ID)
		}
		if e.Relationship == "" {
			add("edge %q has no relationship", e.ID)
		}
		if e.UILineStyle == "" {
			add("edge %q has no ui_line_style", e.ID)
		}
		if e.Confidence < 0 || e.Confidence > 1 {
			add("edge %q confidence %.3f out of range [0,1]", e.ID, e.Confidence)
		}
		if e.Reason == "" {
			add("edge %q has no reason (every edge must explain itself)", e.ID)
		}
		if !deviceIDs[e.Source] {
			add("edge %q references unknown source device %q", e.ID, e.Source)
		}
		if !deviceIDs[e.Target] {
			add("edge %q references unknown target device %q", e.ID, e.Target)
		}
		if IsPhysicalEvidenceEdge(e.Type) && len(e.EvidenceIDs) == 0 {
			add("edge %q claims physical adjacency (%s) but cites no evidence", e.ID, e.Type)
		}
		if !IsPhysicalEvidenceEdge(e.Type) && e.Physical {
			add("edge %q marks physical=true without physical proof edge type %q", e.ID, e.Type)
		}
		checkEvidence(add, "edge "+e.ID, e.EvidenceIDs, evidenceIDs)
	}

	// Summary counts must match the body.
	if r.Summary.DeviceCount != len(r.Devices) {
		add("summary.device_count %d != %d devices", r.Summary.DeviceCount, len(r.Devices))
	}
	if r.Summary.EdgeCount != len(r.Edges) {
		add("summary.edge_count %d != %d edges", r.Summary.EdgeCount, len(r.Edges))
	}
	if r.Summary.EvidenceCount != len(r.Evidence) {
		add("summary.evidence_count %d != %d evidence records", r.Summary.EvidenceCount, len(r.Evidence))
	}

	return problems
}

func checkEvidence(add func(string, ...any), owner string, ids []string, known map[string]bool) {
	for _, id := range ids {
		if !known[id] {
			add("%s references unknown evidence id %q", owner, id)
		}
	}
}
