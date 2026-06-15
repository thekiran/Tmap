package topology

import "github.com/thekiran/iad/pkg/models"

// RoleInput carries the evidence-backed facts the classifier uses. It never
// guesses a role from heuristics that the report cannot justify: "gateway" is set
// only when the device is the agent's proven default gateway, and "router" only
// when it forwards (a route hop beyond it was observed).
type RoleInput struct {
	IsAgent         bool
	IsDefaultGW     bool // device is the agent's default gateway (gateway_route evidence)
	ForwardsTraffic bool // a hop beyond this device was observed (route_hop evidence)
}

// ClassifyRoles returns the evidence-based roles for a device, most specific
// first, always including a concrete fallback ("host"). The result is stable for
// a given input.
func ClassifyRoles(in RoleInput) []string {
	var roles []string
	add := func(r string) {
		for _, e := range roles {
			if e == r {
				return
			}
		}
		roles = append(roles, r)
	}

	if in.IsDefaultGW {
		add(models.RoleGateway)
	}
	if in.ForwardsTraffic {
		add(models.RoleRouter)
	}
	if in.IsAgent {
		add(models.RoleAgent)
	}
	// Every device is at minimum a host on the network.
	add(models.RoleHost)
	return roles
}
