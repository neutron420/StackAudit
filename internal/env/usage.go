package env

import (
	"os"
	"regexp"
	"strings"

	"devdoctor/internal/utils"
)

var (
	patternShell = regexp.MustCompile(`\$\{?([A-Z0-9_]+)\}?`)
	patternNode  = regexp.MustCompile(`process\.env\.([A-Z0-9_]+)`)
	patternGo    = regexp.MustCompile(`os\.Getenv\(\"([A-Z0-9_]+)\"\)`)
	patternGoAlt = regexp.MustCompile(`os\.LookupEnv\(\"([A-Z0-9_]+)\"\)`)
)

func DiscoverUsedVars(root string) (map[string]bool, error) {
	used := map[string]bool{}
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
			".sh":   true,
			".env":  true,
			"*":     true,
		},
		SkipDirs: defaultSkipDirs(),
		Ignore:   ignore,
	})
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.HasSuffix(file, ".env") {
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		content := string(data)
		for _, match := range patternShell.FindAllStringSubmatch(content, -1) {
			used[match[1]] = true
		}
		for _, match := range patternNode.FindAllStringSubmatch(content, -1) {
			used[match[1]] = true
		}
		for _, match := range patternGo.FindAllStringSubmatch(content, -1) {
			used[match[1]] = true
		}
		for _, match := range patternGoAlt.FindAllStringSubmatch(content, -1) {
			used[match[1]] = true
		}
	}

	return used, nil
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
