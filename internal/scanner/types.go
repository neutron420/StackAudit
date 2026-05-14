package scanner

import "time"

type Severity string

func (s Severity) IsAtLeast(severity Severity) bool {
	order := map[Severity]int{
		SeverityCritical: 4,
		SeverityWarning:  3,
		SeverityInfo:     2,
		SeveritySuccess:  1,
	}
	return order[s] >= order[severity]
}

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
	SeveritySuccess  Severity = "success"
)

type Finding struct {
	Severity    Severity `json:"severity"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Remediation string   `json:"remediation,omitempty"`
	File        string   `json:"file,omitempty"`
	Line        int      `json:"line,omitempty"`
	Category    string   `json:"category,omitempty"`
	RuleID      string   `json:"rule_id,omitempty"`
}

type Summary struct {
	Critical int `json:"critical"`
	Warning  int `json:"warning"`
	Info     int `json:"info"`
	Success  int `json:"success"`
	Total    int `json:"total"`
}

type Meta struct {
	RootPath string        `json:"root_path"`
	Duration time.Duration `json:"duration"`
}

type TimeoutOptions struct {
	Default time.Duration
	Modules map[string]time.Duration
}

type Report struct {
	Findings []Finding `json:"findings"`
	Summary  Summary   `json:"summary"`
	Scores   Scores    `json:"scores"`
	Meta     Meta      `json:"meta"`
}

type Scores struct {
	Overall        int `json:"overall"`
	Security       int `json:"security"`
	Infrastructure int `json:"infrastructure"`
	Configuration  int `json:"configuration"`
}
