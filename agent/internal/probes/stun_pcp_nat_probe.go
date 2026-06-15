package probes

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/thekiran/iad/internal/network"
	"github.com/thekiran/iad/pkg/models"
)

type STUNPCPNATProbe struct {
	funcs stunPCPNATFuncs
}

type stunPCPNATFuncs struct {
	publicIP func(context.Context) (string, error)
	gateway  func() (net.IP, error)
	stun     func(context.Context, string) (string, int, error)
	udpProbe func(context.Context, string, string, []byte) bool
}

func (STUNPCPNATProbe) Name() string { return "stun_pcp_nat_probe" }

func (p STUNPCPNATProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())
	f := p.withDefaults()

	var publicIP string
	var stunIP string
	var stunPort int
	if in.Online {
		publicIP, _ = f.publicIP(ctx)
		stunIP, stunPort, _ = f.stun(ctx, "stun.l.google.com:19302")
	}

	pcpReachable := false
	natPMPReachable := false
	if gw, err := f.gateway(); err == nil && isRFC1918IPv4(gw.String()) {
		res.Evidence["gateway"] = gw.String()
		pcpReachable = f.udpProbe(ctx, gw.String(), "5351", []byte{2, 1, 0, 0})
		natPMPReachable = f.udpProbe(ctx, gw.String(), "5351", []byte{0, 0})
	}

	topology := models.NATTopology{
		PublicIP:                   publicIP,
		STUNPublicIP:               stunIP,
		STUNPublicPort:             stunPort,
		PublicIPMatches:            publicIP != "" && stunIP != "" && publicIP == stunIP,
		ExternalPublicIPConsistent: publicIP != "" && stunIP != "" && publicIP == stunIP,
		CGNAT:                      network.IsCGNAT(publicIP),
		PCPReachable:               pcpReachable,
		NATPMPReachable:            natPMPReachable,
		GatewayNATControlReachable: pcpReachable || natPMPReachable,
	}
	topology.Topology = natTopologyLabel(topology)
	res.Evidence["nat_topology"] = topology
	res.Evidence["network_confidence"] = 0.35
	res.Confidence = 0.35
	return res, nil
}

func (p STUNPCPNATProbe) withDefaults() stunPCPNATFuncs {
	f := p.funcs
	if f.publicIP == nil {
		f.publicIP = network.PublicIP
	}
	if f.gateway == nil {
		f.gateway = network.Gateway
	}
	if f.stun == nil {
		f.stun = stunMappedAddress
	}
	if f.udpProbe == nil {
		f.udpProbe = udpGatewayProbe
	}
	return f
}

func stunMappedAddress(ctx context.Context, server string) (string, int, error) {
	var tx [12]byte
	if _, err := rand.Read(tx[:]); err != nil {
		return "", 0, err
	}
	req := make([]byte, 20)
	binary.BigEndian.PutUint16(req[0:2], 0x0001)
	binary.BigEndian.PutUint16(req[2:4], 0)
	binary.BigEndian.PutUint32(req[4:8], 0x2112A442)
	copy(req[8:20], tx[:])
	d := net.Dialer{Timeout: 2 * time.Second}
	conn, err := d.DialContext(ctx, "udp", server)
	if err != nil {
		return "", 0, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := conn.Write(req); err != nil {
		return "", 0, err
	}
	buf := make([]byte, 1500)
	n, err := conn.Read(buf)
	if err != nil {
		return "", 0, err
	}
	return parseSTUNXORMapped(buf[:n], tx[:])
}

func parseSTUNXORMapped(msg, tx []byte) (string, int, error) {
	if len(msg) < 20 {
		return "", 0, fmt.Errorf("short stun response")
	}
	pos := 20
	for pos+4 <= len(msg) {
		attrType := binary.BigEndian.Uint16(msg[pos : pos+2])
		attrLen := int(binary.BigEndian.Uint16(msg[pos+2 : pos+4]))
		pos += 4
		if pos+attrLen > len(msg) {
			break
		}
		if attrType == 0x0020 && attrLen >= 8 {
			family := msg[pos+1]
			xport := binary.BigEndian.Uint16(msg[pos+2:pos+4]) ^ 0x2112
			if family == 0x01 {
				ip := net.IPv4(msg[pos+4]^0x21, msg[pos+5]^0x12, msg[pos+6]^0xA4, msg[pos+7]^0x42)
				return ip.String(), int(xport), nil
			}
		}
		pos += (attrLen + 3) &^ 3
	}
	return "", 0, fmt.Errorf("xor mapped address not found")
}

func udpGatewayProbe(ctx context.Context, ip, port string, payload []byte) bool {
	d := net.Dialer{Timeout: 600 * time.Millisecond}
	conn, err := d.DialContext(ctx, "udp", net.JoinHostPort(ip, port))
	if err != nil {
		return false
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(600 * time.Millisecond))
	if _, err := conn.Write(payload); err != nil {
		return false
	}
	buf := make([]byte, 64)
	_, err = conn.Read(buf)
	return err == nil
}

func natTopologyLabel(n models.NATTopology) string {
	switch {
	case n.CGNAT:
		return "cgnat_possible"
	case n.PublicIP != "" && n.STUNPublicIP != "" && !n.PublicIPMatches:
		return "public_ip_mismatch"
	case n.PCPReachable || n.NATPMPReachable:
		return "gateway_nat_control_reachable"
	case n.STUNPublicIP != "":
		return "stun_observed"
	default:
		return "unknown"
	}
}
