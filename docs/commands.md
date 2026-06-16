# CLI Commands

## scan

Scan the system for cleanup candidates across all enabled categories.

```powershell
broominal scan                          # scan with default preset (quick)
broominal scan --preset standard        # scan with standard preset
broominal scan --preset deep            # scan with deep preset
```

### Flags
* `--preset <quick|standard|deep>`: Temporarily override the scan preset.
* `--json-logs`: Output structured JSON logs to stderr.

Results are printed as a table on the terminal. No files are modified.

---

## clean

Execute cleanup by moving all candidate files to the quarantine directory (or deleting them permanently if quarantine is disabled in config).

```powershell
broominal clean                         # dry-run preview with default preset
broominal clean --yes                   # execute cleanup with default preset
broominal clean --preset deep           # dry-run preview deep preset
broominal clean --preset deep --yes     # execute deep cleanup
```

Without `--yes`, the command prints a summary table of files to clean but does not modify anything.

### Flags
* `--yes`: Execute the cleanup (default is dry-run preview).
* `--preset <quick|standard|deep>`: Clean categories enabled by this preset.

---

## restore

Restore files from a quarantine backup to their original location on disk.

```powershell
broominal restore <id>                  # restore by exact backup ID
broominal restore last                  # restore the most recent backup
broominal restore <prefix>              # restore by unique ID prefix (e.g. "2026-06-16")
broominal restore <id> --force-overwrite # overwrite existing files at destination
```

### Flags
* `--force-overwrite`: Overwrite files that already exist at the destination rather than skipping them.

---

## ui

Launch the interactive TUI. Walk through scan → select → clean with visual feedback.

```powershell
broominal ui
```

---

## doctor

Run health checks and verify the integrity of the configuration, directory permissions, and quarantine manifests.

```powershell
broominal doctor                        # run health checks
broominal doctor --fix --yes            # run automatic fixes (e.g. UAC elevation, purging damaged manifests)
```

### Flags
* `--fix`: Run automatic fixes for issues that support them.
* `--yes`: Skip confirmation prompts when used with `--fix`.

---

## config

Show the current configuration file path and its JSON contents.

```powershell
broominal config
```

---

## quarantine

Manage quarantine backups.

```powershell
broominal quarantine list                       # list all quarantine backups
broominal quarantine clean                      # dry-run preview cleanup of backups older than 30 days
broominal quarantine clean --yes                # delete backups older than 30 days
broominal quarantine clean --yes --max-age-days 7 # delete backups older than 7 days
broominal quarantine delete <id> --yes          # delete a specific backup by ID
broominal quarantine delete --all --yes         # delete all backups
```

### quarantine clean Flags
* `--yes`: Execute the deletion (default is dry-run preview).
* `--max-age-days <days>`: Delete backups older than this many days (default: 30).

### quarantine delete Flags
* `--yes`: Execute the deletion (default is dry-run preview).
* `--all`: Delete all backups.



## report

Scan the system and save a JSON report of all cleanup candidates.

```powershell
broominal report                                # scan and save JSON report to reports dir
broominal report --preset deep                  # run deep preset scan and save report
broominal report --output C:\path\to\report.json # write report to specific file path
broominal report --stdout                       # print report JSON to stdout instead of saving to disk
```

### Flags
* `--preset <quick|standard|deep>`: Preset to use for the report scan.
* `--output <file-path>`: Save the JSON report to the specified file.
* `--stdout`: Output JSON to stdout directly.

---

## path

Manage the broominal executable in the user's system PATH.

```powershell
broominal path status                           # show whether broominal is in PATH
broominal path add                              # add broominal directory to user PATH
broominal path remove                           # remove broominal directory from user PATH
```

Changes take effect after restarting the terminal.

---

## update

Check for updates on GitHub and install the latest release if available.

```powershell
broominal update
```

---

## Global Flags

| Flag | Description |
|------|-------------|
| `--json-logs` | Output logs in JSON format (for structured logging pipelines) |
