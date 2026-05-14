package cicd

import "testing"

func TestScanWorkflowDetectsPullRequestTargetForms(t *testing.T) {
	tests := map[string]string{
		"scalar": "on: pull_request_target\n",
		"list":   "on:\n  - pull_request_target\n",
		"map":    "on:\n  pull_request_target:\n",
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			findings := scanWorkflow("ci.yml", []byte(data))
			for _, finding := range findings {
				if finding.RuleID == "cicd_pull_request_target" {
					return
				}
			}
			t.Fatalf("missing pull_request_target finding in %+v", findings)
		})
	}
}
