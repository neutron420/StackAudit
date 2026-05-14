# StackAudit
### The Local-First Backend Health & Security Audit Tool

[![Go Version](https://img.shields.io/github/go-mod/go-version/neutron420/StackAudit?style=flat-square)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)

StackAudit is a high-performance, developer-first CLI tool designed to audit your backend project's health, security, and infrastructure readiness in seconds. It runs entirely locally, ensuring your secrets never leave your machine.

---

## Features

- **Professional TUI**: A clean, interactive terminal interface for real-time system monitoring and project auditing.
- **Secrets Detection**: Advanced scanning for hardcoded API keys, tokens, and credentials.
- **Docker Security**: Deep analysis of Dockerfiles and Compose files for security best practices.
- **Kubernetes Readiness**: Audit your Kubernetes manifests for resource limits and security contexts.
- **Extensible Plugins**: Add team-specific standards using simple YAML rule sets.
- **High Performance**: Built in Go with high-concurrency module execution.

---

## Installation

### Using GitHub (Recommended)
Download the latest binary for your operating system from the [Releases](https://github.com/neutron420/StackAudit/releases) page.

## Quick Start (One-Command Installation)

Get up and running in seconds with our automated installers. These scripts will download the latest version, install it, and configure your system path automatically.

### Windows (PowerShell)
```powershell
iwr https://raw.githubusercontent.com/neutron420/StackAudit/main/scripts/install.ps1 | iex
```

### macOS / Linux (Bash)
```bash
curl -sSL https://raw.githubusercontent.com/neutron420/StackAudit/main/scripts/install.sh | sh
```

### From Source
```bash
go install github.com/neutron420/stack/cmd/stack@latest
```

### Via NPM
```bash
npm install -g @riteshkumar04/stack-audit
```

## Supported Platforms

| OS | Architecture | Status |
|----|--------------|--------|
| **Windows** | x86_64, arm64, 386 | Fully Supported |
| **macOS** | Apple Silicon (arm64), Intel (x86_64) | Fully Supported |
| **Linux** | All Distros (x86_64, arm64, 386) | Fully Supported |

## Usage

Run the **Interactive Workbench**:
```bash
stack
```

Run a specific scan:
```bash
stack scan redis
```

---

## Configuration

StackAudit works out of the box, but can be customized with a `.stack.yaml` file:

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

We welcome contributions. Whether it is a new scanner module or a bug fix, feel free to open a Pull Request.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

Distributed under the MIT License. See `LICENSE` for more information.

---

<p align="center">Built for developers who care about production health.</p>
