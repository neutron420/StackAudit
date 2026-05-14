package output

import (
	"fmt"
	"strings"

	"stackaudit/internal/scanner"
)

func renderMarkdown(report scanner.Report) string {
	builder := &strings.Builder{}
	fmt.Fprintln(builder, "# StackAudit Report")
	fmt.Fprintln(builder, "")
	fmt.Fprintln(builder, "- Local-only scan. No telemetry. No network calls.")
	fmt.Fprintln(builder, "")
	fmt.Fprintf(builder, "## Scores\n- Overall: %d/100\n- Security: %d/100\n- Infrastructure: %d/100\n- Configuration: %d/100\n\n",
		report.Scores.Overall,
		report.Scores.Security,
		report.Scores.Infrastructure,
		report.Scores.Configuration,
	)
	fmt.Fprintf(builder, "## Summary\n- Critical: %d\n- Warning: %d\n- Info: %d\n- Success: %d\n\n",
		report.Summary.Critical,
		report.Summary.Warning,
		report.Summary.Info,
		report.Summary.Success,
	)
	fmt.Fprintln(builder, "## Findings")
	if len(report.Findings) == 0 {
		fmt.Fprintln(builder, "- No issues found.")
		return builder.String()
	}
	for _, finding := range report.Findings {
		line := fmt.Sprintf("- **%s**: %s", titleCase(string(finding.Severity)), finding.Title)
		if finding.File != "" {
			line += fmt.Sprintf(" (%s)", finding.File)
		}
		fmt.Fprintln(builder, line)
		if finding.Description != "" {
			fmt.Fprintf(builder, "  - %s\n", finding.Description)
		}
		if finding.Remediation != "" {
			fmt.Fprintf(builder, "  - Fix: %s\n", finding.Remediation)
		}
	}
	return builder.String()
}

func titleCase(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
