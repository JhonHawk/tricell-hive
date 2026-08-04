#!/usr/bin/env bash
# deploy-global.sh — deploys tricell-hive's global/ (+ the multi-harness layer
# under harness/) to ~/.claude, ~/.codex, ~/.config/opencode, and ~/.agents/skills.
#
# Faithful port of the procedure documented in ../SKILL.md. Read that file for
# the "why" of each step; this script is the "how".
#
# Safety model:
#   - Defaults to --dry-run. Nothing is written anywhere until --apply is passed.
#   - Even under --apply, manifest-based orphan deletion (and the settings.json /
#     hooks.json purge that rides with it) requires the separate --delete-orphans
#     flag. The script never auto-confirms a delete on its own.
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
readonly AGENTS_SKILLS_HOME="${HOME}/.agents/skills"
# Grok relocates its config root via GROK_HOME (documented in its user guide);
# honor it so a relocated install is not silently deployed to ~/.grok.
readonly GROK_HOME="${GROK_HOME:-${HOME}/.grok}"
readonly GROK_RULES_HOME="${GROK_HOME}/rules"
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
DELETE_ORPHANS=0
FORCE_DELETE_ORPHANS=0
RUN_CLAUDE=1
RUN_CODEX=1
RUN_OPENCODE=1
RUN_GROK=1

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
                    [--delete-orphans] [--force-delete-orphans] [--verbose] [--help]

FLAGS:
  --dry-run          Report what would change; write nothing. DEFAULT.
  --apply            Perform the deploy for real (copies, backup, manifest write).
                      Does NOT by itself delete orphaned files — see --delete-orphans.
  --only SCOPE       Restrict the run to one or more scopes. Repeatable, or
                      comma-separated. One of:
                        claude    -> global/ into ~/.claude
                        codex     -> harness/ codex-specific targets (~/.codex)
                        opencode  -> harness/ opencode-specific targets (~/.config/opencode)
                        grok      -> always-on rules symlinked into ~/.grok/rules
                        harness   -> alias for "codex,opencode,grok"
                        all       -> everything (default when --only is omitted)
  --delete-orphans   Under --apply, actually delete manifest-confirmed orphans
                      (and purge their settings.json/hooks.json entries).
                      Without this flag, orphans are only listed, never removed —
                      this is the confirmation gate the original procedure asks
                      the user for; the script never auto-confirms it. Refused
                      (not deleted) when the orphan set exceeds 20 entries or 25%
                      of the manifest — see --force-delete-orphans.
  --force-delete-orphans
                      Implies --delete-orphans and bypasses the size-based
                      refusal above. Only pass this when a large orphan set is
                      genuinely expected.
  --verbose, -v      Print every file-level action, not just per-category summaries.
  --help, -h         This message.

SCOPE NOTES:
  Backups (tar snapshot of ~/.claude before overwrite) only ever cover
  ~/.claude — there is no equivalent snapshot mechanism for ~/.codex or
  ~/.config/opencode in the source procedure, so the backup step is skipped
  entirely when the "claude" scope is not active.

  The "grok" scope writes only symlinks (no file content), each pointing at the
  deployed rule under ~/.claude/rules — so it needs the "claude" scope to have
  run at least once, and removing a link never touches a rule. Grok's rules
  discovery is NOT recursive and does NOT honor `paths:`, so only always-on
  rules (those without `paths:`) are linked, flattened as `<dir>__<file>.md`.
  Path-scoped rules reach Grok through the router skills instead.

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
            --delete-orphans)
                DELETE_ORPHANS=1
                shift
                ;;
            --force-delete-orphans)
                DELETE_ORPHANS=1
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
                harness)
                    RUN_CODEX=1
                    RUN_OPENCODE=1
                    RUN_GROK=1
                    ;;
                all)
                    RUN_CLAUDE=1
                    RUN_CODEX=1
                    RUN_OPENCODE=1
                    RUN_GROK=1
                    ;;
                *) die "Unknown --only scope: ${s} (expected claude|codex|opencode|grok|harness|all)" ;;
            esac
        done
    fi
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
    if [[ "${RUN_CODEX}" -eq 1 || "${RUN_OPENCODE}" -eq 1 ]]; then
        command -v python3 >/dev/null 2>&1 || missing+=("python3 (needed to rebuild harness/)")
    fi
    if [[ ${#missing[@]} -gt 0 ]]; then
        die "Missing required tool(s): ${missing[*]}"
    fi
    [[ -f "${HOOK_MERGE_FILTER}" ]] || die "Missing bundled filter: ${HOOK_MERGE_FILTER}"
    [[ -f "${HOOK_PURGE_FILTER}" ]] || die "Missing bundled filter: ${HOOK_PURGE_FILTER}"
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
    done < <(find "${REPO_ROOT}/global/hooks" \( -name '*.sh' -o \( -name '*.json' ! -name 'settings-config.json' \) \) -print0 2>/dev/null)
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
    if [[ "${RUN_CODEX}" -eq 1 || "${RUN_OPENCODE}" -eq 1 ]]; then
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
    if [[ "${RUN_CODEX}" -eq 1 || "${RUN_OPENCODE}" -eq 1 ]]; then
        # harness/ is only rebuilt under --apply (step_harness_rebuild), and
        # that rebuild runs AFTER this diff. In dry-run — and even under
        # --apply, since this diff runs before the rebuild — the lines below
        # compare against whatever harness/ currently holds on disk (last
        # commit/build), NOT what `python3 harness/build.py` would produce if
        # canonical sources (global/agents, global/skills) changed since.
        # Deliberately not fixed by moving the rebuild earlier: that would
        # make dry-run mutate files under harness/ (even though repo-internal,
        # it breaks dry-run's "writes nothing, ever" guarantee, and risks a
        # collision if another process is rebuilding the same tree concurrently).
        # SKILL.md's "Before running" section is the compensating control:
        # rebuild + commit harness/ before invoking this script at all.
        log "  (harness/ diff below is pre-rebuild — see SKILL.md 'Before running')"
        diff_category "agents-skills (pre-rebuild)" "${REPO_ROOT}/harness/agents-skills" "${AGENTS_SKILLS_HOME}"
    fi
    if [[ "${RUN_CODEX}" -eq 1 ]]; then
        diff_category "codex-agents (pre-rebuild)" "${REPO_ROOT}/harness/codex/agents" "${CODEX_HOME}/agents" -name '*.toml'
        diff_file "harness AGENTS.md -> codex (pre-rebuild)" "${REPO_ROOT}/harness/AGENTS.md" "${CODEX_HOME}/AGENTS.md"
    fi
    if [[ "${RUN_OPENCODE}" -eq 1 ]]; then
        diff_category "opencode-agents (pre-rebuild)" "${REPO_ROOT}/harness/opencode/agents" "${OPENCODE_HOME}/agents" -name '*.md' ! -name 'README.md'
        diff_category "opencode-commands (pre-rebuild)" "${REPO_ROOT}/harness/opencode/commands" "${OPENCODE_HOME}/commands" -name '*.md'
        diff_category "opencode-rules (pre-rebuild)" "${REPO_ROOT}/harness/opencode/rules" "${OPENCODE_HOME}/rules" -name '*.md' ! -name 'README.md'
        diff_file "harness AGENTS.md -> opencode (pre-rebuild)" "${REPO_ROOT}/harness/AGENTS.md" "${OPENCODE_HOME}/AGENTS.md"
    fi
    if [[ "${RUN_GROK}" -eq 1 ]]; then
        diff_grok_rules
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
# Step 4 (backup) — ~/.claude only; scoped to RUN_CLAUDE.
# ---------------------------------------------------------------------------

BACKUP_FILE=""

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
        grok-rules/*) echo "grok" ;;
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
        CLAUDE.md | rules/* | agents/* | skills/*)
            MAP_SRC="${REPO_ROOT}/global/${rel}"
            MAP_TGT="${CLAUDE_HOME}/${rel}"
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
        *)
            : # unrecognized prefix — MAP_SRC/MAP_TGT stay empty, never touched
            ;;
    esac
}

ORPHANS=()
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
        MANIFEST_TOTAL=$((MANIFEST_TOTAL + 1))
        scope=$(manifest_entry_scope "${rel}")
        case "${scope}" in
            claude) [[ "${RUN_CLAUDE}" -eq 1 ]] || continue ;;
            shared) [[ "${RUN_CODEX}" -eq 1 || "${RUN_OPENCODE}" -eq 1 ]] || continue ;;
            codex) [[ "${RUN_CODEX}" -eq 1 ]] || continue ;;
            opencode) [[ "${RUN_OPENCODE}" -eq 1 ]] || continue ;;
            grok) [[ "${RUN_GROK}" -eq 1 ]] || continue ;;
            *) continue ;; # unknown prefix — never touch
        esac
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

    if [[ "${APPLY}" -eq 1 && "${DELETE_ORPHANS}" -eq 1 ]]; then
        if [[ "${FORCE_DELETE_ORPHANS}" -ne 1 ]] && { [[ ${#ORPHANS[@]} -gt ${ORPHAN_MAX_ABS} ]] || [[ "${pct}" -ge "${ORPHAN_MAX_PCT}" ]]; }; then
            log "REFUSING to delete: ${#ORPHANS[@]} orphans (${pct}%) exceeds the safety threshold (>${ORPHAN_MAX_ABS} entries or >=${ORPHAN_MAX_PCT}% of the manifest). This looks like a broken checkout (missing global/ or harness/ sources) rather than a routine cleanup. Pass --force-delete-orphans to override if this volume of removal is genuinely expected."
            report "orphans: REFUSED deletion — ${#ORPHANS[@]} orphans (${pct}%) over threshold (>${ORPHAN_MAX_ABS} or >=${ORPHAN_MAX_PCT}%); pass --force-delete-orphans to override"
        else
            step_delete_orphans
        fi
    else
        log "Not deleting (pass --apply --delete-orphans to confirm removal)."
        report "orphans: not deleted (requires --apply --delete-orphans)"
    fi
}

step_delete_orphans() {
    log "== Deleting confirmed orphans (--delete-orphans) =="
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
    find "${CLAUDE_HOME}/rules" "${CLAUDE_HOME}/agents" "${CLAUDE_HOME}/skills" "${AGENTS_SKILLS_HOME}" "${GROK_RULES_HOME}" -type d -empty -delete 2>/dev/null || true
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
            find "${REPO_ROOT}/global/hooks" -name '*.json' ! -name 'settings-config.json' -exec cp {} "${CLAUDE_HOME}/hooks/" \;
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
    [[ "${RUN_CODEX}" -eq 1 || "${RUN_OPENCODE}" -eq 1 ]] || return 0
    log "== Harness rebuild =="
    if [[ "${APPLY}" -eq 1 ]]; then
        (cd "${REPO_ROOT}" && python3 harness/build.py)
        local dirty
        dirty=$(cd "${REPO_ROOT}" && git status --porcelain harness/ 2>/dev/null || true)
        if [[ -n "${dirty}" ]]; then
            log "WARNING: harness/ is dirty after rebuild — canonical sources changed without a commit. Deploying anyway; remind the user to commit the diff."
            report "harness rebuild: dirty (uncommitted regenerated output)"
        else
            log "harness/ rebuild clean."
            report "harness rebuild: clean"
        fi
    else
        log "[DRY-RUN] would run: python3 harness/build.py (then check harness/ for a dirty diff)"
        report "harness rebuild: [DRY-RUN]"
    fi
}

step_deploy_shared_harness() {
    [[ "${RUN_CODEX}" -eq 1 || "${RUN_OPENCODE}" -eq 1 ]] || return 0
    deploy_tree "agents-skills (universal skills)" "${REPO_ROOT}/harness/agents-skills" "${AGENTS_SKILLS_HOME}"
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
# Step 13c — Engram #555 detect_project hotfix re-assertion (codex-scoped, temporary)
# ---------------------------------------------------------------------------

step_engram_hotfix() {
    [[ "${RUN_CODEX}" -eq 1 ]] || return 0
    log "== Engram #555 hotfix re-assertion =="

    local ensure_script=""
    if [[ -x "${HOME}/.local/bin/ensure-engram-555" ]]; then
        ensure_script="${HOME}/.local/bin/ensure-engram-555"
    elif [[ -x "${HOME}/.engram/hotfix-555/ensure-engram-555.sh" ]]; then
        ensure_script="${HOME}/.engram/hotfix-555/ensure-engram-555.sh"
    elif [[ -x "${REPO_ROOT}/_support/archive/2026-07-15-engram-555-hotfix/ensure-engram-555.sh" ]]; then
        if [[ "${APPLY}" -eq 1 ]]; then
            # Bare (unwrapped) command would abort the whole deploy under
            # set -e on failure. Step 13c's documented contract is "a failure
            # here is not a deploy abort" — wrap it like its `apply` sibling
            # below.
            if "${REPO_ROOT}/_support/archive/2026-07-15-engram-555-hotfix/ensure-engram-555.sh" install; then
                ensure_script="${HOME}/.local/bin/ensure-engram-555"
            else
                log "WARNING: engram #555 hotfix kit install failed — skipping hotfix re-assertion this run."
                report "engram #555 hotfix: WARNING (install failed)"
                return 0
            fi
        else
            log "[DRY-RUN] would install the engram #555 hotfix kit from _support/archive/2026-07-15-engram-555-hotfix/"
            report "engram #555 hotfix: [DRY-RUN] would install + apply"
            return 0
        fi
    fi

    if [[ -n "${ensure_script}" ]]; then
        if [[ "${APPLY}" -eq 1 ]]; then
            if "${ensure_script}" apply; then
                log "engram #555 hotfix: OK"
                report "engram #555 hotfix: OK"
            else
                log "WARNING: engram #555 ensure failed — memories may split across project aliases until fixed"
                report "engram #555 hotfix: WARNING (ensure failed)"
            fi
        else
            log "[DRY-RUN] would run: ${ensure_script} apply"
            report "engram #555 hotfix: [DRY-RUN] would apply"
        fi
    else
        log "WARNING: engram #555 ensure script not found — see _support/archive/2026-07-15-engram-555-hotfix/"
        report "engram #555 hotfix: WARNING (ensure script not found)"
    fi
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
# Grok scope — flat symlinks into ~/.grok/rules.
#
# Links point at the DEPLOYED rule under ~/.claude/rules, not at the repo:
# both harnesses then read the same bytes, a redeploy updates them at once,
# and a checkout on another branch never silently changes what Grok loads.
# ---------------------------------------------------------------------------

step_deploy_grok() {
    [[ "${RUN_GROK}" -eq 1 ]] || return 0
    log "== Deploy: grok scope (flat rule symlinks) =="

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
                    shared) [[ "${RUN_CODEX}" -eq 1 || "${RUN_OPENCODE}" -eq 1 ]] && continue ;;
                    codex) [[ "${RUN_CODEX}" -eq 1 ]] && continue ;;
                    opencode) [[ "${RUN_OPENCODE}" -eq 1 ]] && continue ;;
                    grok) [[ "${RUN_GROK}" -eq 1 ]] && continue ;;
                    unknown) continue ;;
                esac
                echo "${rel}"
            done <"${MANIFEST}"
        fi

        # Each pipeline below is guarded with `|| true`: under pipefail, `find`
        # returning non-zero because one of several starting paths is missing
        # (e.g. global/skills absent) would otherwise abort the whole manifest
        # write via set -e, even though the pipeline's actual output (from the
        # paths that DO exist) already reached stdout.
        if [[ "${RUN_CLAUDE}" -eq 1 ]]; then
            [[ -f "${REPO_ROOT}/global/CLAUDE.md" ]] && echo "CLAUDE.md"
            find "${REPO_ROOT}/global/rules" "${REPO_ROOT}/global/agents" "${REPO_ROOT}/global/skills" -type f 2>/dev/null | sed "s|^${REPO_ROOT}/global/||" || true
            find "${REPO_ROOT}/global/hooks" -name '*.sh' -print0 2>/dev/null | xargs -0 -n1 basename 2>/dev/null | sed 's|^|hooks/|' || true
        fi
        if [[ "${RUN_CODEX}" -eq 1 || "${RUN_OPENCODE}" -eq 1 ]]; then
            find "${REPO_ROOT}/global/skills" -type f 2>/dev/null | sed "s|^${REPO_ROOT}/global/skills/|agents-skills/|" || true
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
    echo "mode: $([[ ${APPLY} -eq 1 ]] && echo APPLY || echo DRY-RUN) | scopes: claude=${RUN_CLAUDE} codex=${RUN_CODEX} opencode=${RUN_OPENCODE} grok=${RUN_GROK}"
    echo "===================================================================="
    if [[ -s "${REPORT_LOG}" ]]; then
        cat "${REPORT_LOG}"
    fi
    if [[ -n "${BACKUP_FILE}" && "${APPLY}" -eq 1 ]]; then
        echo ""
        echo "Restore with:"
        echo "  tar -xzf ${BACKUP_FILE} -C ${CLAUDE_HOME}/"
    fi
    if [[ "${RUN_CODEX}" -eq 1 || "${RUN_OPENCODE}" -eq 1 ]]; then
        echo ""
        echo "Reminder: harness/{codex,opencode}/*.snippet config merges (plugin/hook"
        echo "registration in opencode.jsonc / config.toml) are ONE-TIME and MANUAL — this"
        echo "deploy never writes those config files. See harness/{codex,opencode}/README.md."
        echo "On a fresh machine, hooks and rules land on disk with no prompt from the tool"
        echo "itself to register them — this reminder is that prompt."
    fi
    echo ""
    echo "Reminder: restart Claude Code / Codex / opencode (or open a new session) to reload."
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

main() {
    parse_args "$@"
    check_deps

    REPORT_LOG=$(mktemp)
    register_tmp "${REPORT_LOG}"

    log "deploy-global.sh — mode: $([[ ${APPLY} -eq 1 ]] && echo APPLY || echo DRY-RUN)"
    log "scopes — claude:${RUN_CLAUDE} codex:${RUN_CODEX} opencode:${RUN_OPENCODE} grok:${RUN_GROK}"

    step_preview
    step_diff
    step_backup
    step_detect_orphans
    step_deploy_claude
    step_deploy_hooks
    step_harness_rebuild
    step_deploy_shared_harness
    step_deploy_codex
    step_deploy_opencode
    step_engram_hotfix
    step_deploy_codex_hooks
    step_deploy_opencode_plugin
    step_deploy_grok
    step_write_manifest

    step_final_report
}

main "$@"
