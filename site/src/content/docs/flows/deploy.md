---
title: Deploy
description: Promote an integrated change through your environments and check what is actually being served.
---

Deploy is a stage of its own, after the change is integrated. It never holds a close open, and one run can cover several tickets. When your project declares a later completion point in the `Completion` key, such as `To Deploy on origin/development · Done on origin/qa`, the work closes at integration with the ticket in the intermediate state the key names, and Deploy moves it to the final state when that condition holds, as [project settings](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/project-settings.md#later-completion-point) sets. If someone else promoted, the next session that sees the condition met applies the move.

## When to use

Use Deploy when you ask to promote or deploy, or to promote the fix of a production defect that cannot wait. It starts from commits already on the base branch and the environments your repository declares in its `## Hive` section. A defect whose fix is not integrated yet goes first to [Research](/flows/research/) while its cause is unknown, or to [Build](/flows/build/) once it is known.

Example requests:

```text wrap
Promote what is on development to the next environment and check it.
Deploy the fix for the checkout error to production now.
```

## What it produces

- A one-line statement of the candidate, the environments in order and the tickets the run carries.
- A check of what the promotion carries, the migrations, the compatibility and the recovery path, and the triggers that would make the agent report a rollback before a production deployment.
- One authorization question per environment the run reaches, with a production deployment kept apart. Asking to promote to a named environment is itself the delivery choice; it covers production only when you name production. For production, the question also names the recovery action of each rollback trigger, with its target and how to revert it. A release that publishes is a critical effect with its own question. An answer you leave open stays unauthorized.
- Promotion one environment at a time, in order, through the declared path: a pull request between environment branches, or, when the environments are not branches, the mechanism your project documents.
- When a deployment to a non-production environment fails because of a defect, the agent fixes it through [Build](/flows/build/), lists the commits that were not in the candidate, checks the new promotion again and promotes it, without asking. When the new promotion carries other commits, it names them and asks once whether each goes in or is reverted on the base branch. It asks only when the fix needs something you have not authorized or a choice that is yours.
- After each deployment, a check that the served version is the one you promoted and that the targets are healthy. A deployment whose identity does not match, whose rollout is unfinished or that shows an error is reported as not verified, never as healthy.
- A review of each promotion pull request limited to the commits not reviewed before their integration.
- A functional walk only when the project declares one, in its `AGENTS.md` or in the plan, or you asked to validate the deployment. The effects the walk needs in that environment are asked before it runs.

The final report lists, per environment, the identity served, the health and the functional acceptance, each as met, failed, blocked or not run. The exact steps and checks are in the [skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/SKILL.md).

## What it does not do

- It does not commit, open pull requests or merge the work itself, nor the fix for a defect that blocks the run, which goes through Build; Build commits it and the stage that integrates it, Verify or Build when verify is skipped, follows [git-workflow](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/git-workflow/SKILL.md). It only opens and merges the promotion pull request when the environments are branches.
- It does not run a recovery action unless the answer you gave to the production question covers it; when a trigger fires and no answer covers it, it reports the trigger and asks.
- It does not reopen a ticket; the promotion is recorded separately.

## Related flows

- [Verify](/flows/verify/) integrates the candidate that Deploy promotes.
- [Changes and loops](/flows/overview/) shows what happens when you ask to add or remove tickets during a run, and where a defect found while promoting returns the work.
- [Close](/flows/close/) closes the work at integration; Deploy comes after it and, after a promotion, only the branch and process cleanup of Close runs.

## FAQ

### What about an urgent fix for production?
The fix enters the base branch and is promoted at once, in order. If the promotion would also carry work you are holding back, the agent stops and offers a hotfix branch only where your branch policy accepts it. The procedure is in the [urgent fixes reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/references/urgent-fixes.md).

### Where is the exact rule?
See the source linked below.

Source: [content/skills/flow-deploy/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/SKILL.md)
