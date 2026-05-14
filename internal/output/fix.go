package output

import (
	"fmt"
	"strings"

	"stackaudit/internal/fix"
)

func RenderFixPlan(plan fix.Plan) string {
	builder := &strings.Builder{}
	fmt.Fprintln(builder, styleHeader.Render("Planned Fixes"))
	fmt.Fprintln(builder, "")
	for _, action := range plan.Actions {
		fmt.Fprintf(builder, "- %s\n", action.Description)
		if len(action.Files) > 0 {
			fmt.Fprintf(builder, "  Files: %s\n", strings.Join(action.Files, ", "))
		}
	}
	if len(plan.Notes) > 0 {
		fmt.Fprintln(builder, "")
		fmt.Fprintln(builder, styleHeader.Render("Suggestions"))
		for _, note := range plan.Notes {
			fmt.Fprintf(builder, "- %s\n", note)
		}
	}
	return builder.String()
}

func RenderFixResults(plan fix.Plan, results []fix.Result) string {
	builder := &strings.Builder{}
	fmt.Fprintln(builder, styleHeader.Render("Fix Results"))
	for _, result := range results {
		status := "OK"
		if result.Error != nil {
			status = "FAILED"
		}
		fmt.Fprintf(builder, "- [%s] %s\n", status, result.Description)
		if result.Error != nil {
			fmt.Fprintf(builder, "  %s\n", result.Error)
		}
	}
	return builder.String()
}

func RenderFixEmpty(plan fix.Plan) string {
	builder := &strings.Builder{}
	fmt.Fprintln(builder, styleHeader.Render("No fixes available"))
	if len(plan.Notes) > 0 {
		fmt.Fprintln(builder, "")
		for _, note := range plan.Notes {
			fmt.Fprintf(builder, "- %s\n", note)
		}
	}
	return builder.String()
}

func Confirm(prompt string) (bool, error) {
	return confirmPrompt(prompt)
}
