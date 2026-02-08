# TUI Guide

Geoffrussy uses Bubble Tea/Bubbles/Lip Gloss for terminal UX.

## Execution Monitor (`develop`)

Interactive monitor appears during `geoffrussy develop`.

Keys:

- `p` pause
- `r` resume
- `s` skip current task
- `f` toggle follow-output
- `?` toggle key help
- `q` quit
- `up/down`, `j/k`, `pgup/pgdn`, `home/end`, `g/G` for log navigation

The monitor shows:

- task/phase progress
- completion percent
- elapsed time
- token in/out counters
- live execution log stream

## Status TUI (`status`)

`geoffrussy status` opens TUI by default.

Use `geoffrussy status --tui=false` for plain text output.

Keys:

- `?` key help
- `q` quit
- scroll keys (same as above)

## Review TUI (`review`)

`geoffrussy review` opens a browsable review dashboard by default.

Use `geoffrussy review --tui=false` for plain text output.

Keys:

- `up/down`, `j/k` navigate
- `enter` drill in/back
- `esc` back/quit
- `q` quit

## Design Language

Monitor/status/review share:

- compact header + status row
- bordered content panes
- consistent color palette and controls
- responsive viewport behavior on terminal resize
