package redis

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
	return "redis"
}

func (s *Scanner) Scan(ctx context.Context, root string, ruleSet rules.RuleSet) ([]scanner.Finding, error) {
	ignore, err := utils.LoadIgnoreMatcher(root)
	if err != nil {
		return nil, err
	}
	files, err := utils.WalkFiles(root, utils.WalkOptions{
		MaxSize: 1_000_000,
		Extensions: map[string]bool{
			".conf": true,
			".env":  true,
			".yaml": true,
			".yml":  true,
		},
		SkipDirs: defaultSkipDirs(),
		Ignore:   ignore,
	})
	if err != nil {
		return nil, err
	}

	findings := []scanner.Finding{}
	for _, file := range files {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		name := strings.ToLower(filepath.Base(file))
		switch {
		case name == "redis.conf":
			findings = append(findings, scanRedisConf(file, string(data))...)
		case name == "docker-compose.yml" || name == "docker-compose.yaml" || name == "compose.yml" || name == "compose.yaml":
			findings = append(findings, scanCompose(file, data)...)
		case strings.Contains(name, ".env"):
			findings = append(findings, scanEnv(file, string(data))...)
		}
	}
	return findings, nil
}

func scanRedisConf(path, content string) []scanner.Finding {
	findings := []scanner.Finding{}
	lines := activeLines(content)
	hasRequirePass := false
	for _, line := range lines {
		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(lower, "requirepass ") && strings.TrimSpace(strings.TrimPrefix(line, "requirepass")) != "":
			hasRequirePass = true
		case strings.HasPrefix(lower, "protected-mode no"):
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityCritical,
				Title:       "Redis protected mode is disabled",
				Description: "Keep protected-mode enabled unless Redis is isolated and authenticated.",
				File:        path,
				Category:    "redis",
				RuleID:      "redis_protected_mode_disabled",
			})
		case strings.HasPrefix(lower, "bind ") && strings.Contains(lower, "0.0.0.0"):
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityWarning,
				Title:       "Redis binds to all interfaces",
				Description: "Bind Redis to private interfaces or localhost unless public access is intentional.",
				File:        path,
				Category:    "redis",
				RuleID:      "redis_bind_all_interfaces",
			})
		case strings.HasPrefix(lower, "appendonly no"):
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityInfo,
				Title:       "Redis append-only persistence is disabled",
				Description: "Enable appendonly for workloads that require stronger durability.",
				File:        path,
				Category:    "redis",
				RuleID:      "redis_appendonly_disabled",
			})
		}
	}
	if !hasRequirePass {
		findings = append(findings, scanner.Finding{
			Severity:    scanner.SeverityWarning,
			Title:       "Redis has no requirepass configured",
			Description: "Set requirepass or ACL users for Redis deployments that are reachable outside localhost.",
			File:        path,
			Category:    "redis",
			RuleID:      "redis_missing_auth",
		})
	}
	return findings
}

func scanCompose(path string, data []byte) []scanner.Finding {
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil
	}
	services, _ := raw["services"].(map[string]interface{})
	findings := []scanner.Finding{}
	for name, entry := range services {
		service, _ := entry.(map[string]interface{})
		image := strings.ToLower(stringValue(service["image"]))
		if !strings.HasPrefix(image, "redis") && !strings.Contains(image, "/redis") {
			continue
		}
		command := lowerCommand(service["command"])
		if !strings.Contains(command, "requirepass") && !strings.Contains(command, "--user") {
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityWarning,
				Title:       fmt.Sprintf("Redis service %s does not configure authentication", name),
				Description: "Configure Redis AUTH/ACLs before exposing the service to other networks.",
				File:        path,
				Category:    "redis",
				RuleID:      "redis_compose_missing_auth",
			})
		}
		for _, port := range stringList(service["ports"]) {
			if strings.Contains(port, "6379") && exposesAllInterfaces(port) {
				findings = append(findings, scanner.Finding{
					Severity:    scanner.SeverityWarning,
					Title:       fmt.Sprintf("Redis service %s exposes port 6379", name),
					Description: "Avoid publishing Redis directly on all interfaces.",
					File:        path,
					Category:    "redis",
					RuleID:      "redis_port_exposed",
				})
			}
		}
	}
	return findings
}

var redisURLPattern = regexp.MustCompile(`(?i)REDIS_URL\s*=\s*redis://([^@\s]+)$`)

func scanEnv(path, content string) []scanner.Finding {
	findings := []scanner.Finding{}
	for _, line := range strings.Split(content, "\n") {
		if redisURLPattern.MatchString(strings.TrimSpace(line)) {
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityWarning,
				Title:       "Redis URL does not include credentials",
				Description: "Use authenticated Redis URLs outside local development.",
				File:        path,
				Category:    "redis",
				RuleID:      "redis_url_no_auth",
			})
		}
	}
	return findings
}

func activeLines(content string) []string {
	lines := []string{}
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		lines = append(lines, trimmed)
	}
	return lines
}

func lowerCommand(value interface{}) string {
	switch v := value.(type) {
	case string:
		return strings.ToLower(v)
	case []interface{}:
		parts := []string{}
		for _, part := range v {
			parts = append(parts, fmt.Sprint(part))
		}
		return strings.ToLower(strings.Join(parts, " "))
	default:
		return ""
	}
}

func stringList(value interface{}) []string {
	list := []string{}
	switch v := value.(type) {
	case []interface{}:
		for _, item := range v {
			list = append(list, fmt.Sprint(item))
		}
	case []string:
		list = append(list, v...)
	}
	return list
}

func exposesAllInterfaces(port string) bool {
	return !strings.HasPrefix(port, "127.0.0.1:") && !strings.HasPrefix(port, "localhost:")
}

func stringValue(value interface{}) string {
	result, _ := value.(string)
	return strings.TrimSpace(result)
}

func defaultSkipDirs() map[string]bool {
	return map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
		".next":        true,
	}
}
