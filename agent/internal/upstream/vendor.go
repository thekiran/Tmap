package upstream

import "strings"

// VendorResult is the outcome of analyzing an HTTP/HTTPS surface. It never
// invents a model name — Vendor is only set when a known keyword is present.
type VendorResult struct {
	Vendor     string
	RouterLike bool
	AdminPanel bool
	ONT        bool
	Modem      bool
	Hints      []string
}

// vendorKeywords maps a display vendor name to substrings that, when present in
// HTTP headers / title / realm / body, indicate that maker. Kept conservative:
// these are well-known router/CPE strings, not guesses.
var vendorKeywords = []struct {
	name string
	keys []string
}{
	{"TP-Link", []string{"tp-link", "tplink", "archer", "tl-wr", "tl-wdr"}},
	{"Keenetic", []string{"keenetic", "ndms"}},
	{"Xiaomi", []string{"xiaomi", "miwifi", "mi router", "mi wifi"}},
	{"Huawei", []string{"huawei", "echolife", "hg8", "ws5", "hilink"}},
	{"ZTE", []string{"zte", "zxhn", "f660", "f670"}},
	{"Zyxel", []string{"zyxel"}},
	{"ASUS", []string{"asus", "asuswrt", "rt-ac", "rt-ax"}},
	{"MikroTik", []string{"mikrotik", "routeros", "webfig"}},
	{"Tenda", []string{"tenda"}},
	{"Netis", []string{"netis"}},
	{"FiberHome", []string{"fiberhome", "an5506"}},
	{"Nokia/Alcatel", []string{"nokia", "alcatel", "isam", "beacon", "g-240", "g-010"}},
	{"Technicolor", []string{"technicolor", "thomson"}},
	{"Arris", []string{"arris", "touchstone"}},
	{"Ubee", []string{"ubee"}},
	{"Sagemcom", []string{"sagemcom", "fast 5", "fast 3"}},
	{"Netgear", []string{"netgear", "nighthawk", "orbi"}},
	{"D-Link", []string{"d-link", "dlink", "dir-"}},
	{"Ubiquiti", []string{"ubiquiti", "unifi", "edgeos", "edgerouter"}},
	{"DrayTek", []string{"draytek", "vigor"}},
}

// adminKeywords indicate a router/admin login or management surface — used to
// flag an admin panel WITHOUT ever attempting a login.
var adminKeywords = []string{
	"login", "log in", "router", "admin", "gateway", "web management",
	"management console", "authentication required", "setup wizard",
	"router settings", "broadband", "control panel", "web ui", "webui",
}

var ontKeywords = []string{"ont", "gpon", "epon", "xpon", "optical network", "olt"}
var modemKeywords = []string{"modem", "docsis", "cable modem", "voice modem", "dsl modem"}

// AnalyzeHTTP inspects the safe, read-only HTTP signals (Server header, page
// title, WWW-Authenticate realm, and a truncated body) and reports vendor and
// router/admin indicators. All inputs are already-captured strings; this does
// no network I/O.
func AnalyzeHTTP(server, title, realm, body string) VendorResult {
	hay := strings.ToLower(strings.Join([]string{server, title, realm, body}, " \n "))
	var res VendorResult

	for _, v := range vendorKeywords {
		for _, k := range v.keys {
			if strings.Contains(hay, k) {
				res.Vendor = v.name
				res.Hints = appendUnique(res.Hints, "vendor:"+k)
				break
			}
		}
		if res.Vendor != "" {
			break
		}
	}

	for _, k := range adminKeywords {
		if strings.Contains(hay, k) {
			res.AdminPanel = true
			res.Hints = appendUnique(res.Hints, "admin:"+k)
			break
		}
	}
	// A WWW-Authenticate realm is itself a (non-login) admin-surface signal.
	if strings.TrimSpace(realm) != "" {
		res.AdminPanel = true
	}

	for _, k := range ontKeywords {
		if strings.Contains(hay, k) {
			res.ONT = true
			res.Hints = appendUnique(res.Hints, "ont:"+k)
			break
		}
	}
	for _, k := range modemKeywords {
		if strings.Contains(hay, k) {
			res.Modem = true
			res.Hints = appendUnique(res.Hints, "modem:"+k)
			break
		}
	}

	res.RouterLike = res.Vendor != "" || res.AdminPanel || res.ONT || res.Modem
	return res
}
