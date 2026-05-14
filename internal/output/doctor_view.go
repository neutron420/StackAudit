package output

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type CheckResult struct {
	Name     string
	Status   string // success, warning, error
	Message  string
	Hint     string
}

func RenderDoctor(results []CheckResult) string {
	var b strings.Builder
	b.WriteString(styleHeader.Render("\nEnvironment Diagnostic Result\n"))
	b.WriteString(strings.Repeat("─", 40) + "\n\n")

	for _, res := range results {
		var icon string
		var style lipgloss.Style

		switch res.Status {
		case "success":
			icon = severityIcon("success")
			style = styleSuccess
		case "warning":
			icon = severityIcon("warning")
			style = styleWarning
		case "error":
			icon = severityIcon("critical")
			style = styleCritical
		}

		title := style.Render(fmt.Sprintf("%s %s", icon, res.Name))
		b.WriteString(title + "\n")
		
		if res.Message != "" {
			b.WriteString(styleMuted.Render("  " + res.Message) + "\n")
		}
		
		if res.Hint != "" && res.Status != "success" {
			b.WriteString(styleInfo.Render("  ➜ Hint: ") + res.Hint + "\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}
