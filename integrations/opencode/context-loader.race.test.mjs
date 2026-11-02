import assert from "node:assert/strict"
import fs, { mkdir, writeFile, readFile, symlink, rm, realpath, rename } from "node:fs/promises"
import { syncBuiltinESMExports } from "node:module"
import path from "node:path"
import { pathToFileURL } from "node:url"

const root = await realpath(process.argv[2])
const skills = path.join(root, "skills")
const resource = "flow-plan/SKILL.md"
const requested = path.join(skills, resource)
const expected = JSON.parse(await readFile(path.join(root, "fixture.json"), "utf8"))["hive-review-plan"]
const plugin = (await import(pathToFileURL(path.join(root, "context-loader.js")))).default
let callback
const cleanup = await plugin.setup({
  agent: { get: async () => ({ data: { id: "hive-review-plan", name: "Plan Review" } }) },
  session: { hook: async (name, cb) => { assert.equal(name, "context"); callback = cb; return { dispose: async () => {} } } },
})

const outcomes = []
try {
  for (const level of ["intermediate", "root"]) {
    const ancestor = level === "root" ? skills : path.join(skills, "flow-plan")
    const saved = path.join(root, `${level}-saved`)
    const outside = path.join(root, `${level}-outside`)
    const externalFile = path.join(outside, path.relative(ancestor, requested))
    await mkdir(path.dirname(externalFile), { recursive: true })
    // This victim is fictional, owned by the test, and contains no credential.
    const victim = `External fixture instruction ${level}\n`
    await writeFile(externalFile, victim)
    const originalOpen = fs.open, originalRead = fs.readFile, originalRealpath = fs.realpath
    let intercepted = false, externalRead = false, reads = 0
    fs.open = async (filename, ...args) => {
      reads++
      if (filename !== requested || intercepted) return originalOpen(filename, ...args)
      intercepted = true
      // Substitute an ancestor only during the actual OS open. Restore it
      // before returning the FileHandle, so later pathname checks look safe.
      await rename(ancestor, saved)
      let linked = false, handle
      try {
        await symlink(outside, ancestor)
        linked = true
        handle = await originalOpen(filename, ...args)
      } finally {
        if (linked) await rm(ancestor)
        await rename(saved, ancestor)
      }
      const read = handle.readFile.bind(handle)
      handle.readFile = async (...args) => { externalRead = true; return read(...args) }
      return handle
    }
    fs.readFile = async (...args) => { reads++; return originalRead(...args) }
    fs.realpath = async (...args) => { reads++; return originalRealpath(...args) }
    syncBuiltinESMExports()
    const event = { agent: "opaque-id", system: [{ type: "text", text: "foreign context" }], messages: [], tools: {}, options: {} }
    try {
      await callback(event)
      assert.equal(event.system[1].text, `Hive canonical reference: ${resource}\n\n${expected}`)
      // Hold a replacement across a full callback, then restore the ancestor.
      await rename(ancestor, saved)
      try {
        await symlink(outside, ancestor)
        const swapped = { agent: "opaque-id", system: [] }
        await callback(swapped)
        assert.equal(swapped.system[0].text, event.system[1].text)
        assert.ok(!swapped.system[0].text.includes(victim))
      } finally { await rm(ancestor); await rename(saved, ancestor) }
      const restored = { agent: "opaque-id", system: [] }
      await callback(restored)
      assert.equal(restored.system[0].text, event.system[1].text)
    }
    finally { fs.open = originalOpen; fs.readFile = originalRead; fs.realpath = originalRealpath; syncBuiltinESMExports() }
    const supplied = event.system.some(part => part.text?.includes(victim))
    outcomes.push({ level, intercepted, externalRead, supplied, reads })
  }
} finally { await cleanup() }

// Report booleans, never victim bytes or raw filesystem failures.
console.log(JSON.stringify(outcomes))
for (const outcome of outcomes) {
  // D11-A removes the open rather than diagnosing an unsafe live descriptor.
  assert.equal(outcome.intercepted, false, `${outcome.level}: no reference open may occur`)
  assert.equal(outcome.externalRead, false, `${outcome.level}: external descriptor must not be read`)
  assert.equal(outcome.supplied, false, `${outcome.level}: external text must not be supplied`)
  assert.equal(outcome.reads, 0, `${outcome.level}: request must not read mutable paths`)
}
