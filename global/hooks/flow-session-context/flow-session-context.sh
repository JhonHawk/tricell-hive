#!/usr/bin/env bash
# flow-session-context.sh — SessionStart (startup|clear) hook, NON-BLOCKING.
#
# Cross-harness session-start context for the flow pack v2. Merges two
# independent, self-gating sections into ONE injected context string:
#
#   1. Flow protocol — ONLY inside a flow workspace (ledger _support/PROJECT.md
#      at or above cwd). A static process-chain map so the model can OFFER the
#      matching /flow-* stage when user intent matches. Never forces a stage;
#      flow skills stay user-gated (global CLAUDE.md > Skill Auto-invocation).
#
#   2. Git hygiene — in ANY git repo (not flow-gated). Deterministic backstop
#      for the session-close ritual (git-workflow.md > Session close): most
#      closes are silent, so the ceremony runs at the next fresh seam instead,
#      injecting pending-hygiene FACTS (locally merged branches, [gone]
#      upstreams). It injects state, never routing; the always-on rule owns
#      what to do.
#
# Either section may be empty; both empty -> silent (exit 0, no injection).
# Local git queries only (no fetch, no network). Advisory only — NEVER blocks,
# exits 0 always.

set -uo pipefail

input=$(cat)
source_evt=$(printf '%s' "$input" | jq -r '.source // empty' 2>/dev/null)
cwd=$(printf '%s' "$input" | jq -r '.cwd // empty' 2>/dev/null)
[ -n "$cwd" ] || cwd="$PWD"

# Only fresh contexts: skip resume (context already has it) and compact (mid-task).
case "$source_evt" in startup|clear) ;; *) exit 0 ;; esac

# ─── Section 1: Flow protocol (flow workspaces only) ─────────────────────────
dir="$cwd"; ledger=""
while [ -n "$dir" ] && [ "$dir" != "/" ]; do
  if [ -f "$dir/_support/PROJECT.md" ]; then ledger="$dir/_support/PROJECT.md"; break; fi
  dir=$(dirname "$dir")
done

flow_section=""
if [ -n "$ledger" ]; then
  flow_section='<flow-process-protocol>
Flow pack v2 — process chain for this workspace. Suggest the matching stage when user intent matches; never force it:
- BRAINSTORM: iterating a feature/business idea -> ask ONCE via a yes/no question: "¿Iniciamos modo brainstorming?" — yes: invoke /flow-brainstorming; no: do NOT re-ask this session (manual /flow-brainstorming only). Contested questions route to /adversarial-research. Output: a business decision, not a spec.
- SPEC: a decided idea needs formalization -> offer /flow-specs.
- PLAN: planning intent ALWAYS uses the harness'"'"'s native plan mechanism; flow-plan captures/adopts the approved plan — never a parallel planning ceremony. ONE plan per unit of work (a mock is a work unit like any fullstack build).
- EXECUTE: a captured plan exists -> offer /flow-build to execute it. Resuming a unit of work from the ledger ("continuemos con X", a PX/part) -> make that offer explicitly ONCE per session, direct route named as the alternative; a no is sticky for the session. A hand-run of the protocol after a no still owes the plan'"'"'s per-task gates: the in-vivo walk'"'"'s executor and close-time substitution reporting (flow-core/references/plan-format.md).
- After a deploy/promotion (git conventions own the flow): offer the in-vivo QA walk (flow-core/references/promotion-playbook.md).
- New greenfield project -> offer /flow-start. Pre-pack project entering the flow -> offer /flow-adopt. Workspace health concerns -> /flow-workspace.
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
    git_section="Pending git hygiene from a previous session (deterministic session-close backstop): ${facts}. The ritual that owns this is git-workflow.md > Session close (standing-authorized: prune confirmed-merged branches, report divergences; unmerged branches are decisions, not noise)."
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

jq -n --arg ctx "$ctx" '{
  hookSpecificOutput: {
    hookEventName: "SessionStart",
    additionalContext: $ctx
  }
}'

exit 0
