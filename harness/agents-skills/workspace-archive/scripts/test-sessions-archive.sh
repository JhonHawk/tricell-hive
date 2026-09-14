#!/usr/bin/env bash
# Checks for sessions-archive.sh. Fixture: a throwaway git repo holding a ledger
# and an `x-specs/sessions/` tree whose folders get commits with controlled
# dates (GIT_AUTHOR_DATE / GIT_COMMITTER_DATE) so ages are deterministic.
# Run from anywhere: bash global/skills/workspace-archive/scripts/test-sessions-archive.sh
set -uo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
SCRIPT="$ROOT/sessions-archive.sh"
fail=0
N=15

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

ok()   { printf 'ok   %s\n' "$1"; }
bad()  { printf 'FAIL %s\n' "$1"; fail=$((fail + 1)); }
check() { # check <name> <condition...>
  local name="$1"; shift
  if "$@"; then ok "$name"; else bad "$name"; fi
}
not() { ! "$@"; }

days_ago() { # days_ago <n> → YYYY-MM-DD
  local ts=$(( $(date +%s) - $1 * 86400 ))
  date -r "$ts" +%Y-%m-%d 2>/dev/null || date -d "@$ts" +%Y-%m-%d
}

# ─── fixture ─────────────────────────────────────────────────────────────────
WS="$WORK/acme"
SPECS="$WS/x-specs"
SESS="$SPECS/sessions"
mkdir -p "$WS/_support" "$SESS"
printf '# acme — Project Ledger\n\n| | |\n|---|---|\n| Tracker | none |\n' > "$WS/_support/PROJECT.md"
git -C "$WS" init -q
git -C "$WS" config user.email test@example.com
git -C "$WS" config user.name test
git -C "$WS" config commit.gpgsign false

plan_body() { # plan_body <status line value>
  cat <<EOF
# Fixture — Plan

Status: $1

<!-- hive-plan:contract:start -->
## Contract

Goal: fixture
<!-- hive-plan:contract:end -->

## Execution
EOF
}

add_session() { # add_session <folder> <age days> <status|-> [extra file]
  local folder="$1" age="$2" status="$3" slug d
  slug="${folder:11}"
  mkdir -p "$SESS/$folder"
  if [ "$status" = "-" ]; then
    printf '# notes\n' > "$SESS/$folder/$slug-findings.md"
  else
    plan_body "$status" > "$SESS/$folder/$slug-plan.md"
  fi
  d=$(days_ago "$age")
  git -C "$WS" add -- "$SESS/$folder"
  GIT_AUTHOR_DATE="${d}T12:00:00" GIT_COMMITTER_DATE="${d}T12:00:00" \
    git -C "$WS" commit -q -m "fixture: $folder"
}

add_session 2026-07-01-old-verified   40 verified
add_session 2026-09-10-new-verified    3 verified
add_session 2026-07-05-old-planned    35 planned
add_session 2026-07-06-old-building   33 building
add_session 2026-07-08-legacy-qa      30 "verified in QA"
add_session 2026-07-09-legacy-phase   30 "done (Fase 1)"
add_session 2026-07-10-legacy-wip     30 "In Progress"
add_session 2026-07-11-noplan-done    30 -
add_session 2026-07-12-noplan-open    30 -
add_session 2026-07-13-old-draft      30 draft
add_session 2026-09-11-new-planned     2 planned
# non-candidates
mkdir -p "$SESS/archived/2026-01-01-already" "$SESS/previously"
printf 'x\n' > "$SESS/archived/2026-01-01-already/already-plan.md"
printf 'x\n' > "$SESS/previously/loose.md"
printf 'loose\n' > "$SESS/loose-file.md"
# a plan whose Status sits INSIDE the contract block (must be reported, never touched)
mkdir -p "$SESS/2026-07-14-legacy-inside"
cat > "$SESS/2026-07-14-legacy-inside/legacy-inside-plan.md" <<'EOF'
# Bad — Plan

<!-- hive-plan:contract:start -->
Status: done
<!-- hive-plan:contract:end -->
EOF

cat > "$SESS/README.md" <<'EOF'
# Sesiones

| Fecha | Slug | Estado | Implementa | Docs |
|---|---|---|---|---|
| 2026-07-01 | [`old-verified`](2026-07-01-old-verified/) | ✅ entregado | E01 | [plan](./2026-07-01-old-verified/old-verified-plan.md) |
| 2026-07-11 | [`noplan-done`](2026-07-11-noplan-done/) | Finalizada | E02 | — |
| 2026-07-12 | [`noplan-open`](2026-07-12-noplan-open/) | en curso | E03 | — |
| 2026-07-13 | [`old-draft`](2026-07-13-old-draft/) | borrador | — | — |
EOF
git -C "$WS" add -A
GIT_AUTHOR_DATE="$(days_ago 1)T12:00:00" GIT_COMMITTER_DATE="$(days_ago 1)T12:00:00" \
  git -C "$WS" commit -q -m "fixture: index + extras"

run() { # run <cwd> <args...> → stdout in $OUT, rc in $RC
  local dir="$1"; shift
  OUT=$(cd "$dir" && bash "$SCRIPT" "$@" 2>"$WORK/err"); RC=$?
}
jqq() { printf '%s' "$OUT" | jq -r "$1"; }
in_bucket() { # in_bucket <bucket> <folder>
  [ "$(jqq ".$1 | map(.folder) | index(\"$2\") != null")" = "true" ]
}

# ─── resolution: ledger walk-up from a nested cwd, and from the specs repo ───
mkdir -p "$WS/app/src"
run "$WS/app/src" --json
check "resolve-ledger-walkup" [ "$RC" -eq 0 ] && check "resolve-ledger-path" [ "$(jqq .sessions)" = "$(cd "$SESS" && pwd -P)" ]
run "$SPECS" --json
check "resolve-cwd-specs-repo" [ "$(jqq .sessions)" = "$(cd "$SESS" && pwd -P)" ]

# ─── (i) nothing resolvable → exit 2 ─────────────────────────────────────────
mkdir -p "$WORK/plain/deep"
run "$WORK/plain/deep" --json
check "i-no-ledger-exit-2" [ "$RC" -eq 2 ]
check "i-no-ledger-message" grep -q 'no se resolvió sessions/' "$WORK/err"
run "$WORK/plain" --days abc
check "usage-bad-days-exit-2" [ "$RC" -eq 2 ]

# ─── archive mode, dry-run ───────────────────────────────────────────────────
run "$SPECS" --json --days "$N"
check "dry-run-rc-0" [ "$RC" -eq 0 ]
check "dry-run-mode" [ "$(jqq .mode)" = "archive" ]
check "dry-run-apply-false" [ "$(jqq .apply)" = "false" ]
check "a-verified-old-archivable"      in_bucket archivable 2026-07-01-old-verified
check "b-verified-new-recent"          [ "$(jqq '.recent.folders | index("2026-09-10-new-verified") != null')" = "true" ]
check "c-planned-old-stalled"          in_bucket stalled 2026-07-05-old-planned
check "c-building-old-stalled"         in_bucket stalled 2026-07-06-old-building
check "d-legacy-review"                in_bucket review 2026-07-08-legacy-qa
check "d-legacy-status-raw"            [ "$(jqq '.review[] | select(.folder == "2026-07-08-legacy-qa") | .status_raw')" = "verified in QA" ]
check "e-noplan-index-signal-archivable" in_bucket archivable 2026-07-11-noplan-done
check "e-noplan-index-state-cell"      [ "$(jqq '.archivable[] | select(.folder == "2026-07-11-noplan-done") | .index_state')" = "Finalizada" ]
check "e-noplan-no-signal-review"      in_bucket review 2026-07-12-noplan-open
check "draft-old-review"               in_bucket review 2026-07-13-old-draft
check "active-new-planned"             [ "$(jqq '.active.folders | index("2026-09-11-new-planned") != null')" = "true" ]
check "j-archived-not-candidate"       [ "$(jqq '[.. | .folder? // empty] | index("2026-01-01-already") == null')" = "true" ]
check "j-previously-not-candidate"     [ "$(jqq '[.. | .folder? // empty] | map(select(. == "previously" or . == "archived")) | length')" = "0" ]
check "age-source-git"                 [ "$(jqq '.archivable[] | select(.folder == "2026-07-01-old-verified") | .age_source')" = "git" ]
check "age-days-git"                   [ "$(jqq '.archivable[] | select(.folder == "2026-07-01-old-verified") | .age_days')" = "40" ]
check "dry-run-nothing-moved"          [ ! -d "$SESS/archived/2026-07-01-old-verified" ]
check "dry-run-tree-clean"             [ -z "$(git -C "$WS" status --porcelain)" ]

# (g) counts: buckets partition the candidates, human output agrees with JSON
total=$(jqq .total)
sum=$(jqq '(.archivable | length) + (.stalled | length) + (.review | length) + .recent.count + .active.count')
check "g-json-total-is-partition" [ "$total" = "$sum" ]
check "g-json-total-12" [ "$total" = "12" ]
arch_n=$(jqq '.archivable | length')
run "$SPECS" --days "$N"
check "g-human-rc-0" [ "$RC" -eq 0 ]
human_arch_rows=$(printf '%s' "$OUT" | awk '/^### 1\./{f=1;next} /^### 2\./{f=0} f && /^\| `/' | wc -l | tr -d ' ')
check "g-human-archivable-rows-match-json" [ "$human_arch_rows" = "$arch_n" ]
check "g-human-mentions-recent-count" grep -q "Cerradas hace ≤ $N días: 1\." <<<"$OUT"
check "g-human-commit-line" grep -q "chore(sessions): archive 2 closed session(s) older than $N days" <<<"$OUT"
check "g-human-spanish-header" grep -q '^## Sesiones' <<<"$OUT"

# threshold moves the line: with --days 60 nothing is archivable, old ones become recent/active
run "$SPECS" --json --days 60
check "days-60-no-archivable" [ "$(jqq '.archivable | length')" = "0" ]
check "days-60-old-planned-active" [ "$(jqq '.active.folders | index("2026-07-05-old-planned") != null')" = "true" ]

# (mtime fallback) a folder never committed → age from the newest file mtime, flagged
mkdir -p "$SESS/2026-06-01-uncommitted"
plan_body verified > "$SESS/2026-06-01-uncommitted/uncommitted-plan.md"
touch -t "$(date -r $(( $(date +%s) - 30 * 86400 )) +%Y%m%d%H%M 2>/dev/null || date -d "@$(( $(date +%s) - 30 * 86400 ))" +%Y%m%d%H%M)" "$SESS/2026-06-01-uncommitted/uncommitted-plan.md"
run "$SPECS" --json --days "$N"
check "mtime-fallback-flagged" [ "$(jqq '.archivable[] | select(.folder == "2026-06-01-uncommitted") | .age_source')" = "mtime" ]
check "mtime-fallback-age-30" [ "$(jqq '.archivable[] | select(.folder == "2026-06-01-uncommitted") | .age_days')" = "30" ]
run "$SPECS" --days "$N"
check "mtime-fallback-human-flag" grep -q "\`2026-06-01-uncommitted\`.*30 d (mtime)" <<<"$OUT"
rm -rf "$SESS/2026-06-01-uncommitted"

# ─── (h)(d) normalize mode ───────────────────────────────────────────────────
run "$SPECS" --json --normalize
check "norm-rc-0" [ "$RC" -eq 0 ]
check "norm-mode" [ "$(jqq .mode)" = "normalize" ]
check "d-norm-qa-to-verified"   [ "$(jqq '.resolved[] | select(.folder == "2026-07-08-legacy-qa") | .normalize_to')" = "verified" ]
check "norm-wip-to-building"    [ "$(jqq '.resolved[] | select(.folder == "2026-07-10-legacy-wip") | .normalize_to')" = "building" ]
check "h-norm-phase-unresolved" in_bucket unresolved 2026-07-09-legacy-phase
check "norm-inside-contract-reported" in_bucket in_contract 2026-07-14-legacy-inside
check "norm-inside-not-resolved" [ "$(jqq '.resolved | map(.folder) | index("2026-07-14-legacy-inside") == null')" = "true" ]
check "norm-dry-run-untouched" grep -q '^Status: verified in QA$' "$SESS/2026-07-08-legacy-qa/legacy-qa-plan.md"
check "norm-dry-run-tree-clean" [ -z "$(git -C "$WS" status --porcelain)" ]
run "$SPECS" --normalize
check "norm-human-has-resolved-row" grep -q "\`2026-07-08-legacy-qa\`.*verified in QA.*verified" <<<"$OUT"
check "norm-human-has-unresolved-row" grep -q "\`2026-07-09-legacy-phase\`" <<<"$OUT"

before=$(md5 -q "$SESS/2026-07-09-legacy-phase/legacy-phase-plan.md" 2>/dev/null || md5sum "$SESS/2026-07-09-legacy-phase/legacy-phase-plan.md" | cut -d' ' -f1)
inside_before=$(cat "$SESS/2026-07-14-legacy-inside/legacy-inside-plan.md")
run "$SPECS" --json --normalize --apply
check "norm-apply-rc-0" [ "$RC" -eq 0 ]
check "norm-apply-count-2" [ "$(jqq '.applied.normalized | length')" = "2" ]
qa_plan="$SESS/2026-07-08-legacy-qa/legacy-qa-plan.md"
check "d-norm-apply-status-line"   grep -q '^Status: verified$' "$qa_plan"
check "d-norm-apply-legacy-comment" grep -q "^<!-- legacy status (normalized $(date +%Y-%m-%d)): verified in QA -->$" "$qa_plan"
check "d-norm-apply-comment-follows-status" [ "$(grep -n -m1 '^Status:' "$qa_plan" | cut -d: -f1)" = "$(( $(grep -n -m1 'legacy status' "$qa_plan" | cut -d: -f1) - 1 ))" ]
check "d-norm-apply-rest-intact" [ "$(grep -c . "$qa_plan")" = "$(( $(plan_body x | grep -c .) + 1 ))" ]
check "norm-apply-wip-building" grep -q '^Status: building$' "$SESS/2026-07-10-legacy-wip/legacy-wip-plan.md"
after=$(md5 -q "$SESS/2026-07-09-legacy-phase/legacy-phase-plan.md" 2>/dev/null || md5sum "$SESS/2026-07-09-legacy-phase/legacy-phase-plan.md" | cut -d' ' -f1)
check "h-norm-apply-phase-untouched" [ "$before" = "$after" ]
check "norm-apply-inside-untouched" [ "$inside_before" = "$(cat "$SESS/2026-07-14-legacy-inside/legacy-inside-plan.md")" ]
check "norm-apply-staged" [ "$(git -C "$WS" diff --cached --name-only | grep -c -- '-plan.md$')" = "2" ]
check "norm-apply-no-unstaged" [ -z "$(git -C "$WS" diff --name-only)" ]
check "norm-apply-no-commit" [ "$(git -C "$WS" log --oneline | wc -l | tr -d ' ')" = "12" ]
norm_commit=$(jqq .commit_message)
check "norm-apply-commit-line" [ "$norm_commit" = "chore(sessions): normalize 2 legacy plan status line(s)" ]
# normalized plans now read as canonical → second normalize run finds nothing to resolve
run "$SPECS" --json --normalize
check "norm-idempotent" [ "$(jqq '.resolved | length')" = "0" ]
git -C "$WS" commit -q -m "$norm_commit"

# ─── (f) archive --apply: moves, rewrites index links, stages, idempotent ────
# after normalization the QA plan is `verified`; the normalize commit is skipped when
# dating the folder, so it is still 30 d old → 3 archivables now
run "$SPECS" --json --days "$N"
check "f-pre-apply-3-archivable" [ "$(jqq '.archivable | length')" = "3" ]
check "f-normalize-commit-does-not-reset-age" [ "$(jqq '.archivable[] | select(.folder == "2026-07-08-legacy-qa") | .age_days')" = "30" ]
run "$SPECS" --days "$N" --apply
check "f-apply-rc-0" [ "$RC" -eq 0 ]
check "f-apply-human-says-applied" grep -q 'APLICADO' <<<"$OUT"
check "f-apply-moved-verified" [ -d "$SESS/archived/2026-07-01-old-verified" ] && check "f-apply-src-gone" [ ! -d "$SESS/2026-07-01-old-verified" ]
check "f-apply-moved-noplan" [ -d "$SESS/archived/2026-07-11-noplan-done" ]
check "f-apply-moved-normalized-qa" [ -d "$SESS/archived/2026-07-08-legacy-qa" ]
check "f-apply-stalled-stays" [ -d "$SESS/2026-07-05-old-planned" ]
check "f-apply-review-stays" [ -d "$SESS/2026-07-12-noplan-open" ]
check "f-apply-index-link-rewritten" grep -q '](archived/2026-07-01-old-verified/)' "$SESS/README.md"
check "f-apply-index-dot-link-rewritten" grep -q '](archived/2026-07-01-old-verified/old-verified-plan.md)' "$SESS/README.md"
check "f-apply-index-noplan-rewritten" grep -q '](archived/2026-07-11-noplan-done/)' "$SESS/README.md"
check "f-apply-index-untouched-row" grep -q '](2026-07-12-noplan-open/)' "$SESS/README.md"
check "f-apply-index-no-double-prefix" not grep -q 'archived/archived' "$SESS/README.md"
check "f-apply-no-bak-file" [ ! -f "$SESS/README.md.bak" ]
check "f-apply-all-staged" [ -z "$(git -C "$WS" status --porcelain | grep -v '^[RAMD] ')" ]
check "f-apply-readme-staged" [ -n "$(git -C "$WS" diff --cached --name-only -- "$SESS/README.md")" ]
check "f-apply-renames-staged" [ "$(git -C "$WS" diff --cached --name-status -M | grep -c '^R')" = "3" ]
check "f-apply-no-commit" [ "$(git -C "$WS" log --oneline | wc -l | tr -d ' ')" = "13" ]
check "f-apply-commit-line" grep -q "chore(sessions): archive 3 closed session(s) older than $N days" <<<"$OUT"
run "$SPECS" --json --days "$N" --apply
check "f-idempotent-rc-0" [ "$RC" -eq 0 ]
check "f-idempotent-zero-archivable" [ "$(jqq '.archivable | length')" = "0" ]
check "f-idempotent-zero-moved" [ "$(jqq '.applied.moved | length')" = "0" ]
check "f-idempotent-archived-still-excluded" [ "$(jqq '[.. | .folder? // empty] | index("2026-07-01-old-verified") == null')" = "true" ]

# ─── --sessions explicit + no git → dry-run works, --apply refuses ───────────
mkdir -p "$WORK/nogit/sessions/2026-01-01-thing"
plan_body verified > "$WORK/nogit/sessions/2026-01-01-thing/thing-plan.md"
run "$WORK" --json --sessions "$WORK/nogit/sessions"
check "sessions-flag-no-git-dry-run" [ "$RC" -eq 0 ]
check "sessions-flag-no-git-mtime-recent" [ "$(jqq '.recent.folders | index("2026-01-01-thing") != null')" = "true" ]
run "$WORK" --sessions "$WORK/nogit/sessions" --apply
check "sessions-flag-no-git-apply-exit-2" [ "$RC" -eq 2 ]
run "$WORK" --sessions "$WORK/does-not-exist"
check "sessions-flag-missing-exit-2" [ "$RC" -eq 2 ]

if [ "$fail" -ne 0 ]; then
  printf '\n%d check(s) failed\n' "$fail"
  exit 1
fi
printf '\nall sessions-archive checks passed\n'
