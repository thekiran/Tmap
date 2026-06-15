//go:build darwin

package platform

const Name = "darwin"

func SupportsPassiveLLDP() bool { return false }
