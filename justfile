# broominal build & dev commands
# run: just <command>

VERSION := "0.0.0-dev"

PWSH := "powershell -NoProfile -NonInteractive"

ACC := "`e[38;2;138;173;244m"
OK  := "`e[38;2;166;218;149m[OK]`e[0m"
RST := "`e[0m"

# build with dev version
build:
    @{{PWSH}} -c 'Write-Host "{{ACC}}==>{{RST}} Building broominal (dev)... " -NoNewline'
    @go build -ldflags "-X main.Version={{VERSION}}" -o broominal.exe ./cmd/broominal
    @{{PWSH}} -c 'Write-Host "{{OK}}"'

# build with release version: just build-release v1.0.0
build-release ver:
    @{{PWSH}} -c 'Write-Host "{{ACC}}==>{{RST}} Building release ({{ver}})... " -NoNewline'
    @{{PWSH}} -c '$env:GOOS = "windows"; $env:GOARCH = "amd64"; go build -ldflags "-X main.Version={{ver}}" -o broominal_windows_amd64.exe ./cmd/broominal'
    @{{PWSH}} -c 'Write-Host "{{OK}}"'

# build and run: just run scan
run +args:
    @just build
    @{{PWSH}} -c 'Write-Host "`n{{ACC}}==>{{RST}} Running broominal.exe {{args}} `n"'
    @./broominal.exe {{args}}

# run all tests
test:
    @{{PWSH}} -c 'Write-Host "{{ACC}}==>{{RST}} Running all tests... `n"'
    @go test ./...
    @{{PWSH}} -c 'Write-Host "{{OK}}"'

# run package tests only (skip TUI)
test-pkg:
    @{{PWSH}} -c 'Write-Host "{{ACC}}==>{{RST}} Running package tests... `n"'
    @go test ./pkg/...
    @{{PWSH}} -c 'Write-Host "{{OK}}"'

# run vet
vet:
    @{{PWSH}} -c 'Write-Host "{{ACC}}==>{{RST}} Running go vet... " -NoNewline'
    @go vet ./...
    @{{PWSH}} -c 'Write-Host "{{OK}}"'

# format code
fmt:
    @{{PWSH}} -c 'Write-Host "{{ACC}}==>{{RST}} Formatting code... " -NoNewline'
    @gofmt -w .
    @{{PWSH}} -c 'Write-Host "{{OK}}"'

# run linter
lint:
    @{{PWSH}} -c 'Write-Host "{{ACC}}==>{{RST}} Running golangci-lint... " -NoNewline'
    @golangci-lint run
    @{{PWSH}} -c 'Write-Host "{{OK}}"'

# full check: fmt → vet → lint → tests
check: fmt vet lint test-pkg

# preview raw changelog (commits since last tag, no AI)
changelog-raw:
    @{{PWSH}} -c 'Write-Host "{{ACC}}==>{{RST}} Generating raw changelog... `n"'
    @{{PWSH}} -File scripts/ai-changelog.ps1 -Raw

# generate beautified changelog via OpenRouter (set $env:OPENROUTER_MODEL to override model)
changelog:
    @{{PWSH}} -c 'Write-Host "{{ACC}}==>{{RST}} Generating AI changelog... `n"'
    @{{PWSH}} -File scripts/ai-changelog.ps1

# generate changelog and save to file
changelog-save file:
    @{{PWSH}} -c 'Write-Host "{{ACC}}==>{{RST}} Generating AI changelog to {{file}}... `n"'
    @{{PWSH}} -File scripts/ai-changelog.ps1 -OutputFile {{file}}

# remove build artifacts
clean:
    @{{PWSH}} -c 'Write-Host "{{ACC}}==>{{RST}} Cleaning build artifacts... " -NoNewline'
    -@del broominal.exe broominal_windows_amd64.exe 2>nul
    @{{PWSH}} -c 'Write-Host "{{OK}}"'
