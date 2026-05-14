package rules

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type RuleSet struct {
	RequiredEnv       map[string]bool
	LatestTagOK       bool
	NoLocalhost       bool
	SeverityOverrides map[string]string
}

type config struct {
	Packs []string `yaml:"packs"`
	Rules []rule   `yaml:"rules"`
}

type rule struct {
	ID         string          `yaml:"id"`
	Severity   string          `yaml:"severity"`
	Env        string          `yaml:"env"`
	Required   *bool           `yaml:"required"`
	Docker     *dockerRule     `yaml:"docker"`
	Production *productionRule `yaml:"production"`
}

type dockerRule struct {
	LatestTag bool `yaml:"latest_tag"`
}

type productionRule struct {
	NoLocalhost bool `yaml:"no_localhost"`
}

func DefaultRuleSet() RuleSet {
	return RuleSet{
		RequiredEnv:       map[string]bool{},
		LatestTagOK:       false,
		NoLocalhost:       true,
		SeverityOverrides: map[string]string{},
	}
}

//go:embed packs/*.yaml
var packFS embed.FS

func LoadFromFile(path string) (RuleSet, error) {
	return Load(path, nil)
}

func Load(path string, packs []string) (RuleSet, error) {
	set := DefaultRuleSet()

	for _, pack := range packs {
		cfg, err := loadPackConfig(pack)
		if err != nil {
			return RuleSet{}, err
		}
		if err := applyConfig(&set, cfg); err != nil {
			return RuleSet{}, err
		}
	}

	if path == "" {
		return set, nil
	}

	parsed, err := loadConfigFile(path)
	if err != nil {
		return RuleSet{}, err
	}

	for _, pack := range parsed.Packs {
		cfg, err := loadPackConfig(pack)
		if err != nil {
			return RuleSet{}, err
		}
		if err := applyConfig(&set, cfg); err != nil {
			return RuleSet{}, err
		}
	}

	if err := applyConfig(&set, parsed); err != nil {
		return RuleSet{}, err
	}

	return set, nil
}

func loadConfigFile(path string) (config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return config{}, err
	}
	parsed := config{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return config{}, err
	}
	return parsed, nil
}

func loadPackConfig(name string) (config, error) {
	if name == "" {
		return config{}, nil
	}
	if isPathLike(name) {
		return loadConfigFile(name)
	}
	path := filepath.ToSlash(filepath.Join("packs", name+".yaml"))
	data, err := packFS.ReadFile(path)
	if err != nil {
		return config{}, fmt.Errorf("rule pack not found: %s", name)
	}
	parsed := config{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return config{}, err
	}
	return parsed, nil
}

func applyConfig(set *RuleSet, parsed config) error {
	for _, rule := range parsed.Rules {
		if rule.Env != "" {
			if rule.Required == nil {
				return fmt.Errorf("env rule for %s missing required flag", rule.Env)
			}
			if rule.Required != nil && *rule.Required {
				set.RequiredEnv[rule.Env] = true
			}
		}
		if rule.Docker != nil {
			set.LatestTagOK = rule.Docker.LatestTag
		}
		if rule.Production != nil {
			set.NoLocalhost = rule.Production.NoLocalhost
		}
		if rule.Severity != "" {
			id := strings.TrimSpace(rule.ID)
			if id == "" {
				id = defaultRuleID(rule)
			}
			if id != "" {
				set.SeverityOverrides[strings.ToLower(id)] = strings.ToLower(strings.TrimSpace(rule.Severity))
			}
		}
	}

	return nil
}

func defaultRuleID(rule rule) string {
	if rule.Env != "" && rule.Required != nil {
		return "env_required"
	}
	if rule.Docker != nil {
		return "docker_latest_tag"
	}
	if rule.Production != nil {
		return "env_localhost"
	}
	return ""
}

func isPathLike(value string) bool {
	if strings.ContainsAny(value, "\\/") {
		return true
	}
	lower := strings.ToLower(value)
	return strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml")
}
