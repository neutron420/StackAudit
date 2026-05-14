package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppliesRulePacksAndSeverityOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yaml")
	data := []byte(`
packs:
  - strict
rules:
  - id: docker_latest_tag
    severity: info
  - env: JWT_SECRET
    required: true
    severity: critical
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	set, err := Load(path, []string{"relaxed"})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if set.SeverityOverrides["env_localhost"] != "critical" {
		t.Fatalf("env_localhost severity = %q, want strict critical", set.SeverityOverrides["env_localhost"])
	}
	if set.SeverityOverrides["docker_latest_tag"] != "info" {
		t.Fatalf("docker_latest_tag severity = %q, want file override info", set.SeverityOverrides["docker_latest_tag"])
	}
	if !set.RequiredEnv["JWT_SECRET"] {
		t.Fatal("expected JWT_SECRET required env rule")
	}
	if set.SeverityOverrides["env_required"] != "critical" {
		t.Fatalf("env_required severity = %q, want critical", set.SeverityOverrides["env_required"])
	}
}

func TestLoadRejectsInvalidSeverity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yaml")
	data := []byte(`
rules:
  - id: docker_latest_tag
    severity: urgent
`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := Load(path, nil); err == nil {
		t.Fatal("Load returned nil error for invalid severity")
	}
}
