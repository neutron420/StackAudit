package projectinit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteStarterSkipsExistingFilesUnlessForced(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, ".stack.yaml")
	if err := os.WriteFile(configPath, []byte("output: json\n"), 0o644); err != nil {
		t.Fatalf("write existing config: %v", err)
	}

	written, err := WriteStarter(root, Options{})
	if err != nil {
		t.Fatalf("WriteStarter returned error: %v", err)
	}
	if containsPath(written, configPath) {
		t.Fatalf("expected existing config to be skipped, written=%v", written)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if string(data) != "output: json\n" {
		t.Fatalf("existing config was overwritten: %q", data)
	}

	written, err = WriteStarter(root, Options{Force: true})
	if err != nil {
		t.Fatalf("WriteStarter force returned error: %v", err)
	}
	if !containsPath(written, configPath) {
		t.Fatalf("expected forced config write, written=%v", written)
	}
}

func TestWriteGitHubActionsCreatesWorkflow(t *testing.T) {
	root := t.TempDir()
	written, err := WriteGitHubActions(root, Options{})
	if err != nil {
		t.Fatalf("WriteGitHubActions returned error: %v", err)
	}
	workflow := filepath.Join(root, ".github", "workflows", "stack.yml")
	if !containsPath(written, workflow) {
		t.Fatalf("workflow not reported as written: %v", written)
	}
	if _, err := os.Stat(workflow); err != nil {
		t.Fatalf("workflow not created: %v", err)
	}
}

func containsPath(paths []string, target string) bool {
	for _, path := range paths {
		if path == target {
			return true
		}
	}
	return false
}
