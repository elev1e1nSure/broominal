# Broominal — Agent Context

## What

Safe, reversible Windows cleanup CLI/TUI in Go. Files are moved to quarantine with JSON manifests, not deleted.

## Key Rules

1. **Quarantine by default** — never `os.Remove` on user files outside `cleaner.go`/`quarantine.go`
2. **Safe zones only** — scanner walks predefined paths from `%TEMP%`, `%LOCALAPPDATA%`, `%APPDATA%`, known app dirs
3. **No telemetry** — only outbound call is GitHub release check
4. **No silent actions** — every destructive action requires user confirmation
5. **Config local** — all data under `%LOCALAPPDATA%\broominal\`

## Architecture

```
cmd/broominal/    CLI (Cobra)
pkg/
  categories/     All categories registry → docs/categories.md
  scanner/        Walk functions for each category
  cleaner/        Move-to-quarantine pipeline
  quarantine/     Move, restore, cleanup with manifests
  config/         JSON config, presets (Quick/Standard/Deep)
  i18n/           EN/RU, auto-detect
  types/          Shared structs
internal/tui/     Bubbletea UI
```

## Adding a Category (4 files)

1. `pkg/categories/categories.go` — add `Def{Name, InternalKey, Risk, MinPreset}`
2. `pkg/scanner/scanner.go` — add `scanXxx(ctx, cfg) ([]Item, error)`
3. `pkg/scanner/scanner_registry.go` — wire `"Name": scanXxx` in `scanFuncs`
4. `pkg/i18n/i18n.go` — add `cat_*` + `cat_desc_*` for EN and RU

Missing registration → startup panic (safe by design).

## Docs

- `docs/commands.md` — CLI reference
- `docs/categories.md` — full category list with paths, risks, presets
- `docs/safety.md` — risk model, quarantine, security guards
- `docs/developing.md` — setup, justfile commands, CI

## Dev Commands

```
just build             # build dev binary
just build-release v1  # build with version
just run scan          # build + run
just test              # all tests
just check             # fmt → vet → lint → test
```

## Code Conventions

- `scanDir()` helper for simple directory walks
- Category items use `types.Item{Category, Path, Size, Risk}`
- Log with `slog` — warnings for skipped paths, no panics
- Atomic writes: `.tmp` → `os.Rename`
- `validateID()` before any quarantine file ops
- Skip symlinks, skip excluded paths from config
- `slog.Warn` for failed paths, return nil items on error

## Commits

* commit after every meaningful change
* format: conventional commits with scope
* `type(scope): description` — e.g. `feat(auth): add JWT`, `fix(api): fix timeout`, `chore(deps): update deps`
* types: feat, fix, chore, docs, style, refactor, perf, test, ci, build
* message in English, no trailing dot
* ask if scope is unclear