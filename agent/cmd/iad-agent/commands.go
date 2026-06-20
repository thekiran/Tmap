package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thekiran/iad/internal/detection"
	"github.com/thekiran/iad/internal/deviceintel"
	"github.com/thekiran/iad/internal/discovery"
	"github.com/thekiran/iad/internal/discovery/nmap"
	"github.com/thekiran/iad/internal/network"
	"github.com/thekiran/iad/internal/probes"
	"github.com/thekiran/iad/internal/system"
	"github.com/thekiran/iad/internal/topology"
	"github.com/thekiran/iad/internal/upstream"
	"github.com/thekiran/iad/pkg/models"
)

func runInterfaces(args []string) error {
	fs := flag.NewFlagSet("interfaces", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "print interfaces as JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ifaces, err := discovery.Interfaces()
	if err != nil {
		return err
	}
	if *asJSON {
		data, err := json.MarshalIndent(ifaces, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
	for _, ifc := range ifaces {
		state := "down"
		if ifc.Up {
			state = "up"
		}
		selected := ""
		if ifc.Selected {
			selected = " selected"
		}
		fmt.Printf("%s\t%s\t%s\tvirtual=%t%s\n", ifc.Name, state, ifc.CIDR, ifc.Virtual, selected)
	}
	return nil
}

func runScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	cidr := fs.String("cidr", "auto", "scan scope: auto or CIDR")
	profile := fs.String("profile", "quick", "scan profile: quick | normal | standard | deep | full")
	output := fs.String("output", "", "write JSON report to this file")
	outputShort := fs.String("o", "", "write JSON report to this file")
	iface := fs.String("interface", "", "interface name to scan from")
	includeVirtual := fs.Bool("include-virtual", false, "allow virtual adapters when selecting an interface")
	classify := fs.Bool("classify", false, "include access-type classification")
	useNmap := fs.Bool("nmap", false, "use optional Nmap service discovery when available")
	nmapBin := fs.String("nmap-bin", "", "path to nmap executable (defaults to PATH lookup)")
	full := fs.Bool("full", false, "write a complete single-file JSON report with all available sections and probes")
	redactionMode := fs.String("redaction-mode", "none", "privacy redaction mode: none or safe_to_share")
	maskPublicIP := fs.Bool("mask-public-ip", false, "mask public IP fields in the JSON report")
	maskMAC := fs.Bool("mask-mac", false, "mask MAC addresses in the JSON report")
	maskHostnames := fs.Bool("mask-hostnames", false, "mask hostnames in the JSON report")
	allowPublic := fs.Bool("allow-public", false, "permit non-private scopes")
	timeout := fs.Duration("timeout", 30*time.Second, "overall scan timeout")
	rulesDir := fs.String("rules", "", "rules directory for --classify")
	enrichUpstream := fs.Bool("enrich-upstream", true, "run the read-only upstream gateway enrichment phase (safe ping + TCP/HTTP/TLS fingerprint of private gateway/CPE candidates)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	profileSet, timeoutSet := false, false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "profile":
			profileSet = true
		case "timeout":
			timeoutSet = true
		}
	})
	if *full {
		applyFullScanDefaults(profileSet, timeoutSet, profile, classify, useNmap, timeout)
	}
	if *output == "" {
		*output = *outputShort
	}
	if !nmap.KnownProfile(*profile) {
		return fatalf("invalid --profile %q (use quick, normal, standard, deep, or full)", *profile)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	opts := discovery.Options{
		RequestedCIDR:  *cidr,
		Profile:        *profile,
		AllowPublic:    *allowPublic,
		InterfaceName:  *iface,
		IncludeVirtual: *includeVirtual,
		Version:        Version,
	}
	if host, err := os.Hostname(); err == nil {
		opts.Hostname = host
	}
	opts.OS = system.Info().OS
	// Wire real gateway + DNS detection so the scan can find the default route
	// (used to pick the correct interface and to draw gateway/default edges) and
	// record the resolvers in the report.
	opts.GatewayFn = network.Gateway
	opts.DNSFn = network.DNSServers
	var nmapWarning *models.Warning
	nmapAvailable := false
	if *useNmap {
		runner := nmap.Runner{Binary: *nmapBin}
		if runner.Available() {
			nmapAvailable = true
			opts.Service = nmap.ServiceScanner{Runner: runner, Profile: *profile}
		} else {
			warning := nmapUnavailableWarning(*nmapBin)
			nmapWarning = &warning
		}
	}

	report, err := discovery.Run(ctx, opts)
	if err != nil {
		return err
	}
	if nmapWarning != nil {
		report.Warnings = append(report.Warnings, *nmapWarning)
		fmt.Fprintln(os.Stderr, "warning:", nmapWarning.Message)
	}
	if *classify {
		classification, err := runClassification(ctx, *profile, *rulesDir)
		if err != nil {
			return err
		}
		report.AccessClassification = &classification
	}
	// Dedicated upstream-gateway enrichment: actively (but read-only) probe the
	// private gateway/upstream/CPE candidates the LAN sweep never reaches (e.g.
	// an off-subnet 192.168.1.1) so they get real reachability/service/HTTP/TLS
	// evidence instead of an almost-empty card. Evidence is appended to the
	// report BEFORE Build so the normal ingestion path surfaces services; the
	// classification/reachability/routing are attached to the device AFTER Build.
	var upstreamIntel map[string]upstream.IntelResult
	if *enrichUpstream {
		// Give enrichment its own bounded budget (independent of how much the
		// discovery phase already consumed) so reachability/HTTP/TLS probes
		// actually get to run. The desktop kills the process on cancel, so a
		// Background-derived deadline never outlives a user cancel.
		ictx, icancel := context.WithTimeout(context.Background(), 20*time.Second)
		upstreamIntel = upstream.EnrichReport(ictx, report, upstream.Options{EnablePing: true})
		icancel()
		for _, r := range upstreamIntel {
			report.Evidence = append(report.Evidence, r.Evidence...)
		}
	}

	deviceIntel := deviceintel.Build(report)
	if len(upstreamIntel) > 0 {
		attachUpstreamIntel(&deviceIntel, upstreamIntel)
	}
	report.DeviceIntel = &deviceIntel
	enrichReport(&report, reportEnrichmentOptions{
		Full:                    *full,
		Profile:                 *profile,
		ClassificationRequested: *classify,
		NmapRequested:           *useNmap,
		NmapAvailable:           nmapAvailable,
		IncludeVirtual:          *includeVirtual,
		RedactionMode:           *redactionMode,
		MaskPublicIP:            *maskPublicIP,
		MaskMAC:                 *maskMAC,
		MaskHostnames:           *maskHostnames,
	})

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if *output != "" {
		return os.WriteFile(*output, data, 0o644)
	}
	fmt.Println(string(data))
	return nil
}

func runValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	input := fs.String("input", "", "JSON topology report to validate")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" {
		return fatalf("validate requires --input <file>")
	}
	data, err := os.ReadFile(*input)
	if err != nil {
		return err
	}
	var report models.ScanReport
	if err := json.Unmarshal(data, &report); err != nil {
		return err
	}
	problems := topology.ValidateReport(report)
	if len(problems) == 0 {
		fmt.Println("valid")
		return nil
	}
	for _, p := range problems {
		fmt.Println(p)
	}
	return fatalf("validation failed with %d problem(s)", len(problems))
}

func runClassification(ctx context.Context, profile, rulesDir string) (models.ScanResult, error) {
	dir, err := resolveRulesDir(rulesDir)
	if err != nil {
		return models.ScanResult{}, err
	}
	mode := models.ModeQuick
	if profile == "deep" {
		mode = models.ModeDeep
	} else if profile == "full" {
		mode = models.ModeFull
	}
	in := models.ScanInput{Mode: mode, Online: true, RulesDir: dir}
	runner := probes.Runner{Probes: probes.Default(in), Timeout: 12 * time.Second}
	results := runner.Run(ctx, in)
	engine, err := detection.NewEngine(dir)
	if err != nil {
		return models.ScanResult{}, err
	}
	return engine.Analyze(in, results), nil
}

func applyFullScanDefaults(profileSet, timeoutSet bool, profile *string, classify, useNmap *bool, timeout *time.Duration) {
	*classify = true
	*useNmap = true
	if !profileSet {
		*profile = "full"
	}
	if !timeoutSet {
		*timeout = 60 * time.Second
	}
}

type reportCapabilityOptions struct {
	Full                    bool
	Profile                 string
	ClassificationRequested bool
	NmapRequested           bool
	NmapAvailable           bool
}

func buildReportCapabilities(report models.ScanReport, opts reportCapabilityOptions) []models.ReportCapability {
	caps := []models.ReportCapability{
		{
			Name:        "full_json_report",
			Category:    "report",
			Status:      "completed",
			OutputPath:  "/",
			Description: "Single JSON document containing scan metadata, topology, evidence, warnings, device intelligence, optional access classification, and this capability manifest.",
		},
		{
			Name:        "interface_inventory",
			Category:    "topology",
			Status:      "completed",
			OutputPath:  "/agent/interfaces",
			Description: "Lists local interfaces, selected scan interface, addresses, gateway, and DNS context.",
		},
		{
			Name:        "topology_scan",
			Category:    "topology",
			Status:      "completed",
			OutputPath:  "/topology,/devices,/edges,/evidence,/summary",
			Description: "Builds the evidence-backed LAN device and iad.topology/v2 graph inside the validated scope.",
		},
		{
			Name:        "topology_v2_graph",
			Category:    "topology",
			Status:      "completed",
			OutputPath:  "/topology/nodes,/topology/edges",
			Description: "Frontend-ready graph with evidence, confidence, explanations, warnings, wireless metadata, and conservative edge classification.",
		},
		{
			Name:        "tcp_lan_sweep",
			Category:    "topology",
			Status:      "completed",
			OutputPath:  "/devices",
			Description: fmt.Sprintf("Uses the %s profile to discover live hosts and common open services.", opts.Profile),
		},
		{
			Name:        "passive_lan_observer",
			Category:    "passive_observation",
			Status:      "available",
			OutputPath:  "/raw_observations",
			Description: "Metadata-only passive LAN observation abstraction is available; payload bodies, cookies, tokens, and credentials are not stored.",
		},
		{
			Name:        "mobile_fingerprint_engine",
			Category:    "device_intel",
			Status:      "completed",
			OutputPath:  "/devices/mobileFingerprint,/device_intel/devices/mobileFingerprint,/topology/nodes/mobileFingerprint",
			Description: "Scores iOS/iPadOS and Android separately from local metadata, keeps evidence and conflicts visible, and avoids unsafe probes or fake conclusions.",
		},
		{
			Name:       "passive_wifi_observer",
			Category:   "wireless",
			Status:     "unsupported",
			OutputPath: "/wireless,/warnings",
			Reason:     "No monitor/radiotap backend is configured by default; unsupported is expected on many Windows adapters.",
		},
	}

	if report.DeviceIntel != nil {
		caps = append(caps, models.ReportCapability{
			Name:        "device_intelligence",
			Category:    "device_intel",
			Status:      "completed",
			OutputPath:  "/device_intel",
			Description: "Normalizes discovered devices into inventory, service, risk, OS, vendor, and UI-oriented records.",
		})
	} else {
		caps = append(caps, models.ReportCapability{
			Name:       "device_intelligence",
			Category:   "device_intel",
			Status:     "skipped",
			OutputPath: "/device_intel",
			Reason:     "device intelligence section was not built for this run",
		})
	}

	switch {
	case opts.NmapRequested && opts.NmapAvailable:
		caps = append(caps, models.ReportCapability{
			Name:        "nmap_service_discovery",
			Category:    "optional_tool",
			Status:      "completed",
			OutputPath:  "/devices/services",
			Description: "Optional Nmap TCP connect service discovery was enabled and merged into device services.",
		})
	case opts.NmapRequested:
		caps = append(caps, models.ReportCapability{
			Name:       "nmap_service_discovery",
			Category:   "optional_tool",
			Status:     "unavailable",
			OutputPath: "/warnings",
			Reason:     "Nmap was requested but no usable nmap binary was found",
		})
	default:
		caps = append(caps, models.ReportCapability{
			Name:     "nmap_service_discovery",
			Category: "optional_tool",
			Status:   "skipped",
			Reason:   "enable with --nmap or --full",
		})
	}

	if report.AccessClassification == nil {
		cap := models.ReportCapability{
			Name:     "access_classification",
			Category: "access_detection",
			Status:   "skipped",
			Reason:   "enable with --classify or --full",
		}
		if opts.ClassificationRequested {
			cap.Reason = "access classification was requested but no classification section was attached"
		}
		caps = append(caps, cap)
		return caps
	}

	caps = append(caps, models.ReportCapability{
		Name:        "access_classification",
		Category:    "access_detection",
		Status:      "completed",
		OutputPath:  "/access_classification",
		Description: "Runs access-type probes and the scoring engine for DSL/VDSL/Fiber/Cable/WISP/Mobile/Satellite/Enterprise candidates.",
	})
	if report.AccessClassification.ModemCollection != nil {
		caps = append(caps, models.ReportCapability{
			Name:        "modem_collection",
			Category:    "access_detection",
			Status:      "completed",
			OutputPath:  "/access_classification/modem_collection",
			Description: "Collects CPE candidates, gateway chain, NAT state, WAN evidence, and missing physical proof in one section.",
		})
	}
	for i, result := range report.AccessClassification.Evidence {
		caps = append(caps, models.ReportCapability{
			Name:       result.ProbeName,
			Category:   "access_probe",
			Status:     capabilityStatusFromProbe(result),
			OutputPath: fmt.Sprintf("/access_classification/evidence/%d", i),
			Reason:     capabilityReasonFromProbe(result),
		})
	}
	if opts.Full {
		caps = append(caps, models.ReportCapability{
			Name:        "full_mode",
			Category:    "report",
			Status:      "completed",
			Description: "The --full flag enabled deep profiling, access classification, device intelligence, Nmap when available, and all JSON report sections.",
		})
	}
	return caps
}

func capabilityStatusFromProbe(result models.ProbeResult) string {
	switch result.Status {
	case models.StatusSuccess:
		return "completed"
	case models.StatusFailed:
		return "failed"
	case models.StatusSkipped:
		return "skipped"
	default:
		if result.Status != "" {
			return result.Status
		}
		return "unknown"
	}
}

func capabilityReasonFromProbe(result models.ProbeResult) string {
	if len(result.Errors) > 0 {
		return strings.Join(result.Errors, "; ")
	}
	if result.Status == models.StatusSkipped && len(result.Hints) > 0 {
		return strings.Join(result.Hints, "; ")
	}
	return ""
}

func nmapUnavailableWarning(binary string) models.Warning {
	message := "--nmap was requested, but nmap was not found in PATH. Install Nmap or pass --nmap-bin <path>; continuing without Nmap service discovery."
	if binary != "" {
		message = fmt.Sprintf("--nmap was requested, but the configured Nmap binary %q was not found; continuing without Nmap service discovery.", binary)
	}
	return models.Warning{
		Code:     "nmap_unavailable",
		Severity: models.SeverityWarning,
		Message:  message,
	}
}

func resolveRulesDir(flagDir string) (string, error) {
	if flagDir != "" {
		return flagDir, nil
	}
	candidates := []string{"rules", filepath.Join("..", "rules"), filepath.Join("..", "..", "rules")}
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
	return "", fatalf("could not locate rules directory; pass --rules <dir>")
}

// attachUpstreamIntel copies the upstream enrichment results onto the matching
// device-intel devices (by IP). Classification confidence only ever raises the
// device confidence — it never lowers an already-stronger signal.
func attachUpstreamIntel(intel *models.DeviceIntelReport, results map[string]upstream.IntelResult) {
	for i := range intel.Devices {
		d := &intel.Devices[i]
		for _, ip := range d.IPAddresses {
			r, ok := results[ip]
			if !ok {
				continue
			}
			reach := r.Reachability
			routing := r.Routing
			d.Reachability = &reach
			d.RoutingEvidence = &routing
			d.ClassificationTags = r.Classification.Tags
			d.IntelEvidence = r.Classification.Evidence
			d.EnrichmentWarnings = appendUniqueStrings(d.EnrichmentWarnings, r.Warnings...)
			if r.Classification.Confidence > d.Confidence {
				d.Confidence = r.Classification.Confidence
			}
			break
		}
	}
}

func appendUniqueStrings(values []string, next ...string) []string {
	seen := make(map[string]bool, len(values))
	for _, v := range values {
		seen[v] = true
	}
	for _, v := range next {
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		values = append(values, v)
	}
	return values
}
