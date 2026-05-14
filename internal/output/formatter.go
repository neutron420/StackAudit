package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"stack/internal/scanner"
)

type Mode string

const (
	ModeTable    Mode = "table"
	ModeJSON     Mode = "json"
	ModeMarkdown Mode = "markdown"
	ModeSARIF    Mode = "sarif"
	ModeHTML     Mode = "html"
)

func ParseMode(value string) (Mode, error) {
	mode := Mode(strings.ToLower(value))
	switch mode {
	case ModeTable, ModeJSON, ModeMarkdown, ModeSARIF, ModeHTML:
		return mode, nil
	default:
		return ModeTable, fmt.Errorf("invalid output mode: %s", value)
	}
}

func Render(report scanner.Report, mode Mode) (string, error) {
	switch mode {
	case ModeJSON:
		payload, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return "", err
		}
		return string(payload), nil
	case ModeSARIF:
		return renderSarif(report)
	case ModeMarkdown:
		return renderMarkdown(report), nil
	case ModeHTML:
		return renderHTML(report)
	default:
		return renderTable(report), nil
	}
}
