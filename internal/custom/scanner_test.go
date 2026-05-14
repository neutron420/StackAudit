package custom

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"devdoctor/internal/rules"
)

func TestCustomScannerFindsConfiguredRule(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte("DEBUG=true\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	module := NewScanner("team", []Rule{
		{
			ID:       "no_debug",
			Title:    "Debug mode enabled",
			Severity: "warning",
			Path:     ".env",
			Contains: "DEBUG=true",
		},
	})
	findings, err := module.Scan(context.Background(), root, rules.DefaultRuleSet())
	if err != nil {
		t.Fatalf("Scan returned error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings length = %d, want 1", len(findings))
	}
	if findings[0].RuleID != "plugin:team:no_debug" || findings[0].Line != 1 {
		t.Fatalf("finding = %+v", findings[0])
	}
}

func TestLoadConfigRejectsInvalidRules(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "plugin.yaml")
	if err := os.WriteFile(path, []byte(`
name: team
rules:
  - id: bad
    title: Bad
    severity: urgent
    contains: nope
`), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	if _, err := loadConfig(path); err == nil {
		t.Fatal("loadConfig returned nil error for invalid severity")
	}
}
