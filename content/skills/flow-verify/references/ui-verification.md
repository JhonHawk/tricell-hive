# Proportional UI verification

Read for a UI change, including its first view, final verification and corrections. Use the available browser automation capability, or the host's browser control when it is absent. Exercise changed screens, relevant states and regions displaced, resized or reflowed, with realistic data; capture what a person needs to review. Check desktop and, when the application supports them and the change affects them, the mobile viewport of 440 × 956 CSS pixels (iPhone 16 Pro Max), and each affected supported theme. Record which widths, themes and states were actually checked.

For a correction, check and recapture only affected screens and states, including any neighbors, widths or themes whose layout can change. Preserve earlier evidence for unaffected views and distinguish the new revision. Repeat a complete mobile journey only when navigation or the flow changed or the user explicitly requests it. Reuse suitable captures already made; do not claim coverage for a width, theme or state that was not exercised.

The caller owns acceptance and review. For publishing evidence and instructions, use [flow-walkthrough](../../flow-walkthrough/SKILL.md); follow [artifact layout and retention](../../workspace-conventions/references/artifact-layout.md) for the evidence destination.
