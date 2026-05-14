package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"stack/internal/rules"
	"stack/internal/scanner"
	"stack/internal/utils"

	"gopkg.in/yaml.v3"
)

type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Name() string {
	return "postgres"
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
		case name == "postgresql.conf":
			findings = append(findings, scanPostgresConf(file, string(data))...)
		case name == "pg_hba.conf":
			findings = append(findings, scanHBA(file, string(data))...)
		case name == "docker-compose.yml" || name == "docker-compose.yaml" || name == "compose.yml" || name == "compose.yaml":
			findings = append(findings, scanCompose(file, data)...)
		case strings.Contains(name, ".env"):
			findings = append(findings, scanEnv(file, string(data))...)
		}
	}
	if len(findings) == 0 {
		findings = append(findings, scanner.Finding{
			Severity:    scanner.SeverityInfo,
			Title:       "No PostgreSQL service found",
			Description: "We couldn't find any PostgreSQL configuration files, environment variables, or Docker services in your project root.",
			Category:    "postgres",
			RuleID:      "postgres_no_service",
		})
	}
	return findings, nil
}

func scanPostgresConf(path, content string) []scanner.Finding {
	findings := []scanner.Finding{}
	for _, line := range activeLines(content) {
		lower := strings.ToLower(strings.ReplaceAll(line, " ", ""))
		if strings.HasPrefix(lower, "listen_addresses=") && strings.Contains(lower, "'*'") {
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityWarning,
				Title:       "PostgreSQL listens on all interfaces",
				Description: "Bind PostgreSQL to private interfaces unless public access is intentional.",
				File:        path,
				Category:    "postgres",
				RuleID:      "postgres_listen_all_interfaces",
			})
		}
		if strings.HasPrefix(lower, "ssl=off") {
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityWarning,
				Title:       "PostgreSQL SSL is disabled",
				Description: "Enable SSL for networked PostgreSQL deployments.",
				File:        path,
				Category:    "postgres",
				RuleID:      "postgres_ssl_disabled",
			})
		}
	}
	return findings
}

func scanHBA(path, content string) []scanner.Finding {
	findings := []scanner.Finding{}
	for _, line := range activeLines(content) {
		fields := strings.Fields(strings.ToLower(line))
		if len(fields) == 0 {
			continue
		}
		if fields[len(fields)-1] == "trust" {
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityCritical,
				Title:       "PostgreSQL allows trust authentication",
				Description: "Use password, SCRAM, certificate, or peer authentication instead of trust.",
				File:        path,
				Category:    "postgres",
				RuleID:      "postgres_trust_auth",
			})
		}
		if strings.Contains(line, "0.0.0.0/0") || strings.Contains(line, "::/0") {
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityWarning,
				Title:       "PostgreSQL access rule allows every address",
				Description: "Restrict pg_hba.conf CIDR ranges to trusted networks.",
				File:        path,
				Category:    "postgres",
				RuleID:      "postgres_hba_open_cidr",
			})
		}
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
		if !strings.HasPrefix(image, "postgres") && !strings.Contains(image, "/postgres") && !strings.Contains(image, "postgis") {
			continue
		}
		env := environmentMap(service["environment"])
		password := strings.TrimSpace(env["POSTGRES_PASSWORD"])
		if password == "" || isWeakPassword(password) {
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityCritical,
				Title:       fmt.Sprintf("PostgreSQL service %s has weak password configuration", name),
				Description: "Set POSTGRES_PASSWORD to a strong secret or source it from a secret manager.",
				File:        path,
				Category:    "postgres",
				RuleID:      "postgres_weak_password",
			})
		}
		for _, port := range stringList(service["ports"]) {
			if strings.Contains(port, "5432") && exposesAllInterfaces(port) {
				findings = append(findings, scanner.Finding{
					Severity:    scanner.SeverityWarning,
					Title:       fmt.Sprintf("PostgreSQL service %s exposes port 5432", name),
					Description: "Avoid publishing PostgreSQL directly on all interfaces.",
					File:        path,
					Category:    "postgres",
					RuleID:      "postgres_port_exposed",
				})
			}
		}
	}
	return findings
}

var databaseURLPattern = regexp.MustCompile(`(?i)(DATABASE_URL|POSTGRES_URL)\s*=.*sslmode=disable`)

func scanEnv(path, content string) []scanner.Finding {
	findings := []scanner.Finding{}
	for _, line := range strings.Split(content, "\n") {
		if databaseURLPattern.MatchString(strings.TrimSpace(line)) {
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityWarning,
				Title:       "PostgreSQL connection disables SSL",
				Description: "Avoid sslmode=disable for deployed PostgreSQL connections.",
				File:        path,
				Category:    "postgres",
				RuleID:      "postgres_url_ssl_disabled",
			})
		}
	}
	return findings
}

func environmentMap(value interface{}) map[string]string {
	result := map[string]string{}
	switch env := value.(type) {
	case map[string]interface{}:
		for key, value := range env {
			result[key] = fmt.Sprint(value)
		}
	case []interface{}:
		for _, item := range env {
			key, value, ok := strings.Cut(fmt.Sprint(item), "=")
			if ok {
				result[key] = value
			}
		}
	}
	return result
}

func isWeakPassword(value string) bool {
	trimmed := strings.Trim(value, "\"'")
	lower := strings.ToLower(trimmed)
	return lower == "postgres" || lower == "password" || lower == "changeme" || lower == "example"
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
