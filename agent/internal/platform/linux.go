//go:build linux

package platform

const Name = "linux"

func SupportsPassiveLLDP() bool { return true }
