package scanner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"stack/internal/health"
	"stack/internal/rules"
)

type Module interface {
	Name() string
	Scan(ctx context.Context, root string, ruleSet rules.RuleSet) ([]Finding, error)
}

type Options struct {
	Timeouts TimeoutOptions
}

func Run(ctx context.Context, root string, ruleSet rules.RuleSet, modules []Module, opts Options) (Report, error) {
	start := time.Now()

	var (
		mu       sync.Mutex
		findings []Finding
		firstErr error
	)

	wg := sync.WaitGroup{}
	for _, module := range modules {
		wg.Add(1)
		mod := module
		go func() {
			defer wg.Done()
			timeout := moduleTimeout(mod.Name(), opts.Timeouts)
			results, err, timedOut := runModule(ctx, mod, root, ruleSet, timeout)
			mu.Lock()
			defer mu.Unlock()
			if timedOut {
				description := "Module exceeded its time budget"
				if timeout > 0 {
					description = fmt.Sprintf("Module exceeded the %s budget", timeout)
				}
				findings = append(findings, Finding{
					Severity:    SeverityWarning,
					Title:       fmt.Sprintf("%s scan timed out", mod.Name()),
					Description: description,
					Category:    mod.Name(),
					RuleID:      "module_timeout",
				})
				return
			}
			if err != nil && firstErr == nil {
				firstErr = fmt.Errorf("%s scan failed: %w", mod.Name(), err)
				return
			}
			findings = append(findings, results...)
		}()
	}
	wg.Wait()

	if firstErr != nil {
		return Report{}, firstErr
	}

	findings = ApplySeverityOverrides(findings, ruleSet.SeverityOverrides)
	findings = ApplyRemediations(findings)
	summary := Summarize(findings)
	healthFindings := make([]health.Finding, 0, len(findings))
	for _, finding := range findings {
		healthFindings = append(healthFindings, health.Finding{
			Severity: string(finding.Severity),
			Category: finding.Category,
		})
	}
	scores := health.CalculateScores(healthFindings)

	return Report{
		Findings: findings,
		Summary:  summary,
		Scores: Scores{
			Overall:        scores.Overall,
			Security:       scores.Security,
			Infrastructure: scores.Infrastructure,
			Configuration:  scores.Configuration,
		},
		Meta: Meta{
			RootPath: root,
			Duration: time.Since(start),
		},
	}, nil
}

func Summarize(findings []Finding) Summary {
	summary := Summary{}
	for _, finding := range findings {
		switch finding.Severity {
		case SeverityCritical:
			summary.Critical++
		case SeverityWarning:
			summary.Warning++
		case SeverityInfo:
			summary.Info++
		case SeveritySuccess:
			summary.Success++
		}
		summary.Total++
	}
	return summary
}
