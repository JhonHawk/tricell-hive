---
name: flow-deploy
description: Promote an integrated candidate through the declared environments and verify what is served: rollback triggers, served-identity and health checks, and urgent production fixes. Use when the user asks to promote or deploy, or for a production defect that cannot wait; not for commits, pull requests or merges.
---

# Promote and verify what is served

## Entry and exit

This is a stage of its own, after integration, and one run can cover several tickets. Enter with an integrated, identified candidate (the commits on the base branch), the repository's declared `Environments`, and the authorization for each environment it will reach; when these exist, enter directly, with no plan or build before it, and a request to promote or deploy is such an entry. Exit with, per environment, the served identity, the health and the functional acceptance each stated as met, failed, or "not run". This skill does not grant effects beyond the user's instructions.

## Prepare

Before resolving needed project settings, read [project settings](../flow-plan/references/project-settings.md). Before a user decision or manual validation request, read [questions](../flow-plan/references/questions.md). Before asking to authorize an effect beyond this machine, read [effect authorization](../flow-plan/references/effect-authorization.md). Before delegating or starting an external-work wait, read [delegation and waits](../flow-research/references/delegation.md). Before progress reports, completion reports or messages containing citations, links or externally defined IDs, read [reporting](../flow-close/references/reporting.md). Before a branch, merge or push that carries a promotion, read [git-workflow](../git-workflow/SKILL.md), which keeps the promotion invariant. Before stored-data changes, read [persistent data changes](../flow-build/references/data-changes.md). For the checks of this stage, read [deployment checks](references/deployment-checks.md). For a production defect that cannot wait, read [urgent fixes](references/urgent-fixes.md).

## Procedure

1. **Name the run.** Say in one line the candidate, the environments in the declared order, and the tickets it carries. Without an `Environments` setting, ask once which environments the request reaches.
2. **Reconcile the promoted candidate.** Do the predeploy reconciliation in [deployment checks](references/deployment-checks.md#predeploy): what the promotion carries, destinations, migrations, compatibility, recovery and reusable prior evidence.
3. **Decide the rollback triggers.** Before a production deployment of code or infrastructure, record them as [deployment checks](references/deployment-checks.md#predeploy) sets, in the plan or in the conversation when there is none, before the first effect.
4. **Get the authorization.** Classify each effect and ask for what is missing as [effect authorization](../flow-plan/references/effect-authorization.md) sets: one answer per environment the run reaches, with a production deployment kept apart. A request to promote or deploy covers its delivery chain, not the effects a later environment's verification needs. An authorization left unanswered stays unauthorized, also under unattended delegation.
5. **Promote and deploy**, one environment at a time and in order, only through the declared promotion path that [git-workflow](../git-workflow/SKILL.md) keeps, and wait for the CI run the promotion triggers as [deployment checks](references/deployment-checks.md#regression-and-release-gates) sets.
6. **Check what is served.** After each deployment, verify the served identity and the health of every affected target as [deployment checks](references/deployment-checks.md#postdeploy) sets. A trigger that fires is reported to the user with its evidence; its recovery action runs only when an existing authorization covers it. A deployment with an unmatched identity, an unfinished rollout or an error observed after it is reported as not verified, never as healthy.
7. **Functional walk gate.** The walk runs only when the project declares it or the user asked to validate the deployment, through the children [deployment checks](references/deployment-checks.md#functional-walk) names; otherwise report it as not run, and do not offer it.
8. **Report** per environment: the identity served, the health and the functional acceptance, each as met, failed, blocked or "not run", the triggers decided, and the tickets that stay out of their completion state until a due check runs. Record the promotion in the project's mechanism, without reopening the ticket. Run [flow-close](../flow-close/SKILL.md) under the conditions that [flow-build](../flow-build/SKILL.md#check-and-hand-off) sets.
