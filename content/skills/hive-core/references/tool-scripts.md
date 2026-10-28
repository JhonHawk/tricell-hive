# Tool scripts

## When to read

Read only when the host offers a tool that runs a script able to call other tools: before writing such a script, and before making three or more read-only tool calls it could batch in one. Reading this reference does not invoke a workflow.

## Batch calls into one script

When the host offers a tool that runs a script able to call other tools, make three or more read-only calls that script can reach, or one whose result you would only filter, sort, or count, from one script that returns only what the answer needs, since every result in the conversation is re-read for the rest of the session. Keep direct a single call whose result you must read whole; skill, reference and role files are read whole, as the global guidance sets; delegate a result whose relevant part needs judgment to find, as [delegation and waits](../../flow-research/references/delegation.md) sets. Calls inside a script may skip the host's approval, so put a write, edit, or state-changing command in one only when it is authorized and confirmed as a direct write, and keep secret values out of the script's text.
