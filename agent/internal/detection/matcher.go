package detection

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Fingerprint is one entry in the modem/router fingerprint database. Match holds
// the substrings that identify the device in a model string; AccessHints and
// Supports feed the scoreboard (hints strongly, supported techs lightly).
type Fingerprint struct {
	Vendor      string   `yaml:"vendor"`
	Model       string   `yaml:"model"`
	Category    string   `yaml:"category"`
	Supports    []string `yaml:"supports"`
	AccessHints []string `yaml:"access_hints"`
	Match       []string `yaml:"match"`
}

type fingerprintFile struct {
	Devices []Fingerprint `yaml:"devices"`
}

// LoadFingerprints reads the fingerprint database file from dir. A missing file
// is not an error (detection just runs without fingerprints).
func LoadFingerprints(dir, name string) ([]Fingerprint, error) {
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ff fingerprintFile
	if err := yaml.Unmarshal(data, &ff); err != nil {
		return nil, err
	}
	return ff.Devices, nil
}

// MatchFingerprint returns the first device whose Match substrings (or Model,
// when Match is empty) appear in text, case-insensitively.
func MatchFingerprint(devices []Fingerprint, text string) (*Fingerprint, bool) {
	t := strings.ToLower(text)
	if strings.TrimSpace(t) == "" {
		return nil, false
	}
	for i := range devices {
		d := &devices[i]
		needles := d.Match
		if len(needles) == 0 && d.Model != "" {
			needles = []string{d.Model}
		}
		for _, n := range needles {
			if n != "" && strings.Contains(t, strings.ToLower(n)) {
				return d, true
			}
		}
	}
	return nil, false
}
