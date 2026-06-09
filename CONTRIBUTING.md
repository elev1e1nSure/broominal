# Contributing

Thanks for considering a contribution to Broominal.

Broominal is a Windows cleanup tool, so safety matters more than the number of supported cleanup targets. Changes that delete, move, or classify files must be conservative, testable, and easy to explain.

## What is useful

- Bug reports with Windows version, command used, and expected vs actual behavior
- New cleanup categories with clear paths and risk level
- Safer restore and quarantine behavior
- Tests for filesystem edge cases
- Documentation fixes
- CI, release, and packaging improvements

## Safety rules

- Prefer quarantine over permanent deletion.
- Never clean `danger` items automatically.
- Treat user files, Downloads, dumps, and app data as `review` unless they are clearly safe.
- Do not touch system directories such as `Windows`, `System32`, or `SysWOW64`.
- Do not clean protected extensions such as `.sys`, `.dll`, `.drv`, or `.ocx`.
- Add or update tests for new scanner categories and cleanup behavior.

## Development

```powershell
# Enable repository hooks
git config core.hooksPath githooks

# Run tests
go test ./...

# Build Windows executable
go build -o broominal.exe ./cmd/broominal

# Run TUI
.\broominal.exe ui
```

## Commit style

Use Conventional Commits:

```text
feat: add vscode cache scanner
fix: handle restore path conflicts
chore: update dependencies
docs: update usage examples
```

The repository includes a `commit-msg` hook that checks the commit type.

## Pull requests

Keep pull requests focused. A scanner category, a restore fix, and a TUI redesign should usually be separate PRs.

Before opening a PR, check:

- `go test ./...` passes
- changed code is formatted with `gofmt`
- user-facing behavior is documented
- new cleanup logic has a clear risk level
- destructive operations have dry-run or quarantine behavior
