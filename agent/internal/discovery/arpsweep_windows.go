//go:build windows

package discovery

import (
	"context"
	"encoding/binary"
	"net"
	"sort"
	"sync"
	"syscall"
	"unsafe"

	"github.com/thekiran/iad/pkg/models"
)

var (
	modIPHLPAPI = syscall.NewLazyDLL("iphlpapi.dll")
	procSendARP = modIPHLPAPI.NewProc("SendARP")
)

// winARPSweeper resolves every in-scope IPv4 address to a MAC via the Windows
// SendARP API (iphlpapi.dll). SendARP issues an ARP request (or reads the cache)
// and requires no elevation, so this finds every device that answers ARP —
// printers, IoT, phones — regardless of open ports or ICMP filtering. This is the
// primary "find all LAN devices" source on Windows.
type winARPSweeper struct {
	Concurrency int
}

func newARPSweeper() ARPSweeper { return winARPSweeper{} }

func (s winARPSweeper) SweepARP(ctx context.Context, scope models.ScanScope) []ARPEntry {
	hosts := HostsInScope(scope)
	if len(hosts) == 0 {
		return nil
	}
	conc := s.Concurrency
	if conc <= 0 {
		conc = 64
	}

	var (
		mu  sync.Mutex
		out []ARPEntry
		wg  sync.WaitGroup
	)
	sem := make(chan struct{}, conc)
	for _, ip := range hosts {
		if ctx.Err() != nil {
			break
		}
		sem <- struct{}{}
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			defer func() { <-sem }()
			if mac, ok := sendARP(ip); ok {
				mu.Lock()
				out = append(out, ARPEntry{IP: ip, MAC: mac})
				mu.Unlock()
			}
		}(ip)
	}
	wg.Wait()
	sort.Slice(out, func(i, j int) bool { return ipLess(out[i].IP, out[j].IP) })
	return out
}

// sendARP resolves ip to a MAC using the SendARP IP Helper API. ok=false when the
// host does not answer ARP (i.e. is not present on the local L2 segment).
func sendARP(ip string) (string, bool) {
	v4 := net.ParseIP(ip).To4()
	if v4 == nil {
		return "", false
	}
	// SendARP wants the destination IPv4 as a DWORD whose in-memory byte order is
	// network order; v4 bytes are already network order.
	dst := binary.LittleEndian.Uint32(v4)
	var mac [8]byte
	macLen := uint32(len(mac))
	ret, _, _ := procSendARP.Call(
		uintptr(dst), 0,
		uintptr(unsafe.Pointer(&mac[0])),
		uintptr(unsafe.Pointer(&macLen)),
	)
	if ret != 0 || macLen < 6 {
		return "", false
	}
	hw := net.HardwareAddr(mac[:6])
	if m, ok := normalizeMAC(hw.String()); ok {
		return m, true
	}
	return "", false
}
