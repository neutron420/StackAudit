package output

import (
	"strings"
	"testing"

	"devdoctor/internal/scanner"
)

func TestRenderHTMLIncludesRemediation(t *testing.T) {
	html, err := renderHTML(scanner.Report{
		Findings: []scanner.Finding{
			{
				Severity:    scanner.SeverityWarning,
				Title:       "Dockerfile uses latest image tag",
				Remediation: "Pin the image.",
				RuleID:      "docker_latest_tag",
			},
		},
	})
	if err != nil {
		t.Fatalf("renderHTML returned error: %v", err)
	}
	if !strings.Contains(html, "<!doctype html>") || !strings.Contains(html, "Pin the image.") {
		t.Fatalf("html output missing expected content:\n%s", html)
	}
}
