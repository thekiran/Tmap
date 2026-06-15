package probes

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/thekiran/iad/internal/network"
	"github.com/thekiran/iad/pkg/models"
)

// ASNProbe looks up the reverse-DNS (PTR) name and RDAP org for the public IP.
// ISP PTR hostnames frequently embed the access technology (e.g.
// "host-x.vdsl.example.net", "lte-...", "ftth-..."), which is a useful hint.
// It makes outbound calls, so it is a deep+online probe only.
type ASNProbe struct{}

func (ASNProbe) Name() string { return "asn_probe" }

// techKeywords maps substrings found in PTR/org strings to an access-type hint.
var techKeywords = map[string]string{
	"ftth": models.TypeFiber, "fibre": models.TypeFiber, "fiber": models.TypeFiber, "gpon": models.TypeGPON,
	"vdsl": models.TypeVDSL, "adsl": models.TypeADSL, "dsl": models.TypeDSL,
	"docsis": models.TypeCable, "cable": models.TypeCable, "kabel": models.TypeCable,
	"lte": models.TypeMobile, "umts": models.TypeMobile, "mobile": models.TypeMobile, "gprs": models.TypeMobile,
	"wisp": models.TypeWISP, "wimax": models.TypeWISP, "wireless": models.TypeWISP,
	"starlink": models.TypeSatellite, "vsat": models.TypeSatellite,
}

func (p ASNProbe) Run(ctx context.Context, in models.ScanInput) (*models.ProbeResult, error) {
	res := newResult(p.Name())

	ip, err := network.PublicIP(ctx)
	if err != nil {
		return res, err
	}
	res.Evidence["public_ip"] = ip

	var scanText strings.Builder

	if names, err := net.DefaultResolver.LookupAddr(ctx, ip); err == nil && len(names) > 0 {
		ptr := strings.TrimSuffix(names[0], ".")
		res.Evidence["ptr"] = ptr
		scanText.WriteString(strings.ToLower(ptr))
		scanText.WriteByte(' ')
	}

	if org := rdapOrg(ctx, ip); org != "" {
		res.Evidence["org"] = org
		scanText.WriteString(strings.ToLower(org))
	}

	for kw, hint := range techKeywords {
		if strings.Contains(scanText.String(), kw) {
			res.Hints = appendUnique(res.Hints, hint)
		}
	}
	if len(res.Hints) > 0 {
		res.Confidence = 0.55
	} else {
		res.Confidence = 0.2
	}
	return res, nil
}

// rdapOrg fetches the RDAP record for ip and returns its network name, used as a
// coarse ISP/org string. Best-effort: any failure returns an empty string.
func rdapOrg(ctx context.Context, ip string) string {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://rdap.org/ip/"+ip, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "application/rdap+json")
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return ""
	}
	var rec struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &rec); err != nil {
		return ""
	}
	return rec.Name
}
