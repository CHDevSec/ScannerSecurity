# SecScan - Security Scanning Orchestrator

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Snyk](https://img.shields.io/badge/Powered%20by-Snyk-4C4A73?style=flat&logo=snyk)](https://snyk.io/)

> A production-grade command-line orchestrator for automated security scanning at scale using Snyk CLI.

SecScan is an enterprise-ready security automation tool designed for security engineers, DevSecOps teams, and CI/CD pipelines. It provides deterministic, repeatable vulnerability scanning across multiple repositories and projects with centralized reporting through Snyk.

---

## 📋 Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Security Model](#security-model)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Scan Modes](#scan-modes)
  - [CLI Reference](#cli-reference)
  - [Examples](#examples)
- [Integration Patterns](#integration-patterns)
- [Troubleshooting](#troubleshooting)
- [Security Best Practices](#security-best-practices)
- [Performance Considerations](#performance-considerations)
- [Contributing](#contributing)
- [License](#license)

---

## 🔍 Overview

SecScan orchestrates Snyk CLI operations to provide automated security scanning capabilities for modern development workflows. Built with Go for performance and reliability, it abstracts the complexity of managing multiple Snyk scans across diverse project structures.

### What Problem Does It Solve?

- **Scale**: Scan hundreds of repositories without manual intervention
- **Consistency**: Standardized scanning approach across all projects
- **Automation**: Seamless CI/CD integration for continuous security monitoring
- **Visibility**: Centralized vulnerability reporting through Snyk dashboard
- **Compliance**: Audit trail and deterministic scanning for compliance requirements

---

## ✨ Key Features

- 🔄 **Multiple Scan Modes**: Local directories, remote repositories, GitLab integration
- 🎯 **Multi-Project Support**: Automatically discovers and scans Go modules (`go.mod`)
- 📊 **Snyk Monitor Integration**: Registers projects in Snyk for ongoing monitoring
- 🔐 **Secure Credential Management**: Environment-based configuration with `.env` support
- 🚀 **CI/CD Ready**: Designed for automation pipelines with exit codes and structured output
- 📝 **Comprehensive Logging**: Debug mode for troubleshooting and audit trails
- ⚙️ **Flexible Configuration**: Environment variables, YAML config, and CLI flags
- 🔧 **Git Integration**: Built-in repository cloning with credential injection

---

## 🏗️ Architecture

```
SecScan
├── cmd/
│   └── main.go              # CLI entry point & command definitions
├── internal/
│   ├── adapters/
│   │   ├── snyk_cli_adapter.go    # Snyk CLI wrapper
│   │   └── go_git_adapter.go      # Git operations handler
│   ├── domain/              # Business logic & models
│   └── ports/               # Interface definitions
├── config/
│   └── .env.example         # Configuration template
└── go.mod                   # Go module dependencies
```

### Component Responsibilities

- **CLI Layer**: Command parsing, flag handling, user interaction (Cobra)
- **Adapter Layer**: External tool integration (Snyk CLI, Git)
- **Domain Layer**: Core business logic, scan orchestration
- **Configuration**: Multi-source config management (Viper, godotenv)

---

## 🛡️ Security Model

### Threat Model & Assumptions

#### Trust Boundaries

- **Execution Context**: Runs with privileges of the invoking user
- **Network Trust**: Assumes secure network connection to Snyk API and Git servers
- **File System**: Requires read access to scan directories, write access to temp directories

#### Secrets in Scope

The following credentials are considered sensitive and must be protected:

| Secret | Purpose | Minimum Scope |
|--------|---------|---------------|
| `SNYK_TOKEN` | Snyk API authentication | Read-only org access |
| `GITLAB_TOKEN` | GitLab API/clone access | Read-only repository access |

#### Security Assumptions

1. Target repositories are accessible with provided credentials
2. Snyk organization has appropriate project creation permissions
3. Execution environment is trusted (not running on compromised systems)
4. Network communication with Snyk API and Git servers uses TLS
5. User executing SecScan has file system permissions for scan operations

### Attack Surface

- **Dependency Chain**: Go modules, Snyk CLI binary
- **Input Validation**: Repository URLs, file paths, configuration values
- **Secret Exposure**: Environment variables, temporary files, process memory
- **Command Injection**: Git URLs, Snyk CLI arguments

### Mitigations Implemented

✅ Credential isolation via environment variables  
✅ Argument sanitization for Snyk CLI calls  
✅ Explicit flag formatting (`--org=<value>`) to prevent parsing ambiguities  
✅ No privileged system operations required  
✅ Temporary directory cleanup after scans  

---

## 📦 Prerequisites

### Required

- **Go 1.24+** - [Installation Guide](https://golang.org/doc/install)
- **Snyk CLI** - Must be installed and available in `$PATH`
  ```bash
  # macOS
  brew install snyk/tap/snyk
  
  # Linux/Windows
  npm install -g snyk
  
  # Verify installation
  snyk --version
  ```

### Optional

- **Git** - For repository cloning (usually pre-installed)
- **GitLab Access** - For `scanGitLab` mode

### Snyk Account Requirements

- Active Snyk account with API access
- Organization created in Snyk dashboard
- API token with appropriate permissions:
  - Minimum: `org.read`, `test.read`
  - Recommended: Add `test.monitor` for project registration

---

## 🚀 Installation

### Method 1: Build from Source (Recommended)

```bash
# Clone the repository
git clone https://github.com/CHDevSec/ScannerSecurity.git
cd ScannerSecurity

# Download dependencies
go mod download

# Build the binary
go build -o secscan ./cmd

# Verify build
./secscan --help
```

### Method 2: Direct Go Install

```bash
go install github.com/CHDevSec/ScannerSecurity/cmd@latest
```

### Method 3: Pre-built Binaries

Download the latest release from the [Releases page](https://github.com/CHDevSec/ScannerSecurity/releases).

---

## ⚙️ Configuration

SecScan supports multiple configuration methods with the following precedence (highest to lowest):

1. Command-line flags
2. Environment variables
3. `.env` file in project root
4. YAML configuration file
5. Default values

### Environment Variables

Create a `.env` file in the project root:

```bash
# Required
SNYK_TOKEN=your_snyk_api_token_here
SNYK_ORG=your-organization-slug

# Optional
SNYK_API_URL=https://api.snyk.io/v1
GITLAB_TOKEN=glpat-xxxxxxxxxxxxxxxxxxxx
LOG_LEVEL=info
```

### YAML Configuration (Optional)

Create `~/.config/repo-scanner/config.yaml`:

```yaml
snyk:
  token: ${SNYK_TOKEN}  # Can reference env vars
  org: your-org-slug
  api_url: https://api.snyk.io/v1

gitlab:
  token: ${GITLAB_TOKEN}
  base_url: https://gitlab.example.com

logging:
  level: info
  format: json
```

### Secure Configuration Practices

⚠️ **NEVER commit secrets to version control**

```bash
# Add to .gitignore
echo ".env" >> .gitignore
echo "config.yaml" >> .gitignore
```

🔒 **Restrict file permissions**

```bash
chmod 600 .env
chmod 600 ~/.config/repo-scanner/config.yaml
```

🔑 **Use secret management systems in production**

- AWS Secrets Manager
- HashiCorp Vault
- Kubernetes Secrets
- GitHub Actions Secrets

---

## 🎯 Usage

### Scan Modes

SecScan provides three primary scanning modes optimized for different workflows:

#### 1. `scanLocal` - Local Directory Scanning

**Use Case**: Scan multiple projects in a local directory structure

**Features**:
- Discovers all `go.mod` files recursively
- Scans each project independently
- Supports bulk operations on checked-out repositories

**Syntax**:
```bash
secscan scanLocal [flags]
```

**Example**:
```bash
secscan scanLocal \
  --dir /workspace/projects \
  --snyk-org "my-organization" \
  --monitor \
  --verbose
```

---

#### 2. `scanMonitor` - Remote Repository Scanning

**Use Case**: Clone and scan remote repositories (CI/CD, automation)

**Features**:
- Automatic repository cloning
- Ephemeral scan (cleans up after completion)
- Batch processing via `--repo-file`

**Syntax**:
```bash
secscan scanMonitor [flags]
```

**Example (Single Repo)**:
```bash
export SNYK_ORG="my-organization"
secscan scanMonitor \
  --repo https://github.com/example/project.git \
  --monitor
```

**Example (Multiple Repos)**:
```bash
# repos.txt contains one repository URL per line
secscan scanMonitor \
  --repo-file repos.txt \
  --monitor \
  --verbose
```

---

#### 3. `scanGitLab` - GitLab Integration

**Use Case**: Scan GitLab projects using project ID

**Features**:
- GitLab API integration
- Automatic metadata retrieval
- Private repository support

**Syntax**:
```bash
secscan scanGitLab [flags]
```

**Example**:
```bash
export GITLAB_TOKEN="glpat-xxxxxxxxxxxxxxxxxxxx"
secscan scanGitLab \
  --project-id 12345 \
  --snyk-org "my-organization" \
  --monitor
```

---

### CLI Reference

#### Global Flags

| Flag | Shorthand | Type | Description |
|------|-----------|------|-------------|
| `--verbose` | `-v` | bool | Enable debug logging |
| `--help` | `-h` | bool | Display help information |

#### `scanLocal` Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--dir` | string | ✅ Yes | Root directory to scan |
| `--snyk-org` | string | ✅ Yes | Snyk organization slug |
| `--monitor` | bool | No | Register projects in Snyk (default: false) |
| `--output-file` | string | No | Write results to file |

#### `scanMonitor` Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--repo` | string | Conditional | Single repository URL to scan |
| `--repo-file` | string | Conditional | File containing repository URLs (one per line) |
| `--monitor` | bool | No | Register projects in Snyk (default: false) |

> **Note**: `--repo` or `--repo-file` must be provided (mutually exclusive)

#### `scanGitLab` Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--project-id` | int | ✅ Yes | GitLab project numeric ID |
| `--snyk-org` | string | ✅ Yes | Snyk organization slug |
| `--monitor` | bool | No | Register projects in Snyk (default: false) |

---

### Examples

#### Example 1: Local Development Scan

```bash
# Scan local workspace without monitoring
secscan scanLocal \
  --dir ~/workspace/microservices \
  --snyk-org "engineering-team"
```

#### Example 2: CI/CD Pipeline Integration

```bash
#!/bin/bash
# ci-security-scan.sh

set -euo pipefail

export SNYK_TOKEN="${SNYK_API_TOKEN}"
export SNYK_ORG="production-org"

secscan scanMonitor \
  --repo "${CI_REPOSITORY_URL}" \
  --monitor \
  --verbose

if [ $? -eq 0 ]; then
  echo "✓ Security scan passed"
else
  echo "✗ Security vulnerabilities detected"
  exit 1
fi
```

#### Example 3: Batch Repository Scanning

```bash
# Create repository list
cat > repos.txt <<EOF
https://github.com/org/service-auth.git
https://github.com/org/service-api.git
https://github.com/org/service-workers.git
EOF

# Scan all repositories
secscan scanMonitor \
  --repo-file repos.txt \
  --monitor \
  --output-file scan-results.json
```

#### Example 4: GitLab Private Project

```bash
export GITLAB_TOKEN="glpat-xxxxx"
export SNYK_TOKEN="xxxxx"

secscan scanGitLab \
  --project-id 789 \
  --snyk-org "platform-security" \
  --monitor \
  --verbose
```

---

## 🔗 Integration Patterns

### GitHub Actions

```yaml
name: Security Scan

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  security-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Install Snyk CLI
        run: npm install -g snyk
      
      - name: Install SecScan
        run: |
          git clone https://github.com/CHDevSec/ScannerSecurity.git
          cd ScannerSecurity
          go build -o secscan ./cmd
          sudo mv secscan /usr/local/bin/
      
      - name: Run Security Scan
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
          SNYK_ORG: ${{ secrets.SNYK_ORG }}
        run: |
          secscan scanMonitor \
            --repo ${{ github.repository }} \
            --monitor \
            --verbose
```

### GitLab CI/CD

```yaml
security-scan:
  stage: security
  image: golang:1.24
  before_script:
    - apt-get update && apt-get install -y npm
    - npm install -g snyk
    - git clone https://github.com/CHDevSec/ScannerSecurity.git
    - cd ScannerSecurity && go build -o secscan ./cmd
    - mv secscan /usr/local/bin/
  script:
    - secscan scanMonitor --repo ${CI_REPOSITORY_URL} --monitor
  variables:
    SNYK_TOKEN: ${SNYK_TOKEN}
    SNYK_ORG: ${SNYK_ORG}
  only:
    - main
    - develop
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any
    
    environment {
        SNYK_TOKEN = credentials('snyk-api-token')
        SNYK_ORG = 'your-org-slug'
    }
    
    stages {
        stage('Security Scan') {
            steps {
                script {
                    sh '''
                        secscan scanMonitor \
                            --repo ${GIT_URL} \
                            --monitor \
                            --verbose
                    '''
                }
            }
        }
    }
    
    post {
        always {
            archiveArtifacts artifacts: '**/scan-results.*', allowEmptyArchive: true
        }
    }
}
```

---

## 🔧 Troubleshooting

### Common Issues

#### 1. "Org true was not found"

**Symptoms**: Snyk CLI returns error about organization `true`

**Root Cause**: Argument parsing ambiguity when passing flags to Snyk CLI

**Solutions**:
- ✅ Ensure `--snyk-org` has a value: `--snyk-org="my-org"`
- ✅ Use the `=` syntax: `--snyk-org=my-org` (not `--snyk-org my-org`)
- ✅ Verify `SNYK_ORG` environment variable is set

**Mitigation Applied**: SecScan now formats Snyk flags as `--org=<value>` internally

---

#### 2. "Snyk command not found"

**Symptoms**: Error executing Snyk CLI

**Solutions**:
```bash
# Verify Snyk is installed
which snyk
snyk --version

# If not installed:
npm install -g snyk

# Verify PATH includes Snyk
echo $PATH | grep -q "$(dirname $(which snyk))" && echo "✓ Snyk in PATH"
```

---

#### 3. Authentication Failures

**Symptoms**: "Unauthorized" or "Invalid token" errors

**Solutions**:
```bash
# Verify token is set
echo $SNYK_TOKEN | cut -c1-10  # Should show first 10 chars

# Test Snyk authentication
snyk auth $SNYK_TOKEN

# Verify organization access
snyk test --org=$SNYK_ORG
```

---

#### 4. No Projects Found (scanLocal)

**Symptoms**: "No go.mod files found" message

**Solutions**:
- Verify directory path is correct: `ls -la /path/to/scan`
- Ensure projects contain `go.mod` files: `find /path -name "go.mod"`
- Check directory permissions: `ls -ld /path/to/scan`

---

#### 5. Git Clone Failures

**Symptoms**: Authentication or network errors during repository cloning

**Solutions**:
```bash
# For private repositories, ensure credentials are configured
git config --global credential.helper store

# For GitLab, verify token has correct scopes
# Required scopes: read_repository, read_api

# Test manual clone
git clone https://gitlab.example.com/project.git
```

---

### Debug Mode

Enable verbose logging for detailed troubleshooting:

```bash
secscan scanLocal --dir /path --snyk-org="org" --verbose 2>&1 | tee debug.log
```

Debug output includes:
- Command arguments passed to Snyk CLI
- Git operations and repository URLs
- Environment variable resolution
- File system operations

---

## 🔐 Security Best Practices

### Credential Management

#### Development Environment

```bash
# Use direnv for automatic environment loading
cat > .envrc <<EOF
export SNYK_TOKEN=$(security find-generic-password -s snyk-token -w)
export SNYK_ORG="dev-organization"
EOF

direnv allow
```

#### Production Environment

```bash
# AWS Secrets Manager
aws secretsmanager get-secret-value \
  --secret-id prod/snyk-token \
  --query SecretString \
  --output text

# Kubernetes Secret
kubectl create secret generic snyk-credentials \
  --from-literal=token=$SNYK_TOKEN \
  --from-literal=org=$SNYK_ORG
```

### Access Control

🔒 **Principle of Least Privilege**

Create dedicated Snyk service accounts with minimal permissions:

```
Snyk Organization → Settings → Service Accounts
├── Role: "Viewer" (for read-only scanning)
└── Scope: Limit to specific projects if possible
```

### Audit Logging

Enable comprehensive logging for compliance:

```bash
# Log all scans with timestamp and user
secscan scanLocal \
  --dir /workspace \
  --snyk-org "audit-org" \
  --verbose 2>&1 | \
  ts '[%Y-%m-%d %H:%M:%S]' | \
  tee -a /var/log/security-scans/$(date +%Y%m%d).log
```

### Network Security

- **TLS Verification**: Ensure Snyk API communications use TLS 1.2+
- **Proxy Support**: Configure `HTTPS_PROXY` if required
- **Firewall Rules**: Whitelist `api.snyk.io` and `api.github.com`

### Vulnerability Response

Establish a clear workflow:

1. **Detection**: SecScan identifies vulnerabilities via Snyk
2. **Triage**: Review findings in Snyk dashboard
3. **Remediation**: Apply patches or updates
4. **Verification**: Re-scan to confirm fixes
5. **Documentation**: Record response in ticketing system

---

## ⚡ Performance Considerations

### Optimization Tips

#### Parallel Scanning

SecScan processes projects sequentially by default. For large-scale operations:

```bash
# Scan repositories in parallel using GNU parallel
cat repos.txt | parallel -j 4 \
  'secscan scanMonitor --repo {} --monitor'
```

#### Resource Limits

```bash
# Limit memory and CPU usage
ulimit -v 4194304  # 4GB memory limit
nice -n 10 secscan scanLocal --dir /large-workspace
```

#### Caching Strategy

- Snyk CLI maintains a local cache in `~/.snyk`
- Keep cache warm for faster subsequent scans
- Consider cache cleanup policies: `snyk config clear`

### Scaling Guidelines

| Projects | Recommended Approach | Estimated Time |
|----------|---------------------|----------------|
| 1-10 | Single SecScan instance | 5-15 minutes |
| 10-100 | Parallel execution (4-8 workers) | 15-45 minutes |
| 100+ | Distributed system with queue | 1-3 hours |

---

## 🤝 Contributing

We welcome contributions from the security and development community!

### Development Setup

```bash
# Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/ScannerSecurity.git
cd ScannerSecurity

# Install development dependencies
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run tests
go test ./...

# Run linters
golangci-lint run
```

### Code Quality Standards

- **Test Coverage**: Maintain >80% test coverage
- **Linting**: Code must pass `golangci-lint` with no warnings
- **Documentation**: Public functions require godoc comments
- **Security**: No hardcoded secrets, validate all inputs

### Roadmap

Planned features and improvements:

- [ ] Add `--snyk-org` flag to `scanMonitor` command
- [ ] Implement comprehensive unit tests for adapters
- [ ] Add support for additional languages (Python, JavaScript, Java)
- [ ] Parallel scanning capabilities (built-in)
- [ ] Results export in multiple formats (JSON, SARIF, CSV)
- [ ] Integration with SIEM systems
- [ ] Web dashboard for scan history
- [ ] Slack/Teams notifications on critical findings

### Submitting Changes

1. Create a feature branch: `git checkout -b feature/my-enhancement`
2. Write tests for new functionality
3. Ensure all tests pass: `go test ./...`
4. Commit with descriptive messages: `git commit -m "feat: add parallel scanning"`
5. Push and create a Pull Request

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/CHDevSec/ScannerSecurity/issues)
- **Discussions**: [GitHub Discussions](https://github.com/CHDevSec/ScannerSecurity/discussions)
- **Security**: Report vulnerabilities via email to security@example.com

---

## 🙏 Acknowledgments

- [Snyk](https://snyk.io/) for providing the security scanning engine
- [Cobra](https://github.com/spf13/cobra) for CLI framework
- [Viper](https://github.com/spf13/viper) for configuration management
- All contributors and security researchers

---

**Built with ❤️ by the Security Team**

*Last Updated: 2025-01-19*
