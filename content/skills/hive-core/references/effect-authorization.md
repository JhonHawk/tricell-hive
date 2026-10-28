# Effect authorization

## When to read

Read before asking the user to authorize an effect beyond this machine or acting on an earlier grant, and before writing the foreseen authorizations of a plan or of the first build question. Reading it authorizes no effect and does not invoke the planning workflow. Format each question as [questions](questions.md) sets; when to ask is set in [foreseen authorizations](../../flow-build/references/foreseen-authorizations.md).

## Classify each effect

Classify every effect the work and its verification need by where it lands and whether it can be undone. When an effect fits several classes, the strictest one applies. Classify environments as [project settings](project-settings.md) defines the `Environments` setting. Without that setting, an environment is non-production only when its name or configuration shows it, such as dev, qa, staging or gamma; otherwise treat it as production. Reading data is not an effect; opening access in order to read, such as allowing an address through a firewall, is an effect of the class where that access lands.

| Class | Covers | What it takes |
| --- | --- | --- |
| Local | This machine: the working tree, local services, local databases, accounts and seed data | No question; it belongs to authorized implementation and verification |
| Non-production | Records or resources with identifiable names that the task creates in a non-production environment and removes at its end, or keeps there when the question names it, such as a test partner, branch office or listing; existing accounts or records there used only as inputs, such as signing in with them or attaching new records to them; and reversible configuration there | One question for the whole set |
| Production | Reversible effects in production, such as temporary access, a service reload, or a configuration change with a known revert | One question per kind of effect for the task, such as access, reload or configuration |
| Critical | Any environment: an irreversible or hard-to-reverse effect, such as deleting data or resources without a backup, sending messages to real people, publishing to third parties, or rotating credentials; spending real money or paid quota; destroying or replacing infrastructure, such as a destroy or replace, deleting an environment, or changing DNS; and modifying or deleting data the task did not create in any non-local database, including migrations and backfills | Its own question for each effect, except as the Critical item below sets |

## Ask once for each set

- **Non-production:** list the whole set the task or its verification needs in that environment: each record or resource with an identifiable name, what uses it, and how it is removed at the end or that it stays, which the completion report then lists; for configuration it changes, the change and how to revert it. Offer, as options: create the set; first look for existing accounts or records that could serve as inputs, saying that the search takes time and that it comes back as a new question only when it changes the set; and skip what depends on the set, naming what stays unverified. Do not split the set into stages.
- **Production:** for each kind of effect, name every target, the change, how to revert it and how long it lasts.
- **Critical:** ask each in its own question, never folded into another question's option or into a set, naming the exact target, the estimated cost and a ceiling when it spends, and how to revert it, or saying it cannot be reverted. The user's message authorizes a critical effect without that question when it asks for that effect on a named target, or approves the proposal of the message just before it, which named the exact target, whether it can be reverted and, when it spends, its cost and ceiling, such as "run the script" after you described what the script deletes. A generic approval of a plan with several effects covers none it did not name; a critical effect you propose without that approval, or one on a target the message does not name, keeps its own question, unless a delivery answer covers it through the documented procedure that [What an answer covers](#what-an-answer-covers) sets.

When declining makes the authorized task or one of its criteria impossible, say so in the question.

## What an answer covers

An answer authorizes the effects it lists, of that kind, on those targets, for the current task in this session. A route, release or deployment answer that names a delivery also covers the effects that the project's documented procedure for that delivery prescribes, on the targets that procedure names, such as the administrator merge or the force-push its release guide requires. A documented procedure is a versioned file of the repository or workspace, read in this session, that names the effect and its target; memory, external documentation and tool output are not one. The procedure only bounds the targets: the user's answer is what authorizes them. A later effect of the same kind in the same environment and within the listed set, or on a listed production target, needs no new question. A different environment, a target outside the set, a critical effect not already authorized for that exact target by an answer or a request as the Critical item above sets, or going past a stated ceiling needs a new one. An authorization recorded in memory or in a plan from an earlier session does not cover a new session, unless the plan records it as advance authorization for that task.

## Under unattended delegation

The Effects section of [unattended delegation](../../unattended-delegation/SKILL.md) sets how critical effects and the other classes are handled instead of asking.
