package output

import (
	"bytes"
	"html/template"
	"time"

	"stackaudit/internal/scanner"
)

const reportHTMLTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>StackAudit Report</title>
  <style>
    body { margin: 0; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #f7f8fa; color: #20242a; }
    main { max-width: 1120px; margin: 0 auto; padding: 32px 20px 48px; }
    header { border-bottom: 1px solid #d9dee7; padding-bottom: 20px; margin-bottom: 24px; }
    h1 { margin: 0 0 8px; font-size: 32px; }
    h2 { margin: 28px 0 12px; font-size: 20px; }
    .muted { color: #697386; }
    .scores, .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 12px; }
    .metric, .finding { background: white; border: 1px solid #d9dee7; border-radius: 8px; padding: 14px; }
    .metric strong { display: block; font-size: 28px; margin-top: 4px; }
    .findings { display: grid; gap: 12px; }
    .finding h3 { margin: 0 0 8px; font-size: 16px; }
    .badge { display: inline-block; border-radius: 999px; padding: 3px 9px; font-size: 12px; font-weight: 700; text-transform: uppercase; }
    .critical { background: #ffe8e8; color: #9b1c1c; }
    .warning { background: #fff4d6; color: #7a4b00; }
    .info { background: #e8f1ff; color: #174ea6; }
    .success { background: #e7f7ed; color: #17643a; }
    code { background: #eef1f5; border-radius: 4px; padding: 2px 5px; }
    .fix { border-left: 3px solid #2563eb; padding-left: 10px; margin-top: 8px; }
  </style>
</head>
<body>
<main>
  <header>
    <h1>StackAudit Report</h1>
    <div class="muted">Local-only scan. No telemetry. Completed in {{ .Duration }}.</div>
  </header>

  <h2>Scores</h2>
  <section class="scores">
    <div class="metric">Overall<strong>{{ .Report.Scores.Overall }}/100</strong></div>
    <div class="metric">Security<strong>{{ .Report.Scores.Security }}/100</strong></div>
    <div class="metric">Infrastructure<strong>{{ .Report.Scores.Infrastructure }}/100</strong></div>
    <div class="metric">Configuration<strong>{{ .Report.Scores.Configuration }}/100</strong></div>
  </section>

  <h2>Summary</h2>
  <section class="summary">
    <div class="metric">Critical<strong>{{ .Report.Summary.Critical }}</strong></div>
    <div class="metric">Warning<strong>{{ .Report.Summary.Warning }}</strong></div>
    <div class="metric">Info<strong>{{ .Report.Summary.Info }}</strong></div>
    <div class="metric">Success<strong>{{ .Report.Summary.Success }}</strong></div>
  </section>

  <h2>Findings</h2>
  <section class="findings">
    {{ if .Report.Findings }}
      {{ range .Report.Findings }}
      <article class="finding">
        <div><span class="badge {{ .Severity }}">{{ .Severity }}</span></div>
        <h3>{{ .Title }}</h3>
        {{ if .Description }}<p>{{ .Description }}</p>{{ end }}
        {{ if .Remediation }}<p class="fix"><strong>Fix:</strong> {{ .Remediation }}</p>{{ end }}
        {{ if .File }}<p class="muted"><code>{{ .File }}{{ if .Line }}:{{ .Line }}{{ end }}</code></p>{{ end }}
        {{ if .RuleID }}<p class="muted">Rule: <code>{{ .RuleID }}</code></p>{{ end }}
      </article>
      {{ end }}
    {{ else }}
      <article class="finding">No issues found.</article>
    {{ end }}
  </section>
</main>
</body>
</html>`

func renderHTML(report scanner.Report) (string, error) {
	tmpl, err := template.New("report").Parse(reportHTMLTemplate)
	if err != nil {
		return "", err
	}
	data := struct {
		Report   scanner.Report
		Duration string
	}{
		Report:   report,
		Duration: report.Meta.Duration.Round(time.Millisecond).String(),
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
