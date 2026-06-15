package scoring

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Rule is one scoring rule loaded from YAML. A rule fires when any of its
// specified condition groups matches, and then applies its score deltas.
type Rule struct {
	ID   string    `yaml:"id"`
	If   Condition `yaml:"if"`
	Then Action    `yaml:"then"`
}

// Condition holds the (all optional) match groups. Within a group the entries
// are OR'd; across groups the rule fires if any group matches.
type Condition struct {
	RouterModelContains  []string `yaml:"router_model_contains"`
	TextContains         []string `yaml:"text_contains"`
	InterfaceNameMatches []string `yaml:"interface_name_matches"`
	HintContains         []string `yaml:"hint_contains"`
}

// Action is what a fired rule contributes: score deltas keyed by access type.
type Action struct {
	AddScore map[string]float64 `yaml:"add_score"`
}

type ruleFile struct {
	Rules []Rule `yaml:"rules"`
}

// MatchContext is the normalized evidence a rule is evaluated against. The
// engine builds it from probe results; keeping it here avoids a scoring→detection
// import cycle.
type MatchContext struct {
	RouterModel string
	Text        string
	Interfaces  []string
	Hints       []string
}

// LoadRules reads and concatenates the rule files (relative to dir). A missing
// file is skipped so optional rule sets are allowed; a malformed file is an
// error so typos surface loudly.
func LoadRules(dir string, files ...string) ([]Rule, error) {
	var all []Rule
	for _, name := range files {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		var rf ruleFile
		if err := yaml.Unmarshal(data, &rf); err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		all = append(all, rf.Rules...)
	}
	return all, nil
}

// Matches reports whether the rule fires against ctx.
func (r Rule) Matches(ctx MatchContext) bool {
	c := r.If
	if anyContainsFold(ctx.RouterModel, c.RouterModelContains) {
		return true
	}
	if anyContainsFold(ctx.Text, c.TextContains) {
		return true
	}
	if anySliceContainsFold(ctx.Interfaces, c.InterfaceNameMatches) {
		return true
	}
	if anySliceContainsFold(ctx.Hints, c.HintContains) {
		return true
	}
	return false
}

// anyContainsFold reports whether haystack contains any needle (case-insensitive).
func anyContainsFold(haystack string, needles []string) bool {
	if len(needles) == 0 {
		return false
	}
	h := strings.ToLower(haystack)
	for _, n := range needles {
		if n != "" && strings.Contains(h, strings.ToLower(n)) {
			return true
		}
	}
	return false
}

// anySliceContainsFold reports whether any element of values contains any needle.
func anySliceContainsFold(values, needles []string) bool {
	if len(needles) == 0 {
		return false
	}
	for _, v := range values {
		if anyContainsFold(v, needles) {
			return true
		}
	}
	return false
}
