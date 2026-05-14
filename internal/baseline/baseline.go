package baseline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"stackaudit/internal/scanner"
)

type Entry struct {
	RuleID   string `json:"rule_id"`
	Category string `json:"category"`
	Title    string `json:"title"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

type Snapshot struct {
	Version  int     `json:"version"`
	Entries  []Entry `json:"entries"`
	RootHint string  `json:"root_hint"`
}

const currentVersion = 1

func Load(path string) (Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, err
	}

	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return Snapshot{}, err
	}

	return snapshot, nil
}

func Write(path string, root string, report scanner.Report) error {
	snapshot := Snapshot{
		Version:  currentVersion,
		Entries:  entriesFromReport(report, root),
		RootHint: filepath.ToSlash(root),
	}

	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, payload, 0o644)
}

func Filter(report scanner.Report, snapshot Snapshot, root string) scanner.Report {
	if len(snapshot.Entries) == 0 {
		return report
	}

	set := map[string]struct{}{}
	for _, entry := range snapshot.Entries {
		key := entryKey(entry)
		set[key] = struct{}{}
	}

	filtered := []scanner.Finding{}
	for _, finding := range report.Findings {
		entry := entryFromFinding(finding, root)
		if _, ok := set[entryKey(entry)]; ok {
			continue
		}
		filtered = append(filtered, finding)
	}

	report.Findings = filtered
	return scanner.Rebuild(report)
}

func entriesFromReport(report scanner.Report, root string) []Entry {
	entries := []Entry{}
	for _, finding := range report.Findings {
		if finding.Severity == scanner.SeveritySuccess {
			continue
		}
		entries = append(entries, entryFromFinding(finding, root))
	}
	return entries
}

func entryFromFinding(finding scanner.Finding, root string) Entry {
	file := normalizePath(root, finding.File)
	return Entry{
		RuleID:   strings.ToLower(strings.TrimSpace(finding.RuleID)),
		Category: strings.ToLower(strings.TrimSpace(finding.Category)),
		Title:    strings.TrimSpace(finding.Title),
		File:     file,
		Line:     finding.Line,
	}
}

func entryKey(entry Entry) string {
	return strings.ToLower(entry.RuleID + "|" + entry.Category + "|" + entry.Title + "|" + entry.File + "|" + lineToString(entry.Line))
}

func normalizePath(root string, path string) string {
	if path == "" {
		return ""
	}
	clean := filepath.Clean(path)
	if root != "" {
		if rel, err := filepath.Rel(root, clean); err == nil {
			if rel != "." && !strings.HasPrefix(rel, "..") {
				return filepath.ToSlash(rel)
			}
		}
	}
	return filepath.ToSlash(clean)
}

func lineToString(value int) string {
	if value <= 0 {
		return "0"
	}
	return strconv.Itoa(value)
}
