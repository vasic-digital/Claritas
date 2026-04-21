# CLAUDE.md -- digital.vasic.claritas

Module-specific guidance for Claude Code.

## Status

**FUNCTIONAL.** 2 packages (types, client) ship tested implementations;
`go test -race ./...` all green. Default archive (3 entries) seeded on
`New()`; `DetectExtraction` detector ships with 13 canonical extraction
patterns.

## Hard rules

1. **NO CI/CD pipelines** -- no `.github/workflows/`, `.gitlab-ci.yml`,
   `Jenkinsfile`, `.travis.yml`, `.circleci/`, or any automated
   pipeline. No Git hooks either. Permanent.
2. **SSH-only for Git** -- `git@github.com:...` / `git@gitlab.com:...`.
3. **Conventional Commits** -- `feat(claritas): ...`, `fix(...)`,
   `docs(...)`, `test(...)`, `refactor(...)`.
4. **Code style** -- `gofmt`, `goimports`, 100-char line ceiling,
   errors always checked and wrapped (`fmt.Errorf("...: %w", err)`).
5. **Resource cap for tests** --
   `GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -count=1 -p 1 -race ./...`

## Purpose

System-prompt extraction detection + leaked-prompt archive. Key
surface: `SearchPrompts`, `GetPromptByID`, `GetByCompany`,
`GetByCategory`, `ComparePrompts`, `GetArchiveStats`, `ExportToFormat`,
`AnalyzeTrends`, `AddEntry`, `Count`, `DetectExtraction`.

## Primary consumer

HelixAgent (`dev.helix.agent`) — red-team / guardrail ingress.

## Testing

```
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -count=1 -p 1 -race ./...
```
