/**
 * flow-session-context — OpenCode 2 plugin adapter
 *
 * The opencode analog of the flow-session-context.sh SessionStart hook (Claude
 * Code / Codex have a real SessionStart event; opencode does not). It injects
 * the same two self-gating sections — flow process protocol + pending git
 * hygiene — into the system prompt, on every model invocation (see below for
 * why "once per session" is the wrong shape in V2).
 *
 * Node builtins only. Advisory: every failure path degrades to no injection,
 * never throws out of the hook.
 *
 * ── Why the import is type-only ──────────────────────────────────────────────
 * The documented form is `import { Plugin } from "@opencode/plugin"` +
 * `Plugin.define({...})`. A local plugin file resolves its imports against the
 * DIRECTORY IT SITS IN, and `~/.config/opencode/` has no node_modules for the
 * v2 scope — measured 2026-09-20 against opencode 2.0.9, where the runtime form
 * dies with `Cannot find package '@opencode/plugin'` while this type-only form
 * loads. `Plugin.define` is the identity function, so the default export below
 * is what `define` would have produced. Both that and the `Plugin.Plugin` type
 * used here (`index.d.ts`: `export * as Plugin from "./plugin.js"`;
 * `plugin.d.ts`: `export interface Plugin`) were read out of the published
 * @opencode/plugin@2.0.11 tarball — the package is NOT on disk anywhere, so
 * re-check with `npm view @opencode/plugin dist.tarball` rather than a local path. Switch to the
 * documented form only if the deploy ever installs the package next to the file.
 *
 * ── V1 is not supported ──────────────────────────────────────────────────────
 * OpenCode 1 expects a NAMED export matching the v1 `Plugin` type and the
 * `experimental.chat.system.transform` hook; OpenCode 2 expects `export default
 * {id, setup}` and rejects the v1 shape with `PluginModule.LoadError: Plugin
 * must export a default definition`. The two shapes do not overlap, so this file
 * targets v2 only and deploy-global gates it on the detected version.
 *
 * ── Why it injects on EVERY invocation, not once per session ─────────────────
 * V1 gated on a per-session Set, so the block reached the first model call and
 * nothing after it. In V2 `system` is rebuilt for every model invocation — the
 * `context` hook runs once per invocation, and opencode's own internal plugins
 * push into `system` unconditionally each time (they would duplicate without
 * limit otherwise). So a once-per-session gate does not prevent duplication; it
 * prevents DELIVERY. What is cached instead is the TEXT: `sessionContextText`
 * runs three synchronous `execFileSync` git probes, and its inputs (the
 * directory) cannot change for a plugin instance, so it is computed lazily once.
 *
 * This also removes the need for post-compaction recovery. V2 does expose a
 * dedicated `compaction` hook, but with the block re-injected on every
 * invocation there is nothing to recover: the next invocation after a compaction
 * carries it again. The PROTOCOL_MARKER guard below is kept only as a safety net
 * in case some path DOES hand back an already-injected array.
 */

import type { Plugin } from "@opencode/plugin"
import { existsSync } from "node:fs"
import { dirname } from "node:path"
import { execFileSync } from "node:child_process"

// ─── Static flow protocol block (module-level: never changes in a session) ───
const FLOW_PROTOCOL = `<flow-process-protocol>
Flow workspace — the process knowledge for this project lives in flow-core references, not in commands. Work the matching playbook conversationally when intent matches; never force ceremony onto a small change:
- IDEA: exploring whether something is worth doing -> converge on proceed/discard/defer with one light decision note (critical-thinking.md). Contested, load-bearing questions -> offer /adversarial-research.
- SPEC: a decided idea needs formalization -> flow-core/references/spec-writing-playbook.md (its business gate is mandatory before delivery is derived).
- PLAN: planning intent uses Hive's portable /flow-plan command; native harness planning remains optional and never grants Hive authorization. ONE plan per unit of work. Format and Preflight: flow-core/references/plan-format.md.
- EXECUTE: an approved plan with pending tasks exists -> offer /flow-build to execute or resume it, ONCE per session, with the direct route named as the alternative; a no is sticky for the session. A hand-run still owes the plan's per-task gates.
- After a deploy/promotion (git conventions own the flow): offer the in-vivo QA walk (flow-core/references/promotion-playbook.md). Epic close / tech-debt baseline -> flow-core/references/audit-playbook.md.
- Greenfield bootstrap -> flow-core/references/bootstrap-playbook.md. Pre-pack project entering the convention -> flow-core/references/migration-playbook.md. Workspace health concerns -> flow-core/references/workspace-hygiene-playbook.md (dispatch workspace-custodian).
</flow-process-protocol>`

const PROTOCOL_MARKER = "<flow-process-protocol>"
const PROTECTED = /^(development|qa|production|main|master)$/

// ─── Helpers ─────────────────────────────────────────────────────────────────

function walkUpForLedger(startDir: string): boolean {
  let cursor = startDir
  while (cursor && cursor !== "/") {
    if (existsSync(`${cursor}/_support/PROJECT.md`)) return true
    cursor = dirname(cursor)
  }
  return false
}

function git(dir: string, args: string[]): string {
  try {
    return execFileSync("git", ["-C", dir, ...args], {
      encoding: "utf8",
      stdio: ["ignore", "pipe", "ignore"],
    }).trim()
  } catch {
    return ""
  }
}

function isGitRepo(dir: string): boolean {
  try {
    execFileSync("git", ["-C", dir, "rev-parse", "--is-inside-work-tree"], {
      stdio: "ignore",
    })
    return true
  } catch {
    return false
  }
}

/** Same integration-target detection as flow-session-context.sh. */
function integrationTarget(dir: string): string {
  try {
    execFileSync("git", ["-C", dir, "show-ref", "--verify", "--quiet", "refs/heads/development"], {
      stdio: "ignore",
    })
    return "development"
  } catch {
    // not development — fall through
  }
  const originHead = git(dir, ["symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"])
  if (originHead) return originHead.replace(/^origin\//, "")
  for (const b of ["main", "master"]) {
    try {
      execFileSync("git", ["-C", dir, "show-ref", "--verify", "--quiet", `refs/heads/${b}`], {
        stdio: "ignore",
      })
      return b
    } catch {
      // keep looking
    }
  }
  return ""
}

/** Pending-git-hygiene facts, or "" when there is nothing to report / not a repo. */
function gitHygieneSection(dir: string): string {
  try {
    if (!isGitRepo(dir)) return ""

    const target = integrationTarget(dir)

    let merged: string[] = []
    if (target) {
      merged = git(dir, ["branch", "--format=%(refname:short)", "--merged", target])
        .split("\n")
        .map((l) => l.trim())
        .filter((l) => l && l !== target && !PROTECTED.test(l))
        .slice(0, 8)
    }

    const gone = git(dir, ["branch", "--format=%(refname:short) %(upstream:track)"])
      .split("\n")
      .filter((l) => l.trim().endsWith("[gone]"))
      .map((l) => l.split(/\s+/)[0])
      .filter((b) => b && !PROTECTED.test(b))
      .slice(0, 8)

    if (merged.length === 0 && gone.length === 0) return ""

    const parts: string[] = []
    if (merged.length > 0) parts.push(`local branches fully merged into ${target}: ${merged.join(" ")}`)
    if (gone.length > 0) parts.push(`branches whose upstream is gone: ${gone.join(" ")}`)

    return `Pending git hygiene from a previous session (deterministic end-of-work-hygiene backstop): ${parts.join("; ")}. The ritual that owns this is git-mechanics.md > End-of-work hygiene (standing-authorized: prune confirmed-merged branches, report divergences; unmerged branches are decisions, not noise).`
  } catch {
    return ""
  }
}

/** The advisory text for a directory, or "" when nothing applies. */
export function sessionContextText(directory: string): string {
  const flowSection = walkUpForLedger(directory) ? FLOW_PROTOCOL : ""
  const gitSection = gitHygieneSection(directory)
  return [flowSection, gitSection].filter((s) => s.length > 0).join("\n\n")
}

// ─── Plugin Export ───────────────────────────────────────────────────────────

const plugin: Plugin.Plugin = {
  id: "flow-session-context",
  async setup(ctx) {
    const directory = ctx.location?.directory
    // No directory means no question this plugin can answer: the ledger walk and
    // the git probes are both relative to it. Falling back to process.cwd() would
    // report ANOTHER repo's pending branches as if they were the session's, so
    // degrade to not registering at all — the advisory contract is "inject
    // nothing", never "inject something wrong".
    if (!directory) return

    // Cached per SESSION, not per plugin instance: `sessionContextText` shells
    // out to git three times, so it must not run on every invocation — but one
    // `setup` serves many sessions (the hook payload carries `sessionID`, and
    // `ctx.location` is where the INSTANCE loaded), and the hygiene section
    // reports live state. Caching for the instance's lifetime would keep
    // reporting a branch the user deleted after acting on the last advisory.
    // The empty string is a valid cached answer (nothing applies here); a throw
    // caches nothing, so the next invocation retries.
    const cache = new Map<string, string>()

    await ctx.session.hook("context", (event) => {
      try {
        // Safety net only — see the header. V2 rebuilds `system` per invocation,
        // so in practice this never fires. It also only covers the flow section:
        // in a git repo with no ledger the injected text is the hygiene section
        // alone, which carries no marker.
        if (event.system.some((p) => typeof p?.text === "string" && p.text.includes(PROTOCOL_MARKER))) return

        const key: string = event.sessionID ?? ""
        let cached = cache.get(key)
        if (cached === undefined) {
          cached = sessionContextText(directory)
          cache.set(key, cached)
        }
        if (cached.length === 0) return

        // `system` is an array of PARTS of one system message, not an array of
        // messages, so push a part — what the documented example and opencode's
        // own plugins do. Never append into the last part: if it carries a
        // `cache` breakpoint, folding volatile text (git branch names) into it
        // changes the cached block's key.
        event.system.push({ type: "text", text: cached })
      } catch {
        // Never crash the hook — degrade to no injection.
      }
    })
  },
}

export default plugin
