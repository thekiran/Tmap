package nmap

import "github.com/thekiran/iad/pkg/models"

// MergeServices merges Nmap services into an existing service list without
// duplicating protocol/port/name tuples. Existing service metadata wins.
func MergeServices(existing []models.Service, discovered []models.Service) []models.Service {
	out := append([]models.Service(nil), existing...)
	seen := map[string]bool{}
	for _, service := range out {
		seen[serviceKey(service)] = true
	}
	for _, service := range discovered {
		key := serviceKey(service)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, service)
	}
	return out
}

// MergeIntoDevice enriches one device-like value with services from a matched
// Nmap host. It is intentionally small; full device identity merge stays in the
// discovery normalizer.
func MergeIntoDevice(device MergeDevice, host Host) MergeDevice {
	mapped := MapHosts([]Host{host})
	if len(mapped) == 0 {
		return device
	}
	device.Services = MergeServices(device.Services, mapped[0].Services)
	return device
}

func serviceKey(service models.Service) string {
	return service.Protocol + ":" + itoa(service.Port) + ":" + service.Name
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
