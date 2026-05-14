package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"devdoctor/internal/rules"
	"devdoctor/internal/scanner"
	"devdoctor/internal/utils"

	"gopkg.in/yaml.v3"
)

type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Name() string {
	return "docker"
}

func (s *Scanner) Scan(ctx context.Context, root string, ruleSet rules.RuleSet) ([]scanner.Finding, error) {
	findings := []scanner.Finding{}
	ignore, err := utils.LoadIgnoreMatcher(root)
	if err != nil {
		return nil, err
	}

	dockerfile := filepath.Join(root, "Dockerfile")
	if ignore == nil || !ignore(dockerfile, false) {
		if _, err := os.Stat(dockerfile); err == nil {
			data, err := os.ReadFile(dockerfile)
			if err != nil {
				return nil, err
			}
			findings = append(findings, scanDockerfile(dockerfile, string(data), ruleSet)...)
		}
	}

	composePath := filepath.Join(root, "docker-compose.yml")
	if ignore == nil || !ignore(composePath, false) {
		if _, err := os.Stat(composePath); err == nil {
			data, err := os.ReadFile(composePath)
			if err != nil {
				return nil, err
			}
			findings = append(findings, scanCompose(composePath, data, ruleSet)...)
		}
	}

	return findings, nil
}

var (
	fromPattern   = regexp.MustCompile(`(?i)^FROM\s+([^\s]+)`)
	userPattern   = regexp.MustCompile(`(?i)^USER\s+(.+)`)
	envPattern    = regexp.MustCompile(`(?i)^ENV\s+(.+)`)
	exposePattern = regexp.MustCompile(`(?i)^EXPOSE\s+(.+)`)
)

func scanDockerfile(path, content string, ruleSet rules.RuleSet) []scanner.Finding {
	findings := []scanner.Finding{}
	lines := strings.Split(content, "\n")
	seenUser := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if matches := fromPattern.FindStringSubmatch(line); len(matches) > 1 {
			image := matches[1]
			if !strings.Contains(image, ":") || strings.HasSuffix(image, ":latest") {
				if !ruleSet.LatestTagOK {
					findings = append(findings, scanner.Finding{
						Severity:    scanner.SeverityWarning,
						Title:       "Dockerfile uses latest image tag",
						Description: fmt.Sprintf("Pin image tags for reproducible builds: %s", image),
						File:        path,
						Category:    "docker",
						RuleID:      "docker_latest_tag",
					})
				}
			}
		}
		if matches := userPattern.FindStringSubmatch(line); len(matches) > 1 {
			seenUser = true
		}
		if matches := envPattern.FindStringSubmatch(line); len(matches) > 1 {
			if strings.Contains(strings.ToUpper(matches[1]), "SECRET") || strings.Contains(strings.ToUpper(matches[1]), "TOKEN") || strings.Contains(strings.ToUpper(matches[1]), "KEY") {
				findings = append(findings, scanner.Finding{
					Severity:    scanner.SeverityWarning,
					Title:       "Potential secret in Dockerfile",
					Description: "Avoid hardcoding secrets in Dockerfile ENV directives.",
					File:        path,
					Category:    "docker",
					RuleID:      "docker_env_secret",
				})
			}
		}
		if matches := exposePattern.FindStringSubmatch(line); len(matches) > 1 {
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityInfo,
				Title:       fmt.Sprintf("Dockerfile exposes port %s", matches[1]),
				Description: "Ensure exposed ports are expected and documented.",
				File:        path,
				Category:    "docker",
				RuleID:      "docker_exposed_port",
			})
		}
	}

	if !seenUser {
		findings = append(findings, scanner.Finding{
			Severity:    scanner.SeverityWarning,
			Title:       "Dockerfile does not set a non-root user",
			Description: "Add a USER directive to reduce container privileges.",
			File:        path,
			Category:    "docker",
			RuleID:      "docker_no_user",
		})
	}

	return findings
}

func scanCompose(path string, data []byte, ruleSet rules.RuleSet) []scanner.Finding {
	findings := []scanner.Finding{}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return []scanner.Finding{
			{
				Severity:    scanner.SeverityCritical,
				Title:       "docker-compose.yml is invalid",
				Description: err.Error(),
				File:        path,
				Category:    "docker",
				RuleID:      "docker_compose_invalid",
			},
		}
	}

	services, ok := raw["services"].(map[string]interface{})
	if !ok {
		return findings
	}

	for name, entry := range services {
		service, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		if _, hasRestart := service["restart"]; !hasRestart {
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityWarning,
				Title:       fmt.Sprintf("Service %s missing restart policy", name),
				Description: "Add restart: unless-stopped to improve reliability.",
				File:        path,
				Category:    "docker",
				RuleID:      "docker_restart_policy",
			})
		}
		if image, ok := service["image"].(string); ok {
			if (!strings.Contains(image, ":") || strings.HasSuffix(image, ":latest")) && !ruleSet.LatestTagOK {
				findings = append(findings, scanner.Finding{
					Severity:    scanner.SeverityWarning,
					Title:       fmt.Sprintf("Service %s uses latest image tag", name),
					Description: "Pin image tags in docker-compose for reproducible deployments.",
					File:        path,
					Category:    "docker",
					RuleID:      "docker_latest_tag:compose",
				})
			}
		}
	}

	return findings
}
