# 🛡️ stack
### The Local-First Backend Health & Security Audit Tool

[![Go Version](https://img.shields.io/github/go-mod/go-version/neutron420/stack?style=flat-square)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)

**stack** is a high-performance, developer-first CLI tool designed to audit your backend project's health, security, and infrastructure readiness in seconds. It runs entirely locally, ensuring your secrets never leave your machine.

---

## ✨ Features

- **Premium TUI**: A beautiful, interactive terminal interface powered by Bubble Tea and Lipgloss.
- **Secrets Guard**: Advanced scanning for hardcoded API keys, tokens, and credentials.
- **Docker Health**: Deep analysis of Dockerfiles and Compose files for security best practices.
- **K8s Readiness**: Audit your Kubernetes manifests for resource limits and security contexts.
- **Extensible Plugins**: Add your own team-specific standards using simple YAML rule sets.
- **Lightning Fast**: Built in Go with high-concurrency module execution.

---

## Quick Start

### Install

```bash
# macOS/Linux
curl -sSL https://stack.io/install.sh | sh

# Windows (PowerShell)
iwr https://stack.io/install.ps1 | iex
```

### Usage

Run a full project health check:

```bash
stack scan
```

Run a specific diagnostic check:

```bash
stack doctor
```

---

## 🛠 Interactive Configuration

stack is ready to go out of the box, but you can customize it with a `.stack.yaml` file:

```yaml
rule_packs:
  - strict
output: table
min_severity: warning
plugins:
  - .stack/plugins/team.yaml
```

---

## Contributing

We love contributions! Whether it's a new scanner module or a bug fix, feel free to open a PR.

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

Distributed under the MIT License. See `LICENSE` for more information.

---

<p align="center">Built for developers who care about production health.</p>
