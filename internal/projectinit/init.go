package projectinit

import (
	"fmt"
	"os"
	"path/filepath"
)

type Options struct {
	Force bool
}

func WriteStarter(root string, opts Options) ([]string, error) {
	files := map[string]string{
		".stack.yaml":  starterConfig,
		".stackignore": starterIgnore,
		filepath.Join(".stack", "plugins", "team.yaml"): starterPlugin,
	}
	return writeFiles(root, files, opts)
}

func WriteGitHubActions(root string, opts Options) ([]string, error) {
	return writeFiles(root, map[string]string{
		filepath.Join(".github", "workflows", "stack.yml"): githubActionsWorkflow,
	}, opts)
}

func WriteEmptyBaseline(root string, opts Options) ([]string, error) {
	return writeFiles(root, map[string]string{
		".stack.baseline.json": "{\n  \"version\": 1,\n  \"entries\": [],\n  \"root_hint\": \".\"\n}\n",
	}, opts)
}

func WriteDemo(root string, opts Options) ([]string, error) {
	return writeFiles(root, map[string]string{
		"Dockerfile": `FROM node:latest
# INSECURE: Running as root
COPY . .
ENV API_KEY="sk_test_12345"
CMD ["npm", "start"]`,
		".env": `DB_PASSWORD=password123
DEBUG=true
STRIPE_KEY=pk_live_ABCDEF`,
		"docker-compose.yml": `version: '3'
services:
  app:
    build: .
    ports:
      - "3000:3000"`,
	}, opts)
}

func writeFiles(root string, files map[string]string, opts Options) ([]string, error) {
	written := []string{}
	for rel, content := range files {
		path := filepath.Join(root, rel)
		if !opts.Force {
			if _, err := os.Stat(path); err == nil {
				continue
			} else if err != nil && !os.IsNotExist(err) {
				return nil, err
			}
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return nil, fmt.Errorf("write %s: %w", path, err)
		}
		written = append(written, path)
	}
	return written, nil
}

const starterConfig = `root: "."
rule_packs:
  - strict
output: table
exit_code: true
min_severity: warning
baseline: .stack.baseline.json
module_timeouts:
  - env=1s
  - secrets=5s
  - docker=2s
  - ci=2s
  - k8s=3s
  - redis=2s
  - postgres=2s
plugins:
  - .stack/plugins/team.yaml
`

const starterIgnore = `node_modules/
dist/
build/
.next/
vendor/
*.generated.*
`

const starterPlugin = `name: team
rules:
  - id: no_debug_env
    title: Debug mode enabled
    description: Disable DEBUG in committed environment files.
    severity: warning
    category: custom
    path: ".env*"
    contains: "DEBUG=true"
`

const githubActionsWorkflow = `name: stack

on:
  pull_request:
  push:
    branches:
      - main

permissions:
  contents: read
  security-events: write

jobs:
  stack:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - name: Install stack
        run: go install ./cmd/stack
      - name: Run stack
        run: stack scan --output sarif --exit-code --min-severity warning > stack.sarif
      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: stack.sarif
`
