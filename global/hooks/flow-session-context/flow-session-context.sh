#!/usr/bin/env bash
# flow-session-context.sh — SessionStart (startup|clear|compact) hook, NON-BLOCKING.
#
# Cross-harness session-start context for the flow pack v2. Two independent,
# self-gating sections on a FRESH context (startup|clear), plus a third,
# narrower payload after a compaction:
#
#   1. Flow protocol — ONLY inside a flow workspace (ledger _support/PROJECT.md
#      at or above cwd). A static process-chain map so the model can OFFER the
#      matching playbook when user intent matches. Never forces a stage; flow
#      skills stay user-gated (global CLAUDE.md > Skill Auto-invocation).
#
#   2. Git hygiene — in ANY git repo (not flow-gated). Deterministic backstop
#      for the end-of-work hygiene ritual (git-mechanics.md > End-of-work hygiene): most
#      closes are silent, so the ceremony runs at the next fresh seam instead,
#      injecting pending-hygiene FACTS (locally merged branches, [gone]
#      upstreams). It injects state, never routing; the always-on rule owns
#      what to do.
#
#   3. Post-compaction recovery (source == compact) — the process map is the
#      ONLY trigger the SPEC step has, and a compaction can drop it from
#      context. Re-injected condensed: recovery, not the full startup load.
#      Git hygiene stays OUT (a end-of-work-hygiene backstop is noise mid-task), and
#      flow-context's per-session markers are cleared so the pending-plan state
#      re-emits on the next prompt instead of being suppressed by a marker that
#      predates the compaction.
#
# Either section may be empty; both empty -> silent (exit 0, no injection).
# Local git queries only (no fetch, no network). Advisory only — NEVER blocks,
# exits 0 always.

set -uo pipefail

input=$(cat)
source_evt=$(printf '%s' "$input" | jq -r '.source // empty' 2>/dev/null)
cwd=$(printf '%s' "$input" | jq -r '.cwd // empty' 2>/dev/null)
session_id=$(printf '%s' "$input" | jq -r '.session_id // empty' 2>/dev/null)
[ -n "$cwd" ] || cwd="$PWD"

# Fresh contexts and compaction. `resume` is skipped: the context already has it.
case "$source_evt" in startup|clear|compact) ;; *) exit 0 ;; esac

emit() {
  jq -n --arg ctx "$1" '{
    hookSpecificOutput: {
      hookEventName: "SessionStart",
      additionalContext: $ctx
    }
  }'
  exit 0
}

# ─── Flow workspace? Walk up for the ledger ──────────────────────────────────
dir="$cwd"; ledger=""
while [ -n "$dir" ] && [ "$dir" != "/" ]; do
  if [ -f "$dir/_support/PROJECT.md" ]; then ledger="$dir/_support/PROJECT.md"; break; fi
  dir=$(dirname "$dir")
done

# ─── Post-compaction: condensed recovery, then done ──────────────────────────
if [ "$source_evt" = "compact" ]; then
  [ -n "$ledger" ] || exit 0

  # Clear flow-context's per-session markers: their "already injected this
  # session" assumption is void once the context was compacted away.
  if [ -n "$session_id" ]; then
    rm -f "${TMPDIR:-/tmp}/claude-flow-context-phase-${session_id}" \
          "${TMPDIR:-/tmp}/claude-flow-context-planning-${session_id}" 2>/dev/null || true
  fi

  emit '<flow-process-protocol source="post-compaction">
Flow workspace — context was compacted; the process map is re-injected because the pre-compaction copy may not have survived. Same rules, condensed. Never force ceremony onto a small change:
- IDEA: exploring whether something is worth doing -> converge on proceed/discard/defer (quality/critical-thinking.md).
- SPEC: a decided idea needs formalization -> flow-core/references/spec-writing-playbook.md. Its business gate is mandatory BEFORE delivery is derived.
- PLAN: planning intent uses the Hive portable /flow-plan command. ONE plan per unit of work (flow-core/references/plan-format.md); native harness planning remains optional and never grants Hive authorization.
- EXECUTE: an approved plan with pending tasks -> offer /flow-build ONCE, naming the direct route as the alternative; never invoke it uninvited.
Pending-plan state re-emits on the next prompt (flow-context markers were cleared). Verify any Status against git before trusting it.
</flow-process-protocol>'
fi

# ─── Section 1: Flow protocol (flow workspaces only) ─────────────────────────
flow_section=""
if [ -n "$ledger" ]; then
  flow_section='<flow-process-protocol>
Flow workspace — the process knowledge for this project lives in flow-core references, not in commands. Work the matching playbook conversationally when intent matches; never force ceremony onto a small change:
- IDEA: exploring whether something is worth doing -> converge on proceed/discard/defer with one light decision note (quality/critical-thinking.md). Contested, load-bearing questions -> offer /adversarial-research.
- SPEC: a decided idea needs formalization -> flow-core/references/spec-writing-playbook.md (its business gate is mandatory before delivery is derived).
- PLAN: planning intent uses Hive'"'"'s portable /flow-plan command — native harness planning is optional and never grants Hive authorization. ONE plan per unit of work. Format and Preflight: flow-core/references/plan-format.md.
- EXECUTE: an approved plan with pending tasks exists -> offer /flow-build to execute or resume it, ONCE per session, with the direct route named as the alternative; a no is sticky for the session. A hand-run still owes the plan'"'"'s per-task gates.
- After a deploy/promotion (git conventions own the flow): offer the in-vivo QA walk (flow-core/references/promotion-playbook.md). Epic close / tech-debt baseline -> flow-core/references/audit-playbook.md.
- Greenfield bootstrap -> flow-core/references/bootstrap-playbook.md. Pre-pack project entering the convention -> flow-core/references/migration-playbook.md. Workspace health concerns -> flow-core/references/workspace-hygiene-playbook.md (dispatch workspace-custodian).
</flow-process-protocol>'
fi

# ─── Section 2: Git hygiene (any git repo) ───────────────────────────────────
git_section=""
if [ -d "$cwd" ] && git -C "$cwd" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  # Long-lived branches (environment branches + trunks) are never prune candidates.
  protected='^(development|qa|production|main|master)$'

  # Integration target: development if it exists, else origin/HEAD, else main/master.
  target=""
  if git -C "$cwd" show-ref --verify --quiet refs/heads/development; then
    target="development"
  else
    target=$(git -C "$cwd" symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed 's|^origin/||')
    if [ -z "$target" ]; then
      for b in main master; do
        if git -C "$cwd" show-ref --verify --quiet "refs/heads/$b"; then target="$b"; break; fi
      done
    fi
  fi

  merged=""
  if [ -n "$target" ]; then
    merged=$(git -C "$cwd" branch --format='%(refname:short)' --merged "$target" 2>/dev/null \
      | grep -Ev "$protected" | grep -Fxv "$target" | head -8)
  fi

  gone=$(git -C "$cwd" branch --format='%(refname:short) %(upstream:track)' 2>/dev/null \
    | awk '$2 == "[gone]" {print $1}' | grep -Ev "$protected" | head -8)

  if [ -n "$merged" ] || [ -n "$gone" ]; then
    facts=""
    if [ -n "$merged" ]; then
      facts="local branches fully merged into ${target}: $(printf '%s' "$merged" | tr '\n' ' ' | sed 's/ *$//')"
    fi
    if [ -n "$gone" ]; then
      [ -n "$facts" ] && facts="${facts}; "
      facts="${facts}branches whose upstream is gone: $(printf '%s' "$gone" | tr '\n' ' ' | sed 's/ *$//')"
    fi
    git_section="Pending git hygiene from a previous session (deterministic end-of-work-hygiene backstop): ${facts}. The ritual that owns this is git-mechanics.md > End-of-work hygiene (standing-authorized: prune confirmed-merged branches, report divergences; unmerged branches are decisions, not noise)."
  fi
fi

# ─── Emit: join whichever sections are non-empty; both empty -> silent ───────
ctx=""
[ -n "$flow_section" ] && ctx="$flow_section"
if [ -n "$git_section" ]; then
  [ -n "$ctx" ] && ctx="${ctx}

"
  ctx="${ctx}${git_section}"
fi

[ -n "$ctx" ] || exit 0

emit "$ctx"
