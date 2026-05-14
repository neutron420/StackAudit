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
	viewport    viewport.Model
	textInput   textinput.Model
	ready       bool
	output      *strings.Builder // Using a pointer to avoid "copy by value" panics
	execute     func(args []string) string
	width       int
	height      int
	stats       string
}

type statsMsg string

func pollStats() tea.Cmd {
	return func() tea.Msg {
		return statsMsg(getLiveStats())
	}
}

func RunSandbox(execute func(args []string) string) error {
	ti := textinput.New()
	ti.Placeholder = "Type a command (scan, doctor, env...)"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40
	ti.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true).Render("stack > ")

	m := sandboxModel{
		textInput: ti,
		execute:   execute,
		output:    &strings.Builder{},
		stats:     "Initializing stats...",
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

func (m sandboxModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, pollStats())
}

func (m sandboxModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case statsMsg:
		m.stats = string(msg)
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
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
				m.output.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Render("\n[SYSTEM] Output saved to stack_output.txt 📄\n"))
				m.viewport.SetContent(m.output.String())
				m.viewport.GotoBottom()
				m.textInput.Reset()
				return m, nil
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
			m.viewport.YPosition = 14
			m.viewport.SetContent("Welcome to the STACK Sandbox. Type a command to begin.")
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

func (m sandboxModel) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// 🏔️ TOP HEADER (Centered & Clean)
	logoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true)
	logo := logoStyle.Render(`_____ _______       _____ _  __
/ ____|__   __|/\   / ____| |/ /
| (___    | |  /  \ | |    | ' / 
 \___ \   | | / /\ \| |    |  <  
 ____) |  | |/ ____ \ |____| . \ 
|_____/   |_/_/    \_\_____|_|\_\`)

	missionText := lipgloss.NewStyle().Foreground(lipgloss.Color("#9499B0")).Render("The Local-First Backend Health & Security Audit Tool")
	statsLine := lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Render(m.stats)

	// Combine header elements
	headerContent := lipgloss.JoinVertical(lipgloss.Center,
		"\n",
		logo,
		missionText,
		statsLine,
		"\n",
	)
	
	header := centerText(headerContent, m.width)

	// 📊 VIEWPORT & INPUT
	footer := fmt.Sprintf("\n%s\n%s", 
		centerText(lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Render("Ctrl Up/Down Scroll • 'copy' Export • Esc/q Quit"), m.width),
		centerText(lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Render("Modules: env, docker, secrets, redis, k8s, cicd, postgres"), m.width),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		m.viewport.View(),
		m.textInput.View(),
		footer,
	)
}

func centerText(str string, width int) string {
	lines := strings.Split(str, "\n")
	var centered []string
	for _, line := range lines {
		contentWidth := lipgloss.Width(line)
		padding := (width - contentWidth) / 2
		if padding < 0 { padding = 0 }
		centered = append(centered, strings.Repeat(" ", padding)+line)
	}
	return strings.Join(centered, "\n")
}

func max(a, b int) int {
	if a > b { return a }
	return b
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
		// Simple CPU load from /proc/loadavg
		cpuOut, _ := os.ReadFile("/proc/loadavg")
		cpuStr = strings.Fields(string(cpuOut))[0] // 1 min load

		// Simple Memory from /proc/meminfo
		memOut, _ := exec.Command("free", "-m").Output()
		lines := strings.Split(string(memOut), "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) > 3 {
				memStr = fields[3] + " MB Free"
			}
		}

	case "darwin": // macOS
		cpuOut, _ := exec.Command("sysctl", "-n", "vm.loadavg").Output()
		cpuStr = strings.Fields(string(cpuOut))[1] // 1 min load

		memOut, _ := exec.Command("sysctl", "-n", "hw.memsize").Output()
		memStr = strings.TrimSpace(string(memOut)) // Simplified
	}

	if cpuStr == "" { cpuStr = "0" }
	if memStr == "" { memStr = "N/A" }

	return fmt.Sprintf("HOST: %s | OS: %s | CPU: %s | MEM: %s", 
		strings.ToUpper(host), strings.ToUpper(runtime.GOOS), cpuStr, memStr)
}
