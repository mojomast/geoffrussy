# geoffrussy

![Build Status](https://img.shields.io/github/actions/workflow/status/mojomast/geoffrussy/ci.yml?branch=master&label=ci&logo=github)
![Go](https://img.shields.io/badge/go-1.24-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

Geoffrussy is a Go 1.24 AI-driven software delivery orchestrator that guides projects through a staged pipeline:

> interview → design → plan → review → develop

It uses SQLite (via `mattn/go-sqlite3`, requires CGO), Cobra for CLI, Bubbletea for TUI, and supports 16 AI providers through an OpenAI-compatible interface pattern.

## Quick highlights

- Structured interview flow that captures requirements and stores them in project state
- Architecture generation and incremental refinement, then dev plan generation (tasks + phases)
- Review step that validates plan quality before execution
- Execution engine with live monitoring, pause/resume, skip, and blocker handling
- Built-in quota, token and cost tracking, and provider rate-limit persistence to SQLite
- MCP server for agent integrations (JSON-RPC over stdio)

## Get started

### Prerequisites

- Go 1.24 (or newer)
- A C toolchain for SQLite (CGO) — required for `mattn/go-sqlite3`

### Install

```bash
git clone https://github.com/mojomast/geoffrussy.git
cd geoffrussy
make build
./bin/geoffrussy version
```

### Quick start (interactive)

```bash
geoffrussy init
geoffrussy config --set-key          # add provider keys
geoffrussy config --set-model        # select models/defaults

geoffrussy interview
geoffrussy design
geoffrussy plan
geoffrussy review
geoffrussy develop
```

### CI / non-interactive

```bash
geoffrussy init --non-interactive --api-key-openai "$OPENAI_KEY"
```

### Useful runtime commands

```bash
geoffrussy status           # interactive TUI by default
geoffrussy status --tui=false
geoffrussy stats
geoffrussy quota --refresh
```

## Providers

Geoffrussy ships adapters for many providers. Examples include: `openai`, `openai-codex`, `anthropic`, `ollama` (local), `opencode` (CLI bridge), `zai`, `kimi`, `firmware`, `requesty`, `openrouter`, `groq`, `together`, `deepinfra`, `fireworks`, `perplexity`, and `mistral`.

## State & files

- Project state DB: `.geoffrussy/state.db` (project-local)
- Runtime logs: `.geoffrussy/logs/`
- Generated architecture JSON: `.geoffrussy/architecture.json`
- User config: `~/.geoffrussy/config.yaml` (API keys stored in system keyring when available)

## Security notes

- Path sanitizer prevents directory traversal and rejects symlinks that resolve outside the project root
- Sensitive data is scrubbed from logs automatically (logger sanitizes common API key/token patterns)

See `docs/archive/reports/SECURITY_AUDIT.md` for a deeper audit.

## MCP server

Start the MCP server to expose the workflow for agents/automation:

```bash
geoffrussy mcp-server --project-path /absolute/path/to/project --debug
```

This server implements a stdio JSON-RPC 2.0 interface used by agent clients.

## Commands (core)

- `init`, `interview`, `design`, `plan`, `review`, `develop`
- `status`, `stats`, `quota`, `checkpoint`, `rollback`, `resume`, `navigate`
- `config`, `mcp-server`, `version`

## Documentation

- `docs/README.md` - documentation index
- `docs/SETUP.md` - installation and provider auth
- `docs/COMMANDS.md` - command reference
- `docs/WORKFLOW.md` - stage-by-stage workflow and state transitions
- `docs/PROVIDERS.md` - provider capabilities and model selection behavior
- `docs/TUI.md` - TUI controls and usage
- `docs/mcp-integration.md` - MCP setup, tools, resources

## Contributing

Contributions welcome — see `CONTRIBUTING.md` for guidelines and workflow.

## Development

```bash
make fmt
make lint
go test ./...
```

## License

MIT. See `LICENSE`.
