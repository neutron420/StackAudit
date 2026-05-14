package output

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"stack/internal/scanner"

	"github.com/charmbracelet/lipgloss"
)

func renderTable(report scanner.Report, showBanner bool) string {
	var b strings.Builder

	if showBanner {
		// Header Box
		headerStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styleBranding.GetForeground()).
			Padding(0, 2).
			Align(lipgloss.Center).
			Width(40)

		headerContent := lipgloss.JoinVertical(lipgloss.Center,
			styleBranding.Render("stack"),
			styleMuted.Render("Production Health Scanner"),
		)
		fmt.Fprintln(&b, headerStyle.Render(headerContent))

		fmt.Fprintln(&b, styleMuted.Render(" Local-only scan • Secrets stay on your machine"))
		fmt.Fprintln(&b, "")

		// Score Rendering
		renderScore := func(label string, score int) string {
			var scoreStyle lipgloss.Style
			if score >= 90 {
				scoreStyle = styleSuccess
			} else if score >= 70 {
				scoreStyle = styleWarning
			} else {
				scoreStyle = styleCritical
			}
			return fmt.Sprintf("%-22s %s", styleHeader.Render(label), scoreStyle.Render(fmt.Sprintf("%d/100", score)))
		}

		fmt.Fprintln(&b, renderScore("Project Health:", report.Scores.Overall))
		fmt.Fprintln(&b, renderScore("Security:", report.Scores.Security))
		fmt.Fprintln(&b, renderScore("Infrastructure:", report.Scores.Infrastructure))
		fmt.Fprintln(&b, renderScore("Configuration:", report.Scores.Configuration))
		fmt.Fprintln(&b, "")

		// Summary Bar
		summary := fmt.Sprintf("%s %d  %s %d  %s %d  %s %d",
			styleCritical.Render("Critical"), report.Summary.Critical,
			styleWarning.Render("Warning"), report.Summary.Warning,
			styleInfo.Render("Info"), report.Summary.Info,
			styleSuccess.Render("Success"), report.Summary.Success,
		)
		fmt.Fprintln(&b, summary)
		fmt.Fprintln(&b, "")
	}

	// Findings
	grouped := groupBySeverity(report.Findings)
	order := []scanner.Severity{scanner.SeverityCritical, scanner.SeverityWarning, scanner.SeverityInfo, scanner.SeveritySuccess}
	for _, severity := range order {
		findings := grouped[severity]
		if len(findings) == 0 {
			continue
		}
		fmt.Fprintf(&b, "%s\n", severityHeader(severity))
		for _, finding := range findings {
			fmt.Fprintf(&b, "  %s %s\n", severityIcon(string(finding.Severity)), finding.Title)
			if finding.Description != "" {
				fmt.Fprintf(&b, "    %s\n", styleMuted.Render(finding.Description))
			}
			if finding.Remediation != "" {
				fmt.Fprintf(&b, "    %s %s\n", styleHeader.Render("Fix:"), styleMuted.Render(finding.Remediation))
			}
			if finding.File != "" {
				location := finding.File
				if finding.Line > 0 {
					location = fmt.Sprintf("%s:%d", finding.File, finding.Line)
				}
				fmt.Fprintf(&b, "    %s\n", styleMuted.Render(location))
			}
		}
		fmt.Fprintln(&b, "")
	}

	fmt.Fprintf(&b, "%s %s\n", styleMuted.Render("Audit completed in"), styleMuted.Render(report.Meta.Duration.Round(time.Millisecond).String()))
	return b.String()
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
