package health

import "testing"

func TestCalculateScores(t *testing.T) {
	findings := []Finding{
		{Severity: "critical", Category: "secrets"},
		{Severity: "warning", Category: "env"},
		{Severity: "info", Category: "docker"},
	}

	scores := CalculateScores(findings)
	if scores.Overall >= 100 {
		t.Fatalf("expected overall score to decrease")
	}
	if scores.Security >= 100 {
		t.Fatalf("expected security score to decrease")
	}
}
