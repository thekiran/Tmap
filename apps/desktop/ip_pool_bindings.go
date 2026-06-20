package main

import "iad-console/ippool"

type appDiscoveryEmitter struct {
	app *App
}

func (e appDiscoveryEmitter) Emit(event string, data any) {
	if e.app != nil {
		e.app.emit(event, data)
	}
}

func (a *App) discoveryPoolLocked() *ippool.Manager {
	if a.ipPool == nil {
		a.ipPool = ippool.New(ippool.DefaultConfig(), appDiscoveryEmitter{app: a}, hideConsole)
	}
	return a.ipPool
}

// StartDiscoveryPool starts the continuous, safe IP discovery pool. It is
// separate from StartScan: StartScan runs the external single-shot iad-agent,
// while this manager emits live per-device discovery:* events.
func (a *App) StartDiscoveryPool(seeds []string, confirmLargeScope bool) error {
	a.poolMu.Lock()
	pool := a.discoveryPoolLocked()
	a.poolMu.Unlock()
	return pool.Start(seeds, confirmLargeScope)
}

func (a *App) StopDiscoveryPool() {
	a.poolMu.Lock()
	pool := a.ipPool
	a.poolMu.Unlock()
	if pool != nil {
		pool.Stop()
	}
}

func (a *App) AddDiscoveryPoolSeed(ip string) error {
	a.poolMu.Lock()
	pool := a.discoveryPoolLocked()
	a.poolMu.Unlock()
	return pool.AddSeed(ip)
}

func (a *App) ClearDiscoveryPoolStale() int {
	a.poolMu.Lock()
	pool := a.ipPool
	a.poolMu.Unlock()
	if pool == nil {
		return 0
	}
	return pool.ClearStale()
}

func (a *App) DiscoveryPoolDevices() []ippool.DevicePoolEntry {
	a.poolMu.Lock()
	pool := a.ipPool
	a.poolMu.Unlock()
	if pool == nil {
		return nil
	}
	return pool.Devices()
}

func (a *App) AddDiscoveryPoolEvidence(ip string, item ippool.EvidenceItem) bool {
	a.poolMu.Lock()
	pool := a.discoveryPoolLocked()
	a.poolMu.Unlock()
	return pool.AddDeviceEvidence(ip, item)
}
