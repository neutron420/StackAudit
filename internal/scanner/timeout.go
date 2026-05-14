package scanner

import (
	"context"
	"strings"
	"time"

	"stackaudit/internal/rules"
)

type moduleResult struct {
	findings []Finding
	err      error
}

func runModule(ctx context.Context, mod Module, root string, ruleSet rules.RuleSet, timeout time.Duration) ([]Finding, error, bool) {
	modCtx := ctx
	var cancel context.CancelFunc
	if timeout > 0 {
		modCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	if cancel != nil {
		defer cancel()
	}

	resultCh := make(chan moduleResult, 1)
	go func() {
		findings, err := mod.Scan(modCtx, root, ruleSet)
		resultCh <- moduleResult{findings: findings, err: err}
	}()

	if timeout <= 0 {
		result := <-resultCh
		return result.findings, result.err, false
	}

	select {
	case result := <-resultCh:
		return result.findings, result.err, false
	case <-modCtx.Done():
		return nil, modCtx.Err(), true
	}
}

func moduleTimeout(name string, opts TimeoutOptions) time.Duration {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if opts.Modules != nil {
		if timeout, ok := opts.Modules[normalized]; ok {
			return timeout
		}
	}
	return opts.Default
}
