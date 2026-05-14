package main

import (
	"fmt"
	"os"
	"strings"

	"stackaudit/internal/scanner"
)

func applyExitCode(report scanner.Report) error {
	if !cfg.ExitCode {
		return nil
	}

	minSeverity, err := parseMinSeverity(cfg.MinSeverity)
	if err != nil {
		return err
	}

	code := exitCodeForReport(report, minSeverity)
	if code > 0 {
		os.Exit(code)
	}
	return nil
}

func parseMinSeverity(value string) (scanner.Severity, error) {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		trimmed = "warning"
	}
	switch trimmed {
	case "critical":
		return scanner.SeverityCritical, nil
	case "warning":
		return scanner.SeverityWarning, nil
	case "info":
		return scanner.SeverityInfo, nil
	default:
		return "", fmt.Errorf("invalid min severity: %s", value)
	}
}

func exitCodeForReport(report scanner.Report, minSeverity scanner.Severity) int {
	highest := scanner.SeveritySuccess
	for _, finding := range report.Findings {
		if severityRank(finding.Severity) > severityRank(highest) {
			highest = finding.Severity
		}
	}

	if severityRank(highest) < severityRank(minSeverity) {
		return 0
	}

	switch highest {
	case scanner.SeverityCritical:
		return 3
	case scanner.SeverityWarning:
		return 2
	case scanner.SeverityInfo:
		return 1
	default:
		return 0
	}
}

func severityRank(sev scanner.Severity) int {
	switch sev {
	case scanner.SeverityCritical:
		return 3
	case scanner.SeverityWarning:
		return 2
	case scanner.SeverityInfo:
		return 1
	default:
		return 0
	}
}
