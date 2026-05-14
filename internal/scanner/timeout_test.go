package scanner

import (
	"context"
	"testing"
	"time"

	"stackaudit/internal/rules"
)

type timedModule struct {
	name  string
	delay time.Duration
}

func (m timedModule) Name() string {
	return m.name
}

func (m timedModule) Scan(ctx context.Context, root string, ruleSet rules.RuleSet) ([]Finding, error) {
	timer := time.NewTimer(m.delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return []Finding{{
			Severity: SeverityInfo,
			Title:    m.name + " finished",
			Category: m.name,
			RuleID:   m.name + "_finished",
		}}, nil
	}
}

func TestRunUsesModuleSpecificTimeout(t *testing.T) {
	report, err := Run(context.Background(), ".", rules.DefaultRuleSet(), []Module{
		timedModule{name: "env", delay: 50 * time.Millisecond},
		timedModule{name: "docker", delay: time.Millisecond},
	}, Options{Timeouts: TimeoutOptions{
		Default: 100 * time.Millisecond,
		Modules: map[string]time.Duration{
			"env": time.Millisecond,
		},
	}})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if len(report.Findings) != 2 {
		t.Fatalf("findings length = %d, want 2", len(report.Findings))
	}

	var timedOut bool
	var dockerFinished bool
	for _, finding := range report.Findings {
		if finding.RuleID == "module_timeout" && finding.Category == "env" {
			timedOut = true
		}
		if finding.RuleID == "docker_finished" {
			dockerFinished = true
		}
	}
	if !timedOut {
		t.Fatal("expected env module timeout finding")
	}
	if !dockerFinished {
		t.Fatal("expected docker module to finish with default timeout")
	}
}
