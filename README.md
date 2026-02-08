# geoffrussy

Geoffrussy is an AI-driven software delivery orchestrator for real projects. It guides work through a staged pipeline, persists state locally, supports many model providers, and can run through CLI or MCP for autonomous agents.

Pipeline:

`interview -> design -> plan -> review -> develop`

## What It Does

- Runs a structured requirements interview and stores answers in project state.
- Generates architecture artifacts and development phases/tasks.
- Reviews plan quality before execution.
- Executes tasks with live monitoring (pause/resume/skip, logs, blockers).
- Tracks token usage, cost, and provider quota/rate snapshots.
- Supports checkpoints/rollback, resume, and stage navigation.
- Exposes the workflow over MCP for agent clients.

## Current Provider Support

Built-in providers include:

- `openai`, `openai-codex`
- `anthropic`
- `zai`, `kimi`, `firmware`, `requesty`
- `openrouter`, `groq`, `together`, `deepinfra`, `fireworks`, `perplexity`, `mistral`
- `ollama` (local)
- `opencode` (CLI bridge)

Use:

```bash
geoffrussy config --list-providers
geoffrussy config --provider-help <provider>
```

## Install

```bash
git clone https://github.com/mojomast/geoffrussy.git
cd geoffrussy
make build
./bin/geoffrussy version
```

Go 1.24+ and a C toolchain for sqlite are required.

## Quick Start

From your project directory:

```bash
geoffrussy init
geoffrussy config --set-key
geoffrussy config --set-model

geoffrussy interview
geoffrussy design
geoffrussy plan
geoffrussy review
geoffrussy develop
```

Useful during execution:

```bash
geoffrussy status           # interactive TUI by default
geoffrussy status --tui=false
geoffrussy stats
geoffrussy quota --refresh
```

## State and Files

- Project state DB: `.geoffrussy/state.db` (project-local)
- Runtime logs: `.geoffrussy/logs/`
- Generated architecture JSON: `.geoffrussy/architecture.json`
- Config: `~/.geoffrussy/config.yaml`

API keys are stored in OS keyring when available, with secure fallback metadata tracked in config.

## Granular Model Assignment

Model defaults can be assigned per workflow sub-step, for example:

- `interview.run`, `interview.followup`, `interview.analysis`, `interview.defaults`
- `design.generate`, `design.refine`
- `devplan.generate`
- `review.phase`
- `develop.execute`, `develop.blocker_analyze`

Use interactive setup:

```bash
geoffrussy config --set-model
```

## MCP Server

Start:

```bash
geoffrussy mcp-server --project-path /absolute/path/to/project --debug
```

The server runs over stdio JSON-RPC 2.0 and exposes tools/resources for interview/design/plan/execute/status/checkpoints.

See `docs/mcp-integration.md` for full details.

## Commands

Core:

- `init`, `interview`, `design`, `plan`, `review`, `develop`
- `status`, `stats`, `quota`
- `checkpoint`, `rollback`, `resume`, `navigate`
- `config`, `mcp-server`, `version`

## Documentation Map

- `docs/README.md` - documentation index
- `docs/SETUP.md` - installation, provider auth, secure key storage
- `docs/COMMANDS.md` - command reference
- `docs/WORKFLOW.md` - stage-by-stage workflow and state transitions
- `docs/PROVIDERS.md` - provider capabilities and model selection behavior
- `docs/TUI.md` - monitor/status/review TUI controls
- `docs/mcp-integration.md` - MCP setup, tools, resources

Historical reports are in `docs/archive/reports/`.

## Development

```bash
make fmt
make lint
go test ./...
```

## License

MIT. See `LICENSE`.
