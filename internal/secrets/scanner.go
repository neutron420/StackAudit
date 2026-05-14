package secrets

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
)

type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Name() string {
	return "secrets"
}

func (s *Scanner) Scan(ctx context.Context, root string, ruleSet rules.RuleSet) ([]scanner.Finding, error) {
	ignore, err := utils.LoadIgnoreMatcher(root)
	if err != nil {
		return nil, err
	}
	files, err := utils.WalkFiles(root, utils.WalkOptions{
		MaxSize: 1_000_000,
		Extensions: map[string]bool{
			".go":   true,
			".js":   true,
			".ts":   true,
			".jsx":  true,
			".tsx":  true,
			".py":   true,
			".rb":   true,
			".java": true,
			".yaml": true,
			".yml":  true,
			".json": true,
			".env":  true,
			".txt":  true,
			"*":     true,
		},
		SkipDirs: defaultSkipDirs(),
		Ignore:   ignore,
	})
	if err != nil {
		return nil, err
	}

	findings := []scanner.Finding{}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		content := string(data)
		findings = append(findings, matchSecretPatterns(file, content)...)
		findings = append(findings, matchHighEntropy(file, content)...)
	}

	if len(findings) == 0 {
		findings = append(findings, scanner.Finding{
			Category:    "secrets",
			Title:       "No secrets issues detected",
			Description: "No hardcoded secrets, API keys, or high-entropy strings were found in your codebase.",
			Severity:    scanner.SeveritySuccess,
		})
	}

	return findings, nil
}

var (
	patternAWSAccess = regexp.MustCompile(`AKIA[0-9A-Z]{16}`)
	patternAWSSecret = regexp.MustCompile(`(?i)aws(.{0,20})?(secret|access)[^A-Za-z0-9]?([A-Za-z0-9/+=]{40})`)
	patternGitHub    = regexp.MustCompile(`ghp_[A-Za-z0-9]{36}`)
	patternStripe    = regexp.MustCompile(`sk_live_[A-Za-z0-9]{24}`)
	patternJWT       = regexp.MustCompile(`(?i)jwt[_-]?secret\s*[:=]\s*['\"]?([A-Za-z0-9_\-]{16,})['\"]?`)
)

func matchSecretPatterns(file, content string) []scanner.Finding {
	patterns := []struct {
		ID      string
		Name    string
		Pattern *regexp.Regexp
	}{
		{"secret_aws_access", "AWS access key", patternAWSAccess},
		{"secret_aws_secret", "AWS secret key", patternAWSSecret},
		{"secret_github", "GitHub token", patternGitHub},
		{"secret_stripe", "Stripe secret", patternStripe},
		{"secret_jwt", "JWT secret", patternJWT},
	}

	findings := []scanner.Finding{}
	for _, entry := range patterns {
		matches := entry.Pattern.FindAllStringSubmatch(content, -1)
		if len(matches) == 0 {
			continue
		}
		for _, match := range matches {
			finding := scanner.Finding{
				Severity:    scanner.SeverityCritical,
				Title:       fmt.Sprintf("Potential %s found", entry.Name),
				Description: "Secrets should not be committed or stored in plain text.",
				File:        file,
				Category:    "secrets",
				RuleID:      fmt.Sprintf("secret_%s", normalizeID(entry.Name)),
			}
			if len(match) > 1 {
				finding.Description = fmt.Sprintf("Potential %s detected: %s", entry.Name, maskSecret(match[len(match)-1]))
			}
			findings = append(findings, finding)
		}
	}

	return findings
}

func matchHighEntropy(file, content string) []scanner.Finding {
	findings := []scanner.Finding{}
	for _, token := range extractTokens(file, content) {
		if !looksLikeSecretCandidate(token) {
			continue
		}
		entropy := ShannonEntropy(token)
		if entropy >= 4.2 {
			findings = append(findings, scanner.Finding{
				Severity:    scanner.SeverityWarning,
				Title:       "High-entropy string detected",
				Description: fmt.Sprintf("Potential secret with entropy %.2f: %s", entropy, maskSecret(token)),
				File:        file,
				Category:    "secrets",
				RuleID:      "secret_high_entropy",
			})
		}
	}
	return findings
}

func extractTokens(path, content string) []string {
	result := []string{}

	for _, token := range extractStringLiterals(content) {
		if token != "" {
			result = append(result, token)
		}
	}

	if strings.EqualFold(filepath.Ext(path), ".env") {
		result = append(result, extractEnvValues(content)...)
	}

	return result
}

func extractStringLiterals(content string) []string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`"(?:\\.|[^"\\])*"`),
		regexp.MustCompile(`'(?:\\.|[^'\\])*'`),
		regexp.MustCompile("`[^`]*`"),
	}

	result := []string{}
	for _, pattern := range patterns {
		matches := pattern.FindAllString(content, -1)
		for _, match := range matches {
			value := strings.Trim(match, "\"'`")
			if value != "" {
				result = append(result, value)
			}
		}
	}
	return result
}

func normalizeID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}

func extractEnvValues(content string) []string {
	lines := strings.Split(content, "\n")
	values := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func looksLikeSecretCandidate(token string) bool {
	if len(token) < 24 {
		return false
	}
	if strings.Contains(token, " ") {
		return false
	}
	if looksLikeRegexLiteral(token) {
		return false
	}
	if strings.Contains(token, "/") && strings.Contains(token, ".") {
		return false
	}
	if looksLikePackagePath(token) {
		return false
	}

	classes := map[string]bool{
		"lower":  false,
		"upper":  false,
		"digit":  false,
		"symbol": false,
	}
	for _, r := range token {
		switch {
		case r >= 'a' && r <= 'z':
			classes["lower"] = true
		case r >= 'A' && r <= 'Z':
			classes["upper"] = true
		case r >= '0' && r <= '9':
			classes["digit"] = true
		default:
			classes["symbol"] = true
		}
	}

	count := 0
	for _, seen := range classes {
		if seen {
			count++
		}
	}

	return count >= 3
}

func looksLikeRegexLiteral(token string) bool {
	if strings.Contains(token, "(?") {
		return true
	}
	if strings.Contains(token, "\\") {
		return true
	}
	if strings.ContainsAny(token, "[](){}^$|*+?") {
		return true
	}
	return false
}

func looksLikePackagePath(token string) bool {
	if strings.Contains(token, "github.com") || strings.Contains(token, "golang.org") || strings.Contains(token, "gopkg.in") || strings.Contains(token, "gitlab.com") || strings.Contains(token, "bitbucket.org") {
		return true
	}

	packagePath := regexp.MustCompile(`^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+`)
	return packagePath.MatchString(token)
}

func maskSecret(value string) string {
	if len(value) <= 8 {
		return "******"
	}
	return value[:4] + "…" + value[len(value)-4:]
}

func defaultSkipDirs() map[string]bool {
	return map[string]bool{
		".git":         true,
		".idea":        true,
		".vscode":      true,
		"node_modules": true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
		"bin":          true,
		".terraform":   true,
	}
}
