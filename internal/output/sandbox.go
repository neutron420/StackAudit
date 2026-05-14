package output

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sandboxModel struct {
	list     list.Model
	choice   string
	quitting bool
	docker   string // "READY", "OFF", "MISSING"
	k8s      string // "READY", "MISSING"
	config   string // "OK", "MISSING"
}

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := styleBranding.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return styleBranding.Copy().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#BD93F9")).
				Padding(0, 1).
				Render(s...)
		}
	}

	fmt.Fprint(w, fn(str))
}

func RunSandbox() (string, error) {
	items := []list.Item{
		item("🚀 Full Health Scan"),
		item("🩺 Environment Doctor"),
		item("🛠️  Interactive Fixer"),
		item("🧪 Create Demo Project"),
		item("📈 View Last Report"),
		item("🚪 EXIT SANDBOX"),
	}

	l := list.New(items, itemDelegate{}, 50, 12)
	l.Title = "WORKBENCH"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#50FA7B")).
		Padding(0, 1).
		Background(lipgloss.Color("#282A36"))

	m := sandboxModel{list: l}
	m.refreshStatus()

	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return "", err
	}

	return final.(sandboxModel).choice, nil
}

func (m sandboxModel) Init() tea.Cmd {
	return nil
}

func (m sandboxModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *sandboxModel) refreshStatus() {
	// Check Docker
	if _, err := exec.LookPath("docker"); err != nil {
		m.docker = "MISSING"
	} else if err := exec.Command("docker", "info").Run(); err != nil {
		m.docker = "OFFLINE"
	} else {
		m.docker = "READY"
	}

	// Check K8s
	if _, err := exec.LookPath("kubectl"); err != nil {
		m.k8s = "MISSING"
	} else {
		m.k8s = "READY"
	}

	// Check Config
	if _, err := os.Stat(".stack.yaml"); os.IsNotExist(err) {
		m.config = "MISSING"
	} else {
		m.config = "OK"
	}
}

func (m sandboxModel) View() string {
	if m.quitting {
		return ""
	}

	// 🎨 Premium Theme & Dimensions
	width := 100 // Fallback
	height := 30 // Fallback

	// 🏔️ TOP HEADER (k9s style)
	headerLeft := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8BE9FD")).
		Render(fmt.Sprintf(
			"Context: %s\nUser:    %s\nCPU:     %s\nMEM:     %s",
			"local", os.Getenv("USERNAME"), "2%", "14%",
		))

	logo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BD93F9")).
		Bold(true).
		Render(`   _____ _______       _____ _  __
  / ____|__   __|/\   / ____| |/ /
 | (___    | |  /  \ | |    | ' / 
  \___ \   | | / /\ \| |    |  <  
  ____) |  | |/ ____ \ |____| . \ 
 |_____/   |_/_/    \_\_____|_|\_\`)

	header := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(40).Render(headerLeft),
		lipgloss.NewStyle().Width(width-40).Align(lipgloss.Right).Render(logo),
	)

	// 📊 MAIN CONTENT
	getStatusStyle := func(status string) lipgloss.Style {
		switch status {
		case "READY", "OK":
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true)
		case "OFFLINE", "MISSING":
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true)
		default:
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C")).Bold(true)
		}
	}

	sidebar := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(lipgloss.Color("#6272A4")).
		Padding(1, 2).
		Width(30).
		Height(height - 12).
		Render(fmt.Sprintf(
			"%s\n\n%s %s\n%s %s\n%s %s",
			lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("#FF79C6")).Render("PROJECT STATUS"),
			"󰄬 Docker:", getStatusStyle(m.docker).Render(m.docker),
			"󰄬 K8s:   ", getStatusStyle(m.k8s).Render(m.k8s),
			"󰄬 Config:", getStatusStyle(m.config).Render(m.config),
		))

	mainArea := lipgloss.NewStyle().
		Padding(1, 4).
		Render(m.list.View())

	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainArea)

	// ⌨️ BOTTOM HOTKEY BAR
	hotkeys := []string{
		lipgloss.NewStyle().Background(lipgloss.Color("#6272A4")).Foreground(lipgloss.Color("#FFFFFF")).Render(" <s> Scan "),
		lipgloss.NewStyle().Background(lipgloss.Color("#6272A4")).Foreground(lipgloss.Color("#FFFFFF")).Render(" <d> Doctor "),
		lipgloss.NewStyle().Background(lipgloss.Color("#6272A4")).Foreground(lipgloss.Color("#FFFFFF")).Render(" <f> Fix "),
		lipgloss.NewStyle().Background(lipgloss.Color("#FF5555")).Foreground(lipgloss.Color("#FFFFFF")).Render(" <q> Quit "),
	}
	footer := lipgloss.NewStyle().
		Background(lipgloss.Color("#44475A")).
		Width(width).
		Padding(0, 1).
		Render(strings.Join(hotkeys, "  "))

	// 🏆 ASSEMBLE EVERYTHING
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Padding(1, 2).Render(header),
		lipgloss.NewStyle().Padding(1, 0).Height(height-8).Render(content),
		footer,
	)
}
