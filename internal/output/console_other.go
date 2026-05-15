//go:build !windows

package output

func disableQuickEdit() {}

func useAltScreen() bool {
	return true
}

func usePlainPrompt() bool {
	return true
}

func usePlainWorkbench() bool {
	return true
}

func isVSCodeTerminal() bool {
	return false
}
