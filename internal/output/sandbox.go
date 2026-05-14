package output

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// sandboxModel is the core of our k9s-style interactive shell
type sandboxModel struct {
	viewport  viewport.Model
	textInput textinput.Model
	ready     bool
	output    *strings.Builder
	execute   func(args []string) string
	width     int
	height    int
	stats     string
}

type statsMsg string

func pollStats() tea.Cmd {
	return func() tea.Msg {
		return statsMsg(getLiveStats())
	}
}

func RunSandbox(execute func(args []string) string) error {
	ti := textinput.New()
	ti.Placeholder = "Type a command (scan, ci, env, docker...)"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50
	ti.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true).Render("stack > ")

	m := sandboxModel{
		textInput: ti,
		execute:   execute,
		output:    &strings.Builder{},
		stats:     "Polling system metrics...",
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

func (m sandboxModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, pollStats())
}

type executeMsg struct {
	content string
}

func (m sandboxModel) runCommand(args []string) tea.Cmd {
	return func() tea.Msg {
		return executeMsg{content: m.execute(args)}
	}
}

func (m sandboxModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case executeMsg:
		if msg.content == "__CLEAR__" {
			m.output.Reset()
			m.viewport.SetContent("Screen cleared. Type a command to begin.")
			m.viewport.GotoTop()
		} else {
			m.output.WriteString(msg.content)
			m.viewport.SetContent(m.output.String())
			m.viewport.GotoBottom()
		}
		return m, nil

	case statsMsg:
		m.stats = string(msg)
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return pollStats()()
		})

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+up", "ctrl+k", "shift+up":
			m.viewport.LineUp(1)
			return m, nil
		case "ctrl+down", "ctrl+j", "shift+down":
			m.viewport.LineDown(1)
			return m, nil
		}

		switch msg.Type {
		case tea.KeyUp:
			if m.textInput.Value() == "" {
				m.viewport.LineUp(1)
				return m, nil
			}
		case tea.KeyDown:
			if m.textInput.Value() == "" {
				m.viewport.LineDown(1)
				return m, nil
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyPgUp:
			m.viewport.ViewUp()
			return m, nil
		case tea.KeyPgDown:
			m.viewport.ViewDown()
			return m, nil
		case tea.KeyEnter:
			input := m.textInput.Value()
			if input == "exit" || input == "quit" || input == "q" {
				return m, tea.Quit
			}
			if input == "copy" {
				os.WriteFile("stack_output.txt", []byte(m.output.String()), 0644)
				m.output.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Render("\n[SYSTEM] Output saved to stack_output.txt\n"))
				m.viewport.SetContent(m.output.String())
				m.viewport.GotoBottom()
				m.textInput.Reset()
				return m, nil
			}
			if input != "" {
				m.output.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Render(fmt.Sprintf("\n> %s\n", input)))
				m.viewport.SetContent(m.output.String() + "\nRunning...")
				m.viewport.GotoBottom()
				
				args := strings.Fields(input)
				m.textInput.Reset()
				return m, m.runCommand(args)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Header = 10 lines, input = 1, footer = 3, borders = 2
		vpHeight := msg.Height - 16
		if vpHeight < 4 {
			vpHeight = 4
		}

		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, vpHeight)
			m.viewport.SetContent("Welcome to the STACK Workbench. Type a command to begin.\n\nAvailable: scan, scan <module>, ci, env, docker, secrets, redis, k8s, postgres\n           help, copy, clear, quit")
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = vpHeight
		}
	}

	m.textInput, tiCmd = m.textInput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m sandboxModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	// --- STYLES ---
	purple := lipgloss.Color("#BD93F9")
	muted := lipgloss.Color("#6272A4")
	cyan := lipgloss.Color("#8BE9FD")

	centerStyle := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center)

	// --- LOGO ---
	logo := lipgloss.NewStyle().Foreground(purple).Bold(true).Render(
		" _____ _______       _____ _  __\n" +
			"/ ____|__   __|/\\   / ____| |/ /\n" +
			"| (___    | |  /  \\ | |    | ' / \n" +
			" \\___ \\   | | / /\\ \\| |    |  <  \n" +
			" ____) |  | |/ ____ \\ |____| . \\ \n" +
			"|_____/   |_/_/    \\_\\_____|_|\\_\\")

	// --- MISSION ---
	mission := lipgloss.NewStyle().Foreground(muted).Render("The Local-First Backend Health & Security Audit Tool")

	// --- STATS ---
	stats := lipgloss.NewStyle().Foreground(cyan).Render(m.stats)

	// --- HEADER BLOCK (centered) ---
	header := lipgloss.JoinVertical(lipgloss.Center,
		"",
		centerStyle.Render(logo),
		centerStyle.Render(mission),
		centerStyle.Render(stats),
	)

	// --- VIEWPORT with border ---
	vpBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true, false).
		BorderForeground(muted).
		Width(m.width - 2).
		Render(m.viewport.View())

	// --- INPUT ---
	inputBar := lipgloss.NewStyle().PaddingLeft(1).Render(m.textInput.View())

	// --- FOOTER ---
	footer := centerStyle.Foreground(muted).Render(
		"Ctrl+Up/Down Scroll  |  'copy' Export  |  'clear' Reset  |  Esc/q Quit\n" +
			"Modules: env, docker, secrets, redis, k8s, cicd, postgres")

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		vpBorder,
		inputBar,
		footer,
	)
}

func getLiveStats() string {
	host, _ := os.Hostname()
	if len(host) > 15 {
		host = host[:12] + "..."
	}

	var cpuStr, memStr string

	switch runtime.GOOS {
	case "windows":
		cmdText := "(Get-CimInstance Win32_Processor).LoadPercentage; [math]::Round((Get-CimInstance Win32_OperatingSystem).FreePhysicalMemory / 1MB, 2)"
		cmd := exec.Command("powershell", "-NoProfile", "-Command", cmdText)
		out, _ := cmd.Output()

		raw := string(out)
		lines := strings.Split(raw, "\n")
		var valid []string
		for _, l := range lines {
			t := strings.TrimSpace(l)
			if t != "" {
				valid = append(valid, t)
			}
		}

		if len(valid) >= 2 {
			cpuStr = valid[0] + "%"
			memStr = valid[1] + " GB FREE"
		}

	case "linux":
		cpuOut, _ := os.ReadFile("/proc/loadavg")
		fields := strings.Fields(string(cpuOut))
		if len(fields) > 0 {
			cpuStr = fields[0]
		}

		memOut, _ := exec.Command("free", "-m").Output()
		lines := strings.Split(string(memOut), "\n")
		if len(lines) > 1 {
			memFields := strings.Fields(lines[1])
			if len(memFields) > 3 {
				memStr = memFields[3] + " MB Free"
			}
		}

	case "darwin":
		cpuOut, _ := exec.Command("sysctl", "-n", "vm.loadavg").Output()
		fields := strings.Fields(string(cpuOut))
		if len(fields) > 1 {
			cpuStr = fields[1]
		}

		memOut, _ := exec.Command("sysctl", "-n", "hw.memsize").Output()
		memStr = strings.TrimSpace(string(memOut))
	}

	if cpuStr == "" {
		cpuStr = "N/A"
	}
	if memStr == "" {
		memStr = "N/A"
	}

	return fmt.Sprintf("HOST: %s  |  OS: %s  |  CPU: %s  |  MEM: %s",
		strings.ToUpper(host), strings.ToUpper(runtime.GOOS), cpuStr, memStr)
}
