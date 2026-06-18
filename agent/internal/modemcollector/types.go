package modemcollector

import (
	"context"
	"net"
	"time"

	"github.com/thekiran/iad/pkg/models"
)

type ProbeStatus string

const (
	ProbeStatusSuccess ProbeStatus = "success"
	ProbeStatusFailed  ProbeStatus = "failed"
	ProbeStatusSkipped ProbeStatus = "skipped"
)

type Probe interface {
	Name() string
	Run(ctx context.Context, target ProbeTarget, store *EvidenceStore) ProbeResult
}

type ProbeTarget struct {
	IP         net.IP
	Role       string
	Source     string
	Private    bool
	Confidence float64
}

type ProbeResult struct {
	ProbeName  string
	TargetIP   string
	Status     ProbeStatus
	Confidence float64
	Evidence   []Observation
	Errors     []ProbeError
	StartedAt  time.Time
	FinishedAt time.Time
}

type Observation struct {
	ID         string
	Source     string
	TargetIP   string
	Kind       string
	Strength   string
	Confidence float64
	Value      any
	Timestamp  time.Time
}

type ProbeError struct {
	Source  string
	Target  string
	Message string
	Timeout bool
}

type BuildInput struct {
	Result models.ScanResult
}
