package scanner

import "stackaudit/internal/health"

func Rebuild(report Report) Report {
	report.Findings = ApplyRemediations(report.Findings)
	report.Summary = Summarize(report.Findings)
	healthFindings := make([]health.Finding, 0, len(report.Findings))
	for _, finding := range report.Findings {
		healthFindings = append(healthFindings, health.Finding{
			Severity: string(finding.Severity),
			Category: finding.Category,
		})
	}
	scores := health.CalculateScores(healthFindings)
	report.Scores = Scores{
		Overall:        scores.Overall,
		Security:       scores.Security,
		Infrastructure: scores.Infrastructure,
		Configuration:  scores.Configuration,
	}
	return report
}
