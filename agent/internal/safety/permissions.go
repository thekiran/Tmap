package safety

import (
	"fmt"

	"github.com/thekiran/iad/internal/model"
)

func ValidateScope(scope model.ScanScope) error {
	if scope.PublicScanning {
		return fmt.Errorf("public scanning is disabled by policy for topology discovery")
	}
	for _, cidr := range scope.ScannedRanges {
		if !IsPrivateCIDR(cidr) {
			return fmt.Errorf("refusing to scan non-private range %s", cidr)
		}
	}
	for _, cidr := range scope.LocalSubnets {
		if !IsPrivateCIDR(cidr) {
			return fmt.Errorf("local subnet %s is not private", cidr)
		}
	}
	return nil
}

func ValidatePrivateTargets(ips []string) error {
	for _, ip := range ips {
		if !IsPrivateIPString(ip) {
			return fmt.Errorf("refusing non-private target %s", ip)
		}
	}
	return nil
}
