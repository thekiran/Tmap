//go:build !windows

package discovery

import (
	"context"

	"github.com/thekiran/iad/pkg/models"
)

// noopARPSweeper is the non-Windows fallback. There, active ARP resolution relies
// on the TCP-connect sweep populating the kernel neighbour table, which is then
// read via OSARPReader (`ip neigh` / `arp -a`). SendARP is a Windows-only API.
type noopARPSweeper struct{}

func newARPSweeper() ARPSweeper { return noopARPSweeper{} }

func (noopARPSweeper) SweepARP(ctx context.Context, scope models.ScanScope) []ARPEntry {
	return nil
}
