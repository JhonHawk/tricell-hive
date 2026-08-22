#!/usr/bin/env bash
# rule-context.sh — PreToolUse, ADVISORY ONLY. Never blocks; always exits 0.
#
# Names the situational rule that applies at the moment it starts applying:
# when a file of a given kind is about to be written, or a tool with its own
# routing policy is about to run.
#
# WHY THIS EXISTS
# Situational rules reach each harness by a different channel, and two of the
# four have no channel at all for the main thread:
#   Claude Code  path-scoped rules load natively via `paths:`
#   opencode     the opencode-rules plugin loads them by glob
#   Codex        no equivalent — nothing loads them
#   Grok         ignores `paths:`, and path-scoped rules are not symlinked
# So on Codex and Grok the main thread has been writing React or Terraform with
# none of those rules present. Subagents get the policy inlined in their own
# prompt; the main thread had nothing. This closes that, and stays useful on
# the other two as a reminder fired with the concrete file in hand.
#
# DELIBERATELY SEPARATE FROM bash-policy.sh
# That hook DENIES (exit 2 / decision JSON) — it is a gate. This one only
# advises. Coupling them would tie a security gate's reliability to extension
# mapping it does not need. Both may match the same tool; that is fine.
#
# Emits the rule NAME, never a path: the file lives under a different root in
# each harness (~/.claude/rules, ~/.agents/skills/language-rules/references,
# the opencode plugin), and a wrong absolute path sends the model hunting.
#
# One reminder per rule per session — a marker file per (session, rule) keeps
# an edit-heavy session from paying the same pointer on every write.

set -uo pipefail

input=$(cat)

session_id=$(printf '%s' "$input" | jq -r '.session_id // .sessionId // empty' 2>/dev/null)
[ -n "$session_id" ] || exit 0

# Dual-runtime: Claude Code (tool_name/tool_input) and Grok Build
# (toolName/toolInput). Codex sends the same shape as Claude Code.
tool_name=$(printf '%s' "$input" | jq -r '.tool_name // .toolName // empty' 2>/dev/null)
[ -n "$tool_name" ] || exit 0

rules=""

add_rule() {
  case " ${rules} " in
    *" $1 "*) ;;            # already queued this call
    *) rules="${rules} $1" ;;
  esac
}

map_path() {
    local path="$1" base
    base=$(basename "$path")

    case "$path" in
      *.tsx|*.jsx) add_rule "typescript-standards.md"; add_rule "react-nextjs.md" ;;
      *.ts|*.js|*.mjs|*.cjs) add_rule "typescript-standards.md" ;;
      *.py) add_rule "python-standards.md" ;;
      *.java|*.kt|*.kts) add_rule "java-kotlin.md" ;;
      *.sql) add_rule "sql-migrations.md" ;;
      *.sh|*.bash) add_rule "shell-standards.md" ;;
      *.tf|*.tfvars) add_rule "iac-devops.md"; add_rule "devops-principles.md" ;;
      *.css|*.scss) add_rule "tailwind.md" ;;
    esac

    # Name-based signals that the extension alone cannot carry.
    case "$base" in
      Dockerfile|Dockerfile.*|docker-compose*.yml|docker-compose*.yaml)
        add_rule "iac-devops.md"; add_rule "devops-principles.md" ;;
      schema.prisma) add_rule "sql-migrations.md" ;;
    esac
    case "$path" in
      */.github/workflows/*) add_rule "iac-devops.md"; add_rule "devops-principles.md" ;;
      *.component.ts|*.service.ts|*.module.ts)
        # Angular and NestJS share these suffixes; the project decides which.
        add_rule "angular-patterns.md or nestjs-patterns.md (whichever the project uses)" ;;
    esac
}

case "$tool_name" in
  Write|Edit|MultiEdit|search_replace|apply_patch)
    path=$(printf '%s' "$input" | jq -r '
      .tool_input.file_path // .toolInput.file_path //
      .tool_input.path // .toolInput.path //
      .tool_input.filePath // .toolInput.filePath // empty' 2>/dev/null)
    [ -n "$path" ] || exit 0
    map_path "$path"
    ;;

  Bash|run_terminal_command|shell)
    command=$(printf '%s' "$input" | jq -r '
      .tool_input.command // .toolInput.command //
      (.tool_input.command? // empty | if type=="array" then join(" ") else . end) // empty' 2>/dev/null)
    if [ -z "$command" ]; then
      # Codex sends shell argv as an array.
      command=$(printf '%s' "$input" | jq -r '
        (.tool_input.command // .toolInput.command // [])
        | if type=="array" then join(" ") else tostring end' 2>/dev/null)
    fi
    [ -n "$command" ] || exit 0

    case "$command" in
      *agent-browser*) add_rule "browser-automation-reference.md" ;;
    esac
    case "$command" in
    esac
    # Codex edits through `apply_patch` inside a shell call, so the file kind is
    # in the patch header rather than in a file_path field. Without this the hook
    # would be inert on the harness that most needs it.
    case "$command" in
      *apply_patch*)
        while IFS= read -r patched; do
          [ -n "$patched" ] && map_path "$patched"
        done < <(printf '%s' "$command" \
          | grep -oE '^\*\*\* (Add|Update|Delete) File: .+$' 2>/dev/null \
          | sed -E 's/^\*\*\* (Add|Update|Delete) File: //')
        ;;
    esac
    ;;

  *) exit 0 ;;
esac

[ -n "${rules// /}" ] || exit 0

# Report each rule at most once per session.
fresh=""
for rule in $rules; do
  slug=$(printf '%s' "$rule" | tr -c 'a-zA-Z0-9' '-')
  marker="${TMPDIR:-/tmp}/claude-rule-context-${session_id}-${slug}"
  [ -f "$marker" ] && continue
  : > "$marker" 2>/dev/null
  fresh="${fresh}${fresh:+, }${rule}"
done

[ -n "$fresh" ] || exit 0

ctx="Situational rules that apply here: ${fresh}. Read them before proceeding unless already loaded this session — where they are not always-on they live in the language-rules skill references (every harness for browser-automation-reference.md) or load by glob (Claude Code, opencode). Advisory: this never blocks."

jq -n --arg ctx "$ctx" '{
  hookSpecificOutput: {
    hookEventName: "PreToolUse",
    additionalContext: $ctx
  }
}'

exit 0
