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

## Adding a Category (4 steps / files)

1. `pkg/categories/categories.go` — add `Def{Name, InternalKey, Risk, MinPreset}`
2. `pkg/scanner/scanner_xxx.go` — implement `scanXxx(ctx, cfg) ([]Item, error)` in the appropriate scanner file (e.g. scanner_windows.go, scanner_browsers.go, etc.)
3. `pkg/scanner/scanner_registry.go` — wire `"Category Name": scanXxx` in `scanFuncs`
4. `pkg/i18n/strings_en.go` & `pkg/i18n/strings_ru.go` — add translations for `cat_*` and `cat_desc_*` for EN and RU

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

## TUI ↔ CLI Parity

Every user-facing logical change in the TUI **must** be reflected in the CLI, and vice versa. This includes:
- New or renamed categories → update both TUI display and CLI `scan`/`clean` output
- New risk level or preset → update both `internal/tui/` and `cmd/broominal/`
- New quarantine actions (delete, clean) → both the restore screen and `quarantine` subcommands
- New path management features → both `internal/tui/screen_config.go` and `cmd/broominal/path.go`

Shared helpers live in `cmd/broominal/util.go` (CLI-only) and `internal/tui/model.go` (TUI-only).
Shared logic that applies to both belongs in the appropriate `pkg/` package, not duplicated.

## TUI Styling & Visual Rules

- **Padding/Margin consistency**: The main application frame (`appFrameStyle`) enforces a strict inner padding of `Padding(0, 2)`.
- **No manual indents**: Do not manually prepend spaces (e.g., `"  " + text`) to raw strings in `screen_*.go` for indentation. Rely entirely on Lipgloss styles and the inherited container padding.
- **Width calculations**: Since the main frame applies `Padding(0, 2)`, the available inner content width for dynamically sized elements (like progress bars or separators) is exactly `m.width - 6` (2 for left/right borders, 4 for left/right padding).
- **Alignment**: Standard text, titles, and list pointers (`> `) should align flush left (column 0 relative to the content block). List items without a pointer start with a `prefix` of 2 spaces (`"  "`) so their text aligns seamlessly.

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

## Commenting

- **English only** — every comment, doc comment, and inline note in the codebase must be in English. Russian-language keys in `pkg/i18n/strings_ru.go` are translation values, not comments, and stay as-is.
- **Explain why, not what** — a comment should describe reasoning, trade-offs, or non-obvious consequences. Skip comments that just restate the next line of code.
- **Godoc on exported identifiers** — Go convention: doc comments on exported types and funcs start with the identifier name and are declarative. Internal comments are free-form but should still be "why" when present.
- **Security-sensitive code must carry a "why"** — every branch that touches path traversal, symlink following, allow-lists, or UAC must explain the threat or invariant it relies on. The reader should not have to reverse-engineer the rule.
- **No stale comments** — when changing code, update or delete the surrounding comments. A wrong comment is worse than no comment.
- **Section dividers are fine** — `// ── Header ──` or `// Categories — names` style markers help navigate long files. Treat them as structure, not as explanations.
- **No AI-style filler** — no "this function does X" for trivial one-liners, no restating the function signature in prose.