package custom

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"stack/internal/rules"
	"stack/internal/scanner"
	"stack/internal/utils"

	"gopkg.in/yaml.v3"
)

type Scanner struct {
	name  string
	rules []Rule
}

type Config struct {
	Name  string `yaml:"name"`
	Rules []Rule `yaml:"rules"`
}

type Rule struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Severity    string   `yaml:"severity"`
	Category    string   `yaml:"category"`
	Path        string   `yaml:"path"`
	Paths       []string `yaml:"paths"`
	Contains    string   `yaml:"contains"`
	Regex       string   `yaml:"regex"`
}

func NewScanner(name string, pluginRules []Rule) *Scanner {
	if strings.TrimSpace(name) == "" {
		name = "custom"
	}
	return &Scanner{name: name, rules: pluginRules}
}

func NewScanners(root string, paths []string) ([]scanner.Module, error) {
	pluginPaths, err := resolvePluginPaths(root, paths)
	if err != nil {
		return nil, err
	}
	modules := []scanner.Module{}
	for _, pluginPath := range pluginPaths {
		cfg, err := loadConfig(pluginPath)
		if err != nil {
			return nil, err
		}
		modules = append(modules, NewScanner(cfg.Name, cfg.Rules))
	}
	return modules, nil
}

func (s *Scanner) Name() string {
	return "plugin:" + s.name
}

func (s *Scanner) Scan(ctx context.Context, root string, ruleSet rules.RuleSet) ([]scanner.Finding, error) {
	if len(s.rules) == 0 {
		return nil, nil
	}
	ignore, err := utils.LoadIgnoreMatcher(root)
	if err != nil {
		return nil, err
	}
	files, err := utils.WalkFiles(root, utils.WalkOptions{
		MaxSize:  1_000_000,
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
		rel, err := filepath.Rel(root, file)
		if err != nil {
			rel = file
		}
		rel = filepath.ToSlash(rel)

		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		content := string(data)
		for _, rule := range s.rules {
			if !ruleApplies(rule, rel) {
				continue
			}
			line, ok, err := matchesRule(rule, content)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			findings = append(findings, findingForRule(s.name, rule, file, line))
		}
	}
	return findings, nil
}

func resolvePluginPaths(root string, values []string) ([]string, error) {
	paths := []string{}
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			if !filepath.IsAbs(item) {
				rootPath := filepath.Join(root, item)
				if _, err := os.Stat(rootPath); err == nil {
					item = rootPath
				}
			}
			paths = append(paths, item)
		}
	}
	if len(paths) == 0 {
		matches, err := filepath.Glob(filepath.Join(root, ".stack", "plugins", "*.yaml"))
		if err != nil {
			return nil, err
		}
		ymlMatches, err := filepath.Glob(filepath.Join(root, ".stack", "plugins", "*.yml"))
		if err != nil {
			return nil, err
		}
		paths = append(matches, ymlMatches...)
	}
	return paths, nil
}

func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if strings.TrimSpace(cfg.Name) == "" {
		cfg.Name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	for i, rule := range cfg.Rules {
		if strings.TrimSpace(rule.ID) == "" {
			return Config{}, fmt.Errorf("%s rule %d missing id", path, i+1)
		}
		if strings.TrimSpace(rule.Title) == "" {
			return Config{}, fmt.Errorf("%s rule %s missing title", path, rule.ID)
		}
		if _, ok := parseSeverity(rule.Severity); !ok {
			return Config{}, fmt.Errorf("%s rule %s has invalid severity %q", path, rule.ID, rule.Severity)
		}
		if strings.TrimSpace(rule.Contains) == "" && strings.TrimSpace(rule.Regex) == "" {
			return Config{}, fmt.Errorf("%s rule %s needs contains or regex", path, rule.ID)
		}
		if strings.TrimSpace(rule.Regex) != "" {
			if _, err := regexp.Compile(rule.Regex); err != nil {
				return Config{}, fmt.Errorf("%s rule %s invalid regex: %w", path, rule.ID, err)
			}
		}
	}
	return cfg, nil
}

func ruleApplies(rule Rule, rel string) bool {
	patterns := rule.Paths
	if rule.Path != "" {
		patterns = append(patterns, rule.Path)
	}
	if len(patterns) == 0 {
		return true
	}
	for _, pattern := range patterns {
		if matchPath(pattern, rel) {
			return true
		}
	}
	return false
}

func matchesRule(rule Rule, content string) (int, bool, error) {
	var compiled *regexp.Regexp
	if strings.TrimSpace(rule.Regex) != "" {
		regex, err := regexp.Compile(rule.Regex)
		if err != nil {
			return 0, false, err
		}
		compiled = regex
	}
	for i, line := range strings.Split(content, "\n") {
		if rule.Contains != "" && strings.Contains(line, rule.Contains) {
			return i + 1, true, nil
		}
		if compiled != nil && compiled.MatchString(line) {
			return i + 1, true, nil
		}
	}
	return 0, false, nil
}

func findingForRule(pluginName string, rule Rule, file string, line int) scanner.Finding {
	severity, _ := parseSeverity(rule.Severity)
	category := strings.TrimSpace(rule.Category)
	if category == "" {
		category = "custom"
	}
	return scanner.Finding{
		Severity:    severity,
		Title:       rule.Title,
		Description: rule.Description,
		File:        file,
		Line:        line,
		Category:    category,
		RuleID:      "plugin:" + pluginName + ":" + rule.ID,
	}
}

func matchPath(pattern, rel string) bool {
	pattern = filepath.ToSlash(strings.TrimSpace(pattern))
	rel = filepath.ToSlash(rel)
	if pattern == "" {
		return false
	}
	if pattern == "**" || pattern == "*" {
		return true
	}
	if strings.HasPrefix(pattern, "**/") {
		suffix := strings.TrimPrefix(pattern, "**/")
		if ok, _ := path.Match(suffix, path.Base(rel)); ok {
			return true
		}
		return strings.HasSuffix(rel, strings.TrimPrefix(suffix, "*"))
	}
	if ok, _ := path.Match(pattern, rel); ok {
		return true
	}
	if !strings.Contains(pattern, "/") {
		if ok, _ := path.Match(pattern, path.Base(rel)); ok {
			return true
		}
	}
	return false
}

func parseSeverity(value string) (scanner.Severity, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(scanner.SeverityCritical):
		return scanner.SeverityCritical, true
	case string(scanner.SeverityWarning):
		return scanner.SeverityWarning, true
	case string(scanner.SeverityInfo):
		return scanner.SeverityInfo, true
	case string(scanner.SeveritySuccess):
		return scanner.SeveritySuccess, true
	default:
		return "", false
	}
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
