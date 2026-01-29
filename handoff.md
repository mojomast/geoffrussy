Handoff — Next Steps

I created this handoff with focused next steps to finish the CLI surface and move toward a release. Tests currently pass (`go test ./...`) but several CLI commands are stubs and need wiring to the implemented core services.

- Short status
  - All Go unit tests pass locally (cached results); the core subsystems (state store, providers, devplan, executor, token/cost tracking) are largely implemented.
  - Several CLI commands in `internal/cli` are placeholders that print TODOs and return errors instead of calling core packages.

- High priority (make the CLI usable)
  1. Implement minimal non-interactive wiring for CLI commands so they call existing services and return meaningful output (graceful fallbacks if provider credentials or TUI not present).
     - Files to update: `internal/cli/interview.go`, `internal/cli/design.go`, `internal/cli/plan.go`, `internal/cli/review.go`, `internal/cli/develop.go`, `internal/cli/checkpoint.go`, `internal/cli/stats.go`, `internal/cli/rollback.go`.
     - Acceptance: each command should return 0 with helpful message or perform basic work using existing packages; `--help` must show flags.
     - Verify: run `go build ./cmd/geoffrussy` then `./geoffrussy <command> --help` and `./geoffrussy review --model=gpt` (or similar) to see non-error output.

- Medium priority (behavioral and integration polish)
  1. Wire provider/model selection and session creation so CLI commands choose the right provider and model mapping.
     - Files to inspect: `internal/provider/*`, `internal/config/config.go`, `internal/cli/*`.
  2. Implement `applyImprovements` in `internal/cli/review.go` to use `reviewer.Reviewer` and persist changes to `state.Store`.
  3. Implement checkpoint creation/listing/rollback using the existing state and git manager (`internal/state`, `internal/git/manager.go`).

- Lower priority (UX, tests, and hardening)
  1. Integrate Bubbletea TUI flows for interactive commands (interview, develop, live monitor) once non-interactive flows work.
  2. Add unit/property tests referenced in `.kiro/specs/geoffrey-ai-agent/tasks.md` for CLI validation, checkpoints, and multi-provider behavior.
  3. Run and fix linter issues (`go vet`, `golangci-lint`) and add CI steps.

- Immediate actionable checklist (what I recommend doing first)
  1. Create a short-lived branch: `git checkout -b feat/cli-wiring`.
  2. Implement minimal wiring for `review` and `checkpoint` (two highest-impact commands).
  3. Run tests and build:

     ```bash
     go test ./...
     go build ./cmd/geoffrussy
     ./geoffrussy review --help
     ./geoffrussy checkpoint --help
     ```

  4. Commit with message: `feat(cli): wire review and checkpoint commands to core services` and push; open a PR with summary and verification steps.

- Suggested PR checklist
  - Tests still pass (`go test ./...`).
  - `go build` succeeds for `cmd/geoffrussy`.
  - Each wired CLI command has a unit-level smoke test or example invocation in the PR description.
  - Document any breaking flags or behavior in `README.md`.

- Contacts & references
  - Tasks spec: `.kiro/specs/geoffrey-ai-agent/tasks.md`
  - CLI stubs (edit these): `internal/cli/interview.go`, `internal/cli/design.go`, `internal/cli/plan.go`, `internal/cli/review.go`, `internal/cli/develop.go`, `internal/cli/checkpoint.go`, `internal/cli/stats.go`.

If you want I can start with the recommended branch and implement the minimal wiring for `review` + `checkpoint` (option 1). Otherwise tell me which commands you'd like me to wire first.
