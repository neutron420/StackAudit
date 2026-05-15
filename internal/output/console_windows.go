package output

import (
	"os"
	"strings"

	"golang.org/x/sys/windows"
)

func disableQuickEdit() {
	h := windows.Handle(os.Stdin.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return
	}

	mode &^= windows.ENABLE_QUICK_EDIT_MODE
	mode &^= windows.ENABLE_INSERT_MODE
	mode |= windows.ENABLE_EXTENDED_FLAGS
	mode |= windows.ENABLE_VIRTUAL_TERMINAL_INPUT

	_ = windows.SetConsoleMode(h, mode)
}

func useAltScreen() bool {
	return false
}

func usePlainPrompt() bool {
	return true
}

func usePlainWorkbench() bool {
	if os.Getenv("STACK_PLAIN_WORKBENCH") == "1" {
		return true
	}
	if isVSCodeTerminal() {
		return true
	}
	return false
}

func isVSCodeTerminal() bool {
	termProgram := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	if termProgram == "vscode" {
		return true
	}
	if os.Getenv("VSCODE_IPC_HOOK_CLI") != "" {
		return true
	}
	return false
}
