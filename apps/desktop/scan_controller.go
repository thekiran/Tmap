package main

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Scan lifecycle events emitted to the frontend. The console subscribes to
// these to drive the live topology dashboard, so the UI never has to block on a
// full scan completing before it can render.
//
//	scan:started     — a scan goroutine has launched
//	scan:progress    — periodic phase/elapsed update while the scan runs
//	topology:updated — the final graph is available (payload carries raw JSON)
//	scan:completed   — the scan finished successfully
//	scan:failed      — the agent errored (payload carries the message)
//	scan:cancelled   — the scan was cancelled by the user
//
// NOTE: the bundled iad-agent binary emits its JSON report only once, at the
// end of a run (it is a single-shot CLI, not a streaming pipeline). We therefore
// cannot surface genuine per-device "topology:partial" events from the real
// agent without rewriting its discovery pipeline — which is out of scope. The
// phase progress below is REAL elapsed-time progress through the agent's known
// stages; the device/link graph arrives in one payload on topology:updated.
const (
	evtScanStarted   = "scan:started"
	evtScanProgress  = "scan:progress"
	evtTopologyDone  = "topology:updated"
	evtScanCompleted = "scan:completed"
	evtScanFailed    = "scan:failed"
	evtScanCancelled = "scan:cancelled"
)

// scanPhases is the ordered set of stages the agent moves through on a --full
// run. The controller advances these on an elapsed-time schedule so the status
// panel shows continuous, truthful progress while the single underlying scan
// runs to completion.
var scanPhases = []string{
	"starting",
	"enumerating_interfaces",
	"arp_sweep",
	"tcp_probe",
	"nmap_services",
	"classifying_access",
	"finalizing",
}

type scanProgressEvent struct {
	ScanID    string `json:"scanId"`
	Timestamp int64  `json:"timestamp"`
	Phase     string `json:"phase"`
	ElapsedMs int64  `json:"elapsedMs"`
	Status    string `json:"status"`
}

type scanResultEvent struct {
	ScanID    string `json:"scanId"`
	Timestamp int64  `json:"timestamp"`
	Phase     string `json:"phase"`
	Raw       string `json:"raw"`
}

type scanFailureEvent struct {
	ScanID    string `json:"scanId"`
	Timestamp int64  `json:"timestamp"`
	Error     string `json:"error"`
}

// StartScan launches the external iad-agent on a background goroutine and
// returns an id immediately. Progress and results are delivered to the frontend
// via Wails events (see the evt* constants). This is the non-blocking entry
// point the live dashboard uses; the legacy blocking RunScan is preserved for
// callers/environments that cannot subscribe to events.
func (a *App) StartScan(mode string, iface string) (string, error) {
	args, err := agentScanArgs(mode, iface)
	if err != nil {
		return "", err
	}
	bin, err := a.resolveAgentBin()
	if err != nil {
		return "", err
	}

	a.scanMu.Lock()
	// Only one managed scan at a time — cancel any predecessor before starting.
	if a.scanCancel != nil {
		a.scanCancel()
	}
	scanID := fmt.Sprintf("scan-%d", time.Now().UnixNano())
	// Allow comfortably more than the agent's own budget so we capture its full
	// output instead of killing it at the wire.
	ctx, cancel := context.WithTimeout(a.ctx, maxScanTimeout+60*time.Second)
	a.scanCancel = cancel
	a.scanID = scanID
	a.scanMu.Unlock()

	started := time.Now()
	a.emit(evtScanStarted, scanProgressEvent{
		ScanID:    scanID,
		Timestamp: started.UnixMilli(),
		Phase:     scanPhases[0],
		Status:    "scanning",
	})

	go a.runScanJob(ctx, cancel, scanID, bin, args, started)
	return scanID, nil
}

// CancelScan cancels the in-flight managed scan, if any. The goroutine emits
// scan:cancelled once the agent process actually stops.
func (a *App) CancelScan() {
	a.scanMu.Lock()
	cancel := a.scanCancel
	a.scanMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// LatestSnapshot returns the raw JSON of the most recent successful scan, or an
// empty string if none has completed. The frontend uses this both as a
// snapshot-request and as a polling fallback when Wails events are unavailable.
func (a *App) LatestSnapshot() string {
	a.scanMu.Lock()
	defer a.scanMu.Unlock()
	return a.lastReport
}

func (a *App) runScanJob(ctx context.Context, cancel context.CancelFunc, scanID, bin string, args []string, started time.Time) {
	defer cancel()

	// Drive phase progress on its own goroutine until the agent returns.
	done := make(chan struct{})
	go a.emitPhases(ctx, scanID, started, done)

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = filepath.Dir(bin)
	hideConsole(cmd)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	close(done)

	now := time.Now().UnixMilli()

	// Cancellation takes precedence: a killed process also reports an error, but
	// the user-facing meaning is "cancelled", not "failed".
	if ctx.Err() == context.Canceled {
		a.emit(evtScanCancelled, scanProgressEvent{
			ScanID: scanID, Timestamp: now, Phase: "cancelled", Status: "cancelled",
		})
		a.clearScan(scanID)
		return
	}

	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			msg = err.Error() + ": " + msg
		} else {
			msg = err.Error()
		}
		a.emit(evtScanFailed, scanFailureEvent{ScanID: scanID, Timestamp: now, Error: msg})
		a.clearScan(scanID)
		return
	}

	raw := string(out)
	a.scanMu.Lock()
	a.lastReport = raw
	a.scanMu.Unlock()

	// Deliver the final graph, then mark the lifecycle complete so the dashboard
	// can merge the topology and update its status panel in one coherent step.
	a.emit(evtTopologyDone, scanResultEvent{ScanID: scanID, Timestamp: now, Phase: "finalizing", Raw: raw})
	a.emit(evtScanCompleted, scanResultEvent{ScanID: scanID, Timestamp: now, Phase: "completed", Raw: raw})
	a.clearScan(scanID)
}

// emitPhases advances the visible scan phase on an elapsed-time schedule. This
// is real progress (it reflects how long the agent has actually been running),
// not fabricated device data.
func (a *App) emitPhases(ctx context.Context, scanID string, started time.Time, done chan struct{}) {
	ticker := time.NewTicker(900 * time.Millisecond)
	defer ticker.Stop()
	budget := maxScanTimeout
	last := len(scanPhases) - 1
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			elapsed := time.Since(started)
			idx := max(int(float64(last)*(float64(elapsed)/float64(budget))), 0)
			// Hold at the penultimate phase until the agent actually returns;
			// "finalizing" is emitted alongside the completed result.
			idx = min(idx, last-1)
			a.emit(evtScanProgress, scanProgressEvent{
				ScanID:    scanID,
				Timestamp: time.Now().UnixMilli(),
				Phase:     scanPhases[idx],
				ElapsedMs: elapsed.Milliseconds(),
				Status:    "scanning",
			})
		}
	}
}

// clearScan releases the in-flight scan bookkeeping, but only if scanID is still
// the current scan (a newer StartScan may have superseded it).
func (a *App) clearScan(scanID string) {
	a.scanMu.Lock()
	if a.scanID == scanID {
		a.scanCancel = nil
		a.scanID = ""
	}
	a.scanMu.Unlock()
}

// emit is a small guard around runtime.EventsEmit so a nil startup context
// (e.g. in tests) is a no-op rather than a panic.
func (a *App) emit(name string, payload any) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, name, payload)
}
