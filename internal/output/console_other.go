//go:build !windows

package output

func disableQuickEdit() {}

func useAltScreen() bool {
	return true
}

func usePlainPrompt() bool {
	return false
}

func usePlainWorkbench() bool {
	return false
}

func isVSCodeTerminal() bool {
	return false
}
