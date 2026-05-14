package redis

import "testing"

func TestScanRedisConfFindsUnsafeSettings(t *testing.T) {
	findings := scanRedisConf("redis.conf", `
protected-mode no
bind 0.0.0.0
appendonly no
`)

	want := map[string]bool{
		"redis_protected_mode_disabled": false,
		"redis_bind_all_interfaces":     false,
		"redis_appendonly_disabled":     false,
		"redis_missing_auth":            false,
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

func TestScanRedisComposeFindsMissingAuthAndPublishedPort(t *testing.T) {
	findings := scanCompose("docker-compose.yml", []byte(`
services:
  cache:
    image: redis:7
    ports:
      - "6379:6379"
`))

	want := map[string]bool{
		"redis_compose_missing_auth": false,
		"redis_port_exposed":         false,
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
