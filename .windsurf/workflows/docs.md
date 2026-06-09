---
description: auto-sync project docs after significant changes
---

# Auto-Docs Workflow

Run this after significant code changes (new commands, packages, TUI screens, config fields, or scan categories) to keep README.md, README.ru.md, and AGENTS.md in sync with the codebase.

## Steps

1. **Detect changes**
   - Run `git diff HEAD~1 --name-only` (or `git diff --cached --name-only` if staging).
   - Also list any uncommitted files that differ from HEAD.

2. **Map files to doc sections**

   | Changed file pattern | Target doc updates |
   |----------------------|-------------------|
   | `cmd/broominal/main.go` | README Usage, Features (new commands/flags) |
   | `pkg/scanner/*.go` | README Features (scan categories), AGENTS.md scanner list |
   | `pkg/quarantine/*.go` | README Features (restore/quarantine), AGENTS.md quarantine notes |
   | `pkg/config/*.go` | README Features (config-driven), AGENTS.md config thresholds |
   | `pkg/doctor/*.go` | README Features (doctor), AGENTS.md doctor checks |
   | `pkg/i18n/*.go` | README Features (i18n), AGENTS.md i18n notes |
   | `internal/tui/*.go` | README Features (TUI screens), AGENTS.md TUI flow/screens |
   | `pkg/risk/*.go` | AGENTS.md risk rules |
   | `pkg/style/*.go` | AGENTS.md CLI style guide |
   | `pkg/types/*.go` | AGENTS.md shared structs |
   | `go.mod` / `go.sum` | Tech stack versions |

3. **Update README.md**
   - **Features table**: ensure every major capability has a row. Match the description to the current code.
   - **Usage**: add/remove commands and flags to match `main.go`. Keep the Powershell code block.
   - **Architecture**: ensure the tree matches `pkg/*` and `internal/*` directories.
   - **Build / Tech stack**: update Go version if `go.mod` changed.

4. **Update AGENTS.md**
   - **Architecture block**: sync with `pkg/*` and `internal/*`. Update per-package descriptions.
   - **Key design decisions**: if quarantine/dry-run/config/TUI/i18n logic changed, update the relevant numbered point.
   - **Extension points**: if new screens, categories, or config fields were added, append/update bullets.
   - **CLI style guide**: if `pkg/style` changed, update the style table.
   - **Tech stack**: sync with `go.mod`.

5. **Update README.ru.md**
   - Mirror all changes from README.md. Preserve Russian translations.
   - If a new feature/usage example has no Russian translation yet, translate it.

6. **Verify**
   - Read back the edited files.
   - Ensure no placeholder text, broken markdown, or stale references remain.
   - Ensure README.ru.md and README.md stay structurally identical.

## Rules

- Do not invent features that do not exist in the code.
- Do not remove rows unless the feature was actually deleted.
- Keep descriptions concise; match the tone of the existing doc.
- When in doubt, read the source file to confirm behavior.
