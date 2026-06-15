//go:build windows

package platform

const Name = "windows"

func SupportsPassiveLLDP() bool { return false }
