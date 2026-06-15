# broominal build & dev commands
# run: just <command>

VERSION := "0.0.0-dev"

# build with dev version
build:
    go build -ldflags "-X main.Version={{VERSION}}" -o broominal.exe ./cmd/broominal

# build with release version: just build-release v1.0.0
build-release ver:
    go build -ldflags "-X main.Version={{ver}}" -o broominal.exe ./cmd/broominal

# build and run: just run scan
run +args:
    @{{just_executable()}} build
    @.\broominal.exe {{args}}

# run all tests
test:
    go test ./...

# run package tests only (skip TUI)
test-pkg:
    go test ./pkg/...

# run vet
vet:
    go vet ./...

# format code
fmt:
    gofmt -w .

# run linter
lint:
    golangci-lint run

# full check: fmt → vet → lint → tests
check: fmt vet lint test-pkg

# preview raw changelog (commits since last tag, no AI)
changelog-raw:
    @pwsh -File scripts/ai-changelog.ps1 -Raw

# generate beautified changelog via OpenRouter (set $env:OPENROUTER_MODEL to override model)
changelog:
    @pwsh -File scripts/ai-changelog.ps1

# generate changelog and save to file
changelog-save file:
    @pwsh -File scripts/ai-changelog.ps1 -OutputFile {{file}}

# remove build artifacts
clean:
    -del broominal.exe 2>nul
