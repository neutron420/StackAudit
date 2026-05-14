package env

import "devdoctor/internal/scanner"

type FileEnv struct {
	Path       string
	Vars       map[string]string
	Duplicates []string
	Empty      []string
	IsExample  bool
	IsProd     bool
}

type Report struct {
	Files       []FileEnv
	UsedVars    map[string]bool
	UnusedVars  map[string][]string
	Findings    []scanner.Finding
	RequiredEnv []string
}
