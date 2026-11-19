---
name: hive-rule-prose
description: "Hive rule leaf: read before writing a commit message, pull request text or documentation. Cuts filler and AI patterns and sets a plain technical style."
disable-model-invocation: true
---

# Write plain technical text

Read before you write a commit message, a pull request title or body, or documentation, and again to check them in verify. Write so a tired engineer understands on the first read.

## Cut the patterns

- Delete filler: "in order to" becomes "to"; drop "it is important to note that", chatbot closers and praise openers.
- Say "is" or "has", not "serves as" or "boasts". Say "use", not "utilize".
- Drop the vocabulary that signals nothing: additionally, crucial, delve, pivotal, showcase, testament, vibrant, landscape.
- Drop trailing phrases that add no fact ("ensuring…", "highlighting…"), and "not just X, but Y": state the point.
- Do not force three items, and do not use "from X to Y" unless both sit on a real scale.
- Give a number or a mechanism where you wrote an adjective. Cut a sentence that could sit unchanged in another project.
- Allow one hedge, "may". Name a source or delete the attribution.
- Avoid dashes as connectors: use a period or a comma. Use a colon only before a list or an example.
- Use straight quotes, sentence-case headings and no decorative emoji. Bold only a lead-in that adds detail, never a label that restates its line.

## Write the sentence

- One idea per sentence. Split a sentence of more than about 25 words, or an instruction of more than about 20.
- Use active voice and name the actor, unless it is unknown.
- Write instructions as commands. Put the condition before the step it guards.
- Name each thing once and repeat the name. Write the real symbol, flag or path, not a description of it.
- Make "it", "they" and "this" point at one noun. Keep "the" and "a".
- Use no metaphor, idiom or figurative verb where a literal phrase exists.

## Pick the document's mode

One document has one mode: a tutorial (steps that each show a result), a how-to (actions toward a goal), a reference (dry facts that mirror the thing described) or an explanation (context and reasons, where opinion is allowed). Split and link instead of mixing.

## Commit messages and pull request text

- A commit subject says what changed in the imperative, in the repository's convention. The body says why, when the diff does not.
- A pull request body is a briefing a reviewer reads in under a minute: what changed, why, how it was checked. Link logs, hash lists and metric tables; do not paste them.
- Make each count or tree claim true at the commit that lands it, with the command that regenerates it.

When a rule here makes a sentence worse, fix the sentence another way. Product interface strings follow the product's copy guide, not this leaf.
