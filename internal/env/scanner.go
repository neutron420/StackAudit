package env

import (
	"context"

	"stack/internal/rules"
	"stack/internal/scanner"
)

type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Name() string {
	return "env"
}

func (s *Scanner) Scan(ctx context.Context, root string, ruleSet rules.RuleSet) ([]scanner.Finding, error) {
	report, err := Scan(ctx, root, ruleSet)
	if err != nil {
		return nil, err
	}
	findings := report.Findings
	if len(findings) == 0 {
		findings = append(findings, scanner.Finding{
			Category:    "env",
			Title:       "No environment files found",
			Description: "We couldn't find any .env or environment configuration files in your project root.",
			Severity:    scanner.SeverityInfo,
		})
	}

	return findings, nil
}
