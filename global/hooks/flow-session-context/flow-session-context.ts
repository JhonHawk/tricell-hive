/**
 * flow-session-context — OpenCode plugin adapter
 *
 * The opencode analog of the flow-session-context.sh SessionStart hook (Claude
 * Code / Codex have a real SessionStart event; opencode does not). It injects
 * the same two self-gating sections — flow process protocol + pending git
 * hygiene — into the system prompt, once per session.
 *
 * Pattern mirrors the Engram opencode plugin (experimental.chat.system.transform,
 * per-session gate, append to the LAST system entry — some models reject
 * multiple system messages) and borrows the module-level static cache +
 * content-marker double-injection guard from the Superpowers plugin.
 *
 * Node builtins only. Advisory: every failure path degrades to no injection,
 * never throws out of the hook.
 */

import type { Plugin } from "@opencode-ai/plugin"
import { existsSync } from "node:fs"
import { dirname } from "node:path"
import { execFileSync } from "node:child_process"

// ─── Static flow protocol block (module-level: never changes in a session) ───
const FLOW_PROTOCOL = `<flow-process-protocol>
Flow workspace — the process knowledge for this project lives in flow-core references, not in commands. Work the matching playbook conversationally when intent matches; never force ceremony onto a small change:
- IDEA: exploring whether something is worth doing -> converge on proceed/discard/defer with one light decision note (quality/critical-thinking.md). Contested, load-bearing questions -> offer /adversarial-research.
- SPEC: a decided idea needs formalization -> flow-core/references/spec-writing-playbook.md (its business gate is mandatory before delivery is derived).
- PLAN: planning intent ALWAYS uses the harness's native plan mechanism; the plan-capture hook adopts the approved plan — never a parallel planning ceremony. ONE plan per unit of work. Format and Preflight: flow-core/references/plan-format.md.
- EXECUTE: a captured plan with pending tasks exists -> offer /flow-build to execute or resume it, ONCE per session, with the direct route named as the alternative; a no is sticky for the session. A hand-run still owes the plan's per-task gates.
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

    return `Pending git hygiene from a previous session (deterministic session-close backstop): ${parts.join("; ")}. The ritual that owns this is git-mechanics.md > Session close (standing-authorized: prune confirmed-merged branches, report divergences; unmerged branches are decisions, not noise).`
  } catch {
    return ""
  }
}

// ─── Plugin Export ───────────────────────────────────────────────────────────

export const FlowSessionContext: Plugin = async (ctx) => {
  const directory = ctx.directory
  // Per-session gate: inject at most once per session id.
  const injectedSessions = new Set<string>()

  return {
    "experimental.chat.system.transform": async (input, output) => {
      try {
        const sessionID: string = input.sessionID ?? ""

        // Primary gate: once per session.
        if (sessionID && injectedSessions.has(sessionID)) return

        // Secondary gate: content-marker guard — skip if opencode passes an
        // already-transformed system array back through (double injection).
        if (output.system.some((s) => typeof s === "string" && s.includes(PROTOCOL_MARKER))) return

        const flowSection = walkUpForLedger(directory) ? FLOW_PROTOCOL : ""
        const gitSection = gitHygieneSection(directory)

        const sections = [flowSection, gitSection].filter((s) => s.length > 0)
        if (sections.length === 0) {
          // Nothing to inject this session — still consume the gate so we don't
          // re-run the git probe on every step.
          if (sessionID) injectedSessions.add(sessionID)
          return
        }

        const ctxText = sections.join("\n\n")
        if (output.system.length > 0) {
          output.system[output.system.length - 1] += "\n\n" + ctxText
        } else {
          output.system.push(ctxText)
        }

        if (sessionID) injectedSessions.add(sessionID)
      } catch {
        // Never crash the hook — degrade to no injection.
      }
    },
  }
}
