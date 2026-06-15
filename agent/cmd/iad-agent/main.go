// Command iad-agent is an authorized local network discovery and topology mapper.
// It is a thin CLI: each subcommand wires existing packages (discovery, topology,
// nmap, detection) together. No discovery, scoring, or inference logic lives here.
//
// Subcommands:
//
//	iad-agent interfaces                                   list local interfaces
//	iad-agent scan --cidr auto --profile quick -o out.json scan a subnet → topology
//	iad-agent validate --input out.json                    validate a saved report
//	iad-agent version                                      print version
package main

import (
	"fmt"
	"os"
)

// Version is the agent version reported in scan output.
const Version = "0.2.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd, args := os.Args[1], os.Args[2:]
	var err error
	switch cmd {
	case "interfaces", "ifaces":
		err = runInterfaces(args)
	case "scan":
		err = runScan(args)
	case "validate":
		err = runValidate(args)
	case "version", "--version", "-v":
		runVersion()
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `iad-agent - authorized local network discovery & topology mapper

Usage:
  iad-agent interfaces [--json]
  iad-agent scan --cidr <auto|CIDR> --profile <quick|standard|deep> --output <file>
                 [--interface <name>] [--include-virtual] [--classify] [--nmap]
                 [--allow-public] [--timeout <dur>]
  iad-agent validate --input <file>
  iad-agent version

Safety:
  By default only the selected PRIVATE subnet is scanned. Public/non-private
  ranges are refused unless --allow-public is set (only use it on networks you
  are authorized to scan). No stealth, evasion, exploitation, or brute force.
`)
}

func runVersion() {
	fmt.Printf("iad-agent %s\n", Version)
}

// fatalf is a small helper for command-local fatal errors with context.
func fatalf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
