# Contributing to StackAudit

First off, thank you for considering contributing to StackAudit! It's people like you that make StackAudit such a great tool for the developer community.

We welcome all types of contributions:
- Reporting a bug
- Suggesting a new feature
- Improving documentation
- Submitting a pull request

## Pull Requests are Welcome!

We highly encourage you to submit Pull Requests. Whether it is a new scanner module, a bug fix, or a performance improvement, we want to see your work!

### Getting Started

1. **Fork the repository** on GitHub.
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/StackAudit.git
   cd StackAudit
   ```
3. **Install dependencies**:
   ```bash
   go mod download
   ```

### Development Workflow

1. **Create a branch** for your work:
   ```bash
   git checkout -b feature/amazing-new-scanner
   ```
2. **Make your changes**. Ensure your code follows the existing style and is well-commented.
3. **Run local tests**:
   ```bash
   go test ./...
   ```
4. **Build the binary** to test your changes:
   ```bash
   go build ./cmd/stack
   ```

### Submission Guidelines

1. **Commit your changes**:
   ```bash
   git commit -m "feat: added a new scanner for [Service Name]"
   ```
2. **Push to your fork**:
   ```bash
   git push origin feature/amazing-new-scanner
   ```
3. **Open a Pull Request** against the `main` branch of the original repository.
4. **Describe your changes** in detail in the PR description.

### Community Code of Conduct

By contributing to this project, you agree to abide by our code of conduct. Please be respectful and professional in all interactions.

---

Thank you for helping us make backend infrastructure safer for everyone!
