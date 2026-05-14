# DevDoctor

DevDoctor is a local-first backend project health scanner for developers. It scans environment files, secrets, Docker configuration, and CI/CD workflows to surface production safety issues.

## Status

The MVP is complete and production-ready for local scanning. Core modules, rules, output modes, and fix planning are in place, with clean architecture for future expansion.

## Why DevDoctor

- 100% local scans with no telemetry or external calls
- Fast startup and low memory usage
- Modular rule engine for extensibility
- Clean, readable diagnostics with a modern terminal UX

## Feature coverage (current)

- Environment variable checks (.env, .env.local, .env.production, .env.example)
- Secret leak detection with regex and entropy heuristics
- Dockerfile and docker-compose.yml analysis
- CI/CD workflow inspection (.github/workflows)
- Health score engine with security/infrastructure/configuration breakdown
- Fix mode with confirmation and backups
- Ignore file support (.devdoctorignore)
- Exit codes by severity for CI gating
- SARIF output for security tooling

## What is still left (nice-to-have roadmap)

These are optional enhancements. The current release already meets the product vision for local-first scanning.

- Configurable rule severities and custom rule packs
- Baseline snapshots to suppress known findings
- Per-module timeouts and budgets
- Kubernetes, Redis, PostgreSQL scanners
- Plugin system for custom scanners
- Git hooks integration (pre-commit, pre-push)

## Screenshots

Add your terminal screenshots here. Replace with real output captures once the CLI is running in your environment.

## Installation

### From source

```bash
git clone <your-repo>
cd devdoctor
go build ./cmd/devdoctor
```

### Go install

```bash
go install ./cmd/devdoctor
```

## Usage

```bash
devdoctor scan
devdoctor env
devdoctor docker
devdoctor ci
devdoctor secrets
devdoctor doctor
devdoctor fix
devdoctor version
```

Common flags:

```bash
devdoctor scan --path .
devdoctor scan --rules ./configs/sample_rules.yaml
devdoctor scan --output table
devdoctor scan --no-tui
```

### Output modes

```bash
devdoctor scan --output table
devdoctor scan --output json
devdoctor scan --output markdown
devdoctor scan --output sarif
```

### Ignore file

Create a .devdoctorignore file at the project root to skip paths. Patterns support * and ? wildcards.

Negation patterns (starting with !) are not supported yet.

```
node_modules/
dist/
*.generated.go
configs/local/*.yml
```

### Exit codes (CI gating)

Use --exit-code with --min-severity to return a non-zero exit code when findings meet or exceed the minimum.

```bash
devdoctor scan --exit-code --min-severity warning
```

Exit codes:

- 0: no findings at or above the minimum
- 1: info
- 2: warning
- 3: critical

### Rules engine

Rules are defined in YAML. Example:

```yaml
rules:
  - env: JWT_SECRET
    required: true

  - docker:
      latest_tag: false

  - production:
      no_localhost: true
```

Use the rule file:

```bash
devdoctor scan --rules ./configs/sample_rules.yaml
```

## Fix mode

`devdoctor fix` generates a plan and asks for confirmation before applying changes. It can:

- Generate .env.example from existing env files
- Remove unused env variables (with backups)
- Add missing restart policies in docker-compose.yml

## Architecture

```
/cmd
/internal
    /scanner
    /env
    /docker
    /cicd
    /secrets
    /rules
    /output
    /health
    /utils
    /fix
/pkg
/configs
```

## Security guarantees

- No AI
- No telemetry
- No cloud calls
- No user tracking
- Secrets never leave the machine

## Extensibility

The architecture supports future modules for Kubernetes, Redis, PostgreSQL, Git hooks, CI integrations, and editor plugins without major refactoring.

## Development

```bash
go test ./...
```

## Contributing

- Keep changes local-first and offline
- Follow idiomatic Go patterns
- Add unit tests for new rules and scanners

## Beautiful CLI guide

DevDoctor already uses Bubble Tea + Lip Gloss for a premium terminal UX. To extend or refine the UI:

- Keep all UI code inside internal/output to avoid leaking terminal concerns into scanners.
- Build a single style palette and reuse it across sections to keep visual consistency.
- Use stable widths and alignment to prevent flicker across rerenders.
- Use short, focused spinners for long operations and immediately render results afterward.
- Group output by severity and keep spacing consistent across sections.
- Avoid heavy animation; prefer crisp borders, clear typography, and clean spacing.
- Ensure output is readable with color disabled by keeping text meaningful without color.
