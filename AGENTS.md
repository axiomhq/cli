# AGENTS.md

This file provides guidance to AI coding assistants (Claude Code, Cursor, GitHub Copilot, etc.) when working with code in this repository.

## Build & Development Commands

```bash
make dep          # Install and verify dependencies
make build        # Build binaries (uses goreleaser snapshot)
make test         # Run all tests with race detection
make lint         # Run golangci-lint
make fmt          # Format code with golangci-lint fmt
make install      # Install binary to $GOPATH/bin
make man          # Generate man pages
make completions  # Generate shell completion scripts
make all          # Run dep, generate, fmt, lint, test, build, man
```

**Run a single test:**
```bash
go test -race -run TestName ./path/to/package
```

**Verbose test output:**
```bash
make test VERBOSE=1
```

## Architecture

### Entry Point
`cmd/axiom/main.go` - Sets up signal handling, creates the Factory, configures survey prompts, and executes the root command.

### Factory Pattern
`internal/cmdutil/factory.go` - The `Factory` struct bundles shared resources (`Config`, `IO`) and provides methods like `Client()` to create authenticated Axiom API clients. It flows through the entire command hierarchy.

### Command Structure
Commands live in `internal/cmd/<command>/` with a consistent pattern:
- `<command>.go` - Main command with `NewCmd(f *cmdutil.Factory)` constructor
- `<command>_<subcommand>.go` - Subcommands

Commands are grouped:
- **Core:** `ingest`, `query`, `stream` - Data operations
- **Management:** `annotation`, `config`, `dataset` - Resource management
- **Additional:** `auth`, `completion`, `version`, `web`

### Configuration
`internal/config/config.go` - TOML-based config (`~/.axiom.toml`) with environment variable overrides via `AXIOM_` prefix (token, org_id, url, deployment). Supports multiple named deployments.

### Client
`internal/client/client.go` - Wraps `axiom-go` SDK with CLI-specific HTTP transport settings. Uses `github.com/axiomhq/axiom-go/axiom` for all API operations.

### Public Packages (`pkg/`)
- `terminal/` - IO abstraction, color schemes, table printing
- `iofmt/` - Output formatting (JSON, table)
- `surveyext/` - Survey prompt extensions and validation
- `doc/` - Documentation generation helpers

### Tools (`tools/`)
- `gen-cli-docs` - Generates man pages
- `loggen` - Test log generation utility

## Key Dependencies
- `github.com/axiomhq/axiom-go` - Official Axiom Go SDK
- `github.com/spf13/cobra` - CLI framework
- `github.com/AlecAivazis/survey/v2` - Interactive prompts

## Commit Message Format
```
internal/cmd/query: fix result formatting
pkg/terminal: add dark mode support
```

## Claude Code Plugin

This repository includes a Claude Code plugin for APL query assistance.

### Skills

| Skill | Command | Description |
|-------|---------|-------------|
| `explore-dataset` | `/explore-dataset <name>` | Discover schema, fields, and patterns |
| `find-traces` | `/find-traces <trace-id>` | Analyze OpenTelemetry traces |
| `detect-anomalies` | `/detect-anomalies <dataset>` | Statistical anomaly detection |

### Installation

```bash
# Add the marketplace from GitHub
claude plugin marketplace add axiomhq/cli

# Install the plugin
claude plugin install axiom-cli@axiom-cli

# Or for local development in this repo
claude plugin marketplace add ./
claude plugin install axiom-cli@axiom-cli --scope local
```

### Reference

The plugin includes comprehensive APL reference documentation in `skills/axiom-apl/references/`:
- `cli.md` - CLI flags and usage
- `operators.md` - APL operators
- `functions.md` - APL functions
- `patterns.md` - Query patterns by use case
- `gotchas.md` - Common mistakes
