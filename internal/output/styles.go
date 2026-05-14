package output

import "github.com/charmbracelet/lipgloss"

var (
	styleHeader   = lipgloss.NewStyle().Bold(true)
	styleCritical = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	styleWarning  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	styleInfo     = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
	styleSuccess  = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	styleMuted    = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

func severityIcon(severity string) string {
	switch severity {
	case "critical":
		return "🔥"
	case "warning":
		return "⚠"
	case "info":
		return "ℹ"
	case "success":
		return "✅"
	default:
		return "•"
	}
}
