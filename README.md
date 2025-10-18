go mod download
# SecScan — Security Scanning Orchestrator

SecScan is a lightweight command-line orchestrator written in Go designed to run Snyk scans against source repositories. It automates cloning, local directory scans, and registering projects in Snyk (via `snyk monitor`). This tool is intended for security engineers and automation pipelines that need deterministic, repeatable Snyk reporting.

## Table of Contents
- Overview
- Threat model & assumptions
- Quick start
- Installation
- Configuration
- Usage
- CLI reference
- Troubleshooting
- Security considerations
- Development notes
- Contribution

## Overview
SecScan centralizes Snyk CLI usage for scanning multiple projects at scale. It provides three primary modes:

- `scanMonitor` — clone a remote repository and run Snyk test + monitor. Intended for CI or server-side automation.
- `scanLocal` — scan a directory containing multiple Go modules (projects that include a `go.mod`). Useful for local analysis and bulk scanning of already-checked-out source.
- `scanGitLab` — scan a GitLab project by ID (internal integrations available).

The tool uses Viper for configuration and `godotenv` to load environment variables from a `.env` file when present.

## Threat model & assumptions
- Trust boundary: this tool runs with the privileges of the user executing it and relies on local Snyk CLI behavior. It does not make privileged system changes.
- Secrets in scope: `SNYK_TOKEN`, `GITLAB_TOKEN` — treat them as secrets and avoid committing them to source control.
- Assumptions: target repositories are accessible to the configured credentials and the Snyk account has the required organization permissions for the `SNYK_ORG` used.

## Quick start
Create a `.env` file in the project root (or export variables in your shell) and run `scanLocal` against a local directory or `scanMonitor` for a remote repo.

Example (local):
```bash
# from project root
set -a; source .env; set +a
go run ./cmd/main.go scanLocal --dir /path/to/codebase --snyk-org="your-org-slug" --monitor
```

Example (remote):
```bash
export SNYK_TOKEN="<token>"
export SNYK_ORG="your-org-slug"
go run ./cmd/main.go scanMonitor --repo https://gitlab.example.com/team/repo.git --monitor
```

## Installation

Prerequisites
- Go 1.24+
- Snyk CLI installed and on PATH

Install dependencies (module-aware):

```bash
cd /path/to/SecScan
go mod download
```

Build a binary:

```bash
go build -o secscan ./cmd
```

## Configuration

Primary configuration is environment-driven. Create a `.env` file or export variables directly:

```ini
SNYK_TOKEN=your_snyk_api_token
SNYK_ORG=your_snyk_org_slug
SNYK_API_URL=https://api.snyk.io/v1  # optional override
GITLAB_TOKEN=glpat-xxxxxxxx         # for private GitLab access if needed
```

Viper will also look for a YAML configuration at `~/.config/repo-scanner/config.yaml` (optional). Viper uses `AutomaticEnv` so exported environment variables are respected.

## Usage

All commands are exposed via `cmd/main.go`.

- scanMonitor
	- Purpose: clone a remote repo and run Snyk test + monitor.
	- Note: `scanMonitor` reads `SNYK_ORG` from the environment; it does not accept `--snyk-org` as a dedicated flag (use `scanLocal` for that).

- scanLocal
	- Purpose: scan a local directory for Go projects (searches for `go.mod`) and run Snyk test + optional monitor per project.
	- Example:
		```bash
		go run ./cmd/main.go scanLocal --dir /path/to/repos --snyk-org="your-org" --monitor
		```

- scanGitLab
	- Purpose: scan a GitLab project by numeric ID; includes integration logic to fetch project metadata before cloning.

### Common flags
- `--snyk-org` — Snyk organization slug/ID (overrides `SNYK_ORG` env for commands that accept it, e.g. `scanLocal`).
- `--monitor` — after `snyk test`, register the project with `snyk monitor`.
- `--dir` — root directory for `scanLocal`.
- `--repo` — repository URL for `scanMonitor`.
- `-v/--verbose` — enable debug logging.

## CLI reference (short)

- `scanLocal` — flags: `--dir`, `--snyk-org`, `--monitor`, `--output-file`, `--verbose`.
- `scanMonitor` — flags: `--repo`, `--repo-file`, `--monitor`, `--verbose`.
- `scanGitLab` — flags: `--project-id`, `--snyk-org`, `--monitor`, etc.

## Troubleshooting

- "Org true was not found": This error typically means the Snyk CLI received `true` as the organization value. Common causes:
	- Passing `--snyk-org` without a value (Cobra/Viper can parse this as boolean `true`).
	- Argument parsing ambiguity when flags and positional arguments are passed incorrectly to the Snyk CLI.

	Mitigations implemented in SecScan:
	- Adapter now formats organization flag as `--org=<org>` to avoid ambiguity.
	- `--json` is passed as a standalone flag and the path is passed as a positional argument.

- Snyk token issues: ensure `SNYK_TOKEN` is exported and valid. `snyk test` will fail if the token is invalid or lacks org permissions.

- Missing `go.mod`: `scanLocal` only detects projects with `go.mod`. If none are found, verify repository layout.

## Security considerations

- Secrets handling: Do not commit `.env` or any token files. Limit file system permissions on any files containing secrets.
- Least privilege: prefer to create a Snyk API token scoped only to the org/projects required for scanning.
- Audit logging: run SecScan from an environment that records command invocations (CI logs, syslog) for accountability.

## Development notes

- The Snyk CLI adapter is implemented in `internal/adapters/snyk_cli_adapter.go`. It shells out to the `snyk` binary and returns combined output.
- Git clone logic is implemented in `internal/adapters/go_git_adapter.go` and supports basic HTTP(S) credential injection via env vars.

Temporary notes from debugging
- During troubleshooting, the adapter was temporarily updated to print debug commands to stderr to confirm argument composition. This was successful and can be removed for production use.

## Contribution

If you'd like to extend SecScan, open an issue or a pull request. Suggested next steps:

- Add `--snyk-org` to `scanMonitor` (if desired) to provide explicit override.
- Add unit/integration tests for the Snyk adapter to validate argument composition across platforms.

---
If you want, I can: remove the temporary debug prints and tidy up placeholder files in `cmd/` or open a PR with the changes.
