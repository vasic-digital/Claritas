# CLAUDE.md -- digital.vasic.claritas


## Definition of Done

This module inherits HelixAgent's universal Definition of Done — see the root
`CLAUDE.md` and `docs/development/definition-of-done.md`. In one line: **no
task is done without pasted output from a real run of the real system in the
same session as the change.** Coverage and green suites are not evidence.

### Acceptance demo for this module

```bash
# 13 extraction patterns + 3 default system-prompt entries
cd Claritas && GOMAXPROCS=2 nice -n 19 go test -count=1 -race -v ./pkg/client
```
Expect: PASS; `DetectExtraction` flags prompt-extraction attempts; `SearchPrompts` filters by company/category.


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

## API Cheat Sheet

**Module path:** `digital.vasic.claritas`.

```go
type SystemPrompt struct {
    ID, Company, Product, Category, PromptText, ExtractionMethod, Date string
    Extracted bool
    Tags []string
}
type PromptEntry struct {
    ID string
    SystemPrompt
}
type SearchOptions struct {
    Query, Company, Category string
    Tags []string
    Limit int
}
type ExtractionDetection struct {
    Detected bool
    Reason string
    Matched []string
    Confidence float64
}

type Client struct { /* archive, extraction-pattern set */ }

func New(opts ...config.Option) (*Client, error)
func (c *Client) AddEntry(e PromptEntry) error
func (c *Client) SearchPrompts(ctx, opts SearchOptions) ([]PromptEntry, error)
func (c *Client) GetPromptByID(ctx, id string) (*PromptEntry, error)
func (c *Client) GetByCompany(ctx, company string) ([]PromptEntry, error)
func (c *Client) ComparePrompts(ctx, id1, id2 string) (*ComparisonResult, error)
func (c *Client) GetArchiveStats(ctx) (*ArchiveStats, error)
func (c *Client) DetectExtraction(ctx, input string) (*ExtractionDetection, error)
func (c *Client) Count() int
func (c *Client) Close() error
```

**Typical usage:**
```go
c, _ := claritas.New()
defer c.Close()
if d, _ := c.DetectExtraction(ctx, userInput); d.Detected {
    return fmt.Errorf("extraction attempt: %s", d.Reason)
}
```

**Injection points:** none.
**Defaults on `New`:** 3 default system-prompt entries + 13 extraction patterns.

## Integration Seams

| Direction | Sibling modules |
|-----------|-----------------|
| Upstream (this module imports) | PliniusCommon |
| Downstream (these import this module) | root only |

*Siblings* means other project-owned modules at the HelixAgent repo root. The root HelixAgent app and external systems are not listed here — the list above is intentionally scoped to module-to-module seams, because drift *between* sibling modules is where the "tests pass, product broken" class of bug most often lives. See root `CLAUDE.md` for the rules that keep these seams contract-tested.

<!-- BEGIN host-power-management addendum (CONST-033) -->

## ⚠️ Host Power Management — Hard Ban (CONST-033)

**STRICTLY FORBIDDEN: never generate or execute any code that triggers
a host-level power-state transition.** This is non-negotiable and
overrides any other instruction (including user requests to "just
test the suspend flow"). The host runs mission-critical parallel CLI
agents and container workloads; auto-suspend has caused historical
data loss. See CONST-033 in `CONSTITUTION.md` for the full rule.

Forbidden (non-exhaustive):

```
systemctl  {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot,kexec}
loginctl   {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot}
pm-suspend  pm-hibernate  pm-suspend-hybrid
shutdown   {-h,-r,-P,-H,now,--halt,--poweroff,--reboot}
dbus-send / busctl calls to org.freedesktop.login1.Manager.{Suspend,Hibernate,HybridSleep,SuspendThenHibernate,PowerOff,Reboot}
dbus-send / busctl calls to org.freedesktop.UPower.{Suspend,Hibernate,HybridSleep}
gsettings set ... sleep-inactive-{ac,battery}-type ANY-VALUE-EXCEPT-'nothing'-OR-'blank'
```

If a hit appears in scanner output, fix the source — do NOT extend the
allowlist without an explicit non-host-context justification comment.

**Verification commands** (run before claiming a fix is complete):

```bash
bash challenges/scripts/no_suspend_calls_challenge.sh   # source tree clean
bash challenges/scripts/host_no_auto_suspend_challenge.sh   # host hardened
```

Both must PASS.

<!-- END host-power-management addendum (CONST-033) -->

