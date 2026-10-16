import { Plugin } from "@opencode/plugin"
import { appendFile, readFile } from "node:fs/promises"
import { createHash } from "node:crypto"

const hash = (text) => createHash("sha256").update(text, "utf8").digest("hex")
const modelName = (model) => model?.providerID && model?.id
  ? `${model.providerID}/${model.id}${model.variant ? `#${model.variant}` : ""}` : "not_observed"

// Only known protocol text fields are inspected in memory. Never walk arbitrary
// keys, headers, URLs, tool arguments or binary/attachment payloads.
function contentText(value) {
  if (typeof value === "string") return [value]
  if (!Array.isArray(value)) return []
  return value.flatMap((part) => {
    if (["text", "input_text", "output_text"].includes(part?.type) && typeof part.text === "string") return [part.text]
    if (part?.type === "tool_result") return contentText(part.content)
    return []
  })
}
export function requestText(body) {
  if (!body || typeof body !== "object") return { protocol: "unsupported", texts: [] }
  if (Array.isArray(body.messages)) {
    const texts = [...contentText(body.system), ...body.messages.flatMap((message) => contentText(message?.content))]
    return { protocol: body.system !== undefined ? "anthropic-messages" : "chat-completions", texts }
  }
  if (Array.isArray(body.input) || typeof body.input === "string") {
    const texts = [...contentText(body.instructions)]
    if (typeof body.input === "string") texts.push(body.input)
    else for (const item of body.input) {
      if (item?.type === "message" || item?.role) texts.push(...contentText(item.content))
      // Tool return text can contain a native read of the reference in A/B.
      if (item?.type === "function_call_output" && typeof item.output === "string") texts.push(item.output)
    }
    return { protocol: "openai-responses", texts }
  }
  if (Array.isArray(body.contents)) {
    const texts = []
    for (const part of [...(body.systemInstruction?.parts ?? []), ...body.contents.flatMap((item) => item?.parts ?? [])]) {
      if (typeof part?.text === "string") texts.push(part.text)
    }
    return { protocol: "gemini-content", texts }
  }
  return { protocol: "unsupported", texts: [] }
}

// OpenCode's read tool prefixes each line with its number ("12: text" in 2.x,
// "00012| text" in 1.x) and adds header or footer lines. Only the copy search
// sees the stripped view; the canonical text itself is never rewritten.
const trimEnd = (text) => text.replace(/[ \t\r\n\f\v]+$/, "")
const stripLinePrefixes = (text) => text.split("\n").map((line) => line.replace(/^[ \t]*\d+[|:][ \t]?/, "")).join("\n")
function findAll(text, needle) {
  const found = []
  let start = 0, position
  while ((position = text.indexOf(needle, start)) !== -1) {
    found.push(text.slice(position, position + needle.length))
    start = position + needle.length
  }
  return found
}

export function inspectReference(texts, resource, canonicalText) {
  const canonical = trimEnd(canonicalText)
  const prefix = `Hive canonical reference: ${resource}\n\n`
  const extracted = []
  let injectedCopies = 0
  for (const text of texts) {
    if (text.includes(prefix)) {
      let start = 0, position
      while ((position = text.indexOf(prefix, start)) !== -1) {
        const contentStart = position + prefix.length
        extracted.push(text.slice(contentStart, contentStart + canonical.length))
        injectedCopies++
        start = contentStart + Math.max(canonical.length, 1)
      }
    } else if (canonical.length > 0) {
      const exact = findAll(text, canonical)
      extracted.push(...(exact.length > 0 ? exact : findAll(stripLinePrefixes(text), canonical)))
    }
  }
  const matches = extracted.length > 0 && extracted.every((text) => text === canonical)
  return { copies: extracted.length, injectedCopies, hash: matches ? hash(extracted[0]) : extracted.length === 1 ? hash(extracted[0]) : "", canonical: matches }
}

export default Plugin.define({
  id: "hive.pilot-observer",
  async setup(ctx) {
    const { output, canonical: filename, resource, role } = ctx.options
    let canonical = "", resourceCode = ""
    try { canonical = await readFile(filename, "utf8") }
    catch { resourceCode = "canonical_unavailable" }
    const sequences = new Map(), snapshots = new Map(), registrations = []
    const emit = async (record) => {
      try { await appendFile(output, `${JSON.stringify({ ...record, ...(resourceCode ? { code: resourceCode } : {}) })}\n`, { mode: 0o600 }) }
      catch { /* Missing evidence blocks the runner; never expose raw errors. */ }
    }
    async function metadata(event) {
      try {
        const [agent, session, inventory] = await Promise.all([(async () => {
          try { return event.agent ? await ctx.agent.get({ agentID: event.agent }) : undefined }
          catch { return undefined }
        })(), ctx.session.get({ sessionID: event.sessionID }), ctx.plugin.list()])
        const agentID = agent?.data?.id || agent?.id || event.agent || ""
        const agentName = agent?.data?.name ?? agent?.name ?? ""
        const sessionInfo = session?.data ?? session
        const expected = ["hive.context-loader", "hive.pilot-observer"]
        const external = inventory.data.filter((plugin) => plugin.source.type !== "builtin")
        const plugins = external.map((plugin) => plugin.id ?? "unknown").sort()
        const inventoryVerified = external.every((plugin) => expected.includes(plugin.id) && plugin.state.status === "active") && plugins.includes("hive.pilot-observer")
        return { sessionID: event.sessionID, parentID: sessionInfo?.parentID ?? "", agent: agentID, agentName, model: modelName(event.model), resource: agentID === role ? resource : "", kind: event.kind ?? "primary", plugins, loaderActive: plugins.includes("hive.context-loader"), ...(inventoryVerified ? {} : { code: "plugin_inventory_unverified" }) }
      } catch {
        return { sessionID: event.sessionID, parentID: "", agent: event.agent ?? "", agentName: "", model: modelName(event.model), resource: "", kind: event.kind ?? "primary", code: "metadata_unverified" }
      }
    }
    registrations.push(await ctx.session.hook("context", async (event) => {
      const meta = await metadata(event)
      const sequence = (sequences.get(event.sessionID) ?? 0) + 1
      sequences.set(event.sessionID, sequence)
      const texts = (event.system ?? []).filter((part) => part.type === "text" && typeof part.text === "string").map((part) => part.text)
      const proof = meta.resource ? inspectReference(texts, resource, canonical) : { copies: 0, hash: "", canonical: false }
      snapshots.set(event.sessionID, { meta, sequence, proof })
      await emit({ ...meta, phase: "context", sequence, ...proof, correlated: false, protocol: "assembled" })
    }))
    registrations.push(await ctx.session.hook("model.request", async (event) => {
      if (event.kind !== "primary") return
      const snapshot = snapshots.get(event.sessionID)
      const meta = await metadata(event)
      if (snapshot) snapshot.modelSeen = snapshot.meta.agent === event.agent && snapshot.meta.model === meta.model
      await emit({ ...meta, phase: "model", sequence: snapshot?.sequence ?? 0, copies: 0, hash: "", canonical: false, correlated: Boolean(snapshot && snapshot.meta.agent === event.agent && snapshot.meta.model === meta.model), protocol: "model" })
    }))
    registrations.push(await ctx.session.hook("http.request", async (event) => {
      const meta = await metadata(event), snapshot = snapshots.get(event.sessionID)
      if (event.kind !== "primary") {
        await emit({ ...meta, phase: "http_auxiliary", sequence: 0, copies: 0, hash: "", canonical: false, correlated: false, protocol: "auxiliary" })
        return
      }
      let parsed
      try { parsed = requestText(await event.request.clone().json()) }
      catch { parsed = { protocol: "unsupported", texts: [] } }
      const proof = meta.resource ? inspectReference(parsed.texts, resource, canonical) : { copies: 0, hash: "", canonical: false }
      // If context held an injected reference, its actual outgoing bytes must
      // still agree. A/B can instead expose a read result in later messages.
      const correlated = Boolean(snapshot && snapshot.modelSeen && snapshot.meta.agent === event.agent && snapshot.meta.model === meta.model &&
        (!meta.resource || !meta.loaderActive || (snapshot.proof.canonical && snapshot.proof.injectedCopies === 1)) &&
        (!snapshot.proof.copies || (snapshot.proof.hash === proof.hash && snapshot.proof.injectedCopies === proof.injectedCopies && snapshot.proof.canonical === proof.canonical)))
      await emit({ ...meta, phase: "http", sequence: snapshot?.sequence ?? 0, ...proof, correlated, protocol: parsed.protocol, ...(parsed.protocol === "unsupported" ? { code: "unsupported_protocol" } : {}) })
    }))
    registrations.push(await ctx.session.hook("experimental.ws.handshake", async (event) => {
      if (event.kind !== "primary") return
      await emit({ ...await metadata(event), phase: "websocket", sequence: sequences.get(event.sessionID) ?? 0, copies: 0, hash: "", canonical: false, correlated: false, protocol: "unsupported", code: "unsupported_transport" })
    }))
    return async () => { for (const registration of registrations) await registration.dispose() }
  },
})
