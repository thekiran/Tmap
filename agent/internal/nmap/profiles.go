package nmap

// Scan profiles. All use TCP connect scans (-sT), which do not require
// root/admin and are the opposite of stealth, and request XML on stdout. None
// enable IDS evasion, decoys, fragmentation, spoofing, or aggressive OS/script
// scanning that needs elevated privileges.
const (
	ProfileQuick    = "quick"
	ProfileStandard = "standard"
	ProfileDeep     = "deep"
)

// profileArgs returns the Nmap arguments for a profile (excluding the target,
// which the runner appends). Unknown profiles fall back to quick.
func profileArgs(profile string) []string {
	switch profile {
	case ProfileDeep:
		// Connect scan of the first 2000 ports with light service/version
		// detection. No -O / -A (those need root and add intrusive probing).
		return []string{"-sT", "-T4", "-p", "1-2000", "-sV", "--version-light", "-oX", "-"}
	case ProfileStandard:
		return []string{"-sT", "-T4", "--top-ports", "1000", "-sV", "--version-light", "-oX", "-"}
	default: // quick
		return []string{"-sT", "-T4", "-F", "-oX", "-"}
	}
}

// KnownProfile reports whether p is a recognized profile name.
func KnownProfile(p string) bool {
	switch p {
	case ProfileQuick, ProfileStandard, ProfileDeep:
		return true
	default:
		return false
	}
}
