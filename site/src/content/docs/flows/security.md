---
title: Security
description: Ask for an independent review of a change's security boundaries and get prioritized findings, without fixes.
---

Security is a review you ask for. An independent reviewer examines the places where a change meets untrusted input, permissions, process execution, public data or dependencies, and reports what it found.

## When to use

Use Security when you want the boundaries of a change reviewed, before you validate it or after it was merged. The agent never starts it on its own. When a change crosses a security boundary, it names the boundary and offers the review in one line, without stopping the work.

Example requests:

```text wrap
Run a security review on this branch before I validate it.
Review the security of the commit we merged yesterday.
```

A declared `Review` value does not count as asking.

## What it produces

A report of findings by priority, from P0 to P3, each with its preconditions, impact and evidence, plus the boundaries covered and not covered and what could not run. The reviewer reads and tests without exploiting anything and reports a secret by its type and location, never its value.

The review runs on the current candidate before you are asked to validate it, or before the push or merge when the delivery has no validation stop. Once you ask, it also covers boundaries that appear later in the same change. For a merged change it runs on the merged commit. The exact placement is in the [skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-security/SKILL.md).

## What it does not do

- It does not fix anything. Fixes for a merged commit go in a follow-up pull request of the same ticket, under the delivery already chosen for it.
- It does not sweep your repositories for a published advisory; that is the [security-advisory skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/security-advisory/SKILL.md).
- It does not check a package before you install it.

## Related flows

- [Verify](/flows/verify/) is the step it usually runs inside; it can also run on a merged commit.
- [Build](/flows/build/) offers the review when a change crosses a boundary.

## FAQ

### Can I ask for it later in the work?
Yes. Ask at any point before the change is merged, or afterwards for the merged commit.

### Where is the exact rule?
See the source linked below.

Source: [content/skills/flow-security/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-security/SKILL.md)
