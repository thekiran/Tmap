package topology

import (
	"sort"

	"github.com/thekiran/iad/pkg/models"
)

// Graph accumulates topology edges between known device IDs. It deduplicates by
// edge ID (source+target+type), keeping the highest-confidence variant, and
// produces deterministically ordered output so reports are reproducible.
type Graph struct {
	deviceIDs map[string]bool
	edges     map[string]models.TopologyEdge
}

// NewGraph returns an empty graph seeded with the known device IDs. Edges whose
// endpoints are not among these IDs are rejected (an edge must connect real,
// discovered devices).
func NewGraph(deviceIDs []string) *Graph {
	set := make(map[string]bool, len(deviceIDs))
	for _, id := range deviceIDs {
		set[id] = true
	}
	return &Graph{deviceIDs: set, edges: map[string]models.TopologyEdge{}}
}

// AddEdge inserts an edge if both endpoints are known devices and the edge is not
// a self-loop. On a duplicate (same ID), the higher-confidence edge wins; on a
// tie the existing one is kept. Returns true if the edge was stored or upgraded.
func (g *Graph) AddEdge(e models.TopologyEdge) bool {
	if e.Source == e.Target {
		return false
	}
	if !g.deviceIDs[e.Source] || !g.deviceIDs[e.Target] {
		return false
	}
	if existing, ok := g.edges[e.ID]; ok {
		if e.Confidence <= existing.Confidence {
			return false
		}
	}
	g.edges[e.ID] = e
	return true
}

// Edges returns all edges sorted by (descending confidence, type, source, target)
// so the most trustworthy links come first and the order is stable.
func (g *Graph) Edges() []models.TopologyEdge {
	out := make([]models.TopologyEdge, 0, len(g.edges))
	for _, e := range g.edges {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Confidence != out[j].Confidence {
			return out[i].Confidence > out[j].Confidence
		}
		if out[i].Type != out[j].Type {
			return out[i].Type < out[j].Type
		}
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		return out[i].Target < out[j].Target
	})
	return out
}

// HasDevice reports whether a device ID is part of the graph.
func (g *Graph) HasDevice(id string) bool { return g.deviceIDs[id] }
