package output

import (
	"fmt"

	"devdoctor/internal/scanner"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type scanDoneMsg struct {
	report scanner.Report
	err    error
}

type scanModel struct {
	spinner spinner.Model
	title   string
	runner  func() (scanner.Report, error)
	report  scanner.Report
	err     error
}

func RunScanSpinner(title string, runner func() (scanner.Report, error)) (scanner.Report, error) {
	m := scanModel{
		spinner: spinner.New(),
		title:   title,
		runner:  runner,
	}
	m.spinner.Spinner = spinner.Pulse
	m.spinner.Style = styleBranding

	p := tea.NewProgram(m, tea.WithoutSignals())
	final, err := p.Run()
	if err != nil {
		return scanner.Report{}, err
	}

	finalModel := final.(scanModel)
	return finalModel.report, finalModel.err
}

func (m scanModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, runScanCmd(m.runner))
}

func runScanCmd(runner func() (scanner.Report, error)) tea.Cmd {
	return func() tea.Msg {
		report, err := runner()
		return scanDoneMsg{report: report, err: err}
	}
}

func (m scanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case scanDoneMsg:
		m.report = typed.report
		m.err = typed.err
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m scanModel) View() string {
	if m.err != nil || m.report.Findings != nil {
		return ""
	}
	return fmt.Sprintf("\n %s %s %s\n", m.spinner.View(), styleHeader.Render(m.title), styleMuted.Render("... this may take a few seconds"))
}
