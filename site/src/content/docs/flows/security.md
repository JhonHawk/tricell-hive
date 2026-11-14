---
title: Security review
description: Ask for an independent review of a change's security boundaries, and check whether a published advisory affects your repositories.
---

Hive has two security tools, and neither starts on its own. One reviews a change you are making. The other checks whether an advisory published elsewhere affects your repositories.

## Reviewing a change

The [hive-review-security](https://github.com/JhonHawk/tricell-hive/blob/master/content/agents/review/hive-review-security.md) role is an independent reviewer that reads and tests but changes nothing. It looks at the places where a change meets a security boundary: authentication, authorization, untrusted input, command or process execution, secrets, cryptography, dependencies and data exposure. It traces a suspected problem from its source to where it does harm, runs the scanners your project already has, and reports a secret by its type and location, never its value.

### When it runs

Only when you ask for it. The agent does not start it by itself and your answer to the code review question does not count as asking. Once you ask, [Verify](/flows/verify/) runs it on the candidate at the start of verification, before you are asked to validate. You can also ask for it on a change that is already merged.

Example requests:

```text wrap
Run a security review on this branch before I validate it.
Review the security of the commit we merged yesterday.
```

### What you get

A report of findings by priority, each with the conditions that make it exploitable, its impact and its evidence, plus what was covered and what could not be checked. The reviewer exploits nothing on live systems, rotates no credentials and fixes no source. Confirmed blocking findings are fixed in the verification round; the rest go to the findings of [Close](/flows/close/).

## Checking an advisory

Use the [security advisory mode](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/references/security-advisory.md) of the Research skill when a CVE, a published advisory or a package compromise makes you ask whether it affects your projects. It reads the advisory from a primary source, inspects the repositories you name without changing them, and reports for each one whether it is affected, mitigated, not affected or unknown, with the file and line that show it. If an application's exposure depends on how its code reaches the vulnerable path, it offers a reachability assessment by the security reviewer, which runs only if you ask. Fixes then happen in each project's own session.

Example request:

```text wrap
Does advisory GHSA-xxxx-xxxx-xxxx affect the repositories under ~/projects/clients?
```

It is not for reviewing a change, and not for checking a package before you install it.

## Related flows

- [Verify](/flows/verify/) is where a requested security review runs.
- [Build](/flows/build/) mentions the review when a change crosses a boundary.
- [Research](/flows/research/) is the method the advisory sweep follows.

## FAQ

### Is a security review part of every change?
No. Ask for it when a change touches a security boundary or you want a second look.

Source: [content/agents/review/hive-review-security.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/agents/review/hive-review-security.md)
