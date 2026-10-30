import { Plugin } from "@opencode/plugin"

const snapshot = __HIVE_CONTEXT_SNAPSHOT_DATA__

const resources = new Map([
  ["hive-review-plan", "flow-plan/SKILL.md"],
  ["hive-review-harness", "harness-audit/SKILL.md"],
])

export default Plugin.define({
  id: "hive.context-loader",
  async setup(ctx) {
    // Identity tracks only our own parts, never matching or removing foreign text.
    const ownParts = new WeakSet()
    const registration = await ctx.session.hook("context", async (event) => {
      let agent
      try { agent = await ctx.agent.get({ agentID: event.agent }) } catch {
        // Markdown agents may not be available through the native lookup yet.
      }
      const role = agent?.data?.id || agent?.id || event.agent
      const resource = resources.get(role)
      if (!resource) return
      const result = snapshot?.schema === 2 && Array.isArray(snapshot.references)
        ? snapshot.references.find(item => item.role === role && item.resource === resource && typeof item.text === "string" && /^[0-9a-f]{64}$/.test(item.sha256))
        : undefined
      const text = !result
        ? `Hive reference not supplied: ${resource} (missing_snapshot). Report the missing dependency and pause work that requires it; do not treat its procedure as loaded.`
        : `Hive canonical reference: ${resource}\n\n${result.text}`
      for (let i = event.system.length - 1; i >= 0; i--) {
        if (ownParts.has(event.system[i])) event.system.splice(i, 1)
      }
      const part = { type: "text", text }
      ownParts.add(part)
      event.system.push(part)
    })
    return () => registration.dispose()
  },
})
