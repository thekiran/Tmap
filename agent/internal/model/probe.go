package model

import "time"

type ProbeStatus string

const (
	ProbeStatusSuccess ProbeStatus = "success"
	ProbeStatusPartial ProbeStatus = "partial"
	ProbeStatusSkipped ProbeStatus = "skipped"
	ProbeStatusFailed  ProbeStatus = "failed"
)

type ProbeInput struct {
	Mode         ScanMode        `json:"mode"`
	Scope        ScanScope       `json:"scope"`
	CandidateIPs []string        `json:"candidate_ips,omitempty"`
	Metadata     map[string]any  `json:"metadata,omitempty"`
	SNMP         *SNMPCredential `json:"snmp,omitempty"`
}

type SNMPCredential struct {
	Community string `json:"community,omitempty"`
	Version   string `json:"version,omitempty"`
	Username  string `json:"username,omitempty"`
}

type ProbeResult struct {
	ProbeName  string         `json:"probe_name"`
	Status     ProbeStatus    `json:"status"`
	StartedAt  time.Time      `json:"started_at"`
	FinishedAt time.Time      `json:"finished_at"`
	Confidence float64        `json:"confidence"`
	Evidence   []Evidence     `json:"evidence,omitempty"`
	Devices    []Device       `json:"devices,omitempty"`
	Edges      []TopologyEdge `json:"edges,omitempty"`
	Raw        map[string]any `json:"raw,omitempty"`
	Errors     []string       `json:"errors,omitempty"`
}
