package postgres

import "testing"

func TestScanPostgresConfigAndHBASurfaceUnsafeSettings(t *testing.T) {
	confFindings := scanPostgresConf("postgresql.conf", `
listen_addresses = '*'
ssl = off
`)
	hbaFindings := scanHBA("pg_hba.conf", `
host all all 0.0.0.0/0 trust
`)
	findings := append(confFindings, hbaFindings...)

	want := map[string]bool{
		"postgres_listen_all_interfaces": false,
		"postgres_ssl_disabled":          false,
		"postgres_trust_auth":            false,
		"postgres_hba_open_cidr":         false,
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

func TestScanPostgresComposeFindsWeakPasswordAndPublishedPort(t *testing.T) {
	findings := scanCompose("docker-compose.yml", []byte(`
services:
  db:
    image: postgres:16
    environment:
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
`))

	want := map[string]bool{
		"postgres_weak_password": false,
		"postgres_port_exposed":  false,
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
