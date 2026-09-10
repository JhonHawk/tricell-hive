#!/usr/bin/env bash
# deploy-global.sh — deploys tricell-hive's global/ (+ the multi-harness layer
# under harness/) to ~/.claude, ~/.codex, ~/.config/opencode, ~/.grok, the PI
# agent directory, and the neutral shared skills owner under ~/.agents/skills.
#
# Faithful port of the procedure documented in ../SKILL.md. Read that file for
# the "why" of each step; this script is the "how".
#
# Safety model:
#   - Defaults to --dry-run. Nothing is written anywhere until --apply is passed.
#   - Under --apply, manifest-confirmed orphans are deleted as part of the run
#     (with the settings.json / hooks.json purge that rides with it): the orphan
#     set is by construction sources the user already removed from the repo, and
#     deploys are user-initiated. --keep-orphans opts out (list-only); the
#     size-based refusal (>20 entries or >=25% of the manifest) stays as the
#     anomaly brake and still requires --force-delete-orphans to override.
#     --delete-orphans is accepted as a deprecated no-op alias.
#   - Targets zsh-invoked bash on macOS + BSD userland: no GNU-only flags, no
#     bash 4+ features (associative arrays, mapfile) — the system /bin/bash on
#     macOS is 3.2.
#
# Usage: see --help.

set -euo pipefail

# set -u catches an UNSET HOME but not an empty one; an empty HOME would
# silently resolve CLAUDE_HOME to "/.claude" below. Refuse outright.
if [[ -z "${HOME:-}" || ! -d "${HOME}" ]]; then
    printf 'ERROR: HOME is unset, empty, or not a directory — refusing to run.\n' >&2
    exit 1
fi

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
# scripts/ -> deploy-global/ -> skills/ -> .claude/ -> repo root
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../../.." && pwd)"
readonly REPO_ROOT
readonly FILTERS_DIR="${SCRIPT_DIR}/../filters"
readonly HOOK_MERGE_FILTER="${FILTERS_DIR}/hook-merge.jq"
readonly HOOK_PURGE_FILTER="${FILTERS_DIR}/hook-purge.jq"

readonly CLAUDE_HOME="${HOME}/.claude"
readonly CODEX_HOME="${HOME}/.codex"
readonly OPENCODE_HOME="${HOME}/.config/opencode"
readonly AGENTS_HOME="${HOME}/.agents"
readonly AGENTS_SKILLS_HOME="${HOME}/.agents/skills"
# Grok relocates its config root via GROK_HOME (documented in its user guide);
# honor it so a relocated install is not silently deployed to ~/.grok.
readonly GROK_HOME="${GROK_HOME:-${HOME}/.grok}"
readonly GROK_RULES_HOME="${GROK_HOME}/rules"
readonly GROK_AGENTS_HOME="${GROK_HOME}/agents"
readonly PI_AGENT_DIR="${PI_CODING_AGENT_DIR:-${HOME}/.pi/agent}"
readonly PI_DEPLOY_HELPER="${REPO_ROOT}/harness/pi/deploy.py"
readonly MANIFEST="${CLAUDE_HOME}/.deploy-manifest"
readonly BACKUP_DIR="${CLAUDE_HOME}/backups"
readonly CODEX_DOC_MAX_BYTES=49152
# Orphan-deletion circuit breaker: refuse rather than delete when the orphan
# set looks like a broken checkout (missing global/harness sources) instead
# of a routine cleanup. See step_detect_orphans.
readonly ORPHAN_MAX_ABS=20
readonly ORPHAN_MAX_PCT=25

# ---------------------------------------------------------------------------
# Globals set by argument parsing
# ---------------------------------------------------------------------------

# APPLY is the single source of truth for dry-run vs. real: APPLY=0 means
# dry-run (the default), APPLY=1 means real. No separate DRY_RUN variable —
# two booleans that must always be kept as exact inverses is how a caller
# accidentally sets --apply and expects --dry-run behavior (or vice versa).
APPLY=0
VERBOSE=0
KEEP_ORPHANS=0
FORCE_DELETE_ORPHANS=0
RUN_CLAUDE=1
RUN_CODEX=1
RUN_OPENCODE=1
RUN_GROK=1
RUN_PI=1

REPORT_LOG=""
TMP_FILES=()

# ---------------------------------------------------------------------------
# Logging — narration to stderr, final structured report to stdout.
# ---------------------------------------------------------------------------

log() { printf '%s\n' "$*" >&2; }
vlog() { [[ "${VERBOSE}" -eq 1 ]] && printf '%s\n' "$*" >&2; return 0; }
die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }
report() { [[ -n "${REPORT_LOG}" ]] && printf '%s\n' "$*" >>"${REPORT_LOG}"; return 0; }

cleanup() {
    local exit_code=$?
    local f
    for f in "${TMP_FILES[@]:-}"; do
        [[ -n "${f}" && -e "${f}" ]] && rm -f "${f}"
    done
    if [[ "${exit_code}" -ne 0 && "${APPLY}" -eq 1 ]]; then
        printf '%s\n' "PARTIAL DEPLOY: selected roots may have different states; no automatic cross-root rollback was attempted." >&2
        if [[ -n "${PI_BACKUP_PATH}" ]]; then
            printf '%s\n' "PI rollback: python3 ${PI_DEPLOY_HELPER} rollback --pi-dir \"${PI_AGENT_DIR}\" --backup-dir \"${PI_BACKUP_PATH}\" --apply" >&2
        fi
        if [[ -n "${SHARED_BACKUP_PATH}" ]]; then
            printf '%s\n' "Shared-skills rollback: python3 ${PI_DEPLOY_HELPER} shared rollback --shared-root \"${AGENTS_HOME}\" --backup-dir \"${SHARED_BACKUP_PATH}\" --apply" >&2
        fi
        if [[ -n "${BACKUP_FILE}" ]]; then
            printf '%s\n' "Claude rollback: tar -xzf \"${BACKUP_FILE}\" -C \"${CLAUDE_HOME}/\"" >&2
        fi
        if [[ -n "${HARNESS_BACKUP_FILE}" ]]; then
            printf '%s\n' "Harness-config rollback: tar -xzf \"${HARNESS_BACKUP_FILE}\" (restore only the selected config roots)" >&2
        fi
    fi
    return "${exit_code}"
}
trap cleanup EXIT

register_tmp() { TMP_FILES+=("$1"); }

# ---------------------------------------------------------------------------
# Help
# ---------------------------------------------------------------------------

print_help() {
    cat <<'EOF'
deploy-global.sh — deploy tricell-hive global/ + harness/ to the local machine.

USAGE:
  deploy-global.sh [--dry-run | --apply] [--only SCOPE[,SCOPE...]]
                    [--keep-orphans] [--force-delete-orphans] [--verbose] [--help]

FLAGS:
  --dry-run          Report what would change; write nothing. DEFAULT.
                      Orphans are only listed, never deleted, in a dry run.
  --apply            Perform the deploy for real (copies, backup, manifest write).
                      Manifest-confirmed orphans are DELETED as part of the run
                      (and their settings.json/hooks.json entries purged) — see
                      --keep-orphans to opt out and the size-based refusal below.
  --only SCOPE       Restrict the run to one or more scopes. Repeatable, or
                      comma-separated. One of:
                        claude    -> global/ into ~/.claude
                        codex     -> harness/ codex-specific targets (~/.codex)
                        opencode  -> harness/ opencode-specific targets (~/.config/opencode)
                        grok      -> always-on rules symlinked into ~/.grok/rules
                                     + generated agents into ~/.grok/agents
                        pi        -> isolated PI layer under $PI_CODING_AGENT_DIR
                        harness   -> alias for "codex,opencode,grok,pi"
                        all       -> everything (default when --only is omitted)
  --keep-orphans     Under --apply, list manifest-confirmed orphans WITHOUT
                      deleting them (the pre-2026-08 default). Their manifest
                      entries are preserved so a later run can still delete them.
                      Orphan deletion is otherwise part of the apply flow: the
                      orphan set is by construction sources the user already
                      removed from the repo, and the full list is printed before
                      deletion. Deletion is refused (not performed) when the set
                      exceeds 20 entries or 25% of the manifest — the anomaly
                      brake; see --force-delete-orphans.
  --force-delete-orphans
                      Bypasses the size-based refusal above. Only pass this when
                      a large orphan set is genuinely expected.
  --delete-orphans   Deprecated no-op alias (deletion is now the --apply
                      default); accepted for compatibility.
  --verbose, -v      Print every file-level action, not just per-category summaries.
  --help, -h         This message.

SCOPE NOTES:
  Backups (tar snapshot of ~/.claude before overwrite) only ever cover
  ~/.claude — there is no equivalent snapshot mechanism for ~/.codex or
  ~/.config/opencode in the source procedure, so the backup step is skipped
  entirely when the "claude" scope is not active.

  The "grok" scope (1) writes rule symlinks into ~/.grok/rules, each pointing
  at the deployed always-on rule under ~/.claude/rules — so it needs the
  "claude" scope to have run at least once for rules; (2) copies generated
  agents from harness/grok/agents into ~/.grok/agents. Grok's rules discovery
  is NOT recursive and does NOT honor `paths:`, so only always-on rules
  (those without `paths:`) are linked, flattened as `<dir>__<file>.md`.
  Path-scoped rules reach Grok through the router skills instead.

  The "pi" scope delegates to the stdlib helper under harness/pi/. It owns the
  PI agent directory and the neutral shared-skills owner under ~/.agents; it
  reads the legacy ~/.claude manifest only for exact, read-only migration.

EXIT STATUS:
  0  ran to completion (dry-run or apply)
  1  bad arguments, missing required tool, or a mutating step failed
EOF
}

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------

parse_args() {
    local only_raw=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --dry-run)
                APPLY=0
                shift
                ;;
            --apply)
                APPLY=1
                shift
                ;;
            --only)
                [[ $# -ge 2 ]] || die "--only requires a value"
                only_raw="${only_raw:+${only_raw},}$2"
                shift 2
                ;;
            --only=*)
                only_raw="${only_raw:+${only_raw},}${1#--only=}"
                shift
                ;;
            --keep-orphans)
                KEEP_ORPHANS=1
                shift
                ;;
            --delete-orphans)
                # Deprecated no-op alias: deletion is the --apply default now.
                printf 'NOTE: --delete-orphans is deprecated (orphan deletion is the --apply default; use --keep-orphans to opt out).\n' >&2
                shift
                ;;
            --force-delete-orphans)
                FORCE_DELETE_ORPHANS=1
                shift
                ;;
            --verbose | -v)
                VERBOSE=1
                shift
                ;;
            --help | -h)
                print_help
                exit 0
                ;;
            *)
                die "Unknown argument: $1 (see --help)"
                ;;
        esac
    done

    if [[ -n "${only_raw}" ]]; then
        RUN_CLAUDE=0
        RUN_CODEX=0
        RUN_OPENCODE=0
        RUN_GROK=0
        RUN_PI=0
        local IFS=','
        local -a scopes
        read -r -a scopes <<<"${only_raw}"
        local s
        for s in "${scopes[@]}"; do
            case "${s}" in
                claude) RUN_CLAUDE=1 ;;
                codex) RUN_CODEX=1 ;;
                opencode) RUN_OPENCODE=1 ;;
                grok) RUN_GROK=1 ;;
                pi) RUN_PI=1 ;;
                harness)
                    RUN_CODEX=1
                    RUN_OPENCODE=1
                    RUN_GROK=1
                    RUN_PI=1
                    ;;
                all)
                    RUN_CLAUDE=1
                    RUN_CODEX=1
                    RUN_OPENCODE=1
                    RUN_GROK=1
                    RUN_PI=1
                    ;;
                *) die "Unknown --only scope: ${s} (expected claude|codex|opencode|grok|pi|harness|all)" ;;
            esac
        done
    fi
}

check_pi_deps() {
    command -v python3 >/dev/null 2>&1 || die "Missing required tool: python3 (needed for the PI deploy helper)"
    command -v shasum >/dev/null 2>&1 || die "Missing required tool: shasum (needed by the PI plan-capture hook)"
    command -v pi >/dev/null 2>&1 || die "Missing required tool: pi (required version 0.85.1; automatic installation is disabled)"
    local pi_version
    pi_version=$(pi --version 2>/dev/null | sed -n '1p')
    [[ "${pi_version}" == "0.85.1" ]] || die "Unsupported PI version: ${pi_version:-unknown} (required 0.85.1; automatic installation is disabled)"
    [[ -f "${PI_DEPLOY_HELPER}" ]] || die "Missing PI deploy helper: ${PI_DEPLOY_HELPER}"
    [[ -f "${REPO_ROOT}/harness/AGENTS.md" ]] || die "Missing generated PI core: ${REPO_ROOT}/harness/AGENTS.md"
    [[ -d "${REPO_ROOT}/harness/pi/agents" ]] || die "Missing generated PI agents: ${REPO_ROOT}/harness/pi/agents"
    [[ -d "${REPO_ROOT}/harness/pi/src" ]] || die "Missing PI runtime source: ${REPO_ROOT}/harness/pi/src"
    [[ -d "${REPO_ROOT}/harness/pi/extensions" ]] || die "Missing PI extension tree: ${REPO_ROOT}/harness/pi/extensions"
    [[ -d "${REPO_ROOT}/global/hooks" ]] || die "Missing canonical hooks: ${REPO_ROOT}/global/hooks"
}

step_deploy_pi() {
    [[ "${RUN_PI}" -eq 1 ]] || return 0
    log "== Deploy: pi scope =="
    local -a cmd=(python3 "${PI_DEPLOY_HELPER}" deploy --repo-root "${REPO_ROOT}" --pi-dir "${PI_AGENT_DIR}")
    if [[ "${APPLY}" -eq 1 ]]; then
        cmd+=(--apply)
    fi
    [[ "${VERBOSE}" -eq 1 ]] && cmd+=(--verbose)
    local output
    if ! output=$("${cmd[@]}" 2>&1); then
        PI_BACKUP_PATH=$(printf '%s\n' "${output}" | sed -n 's/.*after backup \([^:]*\):.*/\1/p' | tail -n 1)
        printf '%s\n' "${output}" >&2
        return 1
    fi
    printf '%s\n' "${output}"
    PI_BACKUP_PATH=$(printf '%s\n' "${output}" | sed -n 's/^backup: //p' | tail -n 1)
    report "PI deploy: completed${PI_BACKUP_PATH:+; backup ${PI_BACKUP_PATH}}"
}

# ---------------------------------------------------------------------------
# Dependency check
# ---------------------------------------------------------------------------

check_deps() {
    # REPO_ROOT sanity guard: if this doesn't look like tricell-hive (bad
    # rebase, worktree, partial clone, script copied elsewhere), every
    # manifest source below resolves "missing" and everything the manifest
    # lists looks orphaned. Refuse before that can happen.
    [[ -f "${REPO_ROOT}/global/CLAUDE.md" ]] || die "REPO_ROOT (${REPO_ROOT}) is missing global/CLAUDE.md — this doesn't look like tricell-hive. Refusing to run."
    [[ -d "${REPO_ROOT}/global/rules" ]] || die "REPO_ROOT (${REPO_ROOT}) is missing global/rules — this doesn't look like tricell-hive. Refusing to run."
    [[ -d "${REPO_ROOT}/global/agents" ]] || die "REPO_ROOT (${REPO_ROOT}) is missing global/agents — this doesn't look like tricell-hive. Refusing to run."

    local missing=()
    command -v jq >/dev/null 2>&1 || missing+=("jq")
    command -v tar >/dev/null 2>&1 || missing+=("tar")
    command -v find >/dev/null 2>&1 || missing+=("find")
    command -v python3 >/dev/null 2>&1 || missing+=("python3 (needed to rebuild harness/ and PI preflight)")
    if [[ ${#missing[@]} -gt 0 ]]; then
        die "Missing required tool(s): ${missing[*]}"
    fi
    [[ -f "${HOOK_MERGE_FILTER}" ]] || die "Missing bundled filter: ${HOOK_MERGE_FILTER}"
    [[ -f "${HOOK_PURGE_FILTER}" ]] || die "Missing bundled filter: ${HOOK_PURGE_FILTER}"
}

shared_scope_selected() {
    [[ "${RUN_PI}" -eq 1 || "${RUN_CODEX}" -eq 1 || "${RUN_OPENCODE}" -eq 1 || "${RUN_GROK}" -eq 1 ]]
}

preflight_parent_chain() {
    local label="$1" path="$2" parent
    parent=$(dirname "${path}")
    while [[ "${parent}" != "/" && "${parent}" != "${HOME}" && -n "${parent}" ]]; do
        if [[ -L "${parent}" ]]; then
            die "${label} target parent is a symlink: ${parent}"
        fi
        parent=$(dirname "${parent}")
    done
}

preflight_target_root() {
    local label="$1" root="$2"
    preflight_parent_chain "${label}" "${root}"
    if [[ -L "${root}" ]]; then
        die "${label} target root is a symlink: ${root}"
    fi
    if [[ -e "${root}" && ! -d "${root}" ]]; then
        die "${label} target root is not a directory: ${root}"
    fi
    if [[ -d "${root}" && ! -w "${root}" ]]; then
        die "${label} target root is not writable: ${root}"
    fi
}

preflight_target_file() {
    local label="$1" target="$2"
    preflight_parent_chain "${label}" "${target}"
    if [[ -L "${target}" ]]; then
        die "${label} target is a symlink: ${target}"
    fi
    if [[ -e "${target}" && ! -f "${target}" ]]; then
        die "${label} target is not a regular file: ${target}"
    fi
    if [[ -e "${target}" && ! -w "${target}" ]]; then
        die "${label} target is not writable: ${target}"
    fi
}

preflight_tree_targets() {
    local label="$1" source_root="$2" target_root="$3"
    [[ -d "${source_root}" ]] || return 0
    local source rel target
    while IFS= read -r -d '' source; do
        rel="${source#"${source_root}"/}"
        target="${target_root}/${rel}"
        preflight_target_file "${label}" "${target}"
    done < <(find "${source_root}" -type f -print0 2>/dev/null)
}

preflight_flat_targets() {
    local label="$1" source_root="$2" target_root="$3"
    [[ -d "${source_root}" ]] || return 0
    local source target
    while IFS= read -r -d '' source; do
        target="${target_root}/$(basename "${source}")"
        preflight_target_file "${label}" "${target}"
    done < <(find "${source_root}" -type f -print0 2>/dev/null)
}

step_preflight_generic_targets() {
    if [[ "${RUN_CLAUDE}" -eq 1 ]]; then
        preflight_target_root "Claude" "${CLAUDE_HOME}"
        preflight_target_root "Claude rules" "${CLAUDE_HOME}/rules"
        preflight_target_root "Claude agents" "${CLAUDE_HOME}/agents"
        preflight_target_root "Claude skills" "${CLAUDE_HOME}/skills"
        preflight_target_root "Claude hooks" "${CLAUDE_HOME}/hooks"
        preflight_target_file "Claude CLAUDE.md" "${CLAUDE_HOME}/CLAUDE.md"
        preflight_target_file "Claude settings.json" "${CLAUDE_HOME}/settings.json"
        preflight_tree_targets "Claude rules" "${REPO_ROOT}/global/rules" "${CLAUDE_HOME}/rules"
        preflight_tree_targets "Claude agents" "${REPO_ROOT}/global/agents" "${CLAUDE_HOME}/agents"
        preflight_tree_targets "Claude skills" "${REPO_ROOT}/global/skills" "${CLAUDE_HOME}/skills"
        preflight_flat_targets "Claude hooks" "${REPO_ROOT}/global/hooks" "${CLAUDE_HOME}/hooks"
    fi
    if [[ "${RUN_CODEX}" -eq 1 ]]; then
        preflight_target_root "Codex" "${CODEX_HOME}"
        preflight_target_root "Codex agents" "${CODEX_HOME}/agents"
        preflight_target_root "Codex hooks" "${CODEX_HOME}"
        preflight_target_file "Codex AGENTS.md" "${CODEX_HOME}/AGENTS.md"
        preflight_target_file "Codex hooks.json" "${CODEX_HOME}/hooks.json"
        preflight_flat_targets "Codex agents" "${REPO_ROOT}/harness/codex/agents" "${CODEX_HOME}/agents"
        local codex_hooks_config hookdir
        while IFS= read -r codex_hooks_config; do
            [[ -n "${codex_hooks_config}" ]] || continue
            hookdir=$(dirname "${codex_hooks_config}")
            preflight_flat_targets "Codex hooks" "${hookdir}" "${CODEX_HOME}/hooks"
        done < <(find "${REPO_ROOT}/global/hooks" -name codex-hooks.json 2>/dev/null)
    fi
    if [[ "${RUN_OPENCODE}" -eq 1 ]]; then
        preflight_target_root "OpenCode" "${OPENCODE_HOME}"
        preflight_target_root "OpenCode agents" "${OPENCODE_HOME}/agents"
        preflight_target_root "OpenCode commands" "${OPENCODE_HOME}/commands"
        preflight_target_root "OpenCode rules" "${OPENCODE_HOME}/rules"
        preflight_target_root "OpenCode plugins" "${OPENCODE_HOME}/plugins"
        preflight_target_file "OpenCode AGENTS.md" "${OPENCODE_HOME}/AGENTS.md"
        preflight_target_file "OpenCode opencode.json" "${OPENCODE_HOME}/opencode.json"
        preflight_tree_targets "OpenCode agents" "${REPO_ROOT}/harness/opencode/agents" "${OPENCODE_HOME}/agents"
        preflight_tree_targets "OpenCode commands" "${REPO_ROOT}/harness/opencode/commands" "${OPENCODE_HOME}/commands"
        preflight_tree_targets "OpenCode rules" "${REPO_ROOT}/harness/opencode/rules" "${OPENCODE_HOME}/rules"
        preflight_target_file "OpenCode session plugin" "${OPENCODE_HOME}/plugins/flow-session-context.ts"
    fi
    if [[ "${RUN_GROK}" -eq 1 ]]; then
        preflight_target_root "Grok" "${GROK_HOME}"
        preflight_target_root "Grok rules" "${GROK_RULES_HOME}"
        preflight_target_root "Grok agents" "${GROK_AGENTS_HOME}"
        preflight_flat_targets "Grok agents" "${REPO_ROOT}/harness/grok/agents" "${GROK_AGENTS_HOME}"
        local grok_rule_relative grok_rule_flat grok_rule_target
        while IFS= read -r grok_rule_relative; do
            [[ -n "${grok_rule_relative}" ]] || continue
            grok_rule_flat="${grok_rule_relative//\//__}"
            grok_rule_target="${GROK_RULES_HOME}/${grok_rule_flat}"
            if [[ -e "${grok_rule_target}" && ! -L "${grok_rule_target}" ]]; then
                die "Grok rule target is not a symlink: ${grok_rule_target}"
            fi
        done < <(grok_always_on_rules)
    fi
}

# All selected roots are validated before the first target write. The helper's
# dry-runs also validate legacy shared ownership and strict conflict handling;
# they never create the neutral manifest or its backup directory.
step_preflight_selected() {
    log "== Preflight: selected scopes =="
    step_preflight_generic_targets
    if [[ "${RUN_PI}" -eq 1 ]]; then
        local -a pi_cmd=(python3 "${PI_DEPLOY_HELPER}" preflight --repo-root "${REPO_ROOT}" --pi-dir "${PI_AGENT_DIR}")
        [[ "${VERBOSE}" -eq 1 ]] && pi_cmd+=(--verbose)
        "${pi_cmd[@]}"
        report "PI preflight: OK"
    fi
    if shared_scope_selected; then
        local -a shared_cmd=(python3 "${PI_DEPLOY_HELPER}" shared deploy --repo-root "${REPO_ROOT}" --shared-root "${AGENTS_HOME}" --legacy-manifest "${MANIFEST}" --dry-run)
        [[ "${VERBOSE}" -eq 1 ]] && shared_cmd+=(--verbose)
        "${shared_cmd[@]}"
        report "shared skills preflight: OK"
    fi
}

# ---------------------------------------------------------------------------
# Small generic helpers
# ---------------------------------------------------------------------------

# count_files_lines DIR [find-args...] -> echoes "FILES LINES"
count_files_lines() {
    local dir="$1"
    shift
    local files=0 lines=0
    if [[ -d "${dir}" ]]; then
        files=$(find "${dir}" -type f "$@" 2>/dev/null | wc -l | tr -d ' ')
        lines=$(find "${dir}" -type f "$@" -print0 2>/dev/null | xargs -0 cat 2>/dev/null | wc -l | tr -d ' ')
    fi
    printf '%s %s' "${files}" "${lines}"
}

# diff_category LABEL SRC_DIR TGT_DIR [find-args...]
diff_category() {
    local label="$1" src="$2" tgt="$3"
    shift 3
    local new=0 modified=0 unchanged=0
    if [[ ! -d "${src}" ]]; then
        log "  ${label}: source not found, skipped"
        return 0
    fi
    while IFS= read -r -d '' f; do
        local rel="${f#"${src}"/}"
        local t="${tgt}/${rel}"
        if [[ ! -e "${t}" ]]; then
            new=$((new + 1))
        elif cmp -s "${f}" "${t}"; then
            unchanged=$((unchanged + 1))
        else
            modified=$((modified + 1))
        fi
    done < <(find "${src}" -type f "$@" -print0 2>/dev/null)
    log "  ${label}: ${new} new, ${modified} modified, ${unchanged} unchanged"
    report "  ${label}: ${new} new, ${modified} modified, ${unchanged} unchanged"
}

# diff_file LABEL SRC_FILE TGT_FILE
diff_file() {
    local label="$1" src="$2" tgt="$3"
    if [[ ! -f "${src}" ]]; then
        log "  ${label}: source not found, skipped"
        return 0
    fi
    if [[ ! -e "${tgt}" ]]; then
        log "  ${label}: NEW"
        report "  ${label}: NEW"
    elif cmp -s "${src}" "${tgt}"; then
        log "  ${label}: unchanged"
        report "  ${label}: unchanged"
    else
        log "  ${label}: MODIFIED"
        report "  ${label}: MODIFIED"
    fi
}

diff_hooks() {
    local new=0 modified=0 unchanged=0
    while IFS= read -r -d '' f; do
        local b t
        b=$(basename "${f}")
        t="${CLAUDE_HOME}/hooks/${b}"
        if [[ ! -e "${t}" ]]; then
            new=$((new + 1))
        elif cmp -s "${f}" "${t}"; then
            unchanged=$((unchanged + 1))
        else
            modified=$((modified + 1))
        fi
    done < <(find "${REPO_ROOT}/global/hooks" \( -name '*.sh' -o \( -name '*.json' ! -name 'settings-config.json' ! -name 'codex-hooks.json' \) \) -print0 2>/dev/null)
    log "  hooks: ${new} new, ${modified} modified, ${unchanged} unchanged"
    report "  hooks: ${new} new, ${modified} modified, ${unchanged} unchanged"
}

# deploy_tree LABEL SRC_DIR TGT_DIR — recursive overlay copy (never deletes).
deploy_tree() {
    local label="$1" src="$2" tgt="$3"
    if [[ ! -d "${src}" ]] || [[ -z "$(find "${src}" -type f -print -quit 2>/dev/null)" ]]; then
        vlog "${label}: source empty or not found, skip"
        return 0
    fi
    local count
    count=$(find "${src}" -type f 2>/dev/null | wc -l | tr -d ' ')
    if [[ "${APPLY}" -eq 1 ]]; then
        mkdir -p "${tgt}"
        cp -R "${src}/." "${tgt}/"
        log "Deployed ${label}: ${count} files -> ${tgt}"
    else
        log "[DRY-RUN] would deploy ${label}: ${count} files -> ${tgt}"
    fi
    report "${label}: ${count} files -> ${tgt}"
}

# deploy_file LABEL SRC_FILE TGT_FILE
deploy_file() {
    local label="$1" src="$2" tgt="$3"
    if [[ ! -f "${src}" ]]; then
        vlog "${label}: source not found, skip"
        return 0
    fi
    if [[ "${APPLY}" -eq 1 ]]; then
        mkdir -p "$(dirname "${tgt}")"
        cp "${src}" "${tgt}"
        log "Deployed ${label} -> ${tgt}"
    else
        log "[DRY-RUN] would deploy ${label} -> ${tgt}"
    fi
    report "${label} -> ${tgt}"
}

# deploy_injected_references — copy build-injected `references/` into the
# Claude skills tree.
#
# Router skills (language-rules, workspace-conventions, memory-policy,
# unattended-delegation, flow-report) carry no `references/` in global/skills:
# harness/build.py injects them from global/rules into harness/agents-skills
# only. That tree feeds ~/.agents/skills (Codex, opencode) — but Grok scans
# ~/.claude/skills and never ~/.agents/skills, so on Grok every router pointed
# at reference files that existed in no path it could see, and the model went
# looking for them by guessing. Copying the injected sets here closes that gap
# and pre-stages the same files for Claude Code.
#
# The comparison is per FILE, not per skill: flow-report owns references in
# global/ AND receives an injected one, so skipping the whole skill would have
# left exactly that file undeployed. A file already present in global/ is never
# overwritten here — convert-skills.py strips frontmatter on its way into
# harness/, so copying it back would deploy the cleaned variant over the source.
deploy_injected_references() {
    local src_root="${REPO_ROOT}/harness/agents-skills"
    [[ -d "${src_root}" ]] || return 0
    local d name f base count=0
    for d in "${src_root}"/*/references/; do
        [[ -d "${d}" ]] || continue
        name=$(basename "$(dirname "${d}")")
        for f in "${d}"*; do
            [[ -f "${f}" ]] || continue
            base=$(basename "${f}")
            # Canonical copy already deployed from global/ — leave it alone.
            [[ -f "${REPO_ROOT}/global/skills/${name}/references/${base}" ]] && continue
            if [[ "${APPLY}" -eq 1 ]]; then
                mkdir -p "${CLAUDE_HOME}/skills/${name}/references"
                cp "${f}" "${CLAUDE_HOME}/skills/${name}/references/${base}"
            fi
            count=$((count + 1))
        done
    done
    if [[ "${count}" -eq 0 ]]; then
        vlog "injected references: none"
        return 0
    fi
    if [[ "${APPLY}" -eq 1 ]]; then
        log "Deployed injected references: ${count} files -> ${CLAUDE_HOME}/skills"
    else
        log "[DRY-RUN] would deploy injected references: ${count} files -> ${CLAUDE_HOME}/skills"
    fi
    report "injected references: ${count} files -> ${CLAUDE_HOME}/skills"
}

# ---------------------------------------------------------------------------
# Grok rule selection
#
# Grok discovers rules by scanning `rules/` NON-recursively and does not honor
# `paths:` — both verified empirically against grok 0.2.118 with marker files
# under a temporary GROK_HOME. Consequences encoded below:
#   - the tree under global/rules/ is invisible to it, so each rule is linked
#     flat as `<subdir>__<file>.md` (file symlinks ARE followed; directory
#     symlinks are not);
#   - a path-scoped rule would become always-on there, so only rules WITHOUT
#     `paths:` are linked. The rest reach Grok through the router skills.
# ---------------------------------------------------------------------------

# rule_is_always_on FILE -> 0 when the rule loads unconditionally.
# `paths:` in the first frontmatter block is the only key that makes a rule
# conditional; every other key (alwaysApply:, …) is documentation, not
# mechanism, and no frontmatter at all means always-on.
rule_is_always_on() {
    local f="$1" first="" fm
    [[ -f "${f}" ]] || return 1
    IFS= read -r first <"${f}" || true
    [[ "${first}" == "---" ]] || return 0
    fm=$(sed -n '2,/^---[[:space:]]*$/p' "${f}")
    grep -q '^paths:' <<<"${fm}" && return 1
    return 0
}

# grok_always_on_rules -> one rel path per line, e.g. "quality/testing.md"
grok_always_on_rules() {
    local f rel
    while IFS= read -r f; do
        [[ -n "${f}" ]] || continue
        rel="${f#"${REPO_ROOT}"/global/rules/}"
        # `__` is the flatten separator; a rule filename containing it would
        # make the flat name ambiguous to reverse in manifest_entry_map. Fail
        # loudly instead of silently mapping it to a nonexistent source.
        case "${rel}" in
            *__*) die "Rule path contains '__', which collides with the Grok flatten separator: ${rel}" ;;
        esac
        if rule_is_always_on "${f}"; then
            printf '%s\n' "${rel}"
        fi
    done < <(find "${REPO_ROOT}/global/rules" -type f -name '*.md' 2>/dev/null | sort)
}

# ---------------------------------------------------------------------------
# Step 1 (preview) — always runs, always read-only.
# ---------------------------------------------------------------------------

step_preview() {
    log "== Preview: source tree =="
    local f l
    if [[ "${RUN_CLAUDE}" -eq 1 ]]; then
        [[ -f "${REPO_ROOT}/global/CLAUDE.md" ]] && log "  CLAUDE.md: 1 file"
        read -r f l <<<"$(count_files_lines "${REPO_ROOT}/global/rules")"
        log "  rules: ${f} files, ${l} lines"
        read -r f l <<<"$(count_files_lines "${REPO_ROOT}/global/agents")"
        log "  agents: ${f} files, ${l} lines"
        read -r f l <<<"$(count_files_lines "${REPO_ROOT}/global/skills")"
        log "  skills: ${f} files, ${l} lines"
        read -r f l <<<"$(count_files_lines "${REPO_ROOT}/global/hooks")"
        log "  hooks: ${f} files, ${l} lines"
    fi
    if [[ "${RUN_CODEX}" -eq 1 || "${RUN_OPENCODE}" -eq 1 || "${RUN_PI}" -eq 1 ]]; then
        log "  (harness/ counts reported after rebuild, in the deploy step)"
    fi
}

# ---------------------------------------------------------------------------
# Step 3 (diff before deploying) — always read-only.
# ---------------------------------------------------------------------------

step_diff() {
    log "== Diff: source vs. deployed =="
    if [[ "${RUN_CLAUDE}" -eq 1 ]]; then
        diff_file "CLAUDE.md" "${REPO_ROOT}/global/CLAUDE.md" "${CLAUDE_HOME}/CLAUDE.md"
        diff_category "rules" "${REPO_ROOT}/global/rules" "${CLAUDE_HOME}/rules"
        diff_category "agents" "${REPO_ROOT}/global/agents" "${CLAUDE_HOME}/agents"
        diff_category "skills" "${REPO_ROOT}/global/skills" "${CLAUDE_HOME}/skills"
        diff_hooks
    fi
    if [[ "${RUN_CODEX}" -eq 1 || "${RUN_OPENCODE}" -eq 1 || "${RUN_PI}" -eq 1 ]]; then
        # step_harness_rebuild already ran the apply build or the read-only
        # parity check before this diff, so the counts describe the generated
        # tree that the deployment steps will consume.
        log "  (harness/ diff below uses the generated tree validated above)"
        diff_category "agents-skills" "${REPO_ROOT}/harness/agents-skills" "${AGENTS_SKILLS_HOME}"
    fi
    if [[ "${RUN_CODEX}" -eq 1 ]]; then
        diff_category "codex-agents" "${REPO_ROOT}/harness/codex/agents" "${CODEX_HOME}/agents" -name '*.toml'
        diff_file "harness AGENTS.md -> codex" "${REPO_ROOT}/harness/AGENTS.md" "${CODEX_HOME}/AGENTS.md"
    fi
    if [[ "${RUN_OPENCODE}" -eq 1 ]]; then
        diff_category "opencode-agents" "${REPO_ROOT}/harness/opencode/agents" "${OPENCODE_HOME}/agents" -name '*.md' ! -name 'README.md'
        diff_category "opencode-commands" "${REPO_ROOT}/harness/opencode/commands" "${OPENCODE_HOME}/commands" -name '*.md'
        diff_category "opencode-rules" "${REPO_ROOT}/harness/opencode/rules" "${OPENCODE_HOME}/rules" -name '*.md' ! -name 'README.md'
        diff_file "harness AGENTS.md -> opencode" "${REPO_ROOT}/harness/AGENTS.md" "${OPENCODE_HOME}/AGENTS.md"
    fi
    if [[ "${RUN_GROK}" -eq 1 ]]; then
        diff_grok_rules
        diff_category "grok-agents" "${REPO_ROOT}/harness/grok/agents" "${GROK_AGENTS_HOME}" -name '*.md' ! -name 'README.md'
    fi
}

diff_grok_rules() {
    local new=0 relink=0 unchanged=0 total=0 rel flat current
    while IFS= read -r rel; do
        [[ -n "${rel}" ]] || continue
        total=$((total + 1))
        flat="${rel//\//__}"
        current=""
        [[ -L "${GROK_RULES_HOME}/${flat}" ]] && current=$(readlink "${GROK_RULES_HOME}/${flat}")
        if [[ "${current}" == "${CLAUDE_HOME}/rules/${rel}" ]]; then
            unchanged=$((unchanged + 1))
        elif [[ -e "${GROK_RULES_HOME}/${flat}" || -L "${GROK_RULES_HOME}/${flat}" ]]; then
            relink=$((relink + 1))
        else
            new=$((new + 1))
        fi
    done < <(grok_always_on_rules)
    log "  grok-rules: ${new} new, ${relink} relinked, ${unchanged} unchanged (${total} always-on of $(find "${REPO_ROOT}/global/rules" -type f -name '*.md' 2>/dev/null | wc -l | tr -d ' ') rules)"
    report "  grok-rules: ${new} new, ${relink} relinked, ${unchanged} unchanged (${total} always-on rules)"
}

# ---------------------------------------------------------------------------
# Step 4 (backup) — ~/.claude tar, scoped to RUN_CLAUDE; step 4b snapshots the
# mutable config files this deploy touches OUTSIDE ~/.claude.
# ---------------------------------------------------------------------------

BACKUP_FILE=""
PI_BACKUP_PATH=""
SHARED_BACKUP_PATH=""

step_backup() {
    [[ "${RUN_CLAUDE}" -eq 1 ]] || return 0
    log "== Backup =="

    local -a tar_args=()
    [[ -f "${CLAUDE_HOME}/CLAUDE.md" ]] && tar_args+=("CLAUDE.md")
    [[ -d "${CLAUDE_HOME}/rules" && -n "$(ls -A "${CLAUDE_HOME}/rules" 2>/dev/null)" ]] && tar_args+=("rules/")
    [[ -d "${CLAUDE_HOME}/agents" && -n "$(ls -A "${CLAUDE_HOME}/agents" 2>/dev/null)" ]] && tar_args+=("agents/")
    [[ -d "${CLAUDE_HOME}/skills" && -n "$(ls -A "${CLAUDE_HOME}/skills" 2>/dev/null)" ]] && tar_args+=("skills/")
    [[ -d "${CLAUDE_HOME}/hooks" && -n "$(ls -A "${CLAUDE_HOME}/hooks" 2>/dev/null)" ]] && tar_args+=("hooks/")
    [[ -f "${CLAUDE_HOME}/settings.json" ]] && tar_args+=("settings.json")

    if [[ ${#tar_args[@]} -eq 0 ]]; then
        log "Nothing to back up (fresh install)"
        report "backup: nothing to back up (fresh install)"
        return 0
    fi

    BACKUP_FILE="${BACKUP_DIR}/global-backup-$(date +%Y%m%d-%H%M%S).tar.gz"
    if [[ "${APPLY}" -eq 1 ]]; then
        mkdir -p "${BACKUP_DIR}"
        tar -czf "${BACKUP_FILE}" -C "${CLAUDE_HOME}" "${tar_args[@]}"
        local size
        size=$(wc -c <"${BACKUP_FILE}" | tr -d ' ')
        log "Backup written: ${BACKUP_FILE} (${size} bytes)"
        report "backup: ${BACKUP_FILE} (${size} bytes)"

        # Keep only the 5 most recent backups.
        local -a old_backups
        old_backups=()
        # shellcheck disable=SC2012  # filenames are script-generated (global-backup-<timestamp>.tar.gz); ls -t is the portable mtime sort on BSD
        while IFS= read -r line; do
            old_backups+=("${line}")
        done < <(ls -t "${BACKUP_DIR}"/global-backup-*.tar.gz 2>/dev/null | tail -n +6)
        if [[ ${#old_backups[@]} -gt 0 ]]; then
            rm -f "${old_backups[@]}"
            log "Pruned ${#old_backups[@]} old backup(s)"
        fi
    else
        log "[DRY-RUN] would back up: ${tar_args[*]} -> ${BACKUP_FILE}"
        report "backup: [DRY-RUN] would write ${BACKUP_FILE} from: ${tar_args[*]}"
    fi
}

# ---------------------------------------------------------------------------
# Step 4b (harness-config backup) — the ~/.claude tar above covers ~/.claude and
# nothing else, yet a codex/opencode run MUTATES two files outside it: the hooks
# merge writes ~/.codex/hooks.json (step 13d), the orphan sweep purges blocks
# from that same file, and the permission merge writes opencode.json. Each merge
# validates into a temp file before moving, so corruption is unlikely — but a
# purge that removes a block the user hand-added had nothing to restore from.
# Scoped per harness: a scope that does not run is not snapshotted.
# ---------------------------------------------------------------------------

HARNESS_BACKUP_FILE=""

step_backup_harness_config() {
    [[ "${RUN_CODEX}" -eq 1 || "${RUN_OPENCODE}" -eq 1 ]] || return 0
    log "== Backup (harness config) =="

    local -a tar_args=()
    local -a restore_hints=()
    if [[ "${RUN_CODEX}" -eq 1 && -f "${CODEX_HOME}/hooks.json" ]]; then
        tar_args+=(-C "${CODEX_HOME}" "hooks.json")
        restore_hints+=("hooks.json -> ${CODEX_HOME}")
    fi
    if [[ "${RUN_OPENCODE}" -eq 1 && -f "${OPENCODE_HOME}/opencode.json" ]]; then
        tar_args+=(-C "${OPENCODE_HOME}" "opencode.json")
        restore_hints+=("opencode.json -> ${OPENCODE_HOME}")
    fi

    if [[ ${#tar_args[@]} -eq 0 ]]; then
        log "No harness config present to back up"
        report "harness-config backup: nothing present to back up"
        return 0
    fi

    HARNESS_BACKUP_FILE="${BACKUP_DIR}/harness-config-backup-$(date +%Y%m%d-%H%M%S).tar.gz"
    if [[ "${APPLY}" -eq 1 ]]; then
        mkdir -p "${BACKUP_DIR}"
        tar -czf "${HARNESS_BACKUP_FILE}" "${tar_args[@]}"
        local size
        size=$(wc -c <"${HARNESS_BACKUP_FILE}" | tr -d ' ')
        log "Harness-config backup written: ${HARNESS_BACKUP_FILE} (${size} bytes)"
        report "harness-config backup: ${HARNESS_BACKUP_FILE} (${size} bytes) — ${restore_hints[*]}"

        # Same retention as the ~/.claude tar: keep the 5 most recent.
        local -a old_backups
        old_backups=()
        # shellcheck disable=SC2012  # filenames are script-generated (harness-config-backup-<timestamp>.tar.gz); ls -t is the portable mtime sort on BSD
        while IFS= read -r line; do
            old_backups+=("${line}")
        done < <(ls -t "${BACKUP_DIR}"/harness-config-backup-*.tar.gz 2>/dev/null | tail -n +6)
        if [[ ${#old_backups[@]} -gt 0 ]]; then
            rm -f "${old_backups[@]}"
            log "Pruned ${#old_backups[@]} old harness-config backup(s)"
        fi
    else
        log "[DRY-RUN] would back up: ${restore_hints[*]} -> ${HARNESS_BACKUP_FILE}"
        report "harness-config backup: [DRY-RUN] would write ${HARNESS_BACKUP_FILE} from: ${restore_hints[*]}"
    fi
}

# ---------------------------------------------------------------------------
# Step 5 (orphan detection + confirmation-gated deletion)
# ---------------------------------------------------------------------------

# manifest_entry_scope REL -> echoes claude|shared|codex|opencode|unknown
manifest_entry_scope() {
    local rel="$1"
    case "${rel}" in
        CLAUDE.md | rules/* | agents/* | skills/* | hooks/*) echo "claude" ;;
        agents-skills/*) echo "shared" ;;
        codex-agents/* | codex-hooks/* | harness-agents/codex/*) echo "codex" ;;
        opencode-agents/* | opencode-commands/* | opencode-plugins/* | harness-agents/opencode/*) echo "opencode" ;;
        grok-rules/* | grok-agents/*) echo "grok" ;;
        *) echo "unknown" ;;
    esac
}

# manifest_entry_map REL — sets MAP_SRC/MAP_TGT globals for REL. Empty
# MAP_SRC (with a non-empty MAP_TGT) is the orphan condition. Empty MAP_TGT
# means an unrecognized prefix — never touched, same "unknown prefix, never
# delete" rule as the original.
#
# NOTE: this used to round-trip through `printf 'src\ttgt' | IFS=$'\t' read`.
# Tab is IFS *whitespace*, so `read` strips leading IFS-whitespace fields —
# when MAP_SRC was empty (exactly the orphan case), the empty first field
# vanished and MAP_TGT silently received MAP_SRC's value instead, leaving the
# real target empty. That broke orphan detection for every find-based branch
# below (hooks/*, codex-agents/*, opencode-agents/*, codex-hooks/*,
# opencode-plugins/*). Globals sidestep the round-trip entirely — no
# delimiter, no read, no subshell.
MAP_SRC=""
MAP_TGT=""

manifest_entry_map() {
    local rel="$1"
    MAP_SRC=""
    MAP_TGT=""
    case "${rel}" in
        CLAUDE.md | rules/* | agents/*)
            MAP_SRC="${REPO_ROOT}/global/${rel}"
            MAP_TGT="${CLAUDE_HOME}/${rel}"
            ;;
        skills/*)
            MAP_SRC="${REPO_ROOT}/global/${rel}"
            MAP_TGT="${CLAUDE_HOME}/${rel}"
            # Router `references/` have no counterpart in global/skills — they
            # are injected into harness/agents-skills by build.py. Without this
            # fallback every injected file reads as an orphan and step 5 offers
            # to delete what the previous step just deployed.
            [[ -e "${MAP_SRC}" ]] || MAP_SRC="${REPO_ROOT}/harness/agents-skills/${rel#skills/}"
            ;;
        hooks/*)
            local bn
            bn=$(basename "${rel}")
            MAP_SRC=$(find "${REPO_ROOT}/global/hooks" -name "${bn}" -print -quit 2>/dev/null)
            MAP_TGT="${CLAUDE_HOME}/${rel}"
            ;;
        agents-skills/*)
            MAP_SRC="${REPO_ROOT}/global/skills/${rel#agents-skills/}"
            MAP_TGT="${AGENTS_SKILLS_HOME}/${rel#agents-skills/}"
            # Same injected-references case as skills/* above: build.py writes
            # them straight into harness/, never into global/skills.
            [[ -e "${MAP_SRC}" ]] || MAP_SRC="${REPO_ROOT}/harness/agents-skills/${rel#agents-skills/}"
            ;;
        codex-agents/*)
            local n
            n=$(basename "${rel}" .toml)
            MAP_SRC=$(find "${REPO_ROOT}/global/agents" -name "${n}.md" -print -quit 2>/dev/null)
            MAP_TGT="${CODEX_HOME}/agents/$(basename "${rel}")"
            ;;
        opencode-agents/*)
            local n
            n=$(basename "${rel}" .md)
            MAP_SRC=$(find "${REPO_ROOT}/global/agents" -name "${n}.md" -print -quit 2>/dev/null)
            MAP_TGT="${OPENCODE_HOME}/agents/$(basename "${rel}")"
            ;;
        opencode-commands/*)
            MAP_SRC="${REPO_ROOT}/harness/opencode/commands/$(basename "${rel}")"
            MAP_TGT="${OPENCODE_HOME}/commands/$(basename "${rel}")"
            ;;
        codex-hooks/*)
            local bn
            bn=$(basename "${rel}")
            MAP_SRC=$(find "${REPO_ROOT}/global/hooks" -name "${bn}" -print -quit 2>/dev/null)
            MAP_TGT="${CODEX_HOME}/hooks/${bn}"
            ;;
        opencode-plugins/*)
            local bn
            bn=$(basename "${rel}")
            MAP_SRC=$(find "${REPO_ROOT}/global/hooks" -name "${bn}" -print -quit 2>/dev/null)
            MAP_TGT="${OPENCODE_HOME}/plugins/${bn}"
            ;;
        harness-agents/codex/*)
            MAP_SRC="${REPO_ROOT}/harness/AGENTS.md"
            MAP_TGT="${CODEX_HOME}/AGENTS.md"
            ;;
        harness-agents/opencode/*)
            MAP_SRC="${REPO_ROOT}/harness/AGENTS.md"
            MAP_TGT="${OPENCODE_HOME}/AGENTS.md"
            ;;
        grok-rules/*)
            # Un-flatten `quality__testing.md` back to `quality/testing.md`.
            # MAP_SRC is deliberately left empty when the rule still exists but
            # is no longer always-on (it gained `paths:`): the link must then be
            # detected as an orphan, since Grok cannot scope it.
            local flat unflat
            flat=$(basename "${rel}")
            unflat="${flat//__//}"
            if rule_is_always_on "${REPO_ROOT}/global/rules/${unflat}"; then
                MAP_SRC="${REPO_ROOT}/global/rules/${unflat}"
            fi
            MAP_TGT="${GROK_RULES_HOME}/${flat}"
            ;;
        grok-agents/*)
            local n
            n=$(basename "${rel}" .md)
            MAP_SRC=$(find "${REPO_ROOT}/global/agents" -name "${n}.md" -print -quit 2>/dev/null)
            MAP_TGT="${GROK_AGENTS_HOME}/$(basename "${rel}")"
            ;;
        *)
            : # unrecognized prefix — MAP_SRC/MAP_TGT stay empty, never touched
            ;;
    esac
}

ORPHANS=()
# Orphans detected this run but NOT deleted (dry run, --keep-orphans, or the
# size breaker refused). step_write_manifest re-emits these entries so a later run
# can still confirm and delete them — dropping them would leave the stale
# files untracked forever.
KEPT_ORPHANS=()
MANIFEST_TOTAL=0

step_detect_orphans() {
    log "== Orphan detection (manifest-based) =="
    ORPHANS=()
    MANIFEST_TOTAL=0
    if [[ ! -f "${MANIFEST}" ]]; then
        log "No manifest found (first run, or unreadable) — skipping orphan detection."
        report "orphans: no manifest present, skipped"
        return 0
    fi

    local rel scope
    while IFS= read -r rel; do
        case "${rel}" in '#'* | '') continue ;; esac
        scope=$(manifest_entry_scope "${rel}")
        case "${scope}" in
            claude) [[ "${RUN_CLAUDE}" -eq 1 ]] || continue ;;
            shared) continue ;; # legacy shared ownership is read-only here
            codex) [[ "${RUN_CODEX}" -eq 1 ]] || continue ;;
            opencode) [[ "${RUN_OPENCODE}" -eq 1 ]] || continue ;;
            grok) [[ "${RUN_GROK}" -eq 1 ]] || continue ;;
            *) continue ;; # unknown prefix — never touch
        esac
        MANIFEST_TOTAL=$((MANIFEST_TOTAL + 1))
        manifest_entry_map "${rel}"
        if [[ -z "${MAP_SRC}" || ! -e "${MAP_SRC}" ]]; then
            # -L as well as -e: a dangling symlink (grok-rules/ pointing at a
            # rule that was removed from ~/.claude/rules) fails -e, and would
            # otherwise survive every cleanup while still being loaded-by-name.
            { [[ -n "${MAP_TGT}" ]] && { [[ -e "${MAP_TGT}" ]] || [[ -L "${MAP_TGT}" ]]; }; } && ORPHANS+=("${rel}")
        fi
    done <"${MANIFEST}"

    if [[ ${#ORPHANS[@]} -eq 0 ]]; then
        log "No orphans found."
        report "orphans: none found"
        return 0
    fi

    local pct=0
    if [[ "${MANIFEST_TOTAL}" -gt 0 ]]; then
        pct=$((${#ORPHANS[@]} * 100 / MANIFEST_TOTAL))
    fi

    log "Found ${#ORPHANS[@]} orphan(s) (repo-managed, source removed from global/harness), ${pct}% of ${MANIFEST_TOTAL} manifest entries:"
    local o
    for o in "${ORPHANS[@]}"; do
        manifest_entry_map "${o}"
        log "  - ${o} -> ${MAP_TGT}"
        report "  - orphan: ${o} -> ${MAP_TGT}"
    done
    report "orphans: ${#ORPHANS[@]} found (${pct}% of ${MANIFEST_TOTAL} manifest entries)"

    if [[ "${APPLY}" -eq 1 && "${KEEP_ORPHANS}" -ne 1 ]]; then
        if [[ "${FORCE_DELETE_ORPHANS}" -ne 1 ]] && { [[ ${#ORPHANS[@]} -gt ${ORPHAN_MAX_ABS} ]] || [[ "${pct}" -ge "${ORPHAN_MAX_PCT}" ]]; }; then
            KEPT_ORPHANS=("${ORPHANS[@]}")
            log "REFUSING to delete: ${#ORPHANS[@]} orphans (${pct}%) exceeds the safety threshold (>${ORPHAN_MAX_ABS} entries or >=${ORPHAN_MAX_PCT}% of the manifest). This looks like a broken checkout (missing global/ or harness/ sources) rather than a routine cleanup. Pass --force-delete-orphans to override if this volume of removal is genuinely expected. Their manifest entries are preserved so a later run can still delete them."
            report "orphans: REFUSED deletion — ${#ORPHANS[@]} orphans (${pct}%) over threshold (>${ORPHAN_MAX_ABS} or >=${ORPHAN_MAX_PCT}%); pass --force-delete-orphans to override (entries preserved in manifest)"
        else
            step_delete_orphans
        fi
    elif [[ "${APPLY}" -eq 1 ]]; then
        KEPT_ORPHANS=("${ORPHANS[@]}")
        log "Not deleting (--keep-orphans). Manifest entries are preserved so a later run can still delete them."
        report "orphans: not deleted (--keep-orphans; entries preserved in manifest)"
    else
        KEPT_ORPHANS=("${ORPHANS[@]}")
        log "Dry run: orphans listed only — an --apply run deletes them (pass --keep-orphans to opt out). Manifest entries are preserved."
        report "orphans: not deleted (dry run; an --apply run deletes them unless --keep-orphans)"
    fi
}

step_delete_orphans() {
    log "== Deleting manifest-confirmed orphans =="
    local rel deleted=0
    for rel in "${ORPHANS[@]}"; do
        manifest_entry_map "${rel}"
        [[ -n "${MAP_TGT}" ]] || continue
        if [[ -e "${MAP_TGT}" ]] || [[ -L "${MAP_TGT}" ]]; then
            rm -rf "${MAP_TGT}"
            log "Deleted orphan: ${MAP_TGT}"
            deleted=$((deleted + 1))
        fi
        case "${rel}" in
            hooks/*)
                purge_hook_block "${CLAUDE_HOME}/settings.json" "$(basename "${rel}")"
                ;;
            codex-hooks/*)
                purge_hook_block "${CODEX_HOME}/hooks.json" "$(basename "${rel}")"
                ;;
        esac
    done
    # Prune now-empty managed directories.
    find "${CLAUDE_HOME}/rules" "${CLAUDE_HOME}/agents" "${CLAUDE_HOME}/skills" "${GROK_RULES_HOME}" -type d -empty -delete 2>/dev/null || true
    report "orphans: ${deleted} deleted"
}

# purge_hook_block SETTINGS_FILE ORPHAN_BASENAME
purge_hook_block() {
    local settings_file="$1" bn="$2"
    [[ -f "${settings_file}" ]] || return 0
    local tmp
    tmp=$(mktemp)
    register_tmp "${tmp}"
    if jq --arg bn "${bn}" -f "${HOOK_PURGE_FILTER}" "${settings_file}" >"${tmp}" && jq empty "${tmp}" 2>/dev/null; then
        mv "${tmp}" "${settings_file}"
        log "${settings_file}: purged orphan hook block for ${bn}"
    else
        rm -f "${tmp}"
        log "WARNING: could not purge ${bn} from ${settings_file} — left unchanged."
    fi
}

# ---------------------------------------------------------------------------
# Steps 6-12 (claude: CLAUDE.md, rules, agents, skills + stale cleanup)
# ---------------------------------------------------------------------------

clean_stale_duplicates() {
    local label="$1" dir="$2"
    [[ -d "${dir}" ]] || return 0
    local cleaned=0
    while IFS= read -r -d '' f; do
        local name
        name=$(basename "${f}")
        if find "${dir}" -mindepth 2 -name "${name}" -print -quit 2>/dev/null | grep -q .; then
            if [[ "${APPLY}" -eq 1 ]]; then
                rm "${f}"
                log "Stale ${label} removed: ${name}"
            else
                log "[DRY-RUN] would remove stale ${label}: ${name}"
            fi
            cleaned=$((cleaned + 1))
        fi
    done < <(find "${dir}" -maxdepth 1 -name '*.md' -print0 2>/dev/null)
    report "${label}: ${cleaned} stale duplicate(s) cleaned"
}

step_deploy_claude() {
    [[ "${RUN_CLAUDE}" -eq 1 ]] || return 0
    log "== Deploy: claude scope =="

    deploy_file "CLAUDE.md" "${REPO_ROOT}/global/CLAUDE.md" "${CLAUDE_HOME}/CLAUDE.md"
    deploy_tree "rules" "${REPO_ROOT}/global/rules" "${CLAUDE_HOME}/rules"
    clean_stale_duplicates "rule" "${CLAUDE_HOME}/rules"
    deploy_tree "agents" "${REPO_ROOT}/global/agents" "${CLAUDE_HOME}/agents"
    clean_stale_duplicates "agent" "${CLAUDE_HOME}/agents"
    deploy_tree "skills" "${REPO_ROOT}/global/skills" "${CLAUDE_HOME}/skills"
    deploy_injected_references

    # Step 12: fallback stale-skill report, report-only, and only meaningful
    # when no manifest existed before this run (manifest-aware orphan
    # detection in step 5 supersedes it from the second deploy onward).
    if [[ ! -f "${MANIFEST}" && -d "${CLAUDE_HOME}/skills" && -d "${REPO_ROOT}/global/skills" ]]; then
        local d name
        for d in "${CLAUDE_HOME}/skills"/*/; do
            [[ -d "${d}" ]] || continue
            name=$(basename "${d}")
            if [[ ! -d "${REPO_ROOT}/global/skills/${name}" ]]; then
                log "NOTE: ${CLAUDE_HOME}/skills/${name} is not in global/skills — possibly manually installed. Not deleted."
            fi
        done
    fi
}

# ---------------------------------------------------------------------------
# Step 13 (hooks + settings.json merge)
# ---------------------------------------------------------------------------

step_deploy_hooks() {
    [[ "${RUN_CLAUDE}" -eq 1 ]] || return 0
    log "== Deploy: hooks (claude) =="

    if [[ -d "${REPO_ROOT}/global/hooks" ]]; then
        local count
        count=$(find "${REPO_ROOT}/global/hooks" -name '*.sh' 2>/dev/null | wc -l | tr -d ' ')
        if [[ "${APPLY}" -eq 1 ]]; then
            mkdir -p "${CLAUDE_HOME}/hooks"
            find "${REPO_ROOT}/global/hooks" -name '*.sh' -exec cp {} "${CLAUDE_HOME}/hooks/" \;
            chmod +x "${CLAUDE_HOME}/hooks/"*.sh 2>/dev/null || true
            # *.json here = runtime config a deployed hook READS from the flat dir
            # (bash-policy.json). Merge INPUTS are excluded like settings-config.json:
            # codex-hooks.json feeds the ~/.codex/hooks.json merge (Codex's native
            # hooks.json discovery — learn.chatgpt.com/docs/hooks); a flattened copy
            # is inert to every harness (Claude registers hooks only via settings.json,
            # no directory auto-scan) and two hooks' files collide on the basename.
            find "${REPO_ROOT}/global/hooks" -name '*.json' ! -name 'settings-config.json' ! -name 'codex-hooks.json' -exec cp {} "${CLAUDE_HOME}/hooks/" \;
            log "Deployed ${count} hook script(s) -> ${CLAUDE_HOME}/hooks"
        else
            log "[DRY-RUN] would deploy ${count} hook script(s) -> ${CLAUDE_HOME}/hooks"
        fi
        report "hooks: ${count} script(s) -> ${CLAUDE_HOME}/hooks"
    fi

    merge_hook_configs \
        "$(find "${REPO_ROOT}/global/hooks" -name settings-config.json 2>/dev/null)" \
        "${CLAUDE_HOME}/settings.json" \
        "claude settings.json"
}

# merge_hook_configs "CONFIG_FILE_LIST(newline)" TARGET_FILE LABEL
merge_hook_configs() {
    local config_list="$1" target="$2" label="$3"
    [[ -n "${config_list}" ]] || return 0

    if [[ "${APPLY}" -eq 0 ]]; then
        local n
        n=$(printf '%s\n' "${config_list}" | grep -c . || true)
        log "[DRY-RUN] would merge hook blocks from ${n} settings-config.json file(s) into ${target}"
        report "${label}: [DRY-RUN] would merge ${n} hook block source(s)"
        return 0
    fi

    [[ -f "${target}" ]] || { mkdir -p "$(dirname "${target}")" && echo '{}' >"${target}"; }

    local managed
    # find -print0 | xargs -0: zsh does NOT word-split an unquoted $var — see
    # SKILL.md's "Shell compatibility" note for why this must stay -print0/xargs -0.
    # `|| managed=""` : a bare assignment whose RHS pipeline fails (e.g. a
    # malformed settings-config.json makes jq exit non-zero) would otherwise
    # abort the whole deploy under set -e. HEAD's documented behavior for a
    # bad hook config is "leave settings.json untouched, fall back to the
    # manual reminder" — not a mid-deploy abort. An empty $managed makes the
    # downstream `jq --argjson m "$managed" ...` fail too, but that failure
    # is inside the if/else below and already degrades to the WARNING path.
    # shellcheck disable=SC2016  # single-quoted jq filter: $c and $e are jq variables, not shell
    managed=$(printf '%s\n' "${config_list}" | tr '\n' '\0' | xargs -0 jq -s \
        'reduce .[] as $c ({}; reduce (($c.hooks // {}) | to_entries[]) as $e (.; .[$e.key] = ((.[$e.key] // []) + $e.value)))') || managed=""

    local tmp
    tmp=$(mktemp)
    register_tmp "${tmp}"
    if jq --argjson m "${managed}" -f "${HOOK_MERGE_FILTER}" "${target}" >"${tmp}" && jq empty "${tmp}" 2>/dev/null; then
        mv "${tmp}" "${target}"
        log "${target}: hook blocks merged (idempotent, additive)."
        report "${label}: merged"
    else
        rm -f "${tmp}"
        log "WARNING: ${target} hook merge failed — left unchanged."
        report "${label}: WARNING merge failed"
    fi
}

# ---------------------------------------------------------------------------
# Step 13b (harness rebuild + universal skills + per-harness agents/commands/rules)
# ---------------------------------------------------------------------------

step_harness_rebuild() {
    # Unconditional — every scope. build.py also assembles the always-on cores
    # (global/CLAUDE.md + harness/AGENTS.md) from global/core-sections/, and this
    # step runs BEFORE step_deploy_claude so the deployed CLAUDE.md is the fresh
    # assembly, not a stale output. Under set -e a failing build (a hand-edited
    # core, or the core-size ratchet refusing unblessed growth) aborts the
    # deploy before anything is copied — and it runs BEFORE step_detect_orphans
    # so the abort also precedes any orphan deletion or settings purge: an
    # aborted run must not have half-mutated the machine.
    log "== Harness rebuild + core assembly =="
    if [[ "${APPLY}" -eq 1 ]]; then
        (cd "${REPO_ROOT}" && python3 harness/build.py)
        local dirty
        dirty=$(cd "${REPO_ROOT}" && git status --porcelain harness/ global/CLAUDE.md 2>/dev/null || true)
        if [[ -n "${dirty}" ]]; then
            log "WARNING: harness/ or global/CLAUDE.md is dirty after rebuild — canonical sources changed without a commit. Deploying anyway; remind the user to commit the diff."
            report "harness rebuild: dirty (uncommitted regenerated output)"
        else
            log "harness/ + cores rebuild clean."
            report "harness rebuild: clean"
        fi
    else
        (cd "${REPO_ROOT}" && python3 harness/build.py --check)
        log "Harness parity check passed."
        report "harness rebuild: parity check passed"
    fi
}

step_deploy_shared_harness() {
    # Universal skills land in ~/.agents/skills (Codex, opencode, Grok all scan it).
    shared_scope_selected || return 0
    log "== Deploy: shared skills owner =="
    local output
    local -a cmd=(python3 "${PI_DEPLOY_HELPER}" shared deploy --repo-root "${REPO_ROOT}" --shared-root "${AGENTS_HOME}" --legacy-manifest "${MANIFEST}")
    if [[ "${APPLY}" -eq 1 ]]; then
        cmd+=(--apply)
    else
        cmd+=(--dry-run)
    fi
    [[ "${VERBOSE}" -eq 1 ]] && cmd+=(--verbose)
    if ! output=$("${cmd[@]}" 2>&1); then
        SHARED_BACKUP_PATH=$(printf '%s\n' "${output}" | sed -n 's/.*after backup \([^:]*\):.*/\1/p' | tail -n 1)
        printf '%s\n' "${output}" >&2
        return 1
    fi
    printf '%s\n' "${output}"
    SHARED_BACKUP_PATH=$(printf '%s\n' "${output}" | sed -n 's/^backup: //p' | tail -n 1)
    report "shared skills deploy: completed${SHARED_BACKUP_PATH:+; backup ${SHARED_BACKUP_PATH}}"
}

deploy_flat_pattern() {
    # deploy_flat_pattern LABEL SRC_DIR TGT_DIR [find-args...]
    local label="$1" src="$2" tgt="$3"
    shift 3
    if [[ ! -d "${src}" ]]; then
        vlog "${label}: source not found, skip"
        return 0
    fi
    local count
    count=$(find "${src}" -type f "$@" 2>/dev/null | wc -l | tr -d ' ')
    [[ "${count}" -gt 0 ]] || return 0
    if [[ "${APPLY}" -eq 1 ]]; then
        mkdir -p "${tgt}"
        find "${src}" -type f "$@" -exec cp {} "${tgt}/" \;
        log "Deployed ${label}: ${count} file(s) -> ${tgt}"
    else
        log "[DRY-RUN] would deploy ${label}: ${count} file(s) -> ${tgt}"
    fi
    report "${label}: ${count} file(s) -> ${tgt}"
}

step_deploy_codex() {
    [[ "${RUN_CODEX}" -eq 1 ]] || return 0
    log "== Deploy: codex scope =="
    deploy_flat_pattern "codex-agents" "${REPO_ROOT}/harness/codex/agents" "${CODEX_HOME}/agents" -name '*.toml'
    deploy_file "harness AGENTS.md -> codex" "${REPO_ROOT}/harness/AGENTS.md" "${CODEX_HOME}/AGENTS.md"

    # Under --apply this checks the just-copied ~/.codex/AGENTS.md (the real
    # deployed state). In dry-run nothing was copied, so checking that path
    # would silently report the OLD file's size — missing exactly the
    # oversized-NEW-file case this check exists to catch. Check the repo
    # source instead when not applying.
    local size_check_file
    if [[ "${APPLY}" -eq 1 ]]; then
        size_check_file="${CODEX_HOME}/AGENTS.md"
    else
        size_check_file="${REPO_ROOT}/harness/AGENTS.md"
    fi
    if [[ -f "${size_check_file}" ]]; then
        local size
        size=$(wc -c <"${size_check_file}" | tr -d ' ')
        if [[ "${size}" -gt "${CODEX_DOC_MAX_BYTES}" ]]; then
            log "WARNING: ${size_check_file} is ${size} bytes, over project_doc_max_bytes (${CODEX_DOC_MAX_BYTES})."
            report "codex AGENTS.md size (${size_check_file}): WARNING ${size} bytes > ${CODEX_DOC_MAX_BYTES}"
        else
            report "codex AGENTS.md size (${size_check_file}): ${size} bytes (OK)"
        fi
    fi

    if [[ "${APPLY}" -eq 1 ]]; then
        if command -v codex >/dev/null 2>&1; then
            if codex --strict-config doctor >/dev/null 2>&1; then
                log "codex config OK"
                report "codex --strict-config doctor: OK"
            else
                log "WARNING: codex --strict-config doctor failed — inspect ~/.codex/config.toml for drift"
                report "codex --strict-config doctor: WARNING (failed)"
            fi
        else
            log "codex CLI not found on PATH — skipping config validation."
            report "codex --strict-config doctor: skipped (codex CLI not found)"
        fi
    else
        log "[DRY-RUN] would run: codex --strict-config doctor"
    fi
}

step_deploy_opencode() {
    [[ "${RUN_OPENCODE}" -eq 1 ]] || return 0
    log "== Deploy: opencode scope =="
    deploy_flat_pattern "opencode-agents" "${REPO_ROOT}/harness/opencode/agents" "${OPENCODE_HOME}/agents" -name '*.md' ! -name 'README.md'
    deploy_flat_pattern "opencode-commands" "${REPO_ROOT}/harness/opencode/commands" "${OPENCODE_HOME}/commands" -name '*.md'
    deploy_flat_pattern "opencode-rules" "${REPO_ROOT}/harness/opencode/rules" "${OPENCODE_HOME}/rules" -name '*.md' ! -name 'README.md'
    deploy_file "harness AGENTS.md -> opencode" "${REPO_ROOT}/harness/AGENTS.md" "${OPENCODE_HOME}/AGENTS.md"
}

# ---------------------------------------------------------------------------
# Step 13d — Codex hooks (register into ~/.codex/hooks.json)
# ---------------------------------------------------------------------------

step_deploy_codex_hooks() {
    [[ "${RUN_CODEX}" -eq 1 ]] || return 0
    log "== Deploy: codex hooks =="

    local found_any=0
    while IFS= read -r cj; do
        [[ -n "${cj}" ]] || continue
        found_any=1
        local hookdir
        hookdir=$(dirname "${cj}")
        local count
        count=$(find "${hookdir}" -name '*.sh' 2>/dev/null | wc -l | tr -d ' ')
        if [[ "${APPLY}" -eq 1 ]]; then
            mkdir -p "${CODEX_HOME}/hooks"
            find "${hookdir}" -name '*.sh' -exec cp {} "${CODEX_HOME}/hooks/" \;
            find "${hookdir}" -name '*.sh' -print0 | while IFS= read -r -d '' s; do
                chmod +x "${CODEX_HOME}/hooks/$(basename "${s}")"
            done
            [[ -f "${CODEX_HOME}/hooks.json" ]] || echo '{}' >"${CODEX_HOME}/hooks.json"
            local managed tmp
            # || managed="": see merge_hook_configs — a bare failing assignment
            # would abort the deploy under set -e instead of degrading to the
            # WARNING branch below.
            managed=$(jq '.hooks // {}' "${cj}") || managed=""
            tmp=$(mktemp)
            register_tmp "${tmp}"
            if jq --argjson m "${managed}" -f "${HOOK_MERGE_FILTER}" "${CODEX_HOME}/hooks.json" >"${tmp}" && jq empty "${tmp}" 2>/dev/null; then
                mv "${tmp}" "${CODEX_HOME}/hooks.json"
                log "${CODEX_HOME}/hooks.json: $(basename "${hookdir}") merged (idempotent, additive)."
                report "codex hooks (${hookdir##*/}): merged, ${count} script(s)"
            else
                rm -f "${tmp}"
                log "WARNING: ${CODEX_HOME}/hooks.json merge failed for $(basename "${hookdir}") — left unchanged."
                report "codex hooks (${hookdir##*/}): WARNING merge failed"
            fi
        else
            log "[DRY-RUN] would deploy ${count} script(s) from ${hookdir} -> ${CODEX_HOME}/hooks, and merge into ${CODEX_HOME}/hooks.json"
            report "codex hooks (${hookdir##*/}): [DRY-RUN]"
        fi
    done < <(find "${REPO_ROOT}/global/hooks" -name codex-hooks.json 2>/dev/null)

    if [[ "${found_any}" -eq 1 ]]; then
        log "Codex trust caveat: a merged hook stays inert until its [hooks.state] slot in ~/.codex/config.toml is present and enabled=true (accepted via the Codex trust prompt, or set manually). This script never edits config.toml."
    fi
}

# ---------------------------------------------------------------------------
# Step 13e — opencode session plugin (auto-loaded by file presence)
# ---------------------------------------------------------------------------

step_deploy_opencode_plugin() {
    [[ "${RUN_OPENCODE}" -eq 1 ]] || return 0
    local src="${REPO_ROOT}/global/hooks/flow-session-context/flow-session-context.ts"
    [[ -f "${src}" ]] || return 0
    log "== Deploy: opencode plugin =="
    if [[ "${APPLY}" -eq 1 ]]; then
        mkdir -p "${OPENCODE_HOME}/plugins"
        cp "${src}" "${OPENCODE_HOME}/plugins/"
        log "opencode plugin deployed: flow-session-context.ts"
        report "opencode plugin: deployed"
    else
        log "[DRY-RUN] would deploy opencode plugin: flow-session-context.ts -> ${OPENCODE_HOME}/plugins/"
        report "opencode plugin: [DRY-RUN]"
    fi
}

# ---------------------------------------------------------------------------
# Step 13f — opencode permission block (surgical, additive merge)
#
# The state-fetcher agent reads flow-core references from ~/.claude/skills, which
# opencode treats as an external directory and gates behind an interactive prompt —
# so status-fetch cannot run unattended there. This merges only the repo-managed
# permission keys; every other key in opencode.json is left untouched, and a rule
# the user already set for the same pattern WINS over the repo's (right operand of
# `+`). Leading `~` in the source is expanded to $HOME at merge time.
#
# jsonc is skipped deliberately: jq cannot round-trip comments, and silently
# stripping the user's comments is worse than reporting the gap.
# ---------------------------------------------------------------------------

step_merge_opencode_permissions() {
    [[ "${RUN_OPENCODE}" -eq 1 ]] || return 0
    local src="${REPO_ROOT}/harness/opencode/permission-config.json"
    [[ -f "${src}" ]] || return 0

    local target="${OPENCODE_HOME}/opencode.json"
    if [[ ! -f "${target}" && -f "${OPENCODE_HOME}/opencode.jsonc" ]]; then
        log "WARNING: only opencode.jsonc found — permission merge skipped (jq cannot preserve comments)."
        report "opencode permissions: WARNING skipped (jsonc, merge by hand)"
        return 0
    fi

    log "== Merge: opencode permission block =="
    if [[ "${APPLY}" -eq 0 ]]; then
        log "[DRY-RUN] would merge repo-managed permission keys into ${target}"
        report "opencode permissions: [DRY-RUN] would merge"
        return 0
    fi

    [[ -f "${target}" ]] || { mkdir -p "${OPENCODE_HOME}" && echo '{}' >"${target}"; }

    local tmp
    tmp=$(mktemp)
    register_tmp "${tmp}"
    # A repo rule is added ONLY when the user has not already declared that pattern
    # (compared with ~ expanded on both sides, so `~/x` and `/Users/me/x` count as the
    # same rule). The user's literal spelling is never rewritten. `permission` set to a
    # bare action string is a valid shape the schema allows — leave it alone entirely.
    # shellcheck disable=SC2016  # single-quoted jq filter: $p/$home/$cat are jq variables, not shell
    if jq --slurpfile p "${src}" --arg home "${HOME}" '
            if ((.permission // {}) | type) != "object" then .
            else
                reduce (($p[0].permission // {}) | to_entries[]) as $cat (.;
                    .permission[$cat.key] =
                        (if ((.permission[$cat.key] // {}) | type) == "object"
                         then ((.permission[$cat.key] // {}) | keys | map(sub("^~"; $home))) as $have
                              | (($cat.value
                                  | with_entries(.key |= sub("^~"; $home))
                                  | with_entries(. as $e | select(($have | index($e.key)) == null)))
                                 + (.permission[$cat.key] // {}))
                         else .permission[$cat.key] end))
            end
        ' "${target}" >"${tmp}" && jq empty "${tmp}" 2>/dev/null; then
        mv "${tmp}" "${target}"
        log "${target}: permission block merged (idempotent, additive; user rules win)."
        report "opencode permissions: merged"
    else
        rm -f "${tmp}"
        log "WARNING: ${target} permission merge failed — left unchanged."
        report "opencode permissions: WARNING merge failed"
    fi
}

# ---------------------------------------------------------------------------
# Grok scope — flat rule symlinks into ~/.grok/rules + agents into ~/.grok/agents.
#
# Rule links point at the DEPLOYED rule under ~/.claude/rules, not at the repo:
# both harnesses then read the same bytes, a redeploy updates them at once,
# and a checkout on another branch never silently changes what Grok loads.
# Agents are real files (Grok frontmatter differs from Claude).
# ---------------------------------------------------------------------------

step_deploy_grok() {
    [[ "${RUN_GROK}" -eq 1 ]] || return 0
    log "== Deploy: grok scope (flat rule symlinks + agents) =="

    local new=0 relinked=0 unchanged=0 undeployed=0
    local rel flat src tgt current
    while IFS= read -r rel; do
        [[ -n "${rel}" ]] || continue
        flat="${rel//\//__}"
        src="${CLAUDE_HOME}/rules/${rel}"
        tgt="${GROK_RULES_HOME}/${flat}"

        if [[ ! -f "${src}" ]]; then
            undeployed=$((undeployed + 1))
            log "  WARNING: ${rel} is not deployed under ${CLAUDE_HOME}/rules — link skipped. Run the claude scope first."
            continue
        fi

        current=""
        [[ -L "${tgt}" ]] && current=$(readlink "${tgt}")
        if [[ "${current}" == "${src}" ]]; then
            unchanged=$((unchanged + 1))
            vlog "  unchanged: ${flat}"
            continue
        fi
        if [[ -e "${tgt}" || -L "${tgt}" ]]; then
            relinked=$((relinked + 1))
        else
            new=$((new + 1))
        fi

        if [[ "${APPLY}" -eq 1 ]]; then
            mkdir -p "${GROK_RULES_HOME}"
            ln -sfn "${src}" "${tgt}"
            vlog "  linked: ${flat} -> ${src}"
        else
            vlog "  [DRY-RUN] would link: ${flat} -> ${src}"
        fi
    done < <(grok_always_on_rules)

    if [[ "${undeployed}" -gt 0 ]]; then
        report "grok-rules: WARNING — ${undeployed} rule(s) not deployed under ${CLAUDE_HOME}/rules, links skipped"
    fi
    if [[ "${APPLY}" -eq 1 ]]; then
        log "Grok rules: ${new} linked, ${relinked} relinked, ${unchanged} already current -> ${GROK_RULES_HOME}"
    else
        log "[DRY-RUN] Grok rules: would link ${new}, relink ${relinked}; ${unchanged} already current -> ${GROK_RULES_HOME}"
    fi
    report "grok-rules: ${new} linked, ${relinked} relinked, ${unchanged} current -> ${GROK_RULES_HOME}"

    # Generated Grok agent definitions (frontmatter differs from Claude — real files).
    deploy_flat_pattern "grok-agents" "${REPO_ROOT}/harness/grok/agents" "${GROK_AGENTS_HOME}" -name '*.md' ! -name 'README.md'
}

# ---------------------------------------------------------------------------
# Step 14 — write manifest. Scope-aware merge: entries for scopes NOT active
# this run are preserved verbatim from the previous manifest instead of being
# dropped, so a partial (--only) run never blinds orphan detection for the
# scopes it didn't touch. See report for why this deviates from a literal
# "always regenerate everything" reading of the source procedure.
# ---------------------------------------------------------------------------

step_write_manifest() {
    log "== Manifest =="
    if [[ "${RUN_CLAUDE}" -eq 0 && "${RUN_CODEX}" -eq 0 && "${RUN_OPENCODE}" -eq 0 && "${RUN_GROK}" -eq 0 ]]; then
        log "PI-only run: leaving legacy ${MANIFEST} untouched; shared skills use the neutral manifest."
        report "manifest: legacy manifest untouched (PI-only)"
        return 0
    fi
    if [[ "${APPLY}" -eq 0 ]]; then
        log "[DRY-RUN] would write ${MANIFEST}"
        report "manifest: [DRY-RUN] not written"
        return 0
    fi

    local tmp
    tmp=$(mktemp)
    register_tmp "${tmp}"

    {
        echo "# deploy-global manifest — files managed by tricell-hive. Do not edit by hand."
        echo "# deployed_at: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
        echo "# source_commit: $(cd "${REPO_ROOT}" && git rev-parse --short HEAD 2>/dev/null || echo unknown)"

        # Preserve entries for scopes not active this run.
        if [[ -f "${MANIFEST}" ]]; then
            local rel scope
            while IFS= read -r rel; do
                case "${rel}" in '#'* | '') continue ;; esac
                scope=$(manifest_entry_scope "${rel}")
                case "${scope}" in
                    claude) [[ "${RUN_CLAUDE}" -eq 1 ]] && continue ;;
                    shared)
                        if [[ "${RUN_CLAUDE}" -eq 1 && "${RUN_CODEX}" -eq 1 && "${RUN_OPENCODE}" -eq 1 && "${RUN_GROK}" -eq 1 && "${RUN_PI}" -eq 1 ]]; then
                            continue
                        fi
                        echo "${rel}"
                        continue
                        ;;
                    codex) [[ "${RUN_CODEX}" -eq 1 ]] && continue ;;
                    opencode) [[ "${RUN_OPENCODE}" -eq 1 ]] && continue ;;
                    grok) [[ "${RUN_GROK}" -eq 1 ]] && continue ;;
                    unknown) continue ;;
                esac
                echo "${rel}"
            done <"${MANIFEST}"
        fi

        # Preserve entries for orphans detected but NOT deleted this run
        # (no --delete-orphans, or the size breaker refused). Their sources
        # are gone, so the fresh scans below will never re-list them; without
        # this block the stale deployed files become untracked forever and no
        # later --delete-orphans run can find them. Guarded expansion: the
        # array is usually empty and set -u would trip on bash < 4.4.
        local kept
        for kept in ${KEPT_ORPHANS[@]+"${KEPT_ORPHANS[@]}"}; do
            echo "${kept}"
        done

        # Each pipeline below is guarded with `|| true`: under pipefail, `find`
        # returning non-zero because one of several starting paths is missing
        # (e.g. global/skills absent) would otherwise abort the whole manifest
        # write via set -e, even though the pipeline's actual output (from the
        # paths that DO exist) already reached stdout.
        if [[ "${RUN_CLAUDE}" -eq 1 ]]; then
            [[ -f "${REPO_ROOT}/global/CLAUDE.md" ]] && echo "CLAUDE.md"
            find "${REPO_ROOT}/global/rules" "${REPO_ROOT}/global/agents" "${REPO_ROOT}/global/skills" -type f 2>/dev/null | sed "s|^${REPO_ROOT}/global/||" || true
            find "${REPO_ROOT}/global/hooks" -name '*.sh' -print0 2>/dev/null | xargs -0 -n1 basename 2>/dev/null | sed 's|^|hooks/|' || true
            # Injected references are deployed into ~/.claude/skills by
            # deploy_injected_references but exist ONLY in the generated tree, so
            # the walk above never saw them. Same skip as that function: a file
            # the canonical global/ copy already provides is not ours to track here.
            local inj_dir inj_skill inj_file inj_base
            for inj_dir in "${REPO_ROOT}/harness/agents-skills"/*/references/; do
                [[ -d "${inj_dir}" ]] || continue
                inj_skill=$(basename "$(dirname "${inj_dir}")")
                for inj_file in "${inj_dir}"*; do
                    [[ -f "${inj_file}" ]] || continue
                    inj_base=$(basename "${inj_file}")
                    [[ -f "${REPO_ROOT}/global/skills/${inj_skill}/references/${inj_base}" ]] && continue
                    echo "skills/${inj_skill}/references/${inj_base}"
                done
            done
        fi
        if [[ "${RUN_CODEX}" -eq 1 ]]; then
            find "${REPO_ROOT}/global/agents" -name '*.md' -print0 2>/dev/null | xargs -0 -n1 basename 2>/dev/null | sed 's|\.md$|.toml|; s|^|codex-agents/|' || true
            [[ -f "${REPO_ROOT}/harness/AGENTS.md" ]] && echo "harness-agents/codex/AGENTS.md"
            find "${REPO_ROOT}/global/hooks" -name codex-hooks.json 2>/dev/null | while IFS= read -r cj; do
                find "$(dirname "${cj}")" -name '*.sh' -print0 2>/dev/null | xargs -0 -n1 basename 2>/dev/null | sed 's|^|codex-hooks/|'
            done || true
        fi
        if [[ "${RUN_OPENCODE}" -eq 1 ]]; then
            find "${REPO_ROOT}/global/agents" -name '*.md' -print0 2>/dev/null | xargs -0 -n1 basename 2>/dev/null | sed 's|^|opencode-agents/|' || true
            find "${REPO_ROOT}/harness/opencode/commands" -name '*.md' -print0 2>/dev/null | xargs -0 -n1 basename 2>/dev/null | sed 's|^|opencode-commands/|' || true
            [[ -f "${REPO_ROOT}/harness/AGENTS.md" ]] && echo "harness-agents/opencode/AGENTS.md"
            find "${REPO_ROOT}/global/hooks" -name '*.ts' -print0 2>/dev/null | xargs -0 -n1 basename 2>/dev/null | sed 's|^|opencode-plugins/|' || true
        fi
        if [[ "${RUN_GROK}" -eq 1 ]]; then
            grok_always_on_rules | sed 's|/|__|g; s|^|grok-rules/|' || true
            find "${REPO_ROOT}/global/agents" -name '*.md' -print0 2>/dev/null | xargs -0 -n1 basename 2>/dev/null | sed 's|^|grok-agents/|' || true
        fi
    } >"${tmp}"

    # --only codex / --only opencode / --only harness on a fresh machine
    # never creates ~/.claude (nothing in those scopes touches it) — but the
    # manifest always lives there. Ensure it exists before the move.
    mkdir -p "${CLAUDE_HOME}"
    mv "${tmp}" "${MANIFEST}"
    local n
    n=$(grep -vc '^#' "${MANIFEST}" || true)
    log "Manifest written: ${MANIFEST} (${n} entries)"
    report "manifest: written, ${n} entries"
}

# ---------------------------------------------------------------------------
# Final report (stdout)
# ---------------------------------------------------------------------------

step_final_report() {
    echo "===================================================================="
    echo "deploy-global.sh report — $(date -u +%Y-%m-%dT%H:%M:%SZ)"
    echo "mode: $([[ ${APPLY} -eq 1 ]] && echo APPLY || echo DRY-RUN) | scopes: claude=${RUN_CLAUDE} codex=${RUN_CODEX} opencode=${RUN_OPENCODE} grok=${RUN_GROK} pi=${RUN_PI}"
    echo "===================================================================="
    if [[ -s "${REPORT_LOG}" ]]; then
        cat "${REPORT_LOG}"
    fi
    if [[ -n "${BACKUP_FILE}" && "${APPLY}" -eq 1 ]]; then
        echo ""
        echo "Restore with:"
        echo "  tar -xzf ${BACKUP_FILE} -C ${CLAUDE_HOME}/"
    fi
    if [[ -n "${HARNESS_BACKUP_FILE}" && "${APPLY}" -eq 1 ]]; then
        echo ""
        echo "Restore harness config with (per file, they have different roots):"
        if [[ "${RUN_CODEX}" -eq 1 ]]; then
            echo "  tar -xzf ${HARNESS_BACKUP_FILE} -C ${CODEX_HOME} hooks.json"
        fi
        if [[ "${RUN_OPENCODE}" -eq 1 ]]; then
            echo "  tar -xzf ${HARNESS_BACKUP_FILE} -C ${OPENCODE_HOME} opencode.json"
        fi
    fi
    if [[ -n "${PI_BACKUP_PATH}" && "${APPLY}" -eq 1 ]]; then
        echo ""
        echo "PI rollback:"
        echo "  python3 ${PI_DEPLOY_HELPER} rollback --pi-dir \"${PI_AGENT_DIR}\" --backup-dir \"${PI_BACKUP_PATH}\" --apply"
    fi
    if [[ -n "${SHARED_BACKUP_PATH}" && "${APPLY}" -eq 1 ]]; then
        echo ""
        echo "Shared-skills rollback:"
        echo "  python3 ${PI_DEPLOY_HELPER} shared rollback --shared-root \"${AGENTS_HOME}\" --backup-dir \"${SHARED_BACKUP_PATH}\" --apply"
    fi
    if [[ "${RUN_CODEX}" -eq 1 || "${RUN_OPENCODE}" -eq 1 ]]; then
        echo ""
        echo "Reminder: harness/{codex,opencode}/*.snippet config merges (plugin/hook"
        echo "registration in opencode.jsonc / config.toml) are ONE-TIME and MANUAL — this"
        echo "deploy writes no part of those files except the opencode permission keys"
        echo "(harness/opencode/permission-config.json). See harness/{codex,opencode}/README.md."
        echo "On a fresh machine, hooks and rules land on disk with no prompt from the tool"
        echo "itself to register them — this reminder is that prompt."
    fi
    echo ""
    echo "Reminder: restart Claude Code / Codex / opencode / Grok / PI (or open a new session) to reload selected scopes."
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

main() {
    parse_args "$@"
    check_deps
    [[ "${RUN_PI}" -eq 1 ]] && check_pi_deps

    REPORT_LOG=$(mktemp)
    register_tmp "${REPORT_LOG}"

    log "deploy-global.sh — mode: $([[ ${APPLY} -eq 1 ]] && echo APPLY || echo DRY-RUN)"
    log "scopes — claude:${RUN_CLAUDE} codex:${RUN_CODEX} opencode:${RUN_OPENCODE} grok:${RUN_GROK} pi:${RUN_PI}"

    step_harness_rebuild
    step_preflight_selected
    step_preview
    step_diff
    step_backup
    step_backup_harness_config
    step_detect_orphans
    step_deploy_pi
    step_deploy_claude
    step_deploy_hooks
    step_deploy_shared_harness
    step_deploy_codex
    step_deploy_opencode
    step_deploy_codex_hooks
    step_deploy_opencode_plugin
    step_merge_opencode_permissions
    step_deploy_grok
    step_write_manifest

    step_final_report
}

main "$@"
