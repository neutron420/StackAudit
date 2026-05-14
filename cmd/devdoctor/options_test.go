package main

import (
	"testing"
	"time"
)

func TestParseModuleTimeoutsSupportsDefaultAndNamedBudgets(t *testing.T) {
	options, err := parseModuleTimeouts([]string{"2s", "env=500ms,secrets=5s,ci=3s"})
	if err != nil {
		t.Fatalf("parseModuleTimeouts returned error: %v", err)
	}

	if options.Default != 2*time.Second {
		t.Fatalf("default timeout = %s, want 2s", options.Default)
	}
	if options.Modules["env"] != 500*time.Millisecond {
		t.Fatalf("env timeout = %s, want 500ms", options.Modules["env"])
	}
	if options.Modules["secrets"] != 5*time.Second {
		t.Fatalf("secrets timeout = %s, want 5s", options.Modules["secrets"])
	}
	if options.Modules["cicd"] != 3*time.Second {
		t.Fatalf("cicd timeout = %s, want 3s", options.Modules["cicd"])
	}
}

func TestParseModuleTimeoutsRejectsInvalidValues(t *testing.T) {
	tests := [][]string{
		{"env="},
		{"=1s"},
		{"env=-1s"},
		{"env=soon"},
	}

	for _, test := range tests {
		if _, err := parseModuleTimeouts(test); err == nil {
			t.Fatalf("parseModuleTimeouts(%v) returned nil error", test)
		}
	}
}
