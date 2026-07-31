#!/usr/bin/env bash
# Detect a multi-repo workspace and (optionally) create a unified .engram/config.json
# at the workspace root and selected child git repos.
#
# Usage:
#   bootstrap-workspace.sh [--detect] [--name NAME] [--root DIR]
#                          [--all] [--only LIST] [--exclude LIST] [dir]
#     --detect       Report detection only; write nothing.
#     --name NAME     Project name. Overrides derivation; required off the canonical layout.
#     --root DIR      Override the detected workspace root.
#     --all           From inside a child repo, target the WHOLE workspace (not just this repo).
#     --only  LIST    Comma-separated child-repo names to include (root always included).
#     --exclude LIST  Comma-separated child-repo names to skip.
#     dir             Starting directory (default: current directory).
#
# Canonical layout = .../projects/<group>/<project>/  → name derived as <group>-<project>.
# A non-canonical parent is NEVER treated as a workspace: from inside a git repo whose
# parent isn't canonical, only that repo can be registered (with --name), never its siblings.
#
# Emits KEY=VALUE on stdout:
#   verdict=workspace | register-repo | needs-name | not-workspace
#   location=workspace_root | child_repo | standalone_repo | single_repo_project | noncanonical_multi | none
#   workspace_root=... | current_repo=... | project_name=... | git_repos=N | reason=...
# Exit (in --detect): 0 workspace/register-repo · 4 needs-name · 3 otherwise.
set -euo pipefail

detect_only=0; override_name=""; override_root=""; want_all=0; only_list=""; exclude_list=""; start=""
while [ $# -gt 0 ]; do
  case "$1" in
    --detect)  detect_only=1; shift ;;
    --name)    override_name="${2:?--name needs a value}"; shift 2 ;;
    --root)    override_root="${2:?--root needs a value}"; shift 2 ;;
    --all)     want_all=1; shift ;;
    --only)    only_list="${2:?--only needs a comma-separated list}"; shift 2 ;;
    --exclude) exclude_list="${2:?--exclude needs a comma-separated list}"; shift 2 ;;
    -*)        echo "unknown flag: $1" >&2; exit 2 ;;
    *)         start="$1"; shift ;;
  esac
done
start="${start:-$PWD}"
if ! start=$(cd "$start" 2>/dev/null && pwd -P); then
  printf 'verdict=not-workspace\nlocation=none\nreason=directory not found\n'; exit 3
fi

# Canonical path → group/project/ws_root.
group=""; project=""; ws_root=""
if [[ "$start" =~ /projects/([^/]+)/([^/]+) ]]; then
  group="${BASH_REMATCH[1]}"; project="${BASH_REMATCH[2]}"
  ws_root="${start%%/projects/*}/projects/$group/$project"
fi

# Are we inside a git repo? (its toplevel)
cwd_repo_root=""
if cwd_repo_root=$(git -C "$start" rev-parse --show-toplevel 2>/dev/null); then
  cwd_repo_root=$(cd "$cwd_repo_root" && pwd -P)
else
  cwd_repo_root=""
fi

count_repos() {
  local r="$1" c n=0
  if [ -d "$r" ]; then
    for c in "$r"/*/; do if [ -e "${c}.git" ]; then n=$((n + 1)); fi; done
  fi
  echo "$n"
}

# Resolve candidate root.
if   [ -n "$override_root" ]; then root="$override_root"
elif [ -n "$ws_root" ];       then root="$ws_root"
elif [ -n "$cwd_repo_root" ]; then root=$(dirname "$cwd_repo_root")
else root="$start"; fi
repos=$(count_repos "$root")

# Classify location.
location=""; current_repo=""
if [ -n "$ws_root" ]; then
  if   [ -n "$cwd_repo_root" ] && [ "$(dirname "$cwd_repo_root")" = "$ws_root" ]; then
    location="child_repo"; current_repo=$(basename "$cwd_repo_root")
  elif [ -n "$cwd_repo_root" ] && [ "$cwd_repo_root" = "$ws_root" ]; then
    location="single_repo_project"
  else
    location="workspace_root"
  fi
elif [ -n "$cwd_repo_root" ]; then
  location="standalone_repo"; current_repo=$(basename "$cwd_repo_root")
elif [ "$repos" -ge 2 ]; then
  location="noncanonical_multi"
else
  location="none"
fi

# Verdict + name.
verdict=""; reason=""; name=""
case "$location" in
  workspace_root|child_repo)
    if [ -n "$override_name" ]; then name="$override_name"
    else name=$(printf '%s' "$group-$project" | tr '[:upper:]' '[:lower:]'); fi
    verdict="workspace"
    if [ "$location" = child_repo ]; then
      reason="canonical workspace projects/$group/$project ($repos repos); inside child repo '$current_repo'"
    else
      reason="canonical workspace projects/$group/$project ($repos repos); at the workspace root"
    fi ;;
  single_repo_project|standalone_repo)
    if [ -n "$override_name" ]; then
      verdict="register-repo"; name="$override_name"
      reason="standalone git repo — will register ONLY this repo as '$name'; the parent is not treated as a workspace"
    else
      verdict="not-workspace"
      # shellcheck disable=SC2016  # the single quotes are literal output; the outer string is double-quoted, so $current_repo does expand
      reason="standalone git repo${current_repo:+ '$current_repo'} — git-remote already gives it a unique project; pass --name only to set a custom name for THIS repo (siblings are never touched)"
    fi ;;
  noncanonical_multi)
    if [ -n "$override_name" ]; then
      verdict="workspace"; name="$override_name"
      reason="non-canonical multi-repo dir ($repos repos); name from --name"
    else
      verdict="needs-name"
      reason="non-canonical multi-repo dir ($repos repos) — re-run with --name <group>-<project> to treat it as a workspace"
    fi ;;
  *)
    verdict="not-workspace"; reason="not a workspace and not a git repo ($repos git repos under $root)" ;;
esac

printf 'verdict=%s\nlocation=%s\nworkspace_root=%s\ncurrent_repo=%s\nproject_name=%s\ngit_repos=%s\nreason=%s\n' \
  "$verdict" "$location" "$root" "$current_repo" "$name" "$repos" "$reason"

if [ "$detect_only" -eq 1 ]; then
  case "$verdict" in
    workspace|register-repo) exit 0 ;;
    needs-name)              exit 4 ;;
    *)                       exit 3 ;;
  esac
fi

# ── write mode ───────────────────────────────────────────────────────────────
if [ "$verdict" != "workspace" ] && [ "$verdict" != "register-repo" ]; then
  echo "Refusing to write — $reason" >&2; exit 3
fi
case "$name" in
  "")         echo "ERROR: empty project name" >&2; exit 1 ;;
  */* | *\\*) echo "ERROR: project name must not contain slashes: $name" >&2; exit 1 ;;
esac

write_config() {
  local dir="$1" cfg existing
  cfg="$dir/.engram/config.json"
  if [ -f "$cfg" ]; then
    existing=$(grep -o '"project_name"[[:space:]]*:[[:space:]]*"[^"]*"' "$cfg" 2>/dev/null || true)
    printf 'skip (exists): %s  ->  %s\n' "$cfg" "${existing:-?}"; return
  fi
  mkdir -p "$dir/.engram"
  printf '{ "project_name": "%s" }\n' "$name" > "$cfg"
  printf 'created:       %s\n' "$cfg"
}

# register-repo: write ONLY the current repo, never the parent.
if [ "$verdict" = "register-repo" ]; then
  printf -- '--- registering single repo as %s ---\n' "$name"
  write_config "$cwd_repo_root"
  printf -- '---\nDone (single repo). Parent left untouched.\n'
  exit 0
fi

should_include() {
  local repo="$1"
  if [ -n "$only_list" ]; then
    case ",$only_list," in *",$repo,"*) return 0 ;; *) return 1 ;; esac
  fi
  if [ -n "$exclude_list" ]; then
    case ",$exclude_list," in *",$repo,"*) return 1 ;; esac
  fi
  return 0
}

# Default scope: at a child repo with no scope flags → only that repo. Otherwise → all (filtered).
this_repo_only=0
if [ "$location" = child_repo ] && [ "$want_all" -eq 0 ] && [ -z "$only_list" ] && [ -z "$exclude_list" ]; then
  this_repo_only=1
fi

printf -- '--- writing configs for %s ---\n' "$name"
write_config "$root"
for child in "$root"/*/; do
  if [ -e "${child}.git" ]; then
    cname=$(basename "${child%/}")
    if [ "$this_repo_only" -eq 1 ]; then
      if [ "$cname" != "$current_repo" ]; then continue; fi
    else
      if ! should_include "$cname"; then continue; fi
    fi
    write_config "${child%/}"
  fi
done
printf -- '---\nDone. Verify from the workspace root: mem_current_project should report source=config.\n'
