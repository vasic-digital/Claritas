# test-coverage.md — digital.vasic.claritas (round-263)

Symbol → test ledger for the Claritas submodule. Every exported and
contract-documented internal symbol is cross-referenced to (a) the unit
test that exercises it, and (b) the multi-locale Challenge that exercises
it against real bilingual extraction-attempt + benign inputs.

Round-263 deep-doc enrichment co-maintained with `README.md`, the
multi-locale fixture at `tests/fixtures/claritas/payloads.json`, the
bilingual Challenge runner at `challenges/runner/main.go`, and the
paired-mutation gate at `challenges/scripts/claritas_describe_challenge.sh`.

> Verbatim 2026-05-19 operator mandate: *"all existing tests and Challenges
> do work in anti-bluff manner - they MUST confirm that all tested codebase
> really works as expected! We had been in position that all tests do
> execute with success and all Challenges as well, but in reality the most
> of the features does not work and can't be used! This MUST NOT be the
> case and execution of tests and Challenges MUST guarantee the quality,
> the completition and full usability by end users of the product!"*

## Exported symbols (package `pkg/client`)

| Symbol | File | Unit test | Challenge section |
|--------|------|-----------|-------------------|
| `Client` (type) | `client.go` | every test in `client_test.go`, `client_extra_test.go` constructs one | every `[archive]*` + `[detect]*` section |
| `ExtractionDetection` (type) | `client.go` | `client_test.go` (return type) | `[detect][benign]`, `[detect][attack]`, `[detect][attack2]` |
| `New` (func) | `client.go` | `TestNew_*` | runner constructs via `claritas.New()` once at startup |
| `NewFromConfig` (func) | `client.go` | `TestNewFromConfig_*` (in extra) | configuration-injection path; not driven by runner |
| `Close` (method) | `client.go` | `TestClient_Close*` | `[invariant][close]` (idempotency assertion) |
| `Config` (method) | `client.go` | `TestClient_Config` | covered indirectly through `New` round-trip |
| `AddEntry` (method) | `client.go` | `TestClient_AddEntry*`, `TestClient_AddEntry_Validation` | `[archive][addentry]` |
| `Count` (method) | `client.go` | `TestClient_Count*` | `[archive][seed]`, `[archive][addentry]` (delta) |
| `SearchPrompts` (method) | `client.go` | `TestClient_SearchPrompts*`, `TestClient_SearchPrompts_Filters` | `[archive][search]` |
| `GetPromptByID` (method) | `client.go` | `TestClient_GetPromptByID*` | `[archive][addentry]` (round-trip) |
| `GetByCompany` (method) | `client.go` | `TestClient_GetByCompany*` | `[archive][bycompany]` (one per seeded company) |
| `GetByCategory` (method) | `client.go` | `TestClient_GetByCategory*` | `[archive][bycategory]` (chat) |
| `ComparePrompts` (method) | `client.go` | `TestClient_ComparePrompts*` | `[archive][compare]` (shared "you" token assertion) |
| `GetArchiveStats` (method) | `client.go` | `TestClient_GetArchiveStats*` | `[archive][stats]` |
| `ExportToFormat` (method) | `client.go` | `TestClient_ExportToFormat*` | `[archive][export]` (JSON round-trip) |
| `AnalyzeTrends` (method) | `client.go` | `TestClient_AnalyzeTrends*` | `[archive][trends]` (single-bucket baseline) |
| `DetectExtraction` (method) | `client.go` | `TestClient_DetectExtraction*` | `[detect][benign]`, `[detect][attack]`, `[detect][attack2]` per locale |

## Exported symbols (package `pkg/types`)

| Symbol | File | Unit test | Challenge section |
|--------|------|-----------|-------------------|
| `SystemPrompt` (type) | `types.go` | `TestSystemPrompt_Validate` | runner does not consume this legacy alias directly; client uses `PromptEntry` |
| `PromptEntry` (type) | `types.go` | `TestPromptEntry_Validate` | every `[archive]*` section returns/inspects these |
| `SearchOptions` (type) | `types.go` | `TestSearchOptions_*` | `[archive][search]` |
| `ArchiveStats` (type) | `types.go` | `TestArchiveStats_*` (in extra) | `[archive][stats]` |
| `ComparisonResult` (type) | `types.go` | `TestComparisonResult_*` (in extra) | `[archive][compare]` |
| `ExportOptions` (type) | `types.go` | covered through `ExportToFormat` | `[archive][export]` |
| `TrendOptions` (type) | `types.go` | `TestTrendOptions_Defaults` | `[archive][trends]` |
| `TrendAnalysis` (type) | `types.go` | covered through `AnalyzeTrends` | `[archive][trends]` |
| `TrendPoint` (type) | `types.go` | covered through `AnalyzeTrends` | `[archive][trends]` |

## Internal helpers (package-private, contract-documented)

These are not exported but are exercised end-to-end through the public
methods above. Listed here so the paired-mutation Challenge can
cross-reference them and detect ledger drift.

| Symbol | File | Exercised via |
|--------|------|---------------|
| `promptMatches` | `client.go` | `SearchPrompts` per `TestClient_SearchPrompts_Filters`; runner `[archive][search]` |
| `tokens` | `client.go` | `ComparePrompts` per `TestClient_ComparePrompts`; runner `[archive][compare]` |
| `toSet` | `client.go` | covered indirectly through `ComparePrompts` |
| `setSlice` | `client.go` | covered indirectly through `GetArchiveStats` |
| `containsFold` | `client.go` | `TestClient_SearchPrompts_Filters` (companies/categories filters); runner `[archive][search]` |
| `anyOverlapFold` | `client.go` | `TestClient_SearchPrompts_Filters` (tags overlap); runner `[archive][search]` |
| `seedDefaults` | `client.go` | implicit in every test via `New()`; runner `[archive][seed]` |

## Behavioural invariants (asserted at runtime)

| Invariant | Unit test | Challenge section |
|-----------|-----------|-------------------|
| Default seed count = 3 (ChatGPT/Claude/Gemini) | `TestClient_Count_DefaultSeed` | `[archive][seed]` (asserted against fixture's `default_total`) |
| `AddEntry` increments `Count` by exactly 1 | `TestClient_AddEntry` | `[archive][addentry]` |
| `GetPromptByID` round-trips the entry after `AddEntry` | `TestClient_GetPromptByID` | `[archive][addentry]` |
| `SearchPrompts(query=chat)` returns >=3 hits on seed | `TestClient_SearchPrompts_QueryChat` (extra) | `[archive][search]` |
| `GetByCompany(seed-company)` returns >=1 entry | `TestClient_GetByCompany_Seed` | `[archive][bycompany]` (×3 companies) |
| `ComparePrompts` similarities are a subset of token-intersection | `TestClient_ComparePrompts` | `[archive][compare]` |
| `ExportToFormat("json")` produces JSON that decodes back to `[]PromptEntry` | `TestClient_ExportToFormat_JSONRoundtrip` | `[archive][export]` |
| `AnalyzeTrends` baseline bucket count = archive size | `TestClient_AnalyzeTrends_Baseline` | `[archive][trends]` |
| Benign prompt → `Detected=false` | `TestClient_DetectExtraction_Benign` | `[detect][benign][*]` per locale |
| Attack prompt → `Detected=true` with `Matched` non-empty | `TestClient_DetectExtraction_Attack` | `[detect][attack][*]` per locale |
| Multi-pattern attack → `Confidence >= 0.7 + 0.05*N` capped at 0.95 | `TestClient_DetectExtraction_ConfidenceFloor` | `[detect][attack2][*]` per locale |
| `Close` is idempotent (no error on second call) | `TestClient_Close_Idempotent` | `[invariant][close]` |

## Multi-locale extraction-attempt matrix

The fixture at `tests/fixtures/claritas/payloads.json` carries 5
locales: `en`, `sr`, `ja`, `ar`, `zh-CN`. Each locale ships:

- **Benign sample** — a polite summarisation request in the local
  language. Must NOT trigger `DetectExtraction` (`benign_should_fire=false`).
- **Attack sample** — a code-switched prompt: native-language carrier
  sentence with an English jailbreak directive embedded. MUST trigger
  `DetectExtraction` because real-world attackers code-switch to
  bypass naive language-detector guards.
- **Attack-second sample** — a multi-pattern stack (DAN persona +
  developer-mode + reveal-prompt + persona-swap) designed to exercise
  the confidence ramp (`0.7 + 0.05*matched`, capped at 0.95).

The fixture also carries `archive_assertions` for the archive surface:
seeded count, companies that MUST be present, the "chat" category
floor, the lowercase token (`you`) that MUST appear in the seeded
ChatGPT-vs-Claude comparison's similarities slice.

## Test execution

```bash
# Unit floor (race-enabled, no cache)
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -count=1 -p 1 -race ./...

# Multi-locale runner (real Claritas exerciser against bilingual fixtures)
go run ./challenges/runner -fixtures tests/fixtures/claritas/payloads.json

# Paired-mutation describe gate
bash challenges/scripts/claritas_describe_challenge.sh
bash challenges/scripts/claritas_describe_challenge.sh --anti-bluff-mutate  # exit 99 expected
```
