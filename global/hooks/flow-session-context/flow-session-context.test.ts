/**
 * Behavior tests for the OpenCode 2 session plugin.
 *
 *   node --experimental-strip-types --test global/hooks/flow-session-context/flow-session-context.test.ts
 *
 * They drive `setup()` with a fake plugin context, so they cover the plugin's
 * own logic — not opencode's loader. That the export SHAPE loads is asserted
 * separately, against the real binary, in
 * `.claude/skills/deploy-global/tests/test_opencode_plugin_compat.py`.
 */

import test from "node:test"
import assert from "node:assert/strict"
import { mkdtempSync, mkdirSync, writeFileSync, rmSync } from "node:fs"
import { tmpdir } from "node:os"
import { join } from "node:path"
import { execFileSync } from "node:child_process"
import plugin from "./flow-session-context.ts"

function repo(opts: { ledger?: boolean; mergedBranch?: boolean } = {}): string {
  const dir = mkdtempSync(join(tmpdir(), "flow-session-context-"))
  if (opts.ledger) {
    mkdirSync(join(dir, "_support"))
    writeFileSync(join(dir, "_support/PROJECT.md"), "| Current phase | specs |\n")
  }
  const git = (...args: string[]) => execFileSync("git", ["-C", dir, ...args], { stdio: "ignore" })
  git("init", "-q", "-b", "master") // pinned: the assertions below name the branch
  git("config", "user.email", "test@example.invalid")
  git("config", "user.name", "test")
  git("commit", "-q", "--allow-empty", "-m", "init")
  if (opts.mergedBranch) git("branch", "feature/done")
  return dir
}

type Part = { type: string; text: string; cache?: unknown }

/** Captures the hook the plugin registers, or leaves it undefined. */
async function load(directory: string | undefined) {
  let hook: ((event: unknown) => void) | undefined
  const ctx = {
    location: directory === undefined ? undefined : { directory },
    session: {
      hook: async (name: string, callback: (event: unknown) => void) => {
        assert.equal(name, "context", "the plugin must register the `context` hook")
        hook = callback
        return { dispose: async () => {} }
      },
    },
  }
  await plugin.setup(ctx as never)
  return {
    registered: hook !== undefined,
    invoke(system: Part[] = [{ type: "text", text: "BASE" }], sessionID = "ses_test") {
      assert.ok(hook, "hook never registered")
      const event = { sessionID, system }
      hook(event)
      return event.system
    },
  }
}

test("exports the V2 default definition", () => {
  assert.equal(plugin.id, "flow-session-context")
  assert.equal(typeof plugin.setup, "function")
})

test("a flow workspace with pending hygiene pushes ONE new part, leaving the others intact", async () => {
  const p = await load(repo({ ledger: true, mergedBranch: true }))
  const system = p.invoke()
  assert.equal(system.length, 2)
  assert.equal(system[0].text, "BASE", "existing parts are never rewritten")
  assert.equal(system[1].type, "text")
  assert.match(system[1].text, /<flow-process-protocol>/)
  assert.match(system[1].text, /local branches fully merged into master: feature\/done/)
})

test("a plain git repo gets the hygiene section only", async () => {
  const p = await load(repo({ mergedBranch: true }))
  const system = p.invoke()
  assert.equal(system.length, 2)
  assert.doesNotMatch(system[1].text, /<flow-process-protocol>/)
  assert.match(system[1].text, /Pending git hygiene/)
})

test("nothing to say leaves the system array untouched", async () => {
  const p = await load(repo())
  assert.deepEqual(p.invoke(), [{ type: "text", text: "BASE" }])
})

test("injects on EVERY invocation, not once per session", async () => {
  // The V1 gate delivered the block to the first model call and nothing after
  // it. V2 rebuilds `system` per invocation, so skipping means not delivering.
  const p = await load(repo({ ledger: true, mergedBranch: true }))
  assert.equal(p.invoke().length, 2)
  assert.equal(p.invoke().length, 2)
  assert.equal(p.invoke().length, 2)
})

test("within one session the git probes run once — later invocations reuse the cached text", async () => {
  const dir = repo({ ledger: true, mergedBranch: true })
  const p = await load(dir)
  const first = p.invoke()[1].text
  rmSync(dir, { recursive: true, force: true }) // no repo left to probe
  assert.equal(p.invoke()[1].text, first)
})

test("the cache is per session — a new session re-probes git", async () => {
  // One `setup` serves many sessions. Caching for the instance's lifetime would
  // keep reporting a branch the user deleted after acting on the last advisory.
  const dir = repo({ ledger: true, mergedBranch: true })
  const p = await load(dir)
  assert.match(p.invoke(undefined, "ses_a")[1].text, /feature\/done/)

  execFileSync("git", ["-C", dir, "branch", "-D", "feature/done"], { stdio: "ignore" })

  assert.match(p.invoke(undefined, "ses_a")[1].text, /feature\/done/, "the same session keeps its cached text")
  assert.doesNotMatch(p.invoke(undefined, "ses_b")[1]?.text ?? "", /feature\/done/, "a new session must see current state")
})

test("an already-injected array is left alone", async () => {
  const p = await load(repo({ ledger: true, mergedBranch: true }))
  const system: Part[] = [{ type: "text", text: "X <flow-process-protocol> Y" }]
  assert.deepEqual(p.invoke(system), system)
})

test("an empty system array gets the part pushed", async () => {
  const p = await load(repo({ ledger: true, mergedBranch: true }))
  const system = p.invoke([])
  assert.equal(system.length, 1)
  assert.match(system[0].text, /<flow-process-protocol>/)
})

test("no directory means no hook at all, never another repo's branches", async () => {
  assert.equal((await load(undefined)).registered, false)
})

test("a throwing guard never escapes the hook", async () => {
  const p = await load(repo({ ledger: true, mergedBranch: true }))
  assert.ok(p.registered)
  assert.doesNotThrow(() =>
    p.invoke(
      Object.defineProperty([] as Part[], "some", {
        get() {
          throw new Error("boom")
        },
      }),
    ),
  )
})

test("a throwing push never escapes the hook", async () => {
  // The guard-throw case above only exercises the FIRST statement in the try.
  // This one reaches past the cache to the push itself — the realistic failure
  // (a frozen or proxied array), and the one that a too-narrow try would miss.
  const p = await load(repo({ ledger: true, mergedBranch: true }))
  const frozen = Object.freeze([{ type: "text", text: "BASE" }]) as Part[]
  assert.doesNotThrow(() => p.invoke(frozen))
  assert.equal(frozen.length, 1, "nothing was appended to the frozen array")
})
