package utils

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

type IgnoreMatcher func(path string, isDir bool) bool

type ignoreRule struct {
	pattern   string
	dirOnly   bool
	matchBase bool
}

func LoadIgnoreMatcher(root string) (IgnoreMatcher, error) {
	ignorePath := filepath.Join(root, ".StackAuditignore")
	data, err := os.ReadFile(ignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	rules := []ignoreRule{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "!") {
			// Negation patterns are not supported yet.
			continue
		}
		dirOnly := strings.HasSuffix(trimmed, "/")
		trimmed = strings.TrimSuffix(trimmed, "/")
		pattern := filepath.ToSlash(trimmed)
		if pattern == "" {
			continue
		}
		rules = append(rules, ignoreRule{
			pattern:   pattern,
			dirOnly:   dirOnly,
			matchBase: !strings.Contains(pattern, "/"),
		})
	}

	if len(rules) == 0 {
		return nil, nil
	}

	return func(p string, isDir bool) bool {
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return false
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return false
		}
		base := path.Base(rel)
		for _, rule := range rules {
			if rule.dirOnly && !isDir {
				continue
			}
			if rule.matchBase {
				if ok, _ := path.Match(rule.pattern, base); ok {
					return true
				}
			}
			if ok, _ := path.Match(rule.pattern, rel); ok {
				return true
			}
		}
		return false
	}, nil
}
