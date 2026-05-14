package output

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"devdoctor/internal/scanner"
)

type sarifLog struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name string `json:"name"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId,omitempty"`
	Level     string          `json:"level,omitempty"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations,omitempty"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           *sarifRegion          `json:"region,omitempty"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int `json:"startLine"`
}

func renderSarif(report scanner.Report) (string, error) {
	results := []sarifResult{}
	for _, finding := range report.Findings {
		if finding.Severity == scanner.SeveritySuccess {
			continue
		}
		level := sarifLevel(finding.Severity)
		message := finding.Title
		if finding.Description != "" {
			message = message + " - " + finding.Description
		}

		result := sarifResult{
			RuleID:  sarifRuleID(finding),
			Level:   level,
			Message: sarifMessage{Text: message},
		}

		if finding.File != "" {
			location := sarifLocation{
				PhysicalLocation: sarifPhysicalLocation{
					ArtifactLocation: sarifArtifactLocation{URI: filepath.ToSlash(finding.File)},
				},
			}
			if finding.Line > 0 {
				location.PhysicalLocation.Region = &sarifRegion{StartLine: finding.Line}
			}
			result.Locations = []sarifLocation{location}
		}

		results = append(results, result)
	}

	log := sarifLog{
		Version: "2.1.0",
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Runs: []sarifRun{
			{
				Tool:    sarifTool{Driver: sarifDriver{Name: "DevDoctor"}},
				Results: results,
			},
		},
	}

	payload, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func sarifLevel(sev scanner.Severity) string {
	switch sev {
	case scanner.SeverityCritical:
		return "error"
	case scanner.SeverityWarning:
		return "warning"
	case scanner.SeverityInfo:
		return "note"
	default:
		return "none"
	}
}

func sarifRuleID(finding scanner.Finding) string {
	if finding.RuleID != "" {
		return finding.RuleID
	}
	base := "DEVDOCTOR"
	if finding.Category != "" {
		base = base + "_" + slugify(finding.Category)
	}
	if finding.Title != "" {
		base = base + "_" + slugify(finding.Title)
	}
	return base
}

func slugify(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "RULE"
	}
	builder := strings.Builder{}
	lastUnderscore := false
	for _, r := range value {
		if r >= 'a' && r <= 'z' {
			builder.WriteByte(byte(r - 32))
			lastUnderscore = false
			continue
		}
		if r >= 'A' && r <= 'Z' {
			builder.WriteByte(byte(r))
			lastUnderscore = false
			continue
		}
		if r >= '0' && r <= '9' {
			builder.WriteByte(byte(r))
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}
	result := strings.Trim(builder.String(), "_")
	if result == "" {
		return "RULE"
	}
	return result
}
