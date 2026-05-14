package scanner

import "strings"

func ApplySeverityOverrides(findings []Finding, overrides map[string]string) []Finding {
	if len(overrides) == 0 {
		return findings
	}

	for i, finding := range findings {
		if finding.RuleID == "" {
			continue
		}
		if severity, ok := overrides[normalizeRuleID(finding.RuleID)]; ok {
			if parsed, ok := parseSeverity(severity); ok {
				findings[i].Severity = parsed
				continue
			}
		}
		if base := baseRuleID(finding.RuleID); base != "" {
			if severity, ok := overrides[normalizeRuleID(base)]; ok {
				if parsed, ok := parseSeverity(severity); ok {
					findings[i].Severity = parsed
				}
			}
		}
	}

	return findings
}

func parseSeverity(value string) (Severity, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(SeverityCritical):
		return SeverityCritical, true
	case string(SeverityWarning):
		return SeverityWarning, true
	case string(SeverityInfo):
		return SeverityInfo, true
	case string(SeveritySuccess):
		return SeveritySuccess, true
	default:
		return "", false
	}
}

func normalizeRuleID(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func baseRuleID(value string) string {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) > 1 {
		return parts[0]
	}
	return ""
}
