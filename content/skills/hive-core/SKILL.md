---
name: hive-core
description: "Shared procedures every Hive flow and the global guidance use: user questions and decision pauses, scope changes, effect authorization, repository settings, language, third-party requests, secrets, persistent memory, evidence tools, tool scripts and long-running loops. Read the reference whose condition below applies; the global guidance and each flow point to them by path."
---

# Shared procedures

This skill holds the procedures that every flow and the global guidance use. It has no workflow of its own: read only the reference whose condition applies, whole, when the condition applies. Reading one grants no effect.

- Before formulating a decision for the user, with or without a native card, and before ending a message with an offer or request that waits on the user's answer, read [questions](references/questions.md).
- When the user asks for something that would change open work, an answer adds to it, or the work finds it needs more than the request asked, read [scope changes](references/scope-changes.md) before acting on it.
- Before asking the user to authorize an effect beyond this machine, acting on an earlier grant, or writing the foreseen authorizations of a plan or of the first build question, read [effect authorization](references/effect-authorization.md).
- Before resolving a project setting the task needs, such as the base branch, tracker, specs location, environments or completion point, or moving a ticket's state, read [project settings](references/project-settings.md).
- Before naming an identifier, an i18n key or a route, writing an example or template, contributing to an external repository, or resuming or relaying text in another language, read [language](references/language.md).
- Before sending a request to a third-party service, read [third-party requests](references/third-party-requests.md).
- Before retrieving or documenting a secret, reading output that may hold one, or when one is exposed, read [secrets](references/secrets.md).
- Before writing persistent memory, or when memories or sources disagree, read [memory](references/memory.md).
- Before looking up external documentation, or locating code in a repository with a `.codegraph/` directory, read [evidence tools](references/evidence-tools.md).
- Before starting or resuming work that repeats cycles until a condition is met, such as an iterative route or unattended work that repeats cycles, read [loops](references/loops.md).
- When the host offers a tool, not a shell, that runs a script able to call other host tools, before writing such a script, or three or more read-only calls it could batch, read [tool scripts](references/tool-scripts.md).
