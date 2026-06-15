// Command iad-agent runs an internet-access detection scan and prints/saves the
// result. It is the CLI front-end for the MVP; a future Tauri UI or local API
// will reuse the same probes + detection engine.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/thekiran/iad/internal/detection"
	"github.com/thekiran/iad/internal/probes"
	"github.com/thekiran/iad/internal/report"
	"github.com/thekiran/iad/pkg/models"
)

func main() {
	mode := flag.String("mode", models.ModeQuick, "scan mode: quick | deep")
	online := flag.Bool("online", true, "allow probes that contact external services (public IP, ASN, traceroute)")
	offline := flag.Bool("offline", false, "disable all online probes (overrides -online)")
	rulesDir := flag.String("rules", "", "directory with rule/fingerprint YAML (auto-detected if empty)")
	out := flag.String("out", "", "write the JSON report to this file")
	flag.Parse()

	if *mode != models.ModeQuick && *mode != models.ModeDeep {
		fmt.Fprintf(os.Stderr, "invalid -mode %q (use quick or deep)\n", *mode)
		os.Exit(2)
	}
	isOnline := *online && !*offline

	dir, err := resolveRulesDir(*rulesDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	engine, err := detection.NewEngine(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load rules:", err)
		os.Exit(1)
	}

	in := models.ScanInput{Mode: *mode, Online: isOnline, RulesDir: dir}
	runner := probes.Runner{Probes: probes.Default(in), Timeout: 12 * time.Second}

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	results := runner.Run(ctx, in)
	result := engine.Analyze(in, results)

	printSummary(result, in, dir)

	if *out != "" {
		if err := report.WriteJSON(*out, result); err != nil {
			fmt.Fprintln(os.Stderr, "failed to write report:", err)
			os.Exit(1)
		}
		fmt.Printf("\nRapor yazıldı: %s\n", *out)
	}
}

// resolveRulesDir returns the rules directory: the flag if set, otherwise the
// first candidate (relative to the cwd and the executable) that contains
// access_rules.yaml.
func resolveRulesDir(flagDir string) (string, error) {
	if flagDir != "" {
		return flagDir, nil
	}
	candidates := []string{"rules", filepath.Join("..", "rules")}
	if exe, err := os.Executable(); err == nil {
		base := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(base, "rules"),
			filepath.Join(base, "..", "rules"),
			filepath.Join(base, "..", "..", "rules"),
		)
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "access_rules.yaml")); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("could not locate rules directory (looked for access_rules.yaml); pass -rules <dir>")
}

func printSummary(r models.ScanResult, in models.ScanInput, dir string) {
	fmt.Println("Internet Access Detector")
	fmt.Println("========================")
	fmt.Printf("Mode: %s | Online: %t | Rules: %s\n", in.Mode, in.Online, dir)

	for _, p := range r.Evidence {
		switch p.ProbeName {
		case "adapter_probe":
			if a := strSlice(p.Evidence["active"]); len(a) > 0 {
				fmt.Printf("Active adapter(s): %v\n", a)
			}
		case "gateway_probe":
			if g, ok := p.Evidence["gateway"].(string); ok {
				fmt.Printf("Gateway: %s\n", g)
			}
		case "public_ip_probe":
			if ip, ok := p.Evidence["public_ip"].(string); ok {
				fmt.Printf("Public IP: %s\n", ip)
			}
		}
	}

	fmt.Printf("\nVerdict: %s  (category: %s)\n", r.PrimaryType, r.Category)
	fmt.Printf("Classification confidence: %.0f%% | Context confidence: %.0f%% | Decision quality: %s\n",
		r.ClassificationConfidence*100, r.ContextConfidence*100, r.DecisionQuality)

	printNetworkContext(r.DetectedNetworkContext)

	fmt.Println("\nScores:")
	for _, ts := range sortedScores(r.Scores) {
		fmt.Printf("  %-10s %.2f\n", ts.Type, ts.Score)
	}

	if len(r.Alternatives) > 0 {
		fmt.Println("\nAlternatives:")
		for _, a := range r.Alternatives {
			fmt.Printf("  %-10s %.0f%%\n", a.Type, a.Score*100)
		}
	}

	if len(r.UncertaintyReasons) > 0 {
		fmt.Println("\nWhy no definite verdict:")
		for _, reason := range r.UncertaintyReasons {
			fmt.Printf("  - %s\n", reason)
		}
	}

	fmt.Println("\nEvidence / explanation:")
	for _, e := range r.Explanation {
		fmt.Printf("  - %s\n", e)
	}

	if len(r.NextBestProbes) > 0 {
		fmt.Println("\nNext best probes:")
		for _, p := range r.NextBestProbes {
			fmt.Printf("  - %s: %s\n", p.ProbeName, p.Reason)
		}
	}

	var failed []string
	for _, p := range r.Evidence {
		if p.Status == models.StatusFailed {
			failed = append(failed, p.ProbeName)
		}
	}
	if len(failed) > 0 {
		fmt.Printf("\nFailed probes (ignored): %v\n", failed)
	}
}

func printNetworkContext(nc *models.NetworkContext) {
	if nc == nil {
		return
	}
	if nc.ISP != "" {
		fmt.Printf("Operator: %s\n", nc.ISP)
	}
	if nc.LocalAccess != "" {
		fmt.Printf("Local access: %s", nc.LocalAccess)
		if nc.MainAdapter != "" {
			fmt.Printf(" (%s)", nc.MainAdapter)
		}
		fmt.Println()
	}
	if len(nc.GatewayChain) > 0 {
		fmt.Printf("Gateway chain: %v", nc.GatewayChain)
		if nc.DoubleNATPossible {
			fmt.Print("  [double NAT possible]")
		}
		fmt.Println()
	}
	fmt.Printf("UPnP modem found: %t | Fingerprint matched: %t\n", nc.UPnPFound, nc.FingerprintMatched)
	printLineProfile(nc.LineProfile)
}

// printLineProfile shows the fine-grained physical-layer reading (VDSL2 profile,
// DSL line stats, DOCSIS, PON optical) when authorized CPE telemetry exposed it.
func printLineProfile(lp *models.LineProfile) {
	if lp == nil {
		return
	}
	label := lp.Subtype
	if label == "" {
		label = lp.Technology
	}
	fmt.Printf("Line profile: %s [read from CPE: %s, confidence %.0f%%]\n", label, lp.Source, lp.Confidence*100)
	if d := lp.DSL; d != nil {
		if d.Profile != "" {
			fmt.Printf("  VDSL2 profile: %s (~%.1f MHz)%s\n", d.Profile, d.ProfileBandMHz, vectoringNote(d.Vectoring))
		}
		if d.SyncDownKbps > 0 || d.SyncUpKbps > 0 {
			fmt.Printf("  Sync rate: %.1f / %.1f Mbps (attainable %.1f / %.1f)\n",
				kbpsToMbps(d.SyncDownKbps), kbpsToMbps(d.SyncUpKbps),
				kbpsToMbps(d.AttainableDownKbps), kbpsToMbps(d.AttainableUpKbps))
		}
		if d.SNRMarginDownDB != 0 || d.AttenuationDownDB != 0 {
			fmt.Printf("  SNR margin: %.1f dB | Attenuation: %.1f dB (downstream)\n", d.SNRMarginDownDB, d.AttenuationDownDB)
		}
	}
	if c := lp.DOCSIS; c != nil {
		fmt.Printf("  DOCSIS: %s | channels %d down / %d up | MER %.1f dB\n", c.Version, c.DownstreamChannels, c.UpstreamChannels, c.SNRMERdB)
	}
	if pon := lp.PON; pon != nil {
		fmt.Printf("  PON: %s | optical Rx %.1f dBm / Tx %.1f dBm | ONT %s\n", pon.Type, pon.RxPowerDBm, pon.TxPowerDBm, pon.ONTModel)
	}
}

func vectoringNote(v bool) string {
	if v {
		return " — vectoring"
	}
	return ""
}

func kbpsToMbps(kbps int64) float64 { return float64(kbps) / 1000.0 }

func sortedScores(scores map[string]float64) []models.TypeScore {
	out := make([]models.TypeScore, 0, len(scores))
	for t, s := range scores {
		out = append(out, models.TypeScore{Type: t, Score: s})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Type < out[j].Type
	})
	return out
}

func strSlice(v any) []string {
	switch arr := v.(type) {
	case []string:
		return arr
	case []any:
		out := make([]string, 0, len(arr))
		for _, e := range arr {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
