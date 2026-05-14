# DevDoctor

DevDoctor is a local-first backend project health scanner for developers. It scans environment files, secrets, Docker, CI/CD, Kubernetes, Redis, and PostgreSQL configuration to surface production safety issues.

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
- Kubernetes manifest checks
- Redis configuration and compose checks
- PostgreSQL configuration and compose checks
- Health score engine with security/infrastructure/configuration breakdown
- Fix mode with confirmation and backups
- Ignore file support (.devdoctorignore)
- Exit codes by severity for CI gating
- SARIF output for security tooling
- Configurable rule severities and embedded/custom rule packs
- Baseline snapshots to suppress known findings
- Per-module timeout budgets
- YAML custom scanner plugins
- Git hook installation for pre-commit and pre-push

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
devdoctor kubernetes
devdoctor redis
devdoctor postgres
devdoctor doctor
devdoctor fix
devdoctor hooks install
devdoctor version
```

Common flags:

```bash
devdoctor scan --path .
devdoctor scan --rules ./configs/sample_rules.yaml
devdoctor scan --rule-pack strict
devdoctor scan --plugin ./configs/sample_plugin.yaml
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

Rules are defined in YAML. Use `severity` to tune any finding by rule ID. Built-in severities are `critical`, `warning`, `info`, and `success`.

```yaml
packs:
  - strict

rules:
  - env: JWT_SECRET
    required: true
    severity: critical

  - id: docker_latest_tag
    severity: info
  - docker:
      latest_tag: false

  - production:
      no_localhost: true
```

Use the rule file:

```bash
devdoctor scan --rules ./configs/sample_rules.yaml
```

Rule packs can be selected from the CLI or composed inside a rules file:

```bash
devdoctor scan --rule-pack relaxed
devdoctor scan --rule-pack strict --rule-pack ./team-rules.yaml
```

Embedded packs:

- `relaxed`: lowers noisy findings for local development
- `strict`: raises production-sensitive findings

### Baseline snapshots

Create a baseline to suppress known findings from future scans:

```bash
devdoctor scan --update-baseline --baseline .devdoctor.baseline.json
devdoctor scan --baseline .devdoctor.baseline.json
```

Baseline filtering happens before summaries, scores, and exit-code checks, so CI only fails on new findings.

### Module timeouts

Set a single budget for every module, or tune individual modules by name:

```bash
devdoctor scan --module-timeout 2s
devdoctor scan --module-timeout env=500ms,secrets=5s,docker=2s,ci=2s
```

When a module exceeds its budget, DevDoctor reports a `module_timeout` warning for that module and keeps the rest of the scan moving.

### Kubernetes, Redis, and PostgreSQL

Run the full scan or target a specific backend surface:

```bash
devdoctor kubernetes
devdoctor redis
devdoctor postgres
```

The Kubernetes scanner inspects YAML manifests with `apiVersion` and `kind`, including Deployments, Pods, StatefulSets, DaemonSets, Jobs, CronJobs, and Services. It flags risky settings such as `latest` images, missing resource limits, privileged containers, root-capable containers, host namespaces, and public LoadBalancer services.

The Redis scanner checks `redis.conf`, compose services, and env files for disabled protected mode, unauthenticated Redis, all-interface binding, exposed port 6379, disabled append-only persistence, and unauthenticated Redis URLs.

The PostgreSQL scanner checks `postgresql.conf`, `pg_hba.conf`, compose services, and env files for trust authentication, open CIDR rules, all-interface listening, SSL disabled, weak compose passwords, exposed port 5432, and connection URLs with `sslmode=disable`.

### Custom scanner plugins

Custom plugins are YAML files that define project-specific text or regex checks. Pass them explicitly with `--plugin`, or place them in `.devdoctor/plugins/*.yaml` to load them automatically during `devdoctor scan`.

```bash
devdoctor scan --plugin ./configs/sample_plugin.yaml
```

Example plugin:

```yaml
name: team
rules:
  - id: no_debug_env
    title: Debug mode enabled
    severity: warning
    category: custom
    path: ".env*"
    contains: "DEBUG=true"

  - id: no_local_admin_email
    title: Local admin email is committed
    severity: info
    paths:
      - "**/*.go"
      - "**/*.ts"
    regex: "admin@example\\.com"
```

### Git hooks

Install managed pre-commit and pre-push hook blocks:

```bash
devdoctor hooks install
devdoctor hooks install --hook pre-commit --command "devdoctor scan --exit-code --min-severity warning"
devdoctor hooks uninstall
```

Existing hook content is preserved. DevDoctor only adds or removes the block between its own hook markers.

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
    /kubernetes
    /redis
    /postgres
    /custom
    /githooks
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

The architecture supports additional scanners, rule packs, custom YAML plugins, CI integrations, and editor plugins without major refactoring.

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
