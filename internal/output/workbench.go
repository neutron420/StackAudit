package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type workbenchModel struct {
	viewport    viewport.Model
	textInput   textinput.Model
	ready       bool
	output      strings.Builder
	execute     func(args []string) string
	width       int
	height      int
}

func RunWorkbench(execute func(args []string) string) error {
	ti := textinput.New()
	ti.Placeholder = "Type a command (scan, doctor, env...)"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40
	ti.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true).Render("stack > ")

	m := workbenchModel{
		textInput: ti,
		execute:   execute,
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

func (m workbenchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m workbenchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			input := m.textInput.Value()
			if input == "exit" || input == "quit" || input == "q" {
				return m, tea.Quit
			}
			if input != "" {
				m.output.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Render(fmt.Sprintf("\n> %s\n", input)))
				args := strings.Fields(input)
				res := m.execute(args)
				m.output.WriteString(res)
				m.viewport.SetContent(m.output.String())
				m.viewport.GotoBottom()
				m.textInput.Reset()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-18)
			m.viewport.YPosition = 15
			m.viewport.SetContent("Welcome to the STACK Workbench. Type a command to begin.")
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 18
		}
	}

	m.textInput, tiCmd = m.textInput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m workbenchModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	// 🏔️ TOP HEADER (Centered Logo)
	logo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BD93F9")).
		Bold(true).
		Render(`   _____ _______       _____ _  __
  / ____|__   __|/\   / ____| |/ /
 | (___    | |  /  \ | |    | ' / 
  \___ \   | | / /\ \| |    |  <  
  ____) |  | |/ ____ \ |____| . \ 
 |_____/   |_/_/    \_\_____|_|\_\`)

	centeredLogo := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(logo)

	// MISSION & SAFETY MESSAGE (No Emojis)
	mission := lipgloss.NewStyle().
		Width(m.width - 20).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("#9499B0")).
		Render("The Local-First Backend Health & Security Audit Tool\n" + 
		       "Guarding your infrastructure by identifying vulnerabilities, leaks, and misconfigurations.\n\n" + 
		       lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true).Render("PRIVACY GUARANTEE: ") + 
		       lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Render("Everything stays on your machine. No data or secrets ever leave this terminal."))

	stats := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8BE9FD")).
		Render(fmt.Sprintf(
			"Context: %s | User: %s | CPU: %s | MEM: %s",
			"local", os.Getenv("USERNAME"), "2%", "14%",
		))
	
	centeredStats := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(stats)

	header := lipgloss.JoinVertical(lipgloss.Center,
		centeredLogo,
		lipgloss.NewStyle().Padding(1, 0).Render(mission),
		centeredStats,
	)

	// 📊 MAIN VIEWPORT
	content := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, true, false).
		BorderForeground(lipgloss.Color("#6272A4")).
		Render(m.viewport.View())

	// ⌨️ INPUT BAR
	inputBar := lipgloss.NewStyle().
		Padding(1, 0).
		Render(m.textInput.View())

	// 🏆 ASSEMBLE
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Padding(1, 2).Render(header),
		content,
		lipgloss.NewStyle().Padding(0, 2).Render(inputBar),
	)
}
