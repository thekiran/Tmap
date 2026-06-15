//go:build windows

package system

import (
	"context"
	"fmt"
	"strconv"
)

// Traceroute runs Windows `tracert` (-d disables reverse DNS so we get raw IPs)
// and returns the ordered hop IPs. tracert often exits non-zero when the final
// host is unreachable yet still prints useful hops, so output is parsed even on
// error as long as something came back.
func Traceroute(ctx context.Context, host string, maxHops int) ([]string, error) {
	out, err := Run(ctx, "tracert", "-d", "-h", strconv.Itoa(maxHops), "-w", "1500", host)
	hops := parseHops(out)
	if len(hops) == 0 && err != nil {
		return nil, fmt.Errorf("tracert failed: %w", err)
	}
	return hops, nil
}
