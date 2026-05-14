package output

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func RunShell(execute func(args []string) error) error {
	scanner := bufio.NewScanner(os.Stdin)

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BD93F9")).
		Bold(true)

	fmt.Println(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B")).
		Bold(true).
		Padding(1, 0).
		Render("Welcome to the STACK Sandbox Shell"))
	fmt.Println(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6272A4")).
		Render("Type 'help' for commands, 'exit' to quit."))

	for {
		fmt.Print(promptStyle.Render("stack > "))
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		args := strings.Fields(input)
		if len(args) == 0 {
			continue
		}

		cmd := args[0]
		if cmd == "exit" || cmd == "quit" || cmd == "q" {
			fmt.Println("Goodbye!")
			break
		}

		if cmd == "clear" || cmd == "cls" {
			fmt.Print("\033[H\033[2J")
			continue
		}

		// Execute the command using the passed function
		err := execute(args)
		if err != nil {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render(fmt.Sprintf("Error: %v", err)))
		}
		fmt.Println() // Add a newline for spacing
	}

	return nil
}
