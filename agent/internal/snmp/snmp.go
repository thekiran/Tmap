// Package snmp defines the interface for read-only SNMP collection used to obtain
// authoritative topology evidence (sysName/sysDescr, interfaces, LLDP neighbour
// tables, bridge/forwarding tables).
//
// SNMP is OFF by default and requires the operator to explicitly supply read-only
// credentials. This package intentionally ships only the interface and a default
// "unavailable" collector: it never guesses community strings, never brute-forces,
// and never fabricates LLDP/CDP/bridge data. A real implementation (e.g. backed by
// gosnmp) is wired in behind the Collector interface only when credentials are
// provided.
package snmp

import (
	"context"
	"errors"
)

// ErrUnavailable indicates SNMP collection is not configured/available. Callers
// must treat this as "no SNMP evidence" and fall back to inferred topology — not
// as an error to paper over with fake data.
var ErrUnavailable = errors.New("snmp: not configured (provide explicit read-only credentials to enable)")

// Version selects the SNMP protocol version.
type Version string

const (
	V2c Version = "v2c"
	V3  Version = "v3"
)

// Credentials are the explicitly-provided, read-only SNMP credentials. There is
// no default community; an empty Credentials yields an unavailable collector.
type Credentials struct {
	Version   Version
	Community string // v2c only
	// v3 USM parameters
	Username     string
	AuthProtocol string // e.g. SHA, SHA256
	AuthPassword string
	PrivProtocol string // e.g. AES, AES256
	PrivPassword string
}

// Configured reports whether the credentials are sufficient to attempt SNMP.
func (c Credentials) Configured() bool {
	switch c.Version {
	case V2c:
		return c.Community != ""
	case V3:
		return c.Username != ""
	default:
		return false
	}
}

// SystemInfo is the SNMPv2-MIB system group.
type SystemInfo struct {
	SysName     string
	SysDescr    string
	SysObjectID string
}

// Interface is one row of the IF-MIB ifTable.
type Interface struct {
	Index int
	Name  string
	Descr string
	MAC   string
	Speed uint64
}

// LLDPNeighbor is one row of the LLDP remote-systems table: a directly attached
// neighbour as the device itself reports it. This is authoritative physical
// adjacency evidence.
type LLDPNeighbor struct {
	LocalPort       string
	RemoteChassisID string
	RemoteSysName   string
	RemotePortID    string
}

// BridgeEntry is one forwarding-database (FDB) entry: a MAC learned on a bridge
// port. Used to derive snmp_bridge edges.
type BridgeEntry struct {
	MAC  string
	Port int
}

// Collector reads read-only topology evidence over SNMP. Every method returns
// ErrUnavailable when SNMP is not configured; it must never invent data.
type Collector interface {
	SystemInfo(ctx context.Context) (SystemInfo, error)
	Interfaces(ctx context.Context) ([]Interface, error)
	LLDPNeighbors(ctx context.Context) ([]LLDPNeighbor, error)
	BridgeTable(ctx context.Context) ([]BridgeEntry, error)
}

// New returns a Collector for the target using the supplied credentials. Until a
// real SNMP backend (e.g. gosnmp) is compiled in, it returns an Unavailable
// collector whenever credentials are missing — and the build here never ships a
// silent fake, so callers always get honest ErrUnavailable rather than fabricated
// neighbours.
func New(target string, creds Credentials) Collector {
	if target == "" || !creds.Configured() {
		return Unavailable{Reason: "no read-only SNMP credentials were provided"}
	}
	// A credentialed backend is intentionally not bundled in this build. Returning
	// Unavailable (rather than fake data) keeps the agent honest; wiring gosnmp
	// here is the single, isolated change needed to light SNMP up.
	return Unavailable{Reason: "SNMP backend not compiled in (provide a gosnmp-backed Collector)"}
}

// Unavailable is the no-op Collector: every method reports ErrUnavailable. It is
// the safe default and is also handy in tests.
type Unavailable struct {
	Reason string
}

func (u Unavailable) err() error {
	if u.Reason != "" {
		return errors.New("snmp: " + u.Reason)
	}
	return ErrUnavailable
}

func (u Unavailable) SystemInfo(context.Context) (SystemInfo, error)     { return SystemInfo{}, u.err() }
func (u Unavailable) Interfaces(context.Context) ([]Interface, error)    { return nil, u.err() }
func (u Unavailable) LLDPNeighbors(context.Context) ([]LLDPNeighbor, error) { return nil, u.err() }
func (u Unavailable) BridgeTable(context.Context) ([]BridgeEntry, error) { return nil, u.err() }
