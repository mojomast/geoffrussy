# Commands Reference

## Core Pipeline

- `geoffrussy init`
- `geoffrussy interview [--resume] [--model <model>]`
- `geoffrussy design [--model <model>] [--refine <guidance>]`
- `geoffrussy plan [--model <model>]`
- `geoffrussy review [--model <model>] [--apply] [--tui=<bool>]`
- `geoffrussy develop [--model <model>] [--phase <phase-id>] [--stop-after-phase]`

## Visibility and Control

- `geoffrussy status [--tui=<bool>] [--verbose]`
- `geoffrussy stats`
- `geoffrussy quota [--refresh]`
- `geoffrussy checkpoint`
- `geoffrussy rollback`
- `geoffrussy resume`
- `geoffrussy navigate --list`

## Provider and Model Config

- `geoffrussy config`
- `geoffrussy config --list-providers`
- `geoffrussy config --set-key`
- `geoffrussy config --set-model`
- `geoffrussy config --provider-help <provider>`

## MCP

- `geoffrussy mcp-server --project-path <path> [--debug]`

## Other

- `geoffrussy version`
