# broominal build & dev commands
# run: just <command>

VERSION := "0.0.0-dev"

ACC := "`e[38;2;138;173;244m"
OK  := "`e[38;2;166;218;149m[OK]`e[0m"
RST := "`e[0m"

# build with dev version
build:
    @pwsh -c 'Write-Host "{{ACC}}==>{{RST}} Building broominal (dev)... " -NoNewline'
    @go build -ldflags "-X main.Version={{VERSION}}" -o broominal.exe ./cmd/broominal
    @pwsh -c 'Write-Host "{{OK}}"'

# build with release version: just build-release v1.0.0
build-release ver:
    @pwsh -c 'Write-Host "{{ACC}}==>{{RST}} Building release ({{ver}})... " -NoNewline'
    @go build -ldflags "-X main.Version={{ver}}" -o broominal.exe ./cmd/broominal
    @pwsh -c 'Write-Host "{{OK}}"'

# build and run: just run scan
run +args:
    @{{just_executable()}} build
    @pwsh -c 'Write-Host "`n{{ACC}}==>{{RST}} Running broominal.exe {{args}} `n"'
    @.\broominal.exe {{args}}

# run all tests
test:
    @pwsh -c 'Write-Host "{{ACC}}==>{{RST}} Running all tests... `n"'
    @go test ./...
    @pwsh -c 'Write-Host "{{OK}}"'

# run package tests only (skip TUI)
test-pkg:
    @pwsh -c 'Write-Host "{{ACC}}==>{{RST}} Running package tests... `n"'
    @go test ./pkg/...
    @pwsh -c 'Write-Host "{{OK}}"'

# run vet
vet:
    @pwsh -c 'Write-Host "{{ACC}}==>{{RST}} Running go vet... " -NoNewline'
    @go vet ./...
    @pwsh -c 'Write-Host "{{OK}}"'

# format code
fmt:
    @pwsh -c 'Write-Host "{{ACC}}==>{{RST}} Formatting code... " -NoNewline'
    @gofmt -w .
    @pwsh -c 'Write-Host "{{OK}}"'

# run linter
lint:
    @pwsh -c 'Write-Host "{{ACC}}==>{{RST}} Running golangci-lint... " -NoNewline'
    @golangci-lint run
    @pwsh -c 'Write-Host "{{OK}}"'

# full check: fmt → vet → lint → tests
check: fmt vet lint test-pkg

# preview raw changelog (commits since last tag, no AI)
changelog-raw:
    @pwsh -c 'Write-Host "{{ACC}}==>{{RST}} Generating raw changelog... `n"'
    @pwsh -File scripts/ai-changelog.ps1 -Raw

# generate beautified changelog via OpenRouter (set $env:OPENROUTER_MODEL to override model)
changelog:
    @pwsh -c 'Write-Host "{{ACC}}==>{{RST}} Generating AI changelog... `n"'
    @pwsh -File scripts/ai-changelog.ps1

# generate changelog and save to file
changelog-save file:
    @pwsh -c 'Write-Host "{{ACC}}==>{{RST}} Generating AI changelog to {{file}}... `n"'
    @pwsh -File scripts/ai-changelog.ps1 -OutputFile {{file}}

# remove build artifacts
clean:
    @pwsh -c 'Write-Host "{{ACC}}==>{{RST}} Cleaning build artifacts... " -NoNewline'
    -@del broominal.exe 2>nul
    @pwsh -c 'Write-Host "{{OK}}"'
