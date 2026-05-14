package githooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallAndUninstallHookBlock(t *testing.T) {
	root := t.TempDir()
	hooksPath := filepath.Join(root, ".git", "hooks")
	if err := os.MkdirAll(hooksPath, 0o755); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}
	hookPath := filepath.Join(hooksPath, "pre-commit")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\n\necho existing\n"), 0o755); err != nil {
		t.Fatalf("write existing hook: %v", err)
	}

	installed, err := Install(root, []string{"pre-commit"}, Options{Command: "StackAudit scan --exit-code"})
	if err != nil {
		t.Fatalf("Install returned error: %v", err)
	}
	if len(installed) != 1 {
		t.Fatalf("installed = %v, want one path", installed)
	}
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "echo existing") || !strings.Contains(content, startMarker) || !strings.Contains(content, "StackAudit scan --exit-code") {
		t.Fatalf("unexpected hook content:\n%s", content)
	}

	updated, err := Uninstall(root, []string{"pre-commit"})
	if err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}
	if len(updated) != 1 {
		t.Fatalf("updated = %v, want one path", updated)
	}
	data, err = os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("read hook after uninstall: %v", err)
	}
	content = string(data)
	if strings.Contains(content, startMarker) || !strings.Contains(content, "echo existing") {
		t.Fatalf("unexpected hook content after uninstall:\n%s", content)
	}
}
