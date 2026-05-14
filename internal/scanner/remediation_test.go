package scanner

import "testing"

func TestApplyRemediationsUsesBaseRuleID(t *testing.T) {
	findings := ApplyRemediations([]Finding{{RuleID: "docker_latest_tag:compose"}})
	if findings[0].Remediation == "" {
		t.Fatal("expected remediation for base rule id")
	}
}
