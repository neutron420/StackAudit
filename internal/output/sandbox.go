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

type sandboxModel struct {
	viewport  viewport.Model
	textInput textinput.Model
	output    strings.Builder
	execute   func([]string) string
	stats     string
	width     int
	height    int
}

func RunSandbox(execute func([]string) string) error {
	m := sandboxModel{
		execute: execute,
		stats:   getLiveStats(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m sandboxModel) Init() tea.Cmd {
	m.textInput = textinput.New()
	m.textInput.Placeholder = "Type a command (scan, ci, env, docker...)"
	m.textInput.Focus()
	m.textInput.CharLimit = 156
	m.textInput.Width = 60
	m.textInput.Prompt = " stack > "
	m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true)

	m.viewport = viewport.New(80, 20)
	m.viewport.SetContent("Welcome to the STACK Workbench. Type a command to begin.\n\nAvailable: scan, scan <module>, ci, env, docker, secrets, redis, k8s, postgres\n           help, copy, clear, quit")

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
			return statsMsg(getLiveStats())
		})

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 12
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
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
	}

	m.textInput, tiCmd = m.textInput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m sandboxModel) View() string {
	logo := `
      ____ _____  _    ____ _  __
     / ___|_   _|/ \  / ___| |/ /
     \___ \ | | / _ \| |   | ' / 
      ___) || |/ ___ \ |___| . \ 
     |____/ |_/_/   \_\____|_|\_\
                                 `

	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BD93F9")).
		Bold(true)

	centeredLogo := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, logoStyle.Render(logo))

	tagline := "The Local-First Backend Health & Security Audit Tool"
	centeredTagline := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, tagline)

	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8BE9FD")).
		Bold(true)
	centeredStats := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, statsStyle.Render(m.stats))

	header := lipgloss.JoinVertical(lipgloss.Center,
		centeredLogo,
		centeredTagline,
		centeredStats,
		strings.Repeat("─", m.width),
	)

	help := lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Render(
		"Ctrl+Up/Down Scroll  |  'copy' Export  |  'clear' Reset  |  Esc/q Quit\n" +
			"         Modules: env, docker, secrets, redis, k8s, cicd, postgres")

	footer := lipgloss.JoinVertical(lipgloss.Left,
		strings.Repeat("─", m.width),
		m.textInput.View(),
		lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		m.viewport.View(),
		footer,
	)
}

type statsMsg string

func pollStats() tea.Cmd {
	return func() tea.Msg {
		return statsMsg(getLiveStats())
	}
}

func getLiveStats() string {
	host, _ := os.Hostname()
	if len(host) > 15 {
		host = host[:12] + "..."
	}

	var cpuStr, memStr, diskStr string

	switch runtime.GOOS {
	case "windows":
		cmdText := "(Get-CimInstance Win32_Processor).LoadPercentage; [math]::Round((Get-CimInstance Win32_OperatingSystem).FreePhysicalMemory / 1MB, 2); [math]::Round((Get-CimInstance Win32_LogicalDisk -Filter \"DeviceID='C:'\").FreeSpace / 1GB, 1)"
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

		if len(valid) >= 3 {
			cpuStr = valid[0] + "%"
			memStr = valid[1] + " GB FREE"
			diskStr = valid[2] + " GB FREE"
		}

	case "linux", "darwin":
		// CPU
		var cpuCmd *exec.Cmd
		if runtime.GOOS == "linux" {
			cpuCmd = exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1\"%\"}'")
		} else {
			cpuCmd = exec.Command("sh", "-c", "top -l 1 | grep 'CPU usage' | awk '{print $3}'")
		}
		cpuOut, _ := cpuCmd.Output()
		cpuStr = strings.TrimSpace(string(cpuOut))

		// MEM
		var memCmd *exec.Cmd
		if runtime.GOOS == "linux" {
			memCmd = exec.Command("sh", "-c", "free -m | grep Mem | awk '{print $4\"MB FREE\"}'")
		} else {
			memCmd = exec.Command("sh", "-c", "vm_stat | perl -ne '/free: +(\\d+)/ && print ($1*4096/1024/1024).\"MB FREE\"'")
		}
		memOut, _ := memCmd.Output()
		memStr = strings.TrimSpace(string(memOut))

		// DISK
		diskCmd := exec.Command("sh", "-c", "df -h / | tail -1 | awk '{print $4\" FREE\"}'")
		diskOut, _ := diskCmd.Output()
		diskStr = strings.TrimSpace(string(diskOut))
	}

	if cpuStr == "" {
		cpuStr = "N/A"
	}
	if memStr == "" {
		memStr = "N/A"
	}
	if diskStr == "" {
		diskStr = "N/A"
	}

	return fmt.Sprintf("HOST: %s  |  OS: %s  |  CPU: %s  |  MEM: %s  |  DISK: %s",
		strings.ToUpper(host), strings.ToUpper(runtime.GOOS), cpuStr, memStr, diskStr)
}
