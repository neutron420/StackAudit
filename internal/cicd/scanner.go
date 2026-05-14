package cicd

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"stackaudit/internal/rules"
	"stackaudit/internal/scanner"
	"stackaudit/internal/utils"

	"gopkg.in/yaml.v3"
)

type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Name() string {
	return "cicd"
}

func (s *Scanner) Scan(ctx context.Context, root string, ruleSet rules.RuleSet) ([]scanner.Finding, error) {
	workflows := filepath.Join(root, ".github", "workflows")
	ignore, err := utils.LoadIgnoreMatcher(root)
	if err != nil {
		return nil, err
	}
	if ignore != nil && ignore(workflows, true) {
		return nil, nil
	}
	entries, err := os.ReadDir(workflows)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	findings := []scanner.Finding{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml") {
			continue
		}
		path := filepath.Join(workflows, name)
		if ignore != nil && ignore(path, false) {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		findings = append(findings, scanWorkflow(path, data)...)
	}

	return findings, nil
}

func scanWorkflow(path string, data []byte) []scanner.Finding {
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return []scanner.Finding{
			{
				Severity:    scanner.SeverityCritical,
				Title:       "Workflow YAML is invalid",
				Description: err.Error(),
				File:        path,
				Category:    "cicd",
				RuleID:      "cicd_invalid_yaml",
			},
		}
	}

	findings := []scanner.Finding{}

	if permissions, ok := raw["permissions"].(string); ok && permissions == "write-all" {
		findings = append(findings, scanner.Finding{
			Severity:    scanner.SeverityWarning,
			Title:       "Workflow uses write-all permissions",
			Description: "Reduce permissions to least privilege.",
			File:        path,
			Category:    "cicd",
			RuleID:      "cicd_permissions_write_all",
		})
	}

	if hasWorkflowEvent(raw["on"], "pull_request_target") {
		findings = append(findings, scanner.Finding{
			Severity:    scanner.SeverityWarning,
			Title:       "Workflow uses pull_request_target",
			Description: "Ensure secrets are not exposed to untrusted forks.",
			File:        path,
			Category:    "cicd",
			RuleID:      "cicd_pull_request_target",
		})
	}

	content := string(data)
	if strings.Contains(content, "actions/checkout@v1") {
		findings = append(findings, scanner.Finding{
			Severity:    scanner.SeverityWarning,
			Title:       "Workflow uses actions/checkout@v1",
			Description: "Upgrade to a supported version of actions/checkout.",
			File:        path,
			Category:    "cicd",
			RuleID:      "cicd_checkout_v1",
		})
	}

	if strings.Contains(content, "${{ secrets.") {
		findings = append(findings, scanner.Finding{
			Severity:    scanner.SeverityInfo,
			Title:       "Workflow references secrets",
			Description: "Ensure secrets are masked and not echoed to logs.",
			File:        path,
			Category:    "cicd",
			RuleID:      "cicd_secrets_reference",
		})
	}

	return findings
}

func hasWorkflowEvent(value interface{}, event string) bool {
	switch on := value.(type) {
	case string:
		return on == event
	case []interface{}:
		for _, item := range on {
			if text, ok := item.(string); ok && text == event {
				return true
			}
		}
	case map[string]interface{}:
		_, ok := on[event]
		return ok
	}
	return false
}
