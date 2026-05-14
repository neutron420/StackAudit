package kubernetes

import (
	"testing"

	"devdoctor/internal/rules"
)

func TestScanManifestFindsUnsafeWorkloadSettings(t *testing.T) {
	data := []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
spec:
  template:
    spec:
      hostNetwork: true
      containers:
        - name: app
          image: example/api:latest
          securityContext:
            privileged: true
`)

	findings := scanManifest("deployment.yaml", data, rules.DefaultRuleSet())

	want := map[string]bool{
		"k8s_host_namespace":       false,
		"k8s_latest_image":         false,
		"k8s_missing_resources":    false,
		"k8s_privileged_container": false,
		"k8s_run_as_non_root":      false,
	}
	for _, finding := range findings {
		if _, ok := want[finding.RuleID]; ok {
			want[finding.RuleID] = true
		}
	}
	for ruleID, found := range want {
		if !found {
			t.Fatalf("missing finding %s in %+v", ruleID, findings)
		}
	}
}

func TestScanManifestSkipsNonKubernetesYAML(t *testing.T) {
	findings := scanManifest("rules.yaml", []byte("rules:\n  - id: docker_latest_tag\n"), rules.DefaultRuleSet())
	if len(findings) != 0 {
		t.Fatalf("findings = %+v, want none", findings)
	}
}
