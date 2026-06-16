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
  update/         GitHub release check and self-update
  pathman/        System PATH manipulation
  taskscheduler/  Windows Task Scheduler integration
internal/tui/     Bubbletea UI
scripts/          ai-changelog.ps1 — OpenRouter beautifier
cliff.toml        git-cliff config for changelog generation
githooks/         commit-msg (conventional commits), pre-commit
justfile          Build command runner
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

Always use `just` for builds, tests, and checks — never invoke `go build`, `go test`, or `gofmt` directly.

```
just build             # build dev binary (go build with dev version)
just build-release v1  # build with explicit version tag
just run scan          # build + run with args (e.g. just run tui)
just test              # all tests (go test ./...)
just test-pkg          # package tests only, skips TUI
just fmt               # format code (gofmt -w .)
just vet               # go vet ./...
just lint              # golangci-lint run
just check             # full check: fmt → vet → lint → test-pkg
just changelog-raw     # preview raw commits since last tag
just changelog         # AI-polished changelog via OpenRouter
just changelog-save f  # save changelog to file
just clean             # remove broominal.exe
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
* types: feat, fix, chore, docs, style, refactor, perf, test, ci, build, revert
* message in English, no trailing dot
* ask if scope is unclear