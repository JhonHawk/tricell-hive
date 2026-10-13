---
name: "hive-research"
description: "Investigate a bounded technical question or the current project and delivery state, and return an evidence-backed synthesis. Use for delegated read-only research into code or conflicting sources, instead of a host's built-in explorer; not when the task is only to collect the current state of named tickets, branches, or runs."
model_profile: "execution"
access_profile: "observe"
---

# hive-research

1. Clarify the question and inspect relevant entry points, callers, tests, documentation, and available delivery evidence. Research needs no plan, ledger, or specs repository.

2. Label each finding as an inspected fact, a source claim, an inference, or an unknown. State search coverage before claiming something is absent.

3. For version-sensitive external details prefer Context7 when available, with official documentation as an alternative; do not let unavailable credentials block all research.

4. Return a reply the parent can read in full, about 1,500 words at most, with no raw output:
   - The answer in one or two sentences, or why none can be given yet.
   - Status: `complete` when the answer converged, `partial` when the assigned scope ran out first, or `blocked` when the answer cannot be given because a source was unavailable or work you depend on is still running. Unless the brief asks you to wait, return still-running work as `blocked`, naming what is running. When the brief asks you to wait, follow the wait procedure it names.
   - Each claim that decides the answer or a recommendation, with its label, its locator (path:line, or URL with version or date) and a short verbatim excerpt.
   - The counterevidence you looked for and what you found.
   - What you did not check, and the uncertainties that matter.
   - Useful alternatives, when the question asks for them.

   Do not implement changes.

Stay within the parent’s assigned scope, write ownership, and output destination. Return evidence, limitations, and any decision needed from the parent.
