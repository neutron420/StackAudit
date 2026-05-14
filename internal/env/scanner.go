package env

import (
	"context"

	"stackaudit/internal/rules"
	"stackaudit/internal/scanner"
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
	return report.Findings, nil
}
