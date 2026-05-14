package githooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	startMarker = "# stack hook:start"
	endMarker   = "# stack hook:end"
)

var validHooks = map[string]bool{
	"pre-commit": true,
	"pre-push":   true,
}

type Options struct {
	Command string
}

func Install(root string, hooks []string, opts Options) ([]string, error) {
	if len(hooks) == 0 {
		hooks = []string{"pre-commit", "pre-push"}
	}
	command := strings.TrimSpace(opts.Command)
	if command == "" {
		command = `stack scan --exit-code --min-severity warning`
	}
	hooksDir, err := hooksDir(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return nil, err
	}

	installed := []string{}
	for _, hook := range hooks {
		if err := validateHook(hook); err != nil {
			return nil, err
		}
		path := filepath.Join(hooksDir, hook)
		current, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		next := upsertBlock(string(current), command)
		if err := os.WriteFile(path, []byte(next), 0o755); err != nil {
			return nil, err
		}
		installed = append(installed, path)
	}
	return installed, nil
}

func Uninstall(root string, hooks []string) ([]string, error) {
	if len(hooks) == 0 {
		hooks = []string{"pre-commit", "pre-push"}
	}
	hooksDir, err := hooksDir(root)
	if err != nil {
		return nil, err
	}
	updated := []string{}
	for _, hook := range hooks {
		if err := validateHook(hook); err != nil {
			return nil, err
		}
		path := filepath.Join(hooksDir, hook)
		current, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		next := removeBlock(string(current))
		if strings.TrimSpace(next) == "" {
			if err := os.Remove(path); err != nil {
				return nil, err
			}
		} else if err := os.WriteFile(path, []byte(next), 0o755); err != nil {
			return nil, err
		}
		updated = append(updated, path)
	}
	return updated, nil
}

func hooksDir(root string) (string, error) {
	gitDir := filepath.Join(root, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%s is not a git repository", root)
		}
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("git worktrees with .git files are not supported yet")
	}
	return filepath.Join(gitDir, "hooks"), nil
}

func validateHook(hook string) error {
	if !validHooks[hook] {
		return fmt.Errorf("unsupported git hook %q", hook)
	}
	return nil
}

func upsertBlock(current string, command string) string {
	block := hookBlock(command)
	without := removeBlock(current)
	trimmed := strings.TrimRight(without, "\n")
	if strings.TrimSpace(trimmed) == "" {
		return "#!/bin/sh\n\n" + block
	}
	if !strings.HasPrefix(trimmed, "#!") {
		trimmed = "#!/bin/sh\n\n" + trimmed
	}
	return trimmed + "\n\n" + block
}

func removeBlock(current string) string {
	start := strings.Index(current, startMarker)
	end := strings.Index(current, endMarker)
	if start == -1 || end == -1 || end < start {
		return current
	}
	end += len(endMarker)
	if end < len(current) && current[end] == '\n' {
		end++
	}
	return strings.TrimRight(current[:start], "\n") + "\n" + strings.TrimLeft(current[end:], "\n")
}

func hookBlock(command string) string {
	return startMarker + `
if command -v stack >/dev/null 2>&1; then
  ` + command + `
else
  echo "stack not found; skipping stack scan"
fi
` + endMarker + `
`
}
