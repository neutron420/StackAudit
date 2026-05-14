package main

import (
	"fmt"
	"strings"
	"time"

	"devdoctor/internal/scanner"
)

func scanOptions() (scanner.Options, error) {
	timeouts, err := parseModuleTimeouts(cfg.ModuleTimeouts)
	if err != nil {
		return scanner.Options{}, err
	}
	return scanner.Options{Timeouts: timeouts}, nil
}

func parseModuleTimeouts(values []string) (scanner.TimeoutOptions, error) {
	options := scanner.TimeoutOptions{Modules: map[string]time.Duration{}}
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}

			name, rawDuration, hasName := strings.Cut(item, "=")
			if !hasName {
				timeout, err := parseTimeoutDuration(item)
				if err != nil {
					return scanner.TimeoutOptions{}, err
				}
				options.Default = timeout
				continue
			}

			name = normalizeModuleName(name)
			if name == "" {
				return scanner.TimeoutOptions{}, fmt.Errorf("module timeout %q is missing a module name", item)
			}
			timeout, err := parseTimeoutDuration(rawDuration)
			if err != nil {
				return scanner.TimeoutOptions{}, fmt.Errorf("module timeout for %s: %w", name, err)
			}
			options.Modules[name] = timeout
		}
	}
	return options, nil
}

func parseTimeoutDuration(value string) (time.Duration, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("empty timeout duration")
	}
	timeout, err := time.ParseDuration(trimmed)
	if err != nil {
		return 0, err
	}
	if timeout < 0 {
		return 0, fmt.Errorf("timeout must be non-negative")
	}
	return timeout, nil
}

func normalizeModuleName(value string) string {
	name := strings.ToLower(strings.TrimSpace(value))
	if name == "ci" {
		return "cicd"
	}
	return name
}
