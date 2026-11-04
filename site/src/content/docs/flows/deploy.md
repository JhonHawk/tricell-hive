---
title: Deploy
description: Promote an integrated change through your environments and check what is actually being served.
---

Deploy is a stage of its own, after the change is integrated. One run can carry several tickets and moves them through the environments your repository declares in its `## Hive` section, one at a time and in order; each ticket moves in the tracker its key belongs to.

## When to use

Use Deploy when you ask to promote or deploy, or to promote the integrated fix of an urgent production defect, as [urgent fixes](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/references/urgent-fixes.md) defines one. It starts from commits already on the base branch. A defect whose fix is not integrated yet goes first to [Research](/flows/research/) while its cause is unknown, or to [Build](/flows/build/) once it is known.

Example requests:

```text wrap
Promote what is on development to the next environment and check it.
Deploy the fix for the checkout error to production now.
```

## What it asks

One authorization question per environment the run reaches, with production kept apart. Asking to promote to a named environment is itself the delivery choice; it covers production only when you name production. The production question also names the recovery action for each rollback trigger, so one answer covers what the agent may do if a signal fires. Publishing a release is a critical effect with its own question unless you asked for it. A functional walk of the environment runs only when your project declares one or you ask for it, and its effects are asked first. The walk uses the accounts and records that already exist there, found through your project's account index; it never creates accounts or test data in that environment on its own, as the [test accounts rule](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md) sets.

## What it produces

- A statement of the candidate, the environments in order and the tickets the run carries.
- A check of what the promotion carries (migrations, compatibility, recovery) and the rollback triggers decided before any production deployment.
- The promotion itself, through a pull request between environment branches or through the mechanism your project documents.
- After each deployment, a check that the served version is the one you promoted and that the targets are healthy. A mismatch, an unfinished rollout or an error is reported as not verified, never as healthy.

When a deployment fails because of a defect, the fix goes through [Build](/flows/build/) and the run starts again with it.

## How it ends

With a report that lists, per environment, the identity served, the health and the functional acceptance, each as met, failed, blocked or not run. [Close](/flows/close/) then ends with its footer, and removes what the run started once you choose to close. The steps and checks are in the [flow-deploy skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/SKILL.md), its [deployment checks](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/references/deployment-checks.md) and its [urgent fixes](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/references/urgent-fixes.md) reference.

## What it does not do

- It does not commit, open pull requests for or merge the work itself; that belongs to the stage that integrates it, following [git-workflow](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/git-workflow/SKILL.md). It opens and merges only the promotion pull request when the environments are branches.
- It does not run a recovery action your production answer does not cover; it reports the trigger and asks.
- It does not skip an environment in the chain.

## Related flows

- [Verify](/flows/verify/) integrates the candidate that Deploy promotes.
- [Close](/flows/close/) closes the work at integration; Deploy comes after it.
- [Changes and loops](/flows/overview/) shows what happens when you add or remove tickets during a run.

## FAQ

### What about an urgent fix for production?
The fix enters the base branch and is promoted at once, in order. If the promotion would also carry work you are holding back, the agent stops and offers a hotfix branch only where your branch policy accepts it.

Source: [content/skills/flow-deploy/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/SKILL.md)
