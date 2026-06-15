package config

import (
	"time"

	"github.com/thekiran/iad/internal/model"
)

type Config struct {
	Mode            model.ScanMode
	GlobalTimeout   time.Duration
	PerProbeTimeout time.Duration
	RateLimit       RateLimit
	SNMP            *model.SNMPCredential
}

type RateLimit struct {
	RequestsPerSecond int
	Burst             int
}

func Default(mode model.ScanMode) Config {
	if mode == "" {
		mode = model.ScanModeSafe
	}
	return Config{
		Mode:            mode,
		GlobalTimeout:   2 * time.Minute,
		PerProbeTimeout: 10 * time.Second,
		RateLimit: RateLimit{
			RequestsPerSecond: 25,
			Burst:             5,
		},
	}
}
