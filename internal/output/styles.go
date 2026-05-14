package output

import "github.com/charmbracelet/lipgloss"

var (
	styleHeader   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
	styleCritical = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4C4C")).Bold(true)
	styleWarning  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB347")).Bold(true)
	styleInfo     = lipgloss.NewStyle().Foreground(lipgloss.Color("#4FC3F7")).Bold(true)
	styleSuccess  = lipgloss.NewStyle().Foreground(lipgloss.Color("#7CFF9B")).Bold(true)
	styleMuted    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
	styleBranding = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true)
)

func severityIcon(severity string) string {
	switch severity {
	case "critical":
		return "󰀦"
	case "warning":
		return "󰀪"
	case "info":
		return "󰋽"
	case "success":
		return "󰄬"
	default:
		return "•"
	}
}
