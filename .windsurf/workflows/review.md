---
auto_execution_mode: 3
description: review code changes
---

review the current code changes for real bugs, security issues, and maintainability problems.

focus on logic errors, missed edge cases, unsafe assumptions, resource leaks, concurrency issues, API contract breaks, caching bugs, and violations of existing project patterns.

inspect only the code needed to understand the change; use parallel tool calls when useful, but do not over-explore.

report only issues you can justify from the code, not guesses or style nitpicks.

include pre-existing bugs only if they are directly relevant or clearly harmful.

for each finding, include severity, affected file/location, why it is a problem, and a concrete fix.

if a specific commit was provided, verify the checked-out state before relying on local files.