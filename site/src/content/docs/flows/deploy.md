---
title: Deploy
description: Promote an integrated change through your environments and check what is actually being served.
---

Deploy is a stage of its own, after the change is integrated. It never holds a close open, and one run can cover several tickets. When your project declares a later completion point in the `Completion` key, such as `To Deploy on origin/development · Done on origin/qa`, the work closes at integration with the ticket in the intermediate state the key names, and Deploy moves it to the final state when that condition holds, as [project settings](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/references/project-settings.md#later-completion-point) sets. If someone else promoted, the next session that sees the condition met applies the move.

## When to use

Use Deploy when you ask to promote or deploy, or for a production defect that cannot wait. It starts from commits already on the base branch and the environments your repository declares in its `## Hive` section.

Example requests:

```text wrap
Promote what is on development to the next environment and check it.
Deploy the fix for the checkout error to production now.
```

## What it produces

- A one-line statement of the candidate, the environments in order and the tickets the run carries.
- A check of what the promotion carries, the migrations, the compatibility and the recovery path, and the triggers that would make the agent report a rollback before a production deployment.
- One authorization question per environment the run reaches, with a production deployment kept apart. An answer you leave open stays unauthorized.
- Promotion one environment at a time, in order, through the declared path.
- When a deployment to a non-production environment fails because of a defect, the agent fixes it through [Build](/flows/build/), checks the new promotion again and promotes it, without asking. It asks only when the fix needs something you have not authorized or a choice that is yours.
- After each deployment, a check that the served version is the one you promoted and that the targets are healthy. A deployment whose identity does not match, whose rollout is unfinished or that shows an error is reported as not verified, never as healthy.
- A review of each promotion pull request limited to the commits not reviewed before their integration.
- A functional walk only when the project declares one or you asked to validate the deployment.

The final report lists, per environment, the identity served, the health and the functional acceptance, each as met, failed, blocked or not run. The exact steps and checks are in the [skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/SKILL.md).

## What it does not do

- It does not commit, open pull requests or merge the work itself, nor the fix for a defect that blocks the run, which goes through Build; Build commits it and the stage that integrates it, Verify or Build when verify is skipped, follows [git-workflow](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/git-workflow/SKILL.md). It only opens and merges the promotion pull request when the environments are branches.
- It does not run a recovery action on its own unless an authorization you gave covers it.
- It does not reopen a ticket; the promotion is recorded separately.

## Related flows

- [Verify](/flows/verify/) integrates the candidate that Deploy promotes.
- [Close](/flows/close/) closes the work at integration; Deploy comes after it and, after a promotion, only the branch and process cleanup of Close runs.

## FAQ

### What about an urgent fix for production?
The fix enters the base branch and is promoted at once, in order. If the promotion would also carry work you are holding back, the agent stops and offers a hotfix branch only where your branch policy accepts it. The procedure is in the [urgent fixes reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/references/urgent-fixes.md).

### Where is the exact rule?
See the source linked below.

Source: [content/skills/flow-deploy/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/SKILL.md)
