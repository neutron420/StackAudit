package baseline

import (
	"testing"

	"stack/internal/scanner"
)

func TestFilterRebuildsSummaryAndScores(t *testing.T) {
	report := scanner.Rebuild(scanner.Report{
		Findings: []scanner.Finding{
			{
				Severity: scanner.SeverityCritical,
				Title:    "Secret exposed",
				Category: "secrets",
				RuleID:   "secret_token",
				File:     "app.go",
				Line:     10,
			},
			{
				Severity: scanner.SeverityWarning,
				Title:    "Missing env",
				Category: "env",
				RuleID:   "env_required:JWT_SECRET",
			},
		},
	})

	filtered := Filter(report, Snapshot{Entries: []Entry{
		{
			RuleID:   "secret_token",
			Category: "secrets",
			Title:    "Secret exposed",
			File:     "app.go",
			Line:     10,
		},
	}}, ".")

	if filtered.Summary.Total != 1 || filtered.Summary.Warning != 1 || filtered.Summary.Critical != 0 {
		t.Fatalf("summary = %+v, want one warning", filtered.Summary)
	}
	if filtered.Scores.Security != 100 {
		t.Fatalf("security score = %d, want 100 after baseline filtering", filtered.Scores.Security)
	}
	if filtered.Scores.Configuration != 95 {
		t.Fatalf("configuration score = %d, want 95 for remaining warning", filtered.Scores.Configuration)
	}
}
