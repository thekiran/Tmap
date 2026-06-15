package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/thekiran/iad/internal/detection"
	"github.com/thekiran/iad/internal/discovery"
	"github.com/thekiran/iad/internal/nmap"
	"github.com/thekiran/iad/internal/probes"
	"github.com/thekiran/iad/internal/system"
	"github.com/thekiran/iad/internal/topology"
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
	profile := fs.String("profile", "quick", "scan profile: quick | standard | deep")
	output := fs.String("output", "", "write JSON report to this file")
	outputShort := fs.String("o", "", "write JSON report to this file")
	iface := fs.String("interface", "", "interface name to scan from")
	includeVirtual := fs.Bool("include-virtual", false, "allow virtual adapters when selecting an interface")
	classify := fs.Bool("classify", false, "include access-type classification")
	useNmap := fs.Bool("nmap", false, "use optional Nmap service discovery when available")
	allowPublic := fs.Bool("allow-public", false, "permit non-private scopes")
	timeout := fs.Duration("timeout", 30*time.Second, "overall scan timeout")
	rulesDir := fs.String("rules", "", "rules directory for --classify")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *output == "" {
		*output = *outputShort
	}
	if *profile != "quick" && *profile != "standard" && *profile != "deep" {
		return fatalf("invalid --profile %q (use quick, standard, or deep)", *profile)
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
	if *useNmap {
		opts.Service = nmapService{runner: nmap.Runner{}, profile: *profile}
	}

	report, err := discovery.Run(ctx, opts)
	if err != nil {
		return err
	}
	if *classify {
		classification, err := runClassification(ctx, *profile, *rulesDir)
		if err != nil {
			return err
		}
		report.AccessClassification = &classification
	}

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

type nmapService struct {
	runner  nmap.Runner
	profile string
}

func (s nmapService) Available() bool {
	return s.runner.Available()
}

func (s nmapService) Scan(ctx context.Context, scope models.ScanScope) ([]discovery.ScannedHost, error) {
	hosts, err := s.runner.Scan(ctx, scope.CIDR, s.profile)
	if err != nil {
		return nil, err
	}
	out := make([]discovery.ScannedHost, 0, len(hosts))
	for _, h := range hosts {
		sh := discovery.ScannedHost{IP: h.IP, MAC: h.MAC, Hostname: h.Hostname}
		for _, p := range h.Ports {
			sh.Services = append(sh.Services, models.Service{
				Port:     p.ID,
				Protocol: p.Protocol,
				State:    "open",
				Name:     p.Service,
				Product:  p.Product,
			})
		}
		out = append(out, sh)
	}
	return out, nil
}
