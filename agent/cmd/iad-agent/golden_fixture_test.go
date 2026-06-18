package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/thekiran/iad/pkg/models"
)

func TestGoldenReportFixtures(t *testing.T) {
	tests := []struct {
		file    string
		profile string
		mode    string
	}{
		{file: "golden_quick_report.json", profile: "quick"},
		{file: "golden_full_report.json", profile: "full"},
		{file: "golden_uncertain_classification.json", profile: "full", mode: "full"},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			path := filepath.Join("..", "..", "tests", "fixtures", tt.file)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var report models.ScanReport
			if err := json.Unmarshal(data, &report); err != nil {
				t.Fatal(err)
			}
			if report.Scope.Profile != tt.profile {
				t.Fatalf("scope.profile = %q, want %q", report.Scope.Profile, tt.profile)
			}
			if tt.mode != "" {
				if report.AccessClassification == nil {
					t.Fatal("expected access_classification")
				}
				if report.AccessClassification.Mode != tt.mode {
					t.Fatalf("access_classification.mode = %q, want %q", report.AccessClassification.Mode, tt.mode)
				}
			}
		})
	}
}
