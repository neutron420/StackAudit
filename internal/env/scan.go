package env

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"stack/internal/rules"
	"stack/internal/scanner"
	"stack/internal/utils"
)

var (
	namePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
)

func Scan(ctx context.Context, root string, ruleSet rules.RuleSet) (Report, error) {
	files := []string{".env", ".env.local", ".env.production", ".env.example"}
	parsed := []FileEnv{}
	ignore, err := utils.LoadIgnoreMatcher(root)
	if err != nil {
		return Report{}, err
	}

	for _, file := range files {
		path := filepath.Join(root, file)
		if ignore != nil && ignore(path, false) {
			continue
		}
		fileEnv, err := parseEnvFile(path)
		if err != nil {
			if utils.IsNotExist(err) {
				continue
			}
			return Report{}, err
		}
		fileEnv.IsExample = file == ".env.example"
		fileEnv.IsProd = file == ".env.production"
		parsed = append(parsed, fileEnv)
	}

	usedVars, err := DiscoverUsedVars(root)
	if err != nil {
		return Report{}, err
	}

	required := requiredVars(ruleSet, parsed)
	unused := map[string][]string{}
	findings := []scanner.Finding{}

	for _, file := range parsed {
		fileUnused := []string{}
		for key, value := range file.Vars {
			if !namePattern.MatchString(key) {
				findings = append(findings, scanner.Finding{
					Severity:    scanner.SeverityWarning,
					Title:       fmt.Sprintf("Inconsistent env var name: %s", key),
					Description: "Use uppercase snake case for environment variables.",
					File:        file.Path,
					Category:    "env",
					RuleID:      "env_inconsistent_name",
				})
			}

			if file.IsProd && ruleSet.NoLocalhost {
				if strings.Contains(strings.ToLower(value), "localhost") || strings.Contains(value, "127.0.0.1") {
					findings = append(findings, scanner.Finding{
						Severity:    scanner.SeverityCritical,
						Title:       fmt.Sprintf("Production env uses localhost in %s", key),
						Description: "Localhost values in production files can break deployments.",
						File:        file.Path,
						Category:    "env",
						RuleID:      "env_localhost",
					})
				}
			}

			if !file.IsExample && !usedVars[key] {
				fileUnused = append(fileUnused, key)
				findings = append(findings, scanner.Finding{
					Severity:    scanner.SeverityInfo,
					Title:       fmt.Sprintf("Unused env var: %s", key),
					Description: "Variable not referenced in codebase.",
					File:        file.Path,
					Category:    "env",
					RuleID:      "env_unused",
				})
			}
		}

		for _, duplicate := range file.Duplicates {
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityWarning,
				Title:       fmt.Sprintf("Duplicate env var: %s", duplicate),
				Description: "Duplicate variables make configuration ambiguous.",
				File:        file.Path,
				Category:    "env",
				RuleID:      "env_duplicate",
			})
		}

		for _, empty := range file.Empty {
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityWarning,
				Title:       fmt.Sprintf("Empty env var: %s", empty),
				Description: "Variable is defined without a value.",
				File:        file.Path,
				Category:    "env",
				RuleID:      "env_empty",
			})
		}

		for _, req := range required {
			if _, ok := file.Vars[req]; !ok && !file.IsExample {
				findings = append(findings, scanner.Finding{
					Severity:    scanner.SeverityWarning,
					Title:       fmt.Sprintf("Missing required env var: %s", req),
					Description: "Required variable is missing from env file.",
					File:        file.Path,
					Category:    "env",
					RuleID:      "env_required:" + req,
				})
			}
		}

		if len(fileUnused) > 0 {
			unused[file.Path] = fileUnused
		}
	}

	return Report{
		Files:       parsed,
		UsedVars:    usedVars,
		UnusedVars:  unused,
		Findings:    findings,
		RequiredEnv: required,
	}, nil
}

func requiredVars(ruleSet rules.RuleSet, parsed []FileEnv) []string {
	required := map[string]bool{}
	for key := range ruleSet.RequiredEnv {
		required[key] = true
	}
	for _, file := range parsed {
		if !file.IsExample {
			continue
		}
		for key := range file.Vars {
			required[key] = true
		}
	}
	result := []string{}
	for key := range required {
		result = append(result, key)
	}
	return result
}
