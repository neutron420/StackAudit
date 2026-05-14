package output

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	choices  []string
	selected map[int]struct{}
	cursor   int
	canceled bool
}

func initialModel(choices []string) model {
	return model{
		choices:  choices,
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.canceled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		case "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	s := styleHeader.Render("\nSelect modules to include in scan:") + "\n"
	s += styleMuted.Render("(space to toggle, enter to confirm, q to cancel)") + "\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = styleInfo.Render(">")
		}

		checked := "[ ]"
		if _, ok := m.selected[i]; ok {
			checked = styleSuccess.Render("[x]")
		}

		s += fmt.Sprintf("%s %s %s\n", cursor, checked, choice)
	}

	return s + "\n"
}

func MultiSelect(choices []string) ([]string, error) {
	m := initialModel(choices)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	m = finalModel.(model)
	if m.canceled {
		return nil, fmt.Errorf("canceled")
	}

	selected := []string{}
	for i := range m.selected {
		selected = append(selected, m.choices[i])
	}

	if len(selected) == 0 && err == nil {
		// If nothing selected, default to all? Or return error?
		// Let's default to all if they just pressed enter on empty
		return choices, nil
	}

	return selected, nil
}
