//go:build !windows

package system

import (
	"context"
	"fmt"
	"strconv"
)

// Traceroute runs the unix `traceroute` (-n disables reverse DNS) and returns
// the ordered hop IPs. On many distros traceroute is installed setuid; if it is
// missing or lacks privileges the caller treats the failure as a skipped probe
// rather than aborting the whole scan.
func Traceroute(ctx context.Context, host string, maxHops int) ([]string, error) {
	out, err := Run(ctx, "traceroute", "-n", "-m", strconv.Itoa(maxHops), host)
	hops := parseHops(out)
	if len(hops) == 0 && err != nil {
		return nil, fmt.Errorf("traceroute failed: %w", err)
	}
	return hops, nil
}
