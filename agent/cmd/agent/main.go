package main

import (
	"context"
	"os"
	"time"

	"github.com/thekiran/iad/internal/config"
	"github.com/thekiran/iad/internal/discovery"
	"github.com/thekiran/iad/internal/model"
	"github.com/thekiran/iad/internal/output"
	"github.com/thekiran/iad/internal/probe"
)

func main() {
	ctx := context.Background()
	mode := model.ScanModeSafe
	cfg := config.Default(mode)
	runner := probe.NewRunner(nil, cfg)
	scope := model.ScanScope{PublicScanning: false}
	results := runner.Run(ctx, model.ProbeInput{Mode: mode, Scope: scope, Metadata: map[string]any{}})
	scan := discovery.BuildScanOutput("scan_"+time.Now().UTC().Format("20060102_150405"), mode, scope, results, model.NetworkContext{}, time.Now().UTC())
	if err := output.ExportJSON(os.Stdout, scan); err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}
