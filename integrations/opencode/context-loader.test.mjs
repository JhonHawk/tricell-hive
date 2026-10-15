import assert from "node:assert/strict"
import fs, { writeFile, readFile, rm, realpath } from "node:fs/promises"
import { syncBuiltinESMExports } from "node:module"
import { createHash } from "node:crypto"
import path from "node:path"
import { pathToFileURL } from "node:url"

const root = await realpath(process.argv[2])
const skills = path.join(root, "skills")
const metadata = path.join(root, "context-loader.json")
const refs = {
  "hive-review-harness": "harness-audit/SKILL.md",
  "hive-review-plan": "flow-plan/references/plan-review.md",
}
const expected = JSON.parse(await readFile(path.join(root, "fixture.json"), "utf8"))
const provenance = JSON.parse(await readFile(metadata, "utf8"))
assert.deepEqual(Object.keys(provenance), ["schema", "references"])
assert.equal(provenance.schema, 2)
assert.deepEqual(provenance.references, Object.entries(refs).map(([role, resource]) => ({ role, resource, sha256: createHash("sha256").update(Buffer.from(expected[role], "utf8")).digest("hex") })))
const plugin = (await import(pathToFileURL(path.join(root, "context-loader.js")))).default
assert.equal(plugin.id, "hive.context-loader")
let callback, disposed = false, role = "hive-review-plan"
const cleanup = await plugin.setup({
  agent: { get: async ({ agentID }) => { assert.equal(agentID, "opaque-agent-id"); return { id: role, name: "Display label" } } },
  session: { hook: async (name, cb) => { assert.equal(name, "context"); callback = cb; return { dispose: async () => { disposed = true } } } },
})
assert.equal(typeof callback, "function", "setup must register a context hook")
const request = () => ({ agent: "opaque-agent-id", sessionID: "same-session", system: [{ type: "text", text: "user-owned" }], messages: [{ text: "persistent" }], tools: { read: {} }, options: { temperature: 0.5 }, model: { id: "unchanged" } })
for (const name of Object.keys(refs)) {
  role = name
  const event = request(), before = structuredClone(event)
  await callback(event)
  assert.equal(event.system.length, 2)
  assert.equal(event.system[1].text, `Hive canonical reference: ${refs[name]}\n\n${expected[name]}`, "exact UTF-8 snapshot including BOM, quotes, escapes and marker text")
  await callback(event)
  assert.equal(event.system.length, 2, "same request must have one own copy")
  assert.deepEqual({ ...event, system: before.system }, before, "only system may change")
  await writeFile(path.join(skills, refs[name]), "Fresh request content\n")
  const next = request()
  const originalRead = fs.readFile, originalOpen = fs.open
  let reads = 0
  fs.readFile = fs.open = async () => { reads++; throw new Error("runtime read forbidden") }
  syncBuiltinESMExports()
  try { await callback(next) }
  finally { fs.readFile = originalRead; fs.open = originalOpen; syncBuiltinESMExports() }
  assert.equal(reads, 0, "no runtime reference or metadata reads")
  assert.equal(next.system[1].text, event.system[1].text, "source edits require regeneration and reload")
  const foreignCopy = request()
  foreignCopy.system.push({ ...next.system[1] })
  await callback(foreignCopy)
  assert.equal(foreignCopy.system.length, 3, "foreign identical text must not be removed")
}
role = "unknown"
const unknown = request(), before = structuredClone(unknown)
await callback(unknown)
assert.deepEqual(unknown, before)
for (const { name, agent, eventAgent, coveredRole, throws } of [
  { name: "ID wins over a different display name", agent: { id: "hive-review-plan", name: "hive-review-harness" }, eventAgent: "opaque-agent-id", coveredRole: "hive-review-plan" },
  { name: "native data ID wins over direct ID and display name", agent: { data: { id: "hive-review-harness", name: "Harness Review" }, id: "hive-review-plan", name: "unknown" }, eventAgent: "opaque-agent-id", coveredRole: "hive-review-harness" },
  { name: "native data ID resolves without name", agent: { data: { id: "hive-review-plan" } }, eventAgent: "opaque-agent-id", coveredRole: "hive-review-plan" },
  { name: "missing ID uses event agent", agent: { name: "Display label" }, eventAgent: "hive-review-harness", coveredRole: "hive-review-harness" },
  { name: "undefined agent uses event agent", eventAgent: "hive-review-plan", coveredRole: "hive-review-plan" },
  { name: "get throwing uses event agent", throws: true, eventAgent: "hive-review-plan", coveredRole: "hive-review-plan" },
  { name: "uncovered direct ID ignores covered display name", agent: { id: "unknown", name: "hive-review-plan" }, eventAgent: "hive-review-plan" },
  { name: "uncovered native ID ignores covered display name", agent: { data: { id: "unknown", name: "hive-review-plan" }, id: "hive-review-plan" }, eventAgent: "hive-review-plan" },
  { name: "get throwing leaves uncovered event ID untouched", throws: true, eventAgent: "unknown" },
]) {
  let identityCallback
  const dispose = await plugin.setup({
    agent: { get: async ({ agentID }) => { assert.equal(agentID, eventAgent); if (throws) throw new Error("Agent not found"); return agent } },
    session: { hook: async (_, cb) => { identityCallback = cb; return { dispose: () => {} } } },
  })
  const event = { ...request(), agent: eventAgent }, before = structuredClone(event)
  await identityCallback(event)
  await identityCallback(event)
  if (coveredRole) {
    assert.equal(event.system.length, 2, `${name}: injects exactly once`)
    assert.equal(event.system[1].text, `Hive canonical reference: ${refs[coveredRole]}\n\n${expected[coveredRole]}`, name)
    assert.deepEqual({ ...event, system: before.system }, before, `${name}: only system changes`)
  } else {
    assert.deepEqual(event, before, `${name}: uncovered role stays untouched`)
  }
  await dispose()
}
role = "hive-review-plan"
// Metadata and sources are offline provenance, not runtime capabilities.
await rm(skills, { recursive: true })
await writeFile(metadata, "invalid metadata PRIVATE_SENTINEL")
const stillSnapshot = request()
await callback(stillSnapshot)
assert.equal(stillSnapshot.system[1].text, `Hive canonical reference: ${refs[role]}\n\n${expected[role]}`)
await cleanup()
assert.equal(disposed, true)
for (const [module, isMissing] of [["reloaded.js", false], ["missing.js", true]]) {
  const replacement = (await import(pathToFileURL(path.join(root, module)))).default
  const dispose = await replacement.setup({
    agent: { get: async () => ({ data: { id: role, name: "Display label" } }) },
    session: { hook: async (_, cb) => { callback = cb; return { dispose: async () => {} } } },
  })
  const event = request()
  await callback(event)
  if (isMissing) {
    assert.ok(event.system[1].text.includes("missing_snapshot"))
    assert.ok(event.system[1].text.length < 512)
    assert.ok(!event.system[1].text.includes(root) && !event.system[1].text.includes("PRIVATE_SENTINEL"))
    await callback(event)
    assert.equal(event.system.length, 2)
    const next = request()
    await callback(next)
    assert.equal(next.system[1].text, event.system[1].text)
    assert.deepEqual({ ...next, system: request().system }, request())
  } else {
    assert.equal(event.system[1].text, `Hive canonical reference: ${refs[role]}\n\nRegenerated reference\n`)
  }
  await dispose()
}
console.log("Both role IDs, exact UTF-8 roundtrip, snapshot freshness/reload, request dedup, unknown role, bounded diagnostic and disposal passed")
