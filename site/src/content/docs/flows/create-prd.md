---
title: Create PRD
description: Write a requirement document for a module or feature, restructure one someone else wrote, update it with its author's answers, or review it against the code, so development builds from rules it does not have to decide.
---

Create PRD writes the requirement document of a module or feature: what it must do and why, as rules that development can build without deciding a business rule on its own. It holds no screen design, which [Create design](/flows/create-design/) draws, no technical how, which [Plan](/flows/plan/) settles, no ticket, which [Create issue](/flows/create-issue/) writes, and no schedule. The skill covers four uses: a new document, one that someone else wrote, an update with the author's answers, and a technical review against the code. It runs in the main conversation, with or without the repository, and changes no application code.

## When to use

Ask for a requirement document, a PRD, or the functional requirements or business rules of a module or feature. Hand over a requirement document to validate or complete, or ask whether one is ready for development. [Hive mode](/flows/hive-mode/) hands the request over when its deliverable is such a document, and the skill then asks its own questions.

Example requests:

```text wrap
Write the PRD for the returns module from these meeting notes.
Here is the requirements document the client sent; restructure it and tell me what is missing.
Is the returns document ready for development? Check it against the code.
```

## What it asks

- At most three questions per round, only what changes a rule and nothing it read answers, each with a recommended answer. The questions whose answer changes the most rules come first, and it stops when the rest would change no rule or you say to go on.
- A business rule you do not decide, because it belongs to someone else or is not known yet, is never assumed. It is written as to decide, with who answers it.
- Where the document goes, when you named no destination, and which document tool to use when the session can write to more than one.
- Whether you approve a version, once it passes its checks and its review. Approval freezes the version, and its rules go to development as written.
- For a document that someone else wrote and you are not its author, no questions in the session: the questions for the author are written into the document and handed to you to send.

## What it produces

- A document in a fixed format: problem, goals, non-goals, user stories, requirements, acceptance criteria, success metrics, open questions and a version history. Each requirement is one rule in an entry of its own, never a row of a table, with a stable code, a priority, a state (decided, assumption, to decide or cancelled), its source, and a verdict against the code (not checked when the session has no code). Screen layouts, technical solutions and dates the source held are left out and named in the report as material for the design, the plan or the schedule.
- A check of the draft against a ten-point ready list (backed by `check_prd.py`), run in the same conversation, and an independent review by a separate reviewer agent, [hive-review-prd](https://github.com/JhonHawk/tricell-hive/blob/master/content/agents/review/hive-review-prd.md), once per version. A finding that needs a decision becomes a to decide rule or an open question; the reviewer never decides a rule.
- The document where it lives: in a shared document tool such as Notion, Google Docs or Claude Docs when the session can write there, otherwise as a Markdown file with a PDF beside it. A version that is approved is copied under a title that says it is not to be edited, and its link goes in the version history. A change after approval is a new version.
- A state that the checks support, never the one the source claims: draft, iterating, approved by business, technical review or ready for development. Only the technical review sets the last one, and only in a session with the repository.
- For a document that someone else wrote, a new document that carries each rule as its author wrote it and names the original. The original is never edited, and the new document starts in iterating whatever the original says about itself.
- A technical review compares each rule with the code and gives it a verdict. Where the code contradicts a rule or lacks what it needs, the question goes to the document's owner, and the rule is never rewritten for them.

The procedure for each part lives in the skill and its references: [PRD format](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-prd/references/prd-format.md), [incoming document](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-prd/references/incoming-document.md), [documents](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-prd/references/documents.md) and [technical review](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-prd/references/technical-review.md).

## How it ends

It reports the document's link or path, its version and state, what holds the next state back (each to decide rule, each open question with who answers it, and each failed check), the assumptions it carries, and what the source held that the document leaves out. It then offers the next step as a reply, such as creating its tickets with [Create issue](/flows/create-issue/), designing its screens, running its technical review in a session with the repository, or planning it. Writing the document starts no work on it.

## What it does not do

- It does not approve a version without your answer, and it does not set a state because the source claims it.
- It does not decide a business rule that nobody decided.
- It does not edit, comment on or move a document it did not create.
- It does not change a document's sharing or permissions, or lock a copy; it says how you can.
- It does not write tickets, draw screens, plan the build or touch application code.

## Related flows

- [Hive mode](/flows/hive-mode/) sends a request for a requirement document here.
- [Create issue](/flows/create-issue/) writes the tickets of an approved document and checks that they cover its rules.
- [Create design](/flows/create-design/) draws its screens, and [Plan](/flows/plan/) settles how it is built.

## FAQ

### Can I use it without the repository?
Yes. Drafting, restructuring and updating need none. Without the code every rule's verdict is not checked, and the technical review waits for a session in the repository.

### Can it work from a document the client sent?
Yes. It reads it whole, keeps each rule as its author wrote it, marks as to decide whatever can be read two ways, and writes the questionnaire for the author.

### Where is the exact rule?
Source: [content/skills/create-prd/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-prd/SKILL.md)
