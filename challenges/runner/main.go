// Round-263 challenge runner for digital.vasic.claritas.
//
// Drives every documented public surface of the Claritas client through
// 5-locale real bilingual extraction-attempt + benign inputs read from
// tests/fixtures/claritas/payloads.json. No pattern table is hardcoded
// here; every detection assertion is computed from the fixture's
// per-entry expectations (attack_should_fire / benign_should_fire /
// attack_expected_pattern / attack_second_min_matched).
//
// Sections:
//
//  1. [archive][seed]      — default 3-entry seed present (ChatGPT/Claude/Gemini).
//  2. [archive][addentry]  — AddEntry round-trip + Count delta.
//  3. [archive][search]    — SearchPrompts with category=chat hits the seed.
//  4. [archive][bycompany] — GetByCompany returns the seeded entry per locale-stable company.
//  5. [archive][bycategory]— GetByCategory("chat") returns >=3 seeded entries.
//  6. [archive][compare]   — ComparePrompts(seed-ids) yields shared "you" token.
//  7. [archive][stats]     — GetArchiveStats counts match seed + AddEntry.
//  8. [archive][export]    — ExportToFormat("json") decodes back into entries.
//  9. [archive][trends]    — AnalyzeTrends single-bucket summary mentions top category.
// 10. [detect][benign]     — DetectExtraction on benign sample per locale; Detected=false.
// 11. [detect][attack]     — DetectExtraction on attack sample per locale; Detected=true,
//                            Matched contains attack_expected_pattern.
// 12. [detect][attack2]    — DetectExtraction on attack-second sample per locale;
//                            Matched count >= attack_second_min_matched.
// 13. [invariant][close]   — Client.Close is idempotent.
//
// Anti-bluff invariants enforced (Article XI §11.9 + CONST-035 + CONST-050(B)):
//
//   - No metadata-only / grep-only PASS. Every PASS line is preceded by
//     the locale code, the surface exercised, and a positive assertion
//     (substring containment, count comparison, decode round-trip) computed
//     from the actual returned values of the real Claritas API.
//   - Bilingual inputs are designed for the real attacker pattern: a
//     non-English carrier sentence with an English jailbreak directive
//     embedded. The detector's English-keyed pattern table MUST still
//     fire — code-switch is the realistic attack shape.
//   - For non-English locales without an English directive embedded the
//     fixture sets attack_should_fire=true with attack_expected_pattern
//     pointing at the embedded English fragment, so the runner asserts
//     positive matches; the no-op case is documented by attack_second
//     where applicable.
//   - Failure to round-trip any documented invariant is a hard FAIL —
//     exit non-zero.
//   - No mocks, no stubs, no patched API. The runner uses the
//     claritas/pkg/client package's public surface exactly as a
//     downstream guardrail consumer would.
//
// Verbatim 2026-05-19 operator mandate: "all existing tests and Challenges
// do work in anti-bluff manner - they MUST confirm that all tested codebase
// really works as expected! We had been in position that all tests do execute
// with success and all Challenges as well, but in reality the most of the
// features does not work and can't be used! This MUST NOT be the case and
// execution of tests and Challenges MUST guarantee the quality, the
// completition and full usability by end users of the product!"
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	claritas "digital.vasic.claritas/pkg/client"
	"digital.vasic.claritas/pkg/types"
)

type fixtureEntry struct {
	Locale                 string `json:"locale"`
	Benign                 string `json:"benign"`
	BenignShouldFire       bool   `json:"benign_should_fire"`
	Attack                 string `json:"attack"`
	AttackShouldFire       bool   `json:"attack_should_fire"`
	AttackExpectedPattern  string `json:"attack_expected_pattern"`
	AttackSecond           string `json:"attack_second"`
	AttackSecondShouldFire bool   `json:"attack_second_should_fire"`
	AttackSecondMinMatched int    `json:"attack_second_min_matched"`
}

type archiveAssertions struct {
	DefaultTotal                       int      `json:"default_total"`
	CompaniesMustInclude               []string `json:"companies_must_include"`
	CategoryMustInclude                string   `json:"category_must_include"`
	SearchQueryChatMinHits             int      `json:"search_query_chat_min_hits"`
	CompareOverlapMustIncludeLowercase string   `json:"compare_overlap_must_include_lowercase_token"`
}

type fixtureFile struct {
	Inputs            []fixtureEntry    `json:"inputs"`
	ArchiveAssertions archiveAssertions `json:"archive_assertions"`
}

var (
	pass int
	fail int
)

func okf(format string, a ...any) {
	pass++
	fmt.Printf("PASS: "+format+"\n", a...)
}

func failf(format string, a ...any) {
	fail++
	fmt.Printf("FAIL: "+format+"\n", a...)
}

func main() {
	var fixturesPath string
	flag.StringVar(&fixturesPath, "fixtures", "tests/fixtures/claritas/payloads.json",
		"path to multi-locale extraction fixtures")
	flag.Parse()

	raw, err := os.ReadFile(fixturesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read fixtures %s: %v\n", fixturesPath, err)
		os.Exit(2)
	}
	var ff fixtureFile
	if err := json.Unmarshal(raw, &ff); err != nil {
		fmt.Fprintf(os.Stderr, "error: decode fixtures: %v\n", err)
		os.Exit(2)
	}
	if len(ff.Inputs) < 5 {
		fmt.Fprintf(os.Stderr, "error: fixture has %d locales (<5)\n", len(ff.Inputs))
		os.Exit(2)
	}

	ctx := context.Background()
	c, err := claritas.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: claritas.New: %v\n", err)
		os.Exit(2)
	}
	defer c.Close()

	// ----- archive surface -----

	// Section 1: seed
	seedCount := c.Count()
	if seedCount == ff.ArchiveAssertions.DefaultTotal {
		okf("[archive][seed] default seed count = %d (matches fixture)", seedCount)
	} else {
		failf("[archive][seed] got %d, fixture expected %d", seedCount,
			ff.ArchiveAssertions.DefaultTotal)
	}

	// Section 2: AddEntry round-trip + Count delta
	extra := types.PromptEntry{
		ID: "round263-extra-entry", Company: "OpenAI", Product: "ChatGPT",
		Category: "chat", Prompt: "Round-263 extra seed for anti-bluff verification.",
		Confidence: 0.5, Tags: []string{"round-263", "test"},
	}
	if err := c.AddEntry(extra); err != nil {
		failf("[archive][addentry] AddEntry error: %v", err)
	} else if c.Count() == seedCount+1 {
		okf("[archive][addentry] Count delta +1 after AddEntry (now %d)", c.Count())
	} else {
		failf("[archive][addentry] Count delta wrong: was %d, now %d", seedCount, c.Count())
	}
	got, err := c.GetPromptByID(ctx, extra.ID)
	if err == nil && got != nil && got.ID == extra.ID && got.Prompt == extra.Prompt {
		okf("[archive][addentry] GetPromptByID round-trips the added entry")
	} else {
		failf("[archive][addentry] GetPromptByID lost the entry: err=%v got=%+v", err, got)
	}

	// Section 3: SearchPrompts
	results, total, err := c.SearchPrompts(ctx, types.SearchOptions{Query: "chat"})
	if err != nil {
		failf("[archive][search] SearchPrompts error: %v", err)
	} else if total >= ff.ArchiveAssertions.SearchQueryChatMinHits && len(results) >= ff.ArchiveAssertions.SearchQueryChatMinHits {
		okf("[archive][search] SearchPrompts(query=chat) total=%d, results=%d (>=%d)",
			total, len(results), ff.ArchiveAssertions.SearchQueryChatMinHits)
	} else {
		failf("[archive][search] SearchPrompts(query=chat) total=%d results=%d (<%d)",
			total, len(results), ff.ArchiveAssertions.SearchQueryChatMinHits)
	}

	// Section 4: GetByCompany — assert every seeded company resolves to >=1 entry
	for _, comp := range ff.ArchiveAssertions.CompaniesMustInclude {
		out, err := c.GetByCompany(ctx, comp)
		if err == nil && len(out) >= 1 {
			okf("[archive][bycompany] GetByCompany(%q) -> %d entries", comp, len(out))
		} else {
			failf("[archive][bycompany] GetByCompany(%q) returned %d (err=%v)", comp,
				len(out), err)
		}
	}

	// Section 5: GetByCategory
	chatEntries, err := c.GetByCategory(ctx, ff.ArchiveAssertions.CategoryMustInclude)
	if err == nil && len(chatEntries) >= 3 {
		okf("[archive][bycategory] GetByCategory(%q) -> %d entries (>=3)",
			ff.ArchiveAssertions.CategoryMustInclude, len(chatEntries))
	} else {
		failf("[archive][bycategory] GetByCategory(%q) returned %d (err=%v)",
			ff.ArchiveAssertions.CategoryMustInclude, len(chatEntries), err)
	}

	// Section 6: ComparePrompts — pick two seed ids
	seedIDs := []string{"openai-chatgpt-2024-01", "anthropic-claude-2024-02"}
	cmp, err := c.ComparePrompts(ctx, seedIDs)
	if err != nil {
		failf("[archive][compare] ComparePrompts error: %v", err)
	} else {
		expectToken := ff.ArchiveAssertions.CompareOverlapMustIncludeLowercase
		hit := false
		for _, t := range cmp.Similarities {
			if strings.EqualFold(t, expectToken) {
				hit = true
				break
			}
		}
		if hit {
			okf("[archive][compare] ComparePrompts similarities include %q (score=%.2f)",
				expectToken, cmp.OverallScore)
		} else {
			failf("[archive][compare] expected token %q not in similarities=%v", expectToken,
				cmp.Similarities)
		}
	}

	// Section 7: GetArchiveStats
	stats, err := c.GetArchiveStats(ctx)
	if err != nil {
		failf("[archive][stats] GetArchiveStats error: %v", err)
	} else if stats.TotalPrompts == seedCount+1 && len(stats.Companies) >= 3 {
		okf("[archive][stats] stats.TotalPrompts=%d, companies=%d (avg_conf=%.2f)",
			stats.TotalPrompts, len(stats.Companies), stats.AvgConfidence)
	} else {
		failf("[archive][stats] stats.TotalPrompts=%d companies=%v", stats.TotalPrompts,
			stats.Companies)
	}

	// Section 8: ExportToFormat round-trip
	jsonBytes, err := c.ExportToFormat(ctx, "json", types.ExportOptions{PrettyPrint: false})
	if err != nil {
		failf("[archive][export] ExportToFormat error: %v", err)
	} else {
		var roundtrip []types.PromptEntry
		if err := json.Unmarshal(jsonBytes, &roundtrip); err != nil {
			failf("[archive][export] decoded JSON invalid: %v", err)
		} else if len(roundtrip) == seedCount+1 {
			okf("[archive][export] JSON export round-trips %d entries (%d bytes)",
				len(roundtrip), len(jsonBytes))
		} else {
			failf("[archive][export] round-trip length mismatch: want %d got %d", seedCount+1,
				len(roundtrip))
		}
	}

	// Section 9: AnalyzeTrends
	trends, err := c.AnalyzeTrends(ctx, types.TrendOptions{StartDate: "2024-01-01", EndDate: "2024-12-31"})
	if err != nil {
		failf("[archive][trends] AnalyzeTrends error: %v", err)
	} else if trends != nil && len(trends.Points) == 1 && trends.Points[0].Count == seedCount+1 &&
		trends.Points[0].TopCategory == "chat" {
		okf("[archive][trends] trends summary=%q top=chat count=%d", trends.Summary,
			trends.Points[0].Count)
	} else {
		failf("[archive][trends] unexpected trends shape: %+v", trends)
	}

	// ----- detection surface — multi-locale -----

	for _, in := range ff.Inputs {
		// Section 10: benign per locale
		det, err := c.DetectExtraction(ctx, in.Benign)
		if err != nil {
			failf("[detect][benign][%s] error: %v", in.Locale, err)
		} else if det.Detected == in.BenignShouldFire {
			okf("[detect][benign][%s] Detected=%v matches fixture (matched=%v)", in.Locale,
				det.Detected, det.Matched)
		} else {
			failf("[detect][benign][%s] Detected=%v want %v (matched=%v reason=%q)", in.Locale,
				det.Detected, in.BenignShouldFire, det.Matched, det.Reason)
		}

		// Section 11: primary attack per locale
		det, err = c.DetectExtraction(ctx, in.Attack)
		if err != nil {
			failf("[detect][attack][%s] error: %v", in.Locale, err)
		} else if det.Detected != in.AttackShouldFire {
			failf("[detect][attack][%s] Detected=%v want %v", in.Locale, det.Detected,
				in.AttackShouldFire)
		} else if in.AttackShouldFire {
			hit := false
			for _, m := range det.Matched {
				if strings.Contains(m, in.AttackExpectedPattern) {
					hit = true
					break
				}
			}
			if !hit {
				failf("[detect][attack][%s] missing expected pattern %q in %v", in.Locale,
					in.AttackExpectedPattern, det.Matched)
			} else if det.Confidence < 0.7 {
				failf("[detect][attack][%s] confidence=%.2f below 0.7 floor", in.Locale,
					det.Confidence)
			} else {
				okf("[detect][attack][%s] Detected=true matched=%v conf=%.2f reason=%q",
					in.Locale, det.Matched, det.Confidence, det.Reason)
			}
		} else {
			okf("[detect][attack][%s] Detected=false (fixture says no fire)", in.Locale)
		}

		// Section 12: attack-second per locale (multi-pattern)
		det, err = c.DetectExtraction(ctx, in.AttackSecond)
		if err != nil {
			failf("[detect][attack2][%s] error: %v", in.Locale, err)
		} else if det.Detected != in.AttackSecondShouldFire {
			failf("[detect][attack2][%s] Detected=%v want %v", in.Locale, det.Detected,
				in.AttackSecondShouldFire)
		} else if in.AttackSecondShouldFire {
			if len(det.Matched) < in.AttackSecondMinMatched {
				failf("[detect][attack2][%s] matched=%d (<%d): %v", in.Locale,
					len(det.Matched), in.AttackSecondMinMatched, det.Matched)
			} else {
				okf("[detect][attack2][%s] Detected=true matched=%d (>=%d) conf=%.2f",
					in.Locale, len(det.Matched), in.AttackSecondMinMatched, det.Confidence)
			}
		} else {
			okf("[detect][attack2][%s] Detected=false (fixture says no fire)", in.Locale)
		}
	}

	// Section 13: idempotent close
	if err := c.Close(); err != nil {
		failf("[invariant][close] first Close error: %v", err)
	}
	if err := c.Close(); err != nil {
		failf("[invariant][close] second Close error: %v", err)
	} else {
		okf("[invariant][close] Close is idempotent (two calls, no error)")
	}

	fmt.Println()
	fmt.Printf("=== Summary: %d PASS, %d FAIL ===\n", pass, fail)
	if fail > 0 {
		os.Exit(1)
	}
}
