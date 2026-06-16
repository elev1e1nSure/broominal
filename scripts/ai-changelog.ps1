<#
.SYNOPSIS
    Beautifies raw changelog/commit list via OpenRouter AI.
.DESCRIPTION
    Takes raw text (git-cliff output, git log, or piped input) and sends it
    to a cheap LLM via OpenRouter to produce polished release notes.
.PARAMETER InputText
    Raw changelog text. If omitted and not piped, auto-generates from git log
    since the last tag.
.PARAMETER FromTag
    Starting tag for git log (e.g. "v1.5.1"). Defaults to latest tag.
.PARAMETER ToRef
    Ending ref for git log. Defaults to HEAD.
.PARAMETER Model
    OpenRouter model ID. Default: google/gemini-2.5-flash.
    Override with env var OPENROUTER_MODEL.
.PARAMETER ApiKey
    OpenRouter API key. Falls back to env var OPENROUTER_API_KEY.
.PARAMETER Raw
    Skip AI beautification — just output the raw input from git-cliff/git log.
    Useful for piping into another tool or previewing.
.PARAMETER OutputFile
    Write result to file instead of stdout.
.EXAMPLE
    git cliff --latest | ./scripts/ai-changelog.ps1
.EXAMPLE
    ./scripts/ai-changelog.ps1 -FromTag v1.5.1 -ToRef v1.6.0
.EXAMPLE
    ./scripts/ai-changelog.ps1 -Raw -FromTag v1.5.1
#>

param(
    [string]$InputText,

    [string]$FromTag,
    [string]$ToRef = "HEAD",

    [string]$Model = $env:OPENROUTER_MODEL,
    [string]$ApiKey = $env:OPENROUTER_API_KEY,

    [switch]$Raw,
    [string]$OutputFile
)

$ErrorActionPreference = "Stop"

if (-not $Model) { $Model = "deepseek/deepseek-v4-flash" }

# collect piped input via $input automatic variable
$piped = $input | Out-String
if ($piped.Trim() -and -not $InputText) {
    $InputText = $piped.Trim()
}

function Get-RawChangelog {
    if ($InputText) { return $InputText }

    if (-not $FromTag) {
        $script:FromTag = git describe --tags --abbrev=0 2>$null
        $FromTag = $script:FromTag
        if (-not $FromTag) {
            Write-Warning "No tags found, using last 50 commits"
            $log = git log --pretty=format:"%s" -n 50
            return $log -join "`n"
        }
    }

    Write-Host "Commits: $FromTag..$ToRef" -ForegroundColor DarkGray
    $log = git log "$FromTag..$ToRef" --pretty=format:"- %s"
    if (-not $log) {
        Write-Error "No commits between $FromTag and $ToRef"
        exit 1
    }
    return $log -join "`n"
}

function Invoke-OpenRouter {
    param([string]$Prompt, [string]$SystemPrompt)

    $body = @{
        model    = $Model
        messages = @(
            @{ role = "system"; content = $SystemPrompt }
            @{ role = "user"; content = $Prompt }
        )
        temperature = 0.3
        max_tokens  = 4096
    } | ConvertTo-Json -Depth 4

    $headers = @{
        "Content-Type"  = "application/json"
        "Authorization" = "Bearer $ApiKey"
        "HTTP-Referer"  = "https://github.com/rusq/broominal"
        "X-Title"       = "broominal-changelog"
    }

    Write-Host "Calling $Model..." -ForegroundColor DarkGray

    try {
        $response = Invoke-RestMethod -Uri "https://openrouter.ai/api/v1/chat/completions" `
            -Method Post -Headers $headers -Body $body -TimeoutSec 120

        return $response.choices[0].message.content
    }
    catch {
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errBody = $reader.ReadToEnd()
            Write-Error "OpenRouter API error: $errBody"
        }
        else {
            Write-Error "Request failed: $_"
        }
        exit 1
    }
}

$rawLog = Get-RawChangelog

if ($Raw) {
    if ($OutputFile) {
        Set-Content -Path $OutputFile -Value $rawLog -Encoding utf8
        Write-Host "Raw log written to $OutputFile" -ForegroundColor Green
    }
    else {
        Write-Output $rawLog
    }
    exit 0
}

if (-not $ApiKey) {
    Write-Error "OPENROUTER_API_KEY is not set. Pass -ApiKey, set env var, or use -Raw."
    exit 1
}

$systemPrompt = @"
You are a changelog editor. Convert raw commit messages into polished release notes that look hand-written.

Core principle: only include changes a user can notice, verify, or benefit from.

Rules:
- Keep the existing structure: ### section headers with emojis, then bullet list. You may rename section headers if they misrepresent the content.
- Rename '🔧 Miscellaneous Tasks' to '🔧 Maintenance'.
- Rewrite each commit into a clear user-facing description. Expand abbreviations, fix grammar, make it natural English.
- Summary line at the top: one sentence for patches, two sentences max for minor releases. Never use grandiose language ("major overhaul", "complete redesign", etc.).
- NEVER include any of these — they provide zero value to users:
  * Version bumps ("bump version to…", "chore: bump")
  * CI/CD config changes, workflow changes, release infrastructure
  * .gitignore changes
  * Agent/config files (CLAUDE.md, AGENTS.md, rules.md, etc.)
  * Dependency updates without a visible fix
  * Branch-only commits, merge commits, revert commits
  * Internal refactors with no visible user effect (e.g. "replaced switch-case with handler map")
  * Dead code removal, "cleanup model fields", "remove unused code"
  * Changelog/infra tooling (git-cliff config, AI beautifier script, etc.)
- Merge truly trivial adjacent commits into one bullet. Keep meaningful changes separate.
- Refactor commits: include ONLY if there's a visible result. Instead of "refactored X" write "X is now more reliable/faster/clearer".
- Do NOT invent features, fixes, or details not present in the commits.
- Return ONLY the final markdown. No preamble, no code fences.
- Never mention AI, LLM, machine generation, or the beautification process itself. The changelog must read as if written by a human.
"@

$previousRelease = ""
if ($FromTag) {
    Write-Host "Fetching previous release notes for $FromTag..." -ForegroundColor DarkGray
    $previousRelease = gh release view $FromTag --json body -q .body 2>$null
}

$userPrompt = ""
if ($previousRelease) {
    $userPrompt += "Here is the previous release for tone and style reference:`n$previousRelease`n`n"
}
$userPrompt += "Here are the commits for this release. Produce polished user-facing release notes:`n`n$rawLog"

$result = Invoke-OpenRouter -Prompt $userPrompt -SystemPrompt $systemPrompt

if ($OutputFile) {
    Set-Content -Path $OutputFile -Value $result -Encoding utf8
    Write-Host "Changelog written to $OutputFile" -ForegroundColor Green
}
else {
    Write-Output $result
}
