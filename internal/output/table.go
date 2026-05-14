package output

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"devdoctor/internal/scanner"
)

func renderTable(report scanner.Report) string {
	builder := &strings.Builder{}
	head := "DevDoctor"
	line := "──────────────────────────────"
	fmt.Fprintf(builder, "╭%s╮\n", line)
	fmt.Fprintf(builder, "│%s│\n", center(head, len(line)))
	fmt.Fprintf(builder, "│%s│\n", center("Backend Health Scanner", len(line)))
	fmt.Fprintf(builder, "╰%s╯\n\n", line)

	fmt.Fprintln(builder, styleMuted.Render("Local-only scan. No telemetry. No network calls. Secrets never leave your machine."))
	fmt.Fprintln(builder, "")

	fmt.Fprintf(builder, "%s %d/100\n", styleHeader.Render("Project Health Score:"), report.Scores.Overall)
	fmt.Fprintf(builder, "%s %d/100\n", styleHeader.Render("Security:"), report.Scores.Security)
	fmt.Fprintf(builder, "%s %d/100\n", styleHeader.Render("Infrastructure:"), report.Scores.Infrastructure)
	fmt.Fprintf(builder, "%s %d/100\n\n", styleHeader.Render("Configuration:"), report.Scores.Configuration)

	fmt.Fprintf(builder, "%s %d  %s %d  %s %d  %s %d\n\n",
		styleCritical.Render("Critical"), report.Summary.Critical,
		styleWarning.Render("Warning"), report.Summary.Warning,
		styleInfo.Render("Info"), report.Summary.Info,
		styleSuccess.Render("Success"), report.Summary.Success,
	)

	grouped := groupBySeverity(report.Findings)
	order := []scanner.Severity{scanner.SeverityCritical, scanner.SeverityWarning, scanner.SeverityInfo, scanner.SeveritySuccess}
	for _, severity := range order {
		findings := grouped[severity]
		if len(findings) == 0 {
			continue
		}
		fmt.Fprintf(builder, "%s\n", severityHeader(severity))
		for _, finding := range findings {
			fmt.Fprintf(builder, "  %s %s\n", severityIcon(string(finding.Severity)), finding.Title)
			if finding.Description != "" {
				fmt.Fprintf(builder, "    %s\n", styleMuted.Render(finding.Description))
			}
			if finding.Remediation != "" {
				fmt.Fprintf(builder, "    %s %s\n", styleHeader.Render("Fix:"), styleMuted.Render(finding.Remediation))
			}
			if finding.File != "" {
				location := finding.File
				if finding.Line > 0 {
					location = fmt.Sprintf("%s:%d", finding.File, finding.Line)
				}
				fmt.Fprintf(builder, "    %s\n", styleMuted.Render(location))
			}
		}
		fmt.Fprintln(builder, "")
	}

	fmt.Fprintf(builder, "%s %s\n", styleMuted.Render("Scan completed in"), styleMuted.Render(report.Meta.Duration.Round(time.Millisecond).String()))
	return builder.String()
}

func groupBySeverity(findings []scanner.Finding) map[scanner.Severity][]scanner.Finding {
	grouped := map[scanner.Severity][]scanner.Finding{}
	for _, finding := range findings {
		grouped[finding.Severity] = append(grouped[finding.Severity], finding)
	}
	for _, list := range grouped {
		sort.SliceStable(list, func(i, j int) bool {
			return list[i].Title < list[j].Title
		})
	}
	return grouped
}

func severityHeader(severity scanner.Severity) string {
	switch severity {
	case scanner.SeverityCritical:
		return styleCritical.Render("Critical")
	case scanner.SeverityWarning:
		return styleWarning.Render("Warning")
	case scanner.SeverityInfo:
		return styleInfo.Render("Info")
	case scanner.SeveritySuccess:
		return styleSuccess.Render("Success")
	default:
		return "Findings"
	}
}

func center(value string, width int) string {
	if len(value) >= width {
		return value
	}
	pad := width - len(value)
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + value + strings.Repeat(" ", right)
}
