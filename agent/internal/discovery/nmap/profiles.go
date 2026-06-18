package nmap

// Scan profiles. All use TCP connect scans (-sT), which do not require
// root/admin and are the opposite of stealth, and request XML on stdout. None
// enable IDS evasion, decoys, fragmentation, spoofing, or aggressive OS/script
// scanning that needs elevated privileges.
const (
	ProfileQuick    = "quick"
	ProfileStandard = "standard"
	ProfileNormal   = "normal"
	ProfileDeep     = "deep"
	ProfileFull     = "full"
)

// ArgsForProfile returns Nmap arguments for a profile, excluding the target.
// Unknown profiles fall back to quick.
func ArgsForProfile(profile string) []string {
	switch profile {
	case ProfileDeep, ProfileFull:
		return []string{"-sT", "-T4", "-p", "1-2000", "-sV", "--version-light", "-oX", "-"}
	case ProfileStandard, ProfileNormal:
		return []string{"-sT", "-T4", "--top-ports", "1000", "-sV", "--version-light", "-oX", "-"}
	default:
		return []string{"-sT", "-T4", "-F", "-oX", "-"}
	}
}

// KnownProfile reports whether p is a recognized profile name.
func KnownProfile(p string) bool {
	switch p {
	case ProfileQuick, ProfileStandard, ProfileNormal, ProfileDeep, ProfileFull:
		return true
	default:
		return false
	}
}
