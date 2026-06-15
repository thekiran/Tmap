package scoring

import "testing"

func TestRuleMatches(t *testing.T) {
	rule := Rule{
		ID: "dsl",
		If: Condition{
			RouterModelContains: []string{"VR400"},
			HintContains:        []string{"Mobile"},
		},
	}

	cases := []struct {
		name string
		ctx  MatchContext
		want bool
	}{
		{"model match (case-insensitive)", MatchContext{RouterModel: "tp-link archer vr400"}, true},
		{"hint match", MatchContext{Hints: []string{"Mobile"}}, true},
		{"no match", MatchContext{RouterModel: "Huawei HG8245H", Text: "gpon"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := rule.Matches(tc.ctx); got != tc.want {
				t.Errorf("Matches = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBoardNormalizeClamps(t *testing.T) {
	b := NewBoard()
	b.Add(map[string]float64{"VDSL": 130, "ADSL": 40})
	got := b.Normalize()
	if got["VDSL"] != 1.0 {
		t.Errorf("VDSL normalized = %v, want 1.0 (clamped)", got["VDSL"])
	}
	if got["ADSL"] != 0.4 {
		t.Errorf("ADSL normalized = %v, want 0.4", got["ADSL"])
	}
}
