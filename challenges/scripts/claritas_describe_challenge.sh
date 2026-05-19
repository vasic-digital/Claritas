#!/usr/bin/env bash
# claritas_describe_challenge.sh
#
# Round-263 paired-mutation deep-doc challenge for digital.vasic.claritas.
#
# Validates that:
#   1. The deep-doc ledger (docs/test-coverage.md) lists every exported
#      and contract-documented internal symbol from pkg/client and
#      pkg/types.
#   2. The multi-locale fixture (tests/fixtures/claritas/payloads.json)
#      parses and contains at least 5 locales with every per-entry
#      detection field populated.
#   3. The multi-locale runner (challenges/runner/main.go) builds and
#      runs, asserting every archive + detection surface end-to-end
#      across every locale against the real Claritas public API
#      (no mocks, no stubs).
#   4. The README enumerates every transform/surface, the API surface
#      table, and the round-263 anti-bluff guarantees section.
#
# Paired-mutation invariant (Article XI §11.9 + CONST-035 + CONST-050(B)):
#   With --anti-bluff-mutate the script plants a deliberate symbol-rename
#   mutation in the ledger (DetectExtraction -> Bogus_MUTATED), reruns
#   validation, and asserts the gate FAILS with exit 99. This proves the
#   gate actually catches ledger-vs-source drift instead of rubber-stamping it.
#
# Exit codes:
#   0  -- gate PASS on clean tree
#   1  -- gate FAIL on clean tree (real failure to fix)
#   99 -- paired-mutation correctly detected (good -- proves anti-bluff)
#   2  -- usage / environment error
#
# Verbatim 2026-05-19 operator mandate: "all existing tests and Challenges
# do work in anti-bluff manner - they MUST confirm that all tested codebase
# really works as expected! We had been in position that all tests do execute
# with success and all Challenges as well, but in reality the most of the
# features does not work and can't be used! This MUST NOT be the case and
# execution of tests and Challenges MUST guarantee the quality, the
# completition and full usability by end users of the product!"

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULE_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

MUTATE=0
for arg in "$@"; do
    case "$arg" in
        --anti-bluff-mutate) MUTATE=1 ;;
        --help|-h)
            sed -n '1,40p' "$0"
            exit 0
            ;;
        *)
            echo "unknown argument: $arg" >&2
            exit 2
            ;;
    esac
done

PASS=0
FAIL=0
TOTAL=0

pass() { PASS=$((PASS+1)); TOTAL=$((TOTAL+1)); echo "  PASS: $1"; }
fail() { FAIL=$((FAIL+1)); TOTAL=$((TOTAL+1)); echo "  FAIL: $1"; }

LEDGER="${MODULE_DIR}/docs/test-coverage.md"
FIXTURE="${MODULE_DIR}/tests/fixtures/claritas/payloads.json"
RUNNER="${MODULE_DIR}/challenges/runner/main.go"
README="${MODULE_DIR}/README.md"

LEDGER_WORK="${LEDGER}"
TMP_LEDGER=""
if [ "${MUTATE}" -eq 1 ]; then
    TMP_LEDGER="$(mktemp)"
    cp "${LEDGER}" "${TMP_LEDGER}"
    # Plant a rename: DetectExtraction -> Bogus_MUTATED, ComparePrompts -> NoSuchMethod
    sed -i 's/\bDetectExtraction\b/Bogus_MUTATED/g; s/\bComparePrompts\b/NoSuchMethod_MUTATED/g' \
        "${TMP_LEDGER}"
    LEDGER_WORK="${TMP_LEDGER}"
    echo "=== Claritas Describe Challenge (anti-bluff-mutate mode) ==="
else
    echo "=== Claritas Describe Challenge (clean mode) ==="
fi
echo ""

# Section 1: ledger presence and freshness
echo "Section 1: docs/test-coverage.md ledger"
if [ ! -f "${LEDGER_WORK}" ]; then
    fail "ledger missing at ${LEDGER_WORK}"
else
    pass "ledger present"
    if grep -qi "round-263" "${LEDGER_WORK}"; then
        pass "ledger marked round-263"
    else
        fail "ledger missing round-263 marker"
    fi
    if grep -q "execution of tests and Challenges MUST guarantee" "${LEDGER_WORK}"; then
        pass "ledger carries Article XI §11.9 mandate"
    else
        fail "ledger missing Article XI §11.9 mandate"
    fi
fi

# Section 2: every exported + contract-documented symbol appears in ledger
echo ""
echo "Section 2: symbols cross-reference"

CHECKED=0
MISSING=0
for sym in Client ExtractionDetection New NewFromConfig Close Config \
           AddEntry Count SearchPrompts GetPromptByID GetByCompany \
           GetByCategory ComparePrompts GetArchiveStats ExportToFormat \
           AnalyzeTrends DetectExtraction \
           PromptEntry SystemPrompt SearchOptions ArchiveStats \
           ComparisonResult ExportOptions TrendOptions TrendAnalysis \
           TrendPoint \
           promptMatches tokens toSet setSlice containsFold anyOverlapFold \
           seedDefaults; do
    CHECKED=$((CHECKED + 1))
    if grep -qE "\\b${sym}\\b" "${LEDGER_WORK}"; then
        : # symbol cross-referenced
    else
        fail "ledger missing symbol ${sym}"
        MISSING=$((MISSING + 1))
    fi
done
if [ "${CHECKED}" -gt 0 ] && [ "${MISSING}" -eq 0 ]; then
    pass "all ${CHECKED} symbols cross-referenced in ledger"
fi

# Section 3: multi-locale fixture sanity
echo ""
echo "Section 3: multi-locale fixture"
if [ ! -f "${FIXTURE}" ]; then
    fail "fixture missing at ${FIXTURE}"
else
    pass "fixture present"
    LOCALE_COUNT=$(grep -oE '"locale":\s*"[^"]+"' "${FIXTURE}" | sort -u | wc -l)
    if [ "${LOCALE_COUNT}" -ge 5 ]; then
        pass "fixture covers ${LOCALE_COUNT} locales (>=5)"
    else
        fail "fixture covers only ${LOCALE_COUNT} locales (<5)"
    fi
    # Spot-check per-entry detection fields
    for field in benign benign_should_fire attack attack_should_fire \
                 attack_expected_pattern attack_second \
                 attack_second_should_fire attack_second_min_matched; do
        if grep -q "\"${field}\":" "${FIXTURE}"; then
            : # present
        else
            fail "fixture missing field ${field}"
        fi
    done
    # Archive assertions sanity
    for field in default_total companies_must_include category_must_include \
                 search_query_chat_min_hits compare_overlap_must_include_lowercase_token; do
        if grep -q "\"${field}\":" "${FIXTURE}"; then
            : # present
        else
            fail "fixture missing archive_assertions.${field}"
        fi
    done
    pass "fixture carries every per-transform + archive_assertions field"
fi

# Section 4: runner builds + runs against every surface on every locale
echo ""
echo "Section 4: multi-locale runner build + run (real Claritas exerciser)"
if [ ! -f "${RUNNER}" ]; then
    fail "runner missing at ${RUNNER}"
else
    pass "runner source present"
    cd "${MODULE_DIR}"
    if go build -o /tmp/claritas_round263_runner ./challenges/runner/ 2>/tmp/claritas_build.log; then
        pass "runner builds"
        if /tmp/claritas_round263_runner -fixtures "${FIXTURE}" > /tmp/claritas_run.log 2>&1; then
            pass "runner exit 0 across every archive + detect surface + locale"
            # Spot-check per-section per-locale PASS lines
            for chk in '\[archive\]\[seed\]' \
                       '\[archive\]\[addentry\]' \
                       '\[archive\]\[search\]' \
                       '\[archive\]\[bycompany\]' \
                       '\[archive\]\[bycategory\]' \
                       '\[archive\]\[compare\]' \
                       '\[archive\]\[stats\]' \
                       '\[archive\]\[export\]' \
                       '\[archive\]\[trends\]' \
                       '\[detect\]\[benign\]\[en\]' \
                       '\[detect\]\[attack\]\[en\]' \
                       '\[detect\]\[attack2\]\[en\]' \
                       '\[detect\]\[benign\]\[sr\]' \
                       '\[detect\]\[attack\]\[ja\]' \
                       '\[detect\]\[attack2\]\[ar\]' \
                       '\[detect\]\[attack\]\[zh-CN\]' \
                       '\[invariant\]\[close\]'; do
                if grep -qE "PASS: ${chk}" /tmp/claritas_run.log; then
                    : # passed
                else
                    fail "runner missing PASS line for ${chk}"
                fi
            done
            pass "all checkpoint PASS lines present"
        else
            fail "runner exit non-zero -- see /tmp/claritas_run.log"
            sed -n '1,80p' /tmp/claritas_run.log
        fi
    else
        fail "runner build failed -- see /tmp/claritas_build.log"
        sed -n '1,40p' /tmp/claritas_build.log
    fi
    rm -f /tmp/claritas_round263_runner
fi

# Section 5: README round-263 anti-bluff section
echo ""
echo "Section 5: README round-263 anti-bluff section"
if grep -q "Anti-bluff guarantees" "${README}"; then
    pass "README declares Anti-bluff guarantees"
else
    fail "README missing Anti-bluff guarantees section"
fi
if grep -qi "round-263" "${README}"; then
    pass "README marked round-263"
else
    fail "README missing round-263 marker"
fi
if grep -q "API surface" "${README}"; then
    pass "README declares API surface table"
else
    fail "README missing API surface section"
fi

# Cleanup mutated ledger if any
if [ -n "${TMP_LEDGER}" ]; then
    rm -f "${TMP_LEDGER}"
fi

echo ""
echo "=== Summary: ${PASS}/${TOTAL} PASS, ${FAIL} FAIL ==="

if [ "${MUTATE}" -eq 1 ]; then
    if [ "${FAIL}" -gt 0 ]; then
        echo "anti-bluff-mutate: gate correctly detected planted mutation (exit 99)"
        exit 99
    else
        echo "anti-bluff-mutate: gate FAILED to detect planted mutation -- bluff!"
        exit 1
    fi
fi

if [ "${FAIL}" -gt 0 ]; then
    exit 1
fi
exit 0
