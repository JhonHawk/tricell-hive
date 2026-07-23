---
name: deploy-global
description: Deploy the full global/ directory (CLAUDE.md, rules, agents, skills, hooks) to ~/.claude/ with automatic backup, plus the multi-harness layer — universal skills to ~/.agents/skills/, generated agents to ~/.codex/agents/ and ~/.config/opencode/agents/, opencode commands, and harness/AGENTS.md to ~/.codex/AGENTS.md + ~/.config/opencode/AGENTS.md. Also idempotently merges hook blocks into ~/.claude/settings.json (additive, never overwriting preferences). Use when ready to deploy config changes.
disable-model-invocation: true
---

Deploy the entire `global/` directory to `~/.claude/` (CLAUDE.md, rules, agents, skills, hooks), then the multi-harness layer (Codex + opencode) derived from the same sources.

## Steps

1. List all files recursively in `global/` — show what will be deployed grouped by type (config, rules, agents, hooks) with line counts.
2. **Ensure target structures exist** — on a fresh machine nothing may exist:
   ```bash
   mkdir -p ~/.claude/{rules,agents,skills,hooks,backups}
   mkdir -p ~/.agents/skills ~/.codex/{agents,hooks} ~/.config/opencode/{agents,commands,plugins}
   ```
3. **Diff BEFORE deploying** — compare each source file against its deployed counterpart. Show a summary of what will change: new files, modified files, and unchanged files. This MUST happen before any copy operation. If a target file does not exist, report it as NEW.
4. **Backup existing config** before overwriting:
   - Build the tar archive dynamically — only include paths that exist. Use an array so each path is a separate argument regardless of execution context:
     ```bash
     tar_args=()
     [ -f ~/.claude/CLAUDE.md ] && tar_args+=(CLAUDE.md)
     [ -d ~/.claude/rules ] && [ "$(ls -A ~/.claude/rules 2>/dev/null)" ] && tar_args+=(rules/)
     [ -d ~/.claude/agents ] && [ "$(ls -A ~/.claude/agents 2>/dev/null)" ] && tar_args+=(agents/)
     [ -d ~/.claude/skills ] && [ "$(ls -A ~/.claude/skills 2>/dev/null)" ] && tar_args+=(skills/)
     [ -d ~/.claude/hooks ] && [ "$(ls -A ~/.claude/hooks 2>/dev/null)" ] && tar_args+=(hooks/)
     [ -f ~/.claude/settings.json ] && tar_args+=(settings.json)
     if [ ${#tar_args[@]} -gt 0 ]; then
       tar -czf ~/.claude/backups/global-backup-$(date +%Y%m%d-%H%M%S).tar.gz -C ~/.claude "${tar_args[@]}"
     else
       echo "Nothing to back up (fresh install)"
     fi
     ```
   - Keep only the 5 most recent backups — delete older ones.
5. **Remove orphans via manifest**: `~/.claude/.deploy-manifest` records every file the *previous* deploy copied. Anything listed there that no longer has a source in `global/` was repo-managed and lost its source — safe to delete. Files never listed (manually installed, third-party) are unreachable by construction.
   ```bash
   manifest=~/.claude/.deploy-manifest
   orphans=()
   if [ -f "$manifest" ]; then
     while IFS= read -r rel; do
       case "$rel" in '#'*|'') continue ;; esac
       case "$rel" in
         CLAUDE.md|rules/*|agents/*|skills/*) src="global/$rel"; tgt=~/.claude/"$rel" ;;
         hooks/*) src=$(find global/hooks -name "$(basename "$rel")" -print -quit 2>/dev/null); tgt=~/.claude/"$rel" ;;
         agents-skills/*) src="global/skills/${rel#agents-skills/}"; tgt=~/.agents/skills/"${rel#agents-skills/}" ;;
         codex-agents/*) n=$(basename "$rel" .toml); src=$(find global/agents -name "$n.md" -print -quit 2>/dev/null); tgt=~/.codex/agents/"$(basename "$rel")" ;;
         opencode-agents/*) n=$(basename "$rel" .md); src=$(find global/agents -name "$n.md" -print -quit 2>/dev/null); tgt=~/.config/opencode/agents/"$(basename "$rel")" ;;
         opencode-commands/*) src="harness/opencode/commands/$(basename "$rel")"; tgt=~/.config/opencode/commands/"$(basename "$rel")" ;;
         codex-hooks/*) src=$(find global/hooks -name "$(basename "$rel")" -print -quit 2>/dev/null); tgt=~/.codex/hooks/"$(basename "$rel")" ;;
         opencode-plugins/*) src=$(find global/hooks -name "$(basename "$rel")" -print -quit 2>/dev/null); tgt=~/.config/opencode/plugins/"$(basename "$rel")" ;;
         harness-agents/codex/*) src="harness/AGENTS.md"; tgt=~/.codex/AGENTS.md ;;
         harness-agents/opencode/*) src="harness/AGENTS.md"; tgt=~/.config/opencode/AGENTS.md ;;
         *) continue ;;   # unknown prefix — never delete
       esac
       if [ -z "$src" ] || [ ! -e "$src" ]; then
         [ -e "$tgt" ] && orphans+=("$rel")
       fi
     done < "$manifest"
   fi
   printf '%s\n' "${orphans[@]}"
   ```
   - **Orphans found** → list them and confirm deletion with the user (one `AskUserQuestion`, fold into the deploy confirmation if one is already planned). On confirmation: `rm` each orphan at its mapped target path (`tgt` per the case above — spans `~/.claude/`, `~/.agents/skills/`, `~/.codex/agents/`, `~/.config/opencode/`), then prune empty dirs: `find ~/.claude/rules ~/.claude/agents ~/.claude/skills ~/.agents/skills -type d -empty -delete 2>/dev/null`.
   - **Hook orphans also purge their `settings.json` block.** A removed `.sh` must not leave a dangling hook entry. For each orphan whose `rel` is `hooks/*.sh`, purge from `~/.claude/settings.json` any hook entry whose inner `command` references that basename — this is safe *because the manifest confirmed the hook was repo-managed*; the user's own hooks (never in the manifest) are never matched. Same temp+validate+move discipline as the merge; backup is covered by step 4.
     ```bash
     # $orphan_basename is the basename of a hooks/*.sh orphan, e.g. pre-push-lint-reminder.sh
     # The filter is a bundled resource (repo-relative; the skill runs from the repo root).
     tmp=$(mktemp)
     if jq --arg bn "$orphan_basename" -f .claude/skills/deploy-global/filters/hook-purge.jq ~/.claude/settings.json > "$tmp" && jq empty "$tmp" 2>/dev/null; then
       mv "$tmp" ~/.claude/settings.json
       echo "settings.json: purged orphan hook block for $orphan_basename"
     else
       rm -f "$tmp"
       echo "WARNING: could not purge $orphan_basename from settings.json — left unchanged."
     fi
     ```
   - **Codex hook orphans purge `~/.codex/hooks.json` the same way.** A removed `codex-hooks/*.sh` orphan (source gone from `global/hooks/`) must not leave a dangling entry in Codex's hooks file. Because it shares the Claude schema (`.hooks.<Event>[].hooks[].command`), reuse the SAME `hook-purge.jq` filter keyed on the orphan basename, with the identical temp+validate+move discipline:
     ```bash
     # $orphan_basename is the basename of a codex-hooks/*.sh orphan, e.g. flow-session-context.sh
     if [ -f ~/.codex/hooks.json ]; then
       tmp=$(mktemp)
       if jq --arg bn "$orphan_basename" -f .claude/skills/deploy-global/filters/hook-purge.jq ~/.codex/hooks.json > "$tmp" && jq empty "$tmp" 2>/dev/null; then
         mv "$tmp" ~/.codex/hooks.json
         echo "~/.codex/hooks.json: purged orphan hook block for $orphan_basename"
       else
         rm -f "$tmp"
         echo "WARNING: could not purge $orphan_basename from ~/.codex/hooks.json — left unchanged."
       fi
     fi
     ```
     An `opencode-plugins/*.ts` orphan needs no config purge — opencode loads local plugins by file presence, so removing the `.ts` fully deregisters it.
   - **No manifest (first run) or unreadable** → skip deletion entirely and say so. Degrade to no-deletion, never to guessing.
6. **Deploy CLAUDE.md**: copy `global/CLAUDE.md` → `~/.claude/CLAUDE.md`.
7. **Deploy rules**: copy all files from `global/rules/` → `~/.claude/rules/`, preserving subdirectory structure.
8. **Clean stale rules**: find `.md` files at `~/.claude/rules/` root level whose filename also exists in a subdirectory. Use `find` instead of globs to avoid zsh `nomatch` errors:
   ```bash
   find ~/.claude/rules -maxdepth 1 -name '*.md' -print0 2>/dev/null | while IFS= read -r -d '' f; do
     name=$(basename "$f")
     if find ~/.claude/rules -mindepth 2 -name "$name" -print -quit | grep -q .; then
       rm "$f"
       echo "Stale rule removed: $name"
     fi
   done
   ```
   Report which files were cleaned, if any.
9. **Deploy agents**: copy `global/agents/` subdirectory structure to `~/.claude/agents/` **preserving subdirectories** (design/, development/, review/, quality/, ops/, docs/). Claude Code discovers agents recursively.
10. **Clean stale agents**: same logic as rules — use `find` instead of globs.
11. **Deploy skills**: if `global/skills/` exists and is not empty, copy subdirectory structure to `~/.claude/skills/` **preserving subdirectories**. Each skill is a folder containing a `SKILL.md` file and optional supporting files:
    ```bash
    if [ -d global/skills ] && [ "$(ls -A global/skills 2>/dev/null)" ]; then
      for skill_dir in global/skills/*/; do
        skill_name=$(basename "$skill_dir")
        mkdir -p ~/.claude/skills/"$skill_name"
        cp -R "$skill_dir"* ~/.claude/skills/"$skill_name"/
      done
    fi
    ```
12. **Clean stale skills (fallback — only when no manifest exists)**: find skill directories in `~/.claude/skills/` that exist at the target but NOT in `global/skills/`. Only report them — do NOT auto-delete, since the user may have manually installed skills (e.g., symlinked design skills, context7-mcp). From the second manifest-aware deploy onward, step 5 already handles repo-managed orphans precisely; skip this step when a manifest was present.
13. **Deploy hooks** (optional — as of Claude Code 2.1.136+ plan mode is harness-enforced, so most hooks are no longer necessary; this step deploys whatever is in `global/hooks/` if anything): copy `.sh` files from each `global/hooks/<hook-folder>/` to `~/.claude/hooks/` (flat destination — no subdirectories). Make them executable:
    ```bash
    find global/hooks -name '*.sh' -exec cp {} ~/.claude/hooks/ \;
    chmod +x ~/.claude/hooks/*.sh
    find global/hooks -name '*.json' ! -name 'settings-config.json' -exec cp {} ~/.claude/hooks/ \;
    ```
    Hook data files (e.g. `code-search-routing.json`) deploy next to their script; `settings-config.json` is merge-only (below) and `README.md` files are NOT deployed — they are in-repo documentation only.

    **Register hooks in `settings.json` (idempotent, additive merge).** The `.sh`
    copy alone does nothing until the hook is registered. Merge every hook block from
    `global/hooks/*/settings-config.json` into `~/.claude/settings.json` WITHOUT
    overwriting it — add what's missing, leave existing entries (and all non-hook
    preference keys) untouched, and never duplicate on re-deploy. Identity is the
    inner `command` string (the `.sh` path is unique and stable): a managed entry is
    upserted (old version dropped, new appended), so timeout/matcher edits propagate
    while the user's own hooks are preserved. Backup is covered by step 4. Writes to a
    temp file, validates it parses, then moves into place; on any failure it leaves
    `settings.json` untouched and falls back to the manual reminder. Assumes hook paths
    have no spaces (true for `global/hooks/*/`).
    ```bash
    # find -print0 | xargs -0: zsh does NOT word-split an unquoted $var — a plain
    # `jq -s ... $configs` receives ONE newline-joined pseudo-path, jq errors, and
    # the merge silently no-ops (bug found in vivo 2026-07-10).
    if find global/hooks -name settings-config.json 2>/dev/null | grep -q .; then
      [ -f ~/.claude/settings.json ] || echo '{}' > ~/.claude/settings.json
      managed=$(find global/hooks -name settings-config.json -print0 2>/dev/null | xargs -0 jq -s 'reduce .[] as $c ({}; reduce (($c.hooks // {}) | to_entries[]) as $e (.; .[$e.key] = ((.[$e.key] // []) + $e.value)))')
      tmp=$(mktemp)
      # The merge filter is a bundled resource (repo-relative; the skill runs from the repo root).
      if jq --argjson m "$managed" -f .claude/skills/deploy-global/filters/hook-merge.jq ~/.claude/settings.json > "$tmp" && jq empty "$tmp" 2>/dev/null; then
        mv "$tmp" ~/.claude/settings.json
        echo "settings.json: hook blocks merged (idempotent, additive)."
      else
        rm -f "$tmp"
        echo "WARNING: settings.json hook merge failed — left unchanged. Register manually from settings-config.json."
      fi
    fi
    ```
    The reverse — a hook deleted from `global/hooks/` — is handled by step 5, which
    purges the dangling block from `settings.json` (manifest confirms it was repo-managed).
13b. **Deploy the multi-harness layer** — the generated trees are VERSIONED under
    `harness/`; rebuild first, then copy committed artifacts:
    ```bash
    # Regenerate from canonical sources (global/agents, global/skills)
    python3 harness/build.py
    # A dirty harness/ after build = canonical sources changed without a rebuild.
    # Continue the deploy, but report it and remind the user to commit the diff.
    git status --porcelain harness/

    # Universal skills (cleaned; Codex + opencode read ~/.agents/skills)
    for skill_dir in harness/agents-skills/*/; do
      skill_name=$(basename "$skill_dir")
      mkdir -p ~/.agents/skills/"$skill_name"
      cp -R "$skill_dir"* ~/.agents/skills/"$skill_name"/
    done
    # Generated agents — never hand-edited
    cp harness/codex/agents/*.toml ~/.codex/agents/
    find harness/opencode/agents -name '*.md' ! -name 'README.md' -exec cp {} ~/.config/opencode/agents/ \;
    # opencode command wrappers
    cp harness/opencode/commands/*.md ~/.config/opencode/commands/
    # Path-scoped language rules (generated; loaded by the opencode-rules plugin,
    # pinned in opencode.jsonc.snippet — the plugin entry itself is a one-time
    # manual config merge, never written here)
    mkdir -p ~/.config/opencode/rules
    find harness/opencode/rules -name '*.md' ! -name 'README.md' -exec cp {} ~/.config/opencode/rules/ \;
    # Shared cross-harness guidance — REPLACES the target files (they are fully
    # repo-managed from the first deploy; diff per step 3 before overwriting).
    # Both harnesses get the shared file VERBATIM. The Engram protocol reaches
    # Codex via the Engram Codex plugin's bundled hooks (SessionStart et al.) —
    # the former engram-memory-tail.md concat was removed 2026-07-10 as a
    # duplicate once that plugin shipped.
    cp harness/AGENTS.md ~/.config/opencode/AGENTS.md
    cp harness/AGENTS.md ~/.codex/AGENTS.md
    ```
    Config snippets (`harness/{codex,opencode}/*.snippet`) are one-time manual merges —
    point the user at the READMEs, never write their config files. Verify
    `~/.codex/AGENTS.md` stays under `project_doc_max_bytes` (49152) from the snippet.
    Close the Codex leg by validating its config still parses cleanly (`--strict-config`
    errors on unrecognized keys — catches config drift the deploy would otherwise mask):
    ```bash
    # `features list` rejects the flag; `doctor` accepts it (verified 0.144.0)
    codex --strict-config doctor >/dev/null 2>&1 && echo "codex config OK" \
      || echo "WARNING: codex --strict-config doctor failed — inspect ~/.codex/config.toml for drift"
    ```
13c. **Re-assert Engram #555 detect_project hotfix** (temporary — remove when upstream ships https://github.com/Gentleman-Programming/engram/issues/555). Plugin updates rewrite `~/.codex/plugins/cache/engram/**` and silently drop the local patch. Prefer the durable installer (LaunchAgent + kit under `~/.engram/hotfix-555/`); always re-run apply so this deploy leaves the machine healthy even if the watcher missed an update:
    ```bash
    ensure_script=""
    if [ -x "$HOME/.local/bin/ensure-engram-555" ]; then
      ensure_script="$HOME/.local/bin/ensure-engram-555"
    elif [ -x "$HOME/.engram/hotfix-555/ensure-engram-555.sh" ]; then
      ensure_script="$HOME/.engram/hotfix-555/ensure-engram-555.sh"
    elif [ -x "_support/archive/2026-07-15-engram-555-hotfix/ensure-engram-555.sh" ]; then
      # First machine / kit never installed: install durable copy + watcher from the archive.
      _support/archive/2026-07-15-engram-555-hotfix/ensure-engram-555.sh install
      ensure_script="$HOME/.local/bin/ensure-engram-555"
    fi
    if [ -n "$ensure_script" ]; then
      "$ensure_script" apply && echo "engram #555 hotfix: OK" \
        || echo "WARNING: engram #555 ensure failed — memories may split across project aliases until fixed"
    else
      echo "WARNING: engram #555 ensure script not found — see _support/archive/2026-07-15-engram-555-hotfix/"
    fi
    ```
    Report the ensure outcome in step 15. Do not treat a failure as a deploy abort — the rest of the deploy is independent.
13d. **Deploy the Codex hooks.** Every hook folder shipping a `codex-hooks.json` reaches Codex:
    copy its scripts into `~/.codex/hooks/` and register them in `~/.codex/hooks.json` with the
    same additive-merge discipline as step 13 — Codex's hooks file shares the Claude schema
    (`.hooks.<Event>[].hooks[].command`), so the merge is keyed on the inner `command` string
    and re-deploys never duplicate:
    ```bash
    mkdir -p ~/.codex/hooks
    find global/hooks -name codex-hooks.json 2>/dev/null | while IFS= read -r cj; do
      hookdir=$(dirname "$cj")
      find "$hookdir" -name '*.sh' -exec cp {} ~/.codex/hooks/ \; -exec sh -c 'chmod +x ~/.codex/hooks/"$(basename "$1")"' _ {} \;
      [ -f ~/.codex/hooks.json ] || echo '{}' > ~/.codex/hooks.json
      managed=$(jq '.hooks // {}' "$cj")
      tmp=$(mktemp)
      if jq --argjson m "$managed" -f .claude/skills/deploy-global/filters/hook-merge.jq ~/.codex/hooks.json > "$tmp" && jq empty "$tmp" 2>/dev/null; then
        mv "$tmp" ~/.codex/hooks.json
        echo "~/.codex/hooks.json: $(basename "$hookdir") merged (idempotent, additive)."
      else
        rm -f "$tmp"
        echo "WARNING: ~/.codex/hooks.json merge failed for $(basename "$hookdir") — left unchanged. Register manually from its codex-hooks.json."
      fi
    done
    ```
    **Codex guards hooks with `[hooks.state]` in `~/.codex/config.toml`** — a `trusted_hash` per
    command plus a per-hook `enabled` flag. The deploy NEVER edits `config.toml`. Report in step
    15: if the new hook's slot is missing or `enabled = false`, the merged hook stays inert until
    the user runs Codex once and accepts the trust prompt (or sets `enabled = true`).
13e. **Deploy the opencode session plugin.** opencode auto-loads local plugins by file presence —
    no `opencode.json` edit is needed:
    ```bash
    if [ -f global/hooks/flow-session-context/flow-session-context.ts ]; then
      cp global/hooks/flow-session-context/flow-session-context.ts ~/.config/opencode/plugins/
      echo "opencode plugin deployed: flow-session-context.ts"
    fi
    ```
14. **Write the manifest** — after all copies succeed, record exactly what this deploy manages, mapped to *deployed* paths (relative to `~/.claude/`; hooks flattened to match step 13's flat copy):
    ```bash
    {
      echo "# deploy-global manifest — files managed by tricell-hive. Do not edit by hand."
      echo "# deployed_at: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
      echo "# source_commit: $(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
      [ -f global/CLAUDE.md ] && echo "CLAUDE.md"
      find global/rules global/agents global/skills -type f 2>/dev/null | sed 's|^global/||'
      find global/hooks -name '*.sh' -print0 2>/dev/null | xargs -0 -n1 basename 2>/dev/null | sed 's|^|hooks/|'
      # multi-harness layer
      find global/skills -type f 2>/dev/null | sed 's|^global/skills/|agents-skills/|'
      find global/agents -name '*.md' -print0 2>/dev/null | xargs -0 -n1 basename 2>/dev/null | sed 's|\.md$|.toml|; s|^|codex-agents/|'
      find global/agents -name '*.md' -print0 2>/dev/null | xargs -0 -n1 basename 2>/dev/null | sed 's|^|opencode-agents/|'
      find harness/opencode/commands -name '*.md' -print0 2>/dev/null | xargs -0 -n1 basename 2>/dev/null | sed 's|^|opencode-commands/|'
      [ -f harness/AGENTS.md ] && printf 'harness-agents/codex/AGENTS.md\nharness-agents/opencode/AGENTS.md\n'
      # cross-harness session hook: Codex hooks (.sh in any folder shipping a codex-hooks.json) + opencode plugin (.ts)
      find global/hooks -name codex-hooks.json 2>/dev/null | while IFS= read -r cj; do
        find "$(dirname "$cj")" -name '*.sh' -print0 2>/dev/null | xargs -0 -n1 basename 2>/dev/null | sed 's|^|codex-hooks/|'
      done
      find global/hooks -name '*.ts' -print0 2>/dev/null | xargs -0 -n1 basename 2>/dev/null | sed 's|^|opencode-plugins/|'
    } > ~/.claude/.deploy-manifest
    ```
    Write the manifest only after a successful deploy — if the deploy aborted midway, leave the previous manifest untouched so the next run still sees the last known-good state.
15. Report:
    - What was deployed per category (config, rules, agents, skills, hooks) with counts
    - Multi-harness layer: universal skills count (~/.agents/skills), generated agents per harness, opencode commands, and the shared `AGENTS.md` deployed verbatim to `~/.config/opencode/AGENTS.md` and `~/.codex/AGENTS.md`; remind that config snippets (READMEs in `harness/{codex,opencode}/`) are one-time manual merges
    - Orphans removed via manifest (step 5) and stale files cleaned (if any); note when no manifest existed yet
    - Backup location and size
    - Diff summary from step 2 (new, modified, unchanged)
    - For hooks: report the settings.json merge outcome (blocks registered, or the WARNING fallback if the merge failed). No manual step is needed when the merge succeeded
    - For the cross-harness session hook (steps 13d/13e): report the `~/.codex/hooks.json` merge outcome and the opencode plugin copy. **Codex trust caveat:** if the hook's `[hooks.state]` slot in `~/.codex/config.toml` is missing or `enabled = false`, the merged hook stays inert until the user runs Codex once and accepts the trust prompt (or sets `enabled = true`) — the deploy never edits `config.toml`
    - Engram #555 ensure (step 13c): OK / WARNING — temporary until upstream ships the fix
    - Remind the user to restart Claude Code or open a new session to reload
    - Show how to restore:
      ```bash
      tar -xzf ~/.claude/backups/<backup-file>.tar.gz -C ~/.claude/
      ```

## Shell compatibility

This skill runs in both bash and zsh. Avoid bare globs that fail in zsh when there are no matches:
- **Use `find` instead of `for f in *.md`** for stale-file cleanup.
- **Check directory existence** before including paths in tar commands.
- **Use `2>/dev/null` on find** to suppress permission or missing-path errors.
- **Never inline multi-line jq programs.** The two `settings.json` filters ship as bundled resources under `.claude/skills/deploy-global/filters/` and are invoked with `jq -f` — an inline single-quoted multi-line program breaks when the executing agent flattens the command with `\` continuations (the backslashes land inside the quoted program and jq fails to compile). Single-line jq stays inline.

## Rules
- **Never delete** files from `~/.claude/rules/`, `~/.claude/agents/`, `~/.claude/skills/`, or `~/.claude/hooks/` that are NOT managed by this repo. Only overwrite files that exist in `global/`. The manifest defines "managed": a path may be deleted only if the previous manifest lists it (step 5) — and only after user confirmation.
- **Exception: stale duplicates.** If a file exists at root level AND inside a subdirectory with the same name, the root copy is stale and must be removed. Claude Code discovers rules recursively — a duplicate means the rule loads twice, wasting context window.
- **Manifest failure degrades to no-deletion.** Missing, unreadable, or malformed manifest → skip step 5 with a warning and continue the deploy. Entries with prefixes outside `CLAUDE.md|rules/|agents/|skills/|hooks/` are ignored, so a corrupted manifest can never reach arbitrary paths.
- Step 2 guarantees all directories exist — no need to check individually in later steps.
- Always create a backup before any overwrite.
- Show a diff summary for every file being overwritten.
- **`~/.claude/settings.json` is never overwritten — only surgically edited for repo-managed hooks.** It holds the user's preferences. Exactly two writes are permitted, both confined to the `hooks` key, both backed up (step 4) and validated (temp + `jq empty` before move): (1) the idempotent additive merge in step 13 (upsert by `command` — add/update repo hooks, preserve every existing entry and all non-hook keys, never duplicate); (2) the orphan purge in step 5 (remove a hook block only when the manifest confirms the `.sh` was repo-managed and its source is gone). Non-hook keys and the user's own hooks are off-limits — touch nothing those two operations don't own.
