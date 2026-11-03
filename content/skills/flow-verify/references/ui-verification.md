# Proportional UI verification

Read for a UI change, including its first view, final verification and corrections. Drive the browser as [Browser tool](#browser-tool) sets. Exercise changed screens, relevant states and regions displaced, resized or reflowed, with realistic data; capture what a person needs to review. Check desktop and, whenever the application has a responsive layout (a breakpoint or a viewport meta tag) and the diff changes a rendered component, style or layout, the mobile viewport of 440 × 956 CSS pixels (iPhone 16 Pro Max), and each affected supported theme. Record which widths, themes and states were actually checked.

For a correction, check and recapture only affected screens and states, including any neighbors, widths or themes whose layout can change. Preserve earlier evidence for unaffected views and distinguish the new revision. Repeat a complete mobile journey only when navigation or the flow changed or the user explicitly requests it. Reuse suitable captures already made; do not claim coverage for a width, theme or state that was not exercised.

## Automated accessibility audit

Read when reviewing the UX of a UI change. The audit measures; it does not replace the keyboard, focus and interaction checks of the review.

- With `agent-browser`, run `agent-browser a11y --tags wcag2a,wcag2aa,wcag21a,wcag21aa,wcag22aa,best-practice --json` on each reviewed screen and state, including each state reached by interaction, such as an open dialog, popover, menu, panel or mobile navigation, since the audit sees only what is rendered. Before each run, wait for the state to finish loading with `agent-browser wait --load networkidle` or a selector of that state, not a fixed delay. Its embedded axe-core needs no network and is not blocked by the page's content security policy. Without `agent-browser`, run the audit through the browser tool only when it can execute axe-core, and report the audit as not run otherwise.
- Report a violation tagged `wcag*` as a defect with the priority its impact on users sets. Report a violation tagged only `best-practice`, such as `region`, as a recommendation, not as a WCAG failure.
- Treat each `incomplete` result as pending manual review, not as a pass; check it with rendered colors, keyboard or focus as the rule requires, and name the ones left unchecked. An `incomplete` result that comes from an intended pattern of the project's UI library, such as focus guards, is noted once and not reported as a defect.
- Count affected elements from `nodeCount`, since the JSON lists at most 10 nodes per rule, and record `axeVersion` with the evidence.
- axe-core finds on average 57% of WCAG issues, by its maintainer's measure: a run with no violations is not evidence of conformance, and the report never claims conformance from it.
- When the project guidance declares a UI linter, also run it on the running URL, with JSON output when it has one. Report its findings on contrast, overflow, structure, broken images and script errors as measured defects, and its aesthetic or style rules as informational only, since the project's design system governs the look.

## Browser tool

- Use the `agent-browser` skill through the shell when it is installed (`command -v agent-browser`): its snapshots are smaller than an MCP browser's, and a stable flow chains in one command. An installed MCP browser server is not a reason to prefer it. A child uses the tool its brief assigns.
- Use Chrome DevTools MCP only for Lighthouse audits, performance traces, CPU throttling and heap snapshots, or for an independent reproduction after `agent-browser` failed twice. Without `agent-browser`, use the host's browser tool. Report which tool drove each check.
- `agent-browser` has no timeout for `open`, `eval`, `snapshot`, `screenshot` or `close`, and one stuck command blocks its session. Run each command through a 45-second limit: `timeout 45 agent-browser …`, or `gtimeout 45`, or `perl -e 'alarm shift; exec @ARGV' 45 agent-browser …` when neither exists. When the limit expires, run `agent-browser doctor`, close that session and continue in a fresh one; do not resend the command into it.
- Give each flow its own `--session` and close every session you opened with `agent-browser close --session <name>` when the flow ends.

The caller owns acceptance and review. For publishing evidence and instructions, use [flow-walkthrough](../../flow-walkthrough/SKILL.md); follow [artifact layout and retention](../../workspace-conventions/references/artifact-layout.md) for the evidence destination.
