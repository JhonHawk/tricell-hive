# Hive history and technical provenance

Prepared on 2026-10-02, America/Mexico_City. This document describes five milestones preserved in Hive's history and separates the content of the rules, the dates recorded by Git, and the time corroboration available on GitHub. It documents the project's evolution; it does not show who invented these practices, their priority over other projects, or independence from other projects. No comparative study was carried out.

## History that can be cross-checked

The original frozen base is `39dfdb62fc73a91f373f3e46769891f329cea405`; the curated historical base is `c71d6167b10d921cf55cc03b0f8a2ed786364095`. Both contain the 934 historical commits and 70 merges. The documentation was added afterwards through commit `fa5528e65480d1ee2a1daba052d5c3fc3f76de27`, without changing those 934 commits. The curation preserves names, dates with time zones, parents, and their order, including the commits that ended up empty. It replaces emails with verified GitHub noreply addresses, excludes private material, and adds classified attributions. These changes produce new commit identifiers; 51 signature headers were removed, without checking their cryptographic validity.

The five commit pairs below belong to the original/curated map verified during preparation. The files of the five curated milestones were checked through the authenticated API of the private repository; anonymous access depends on visibility. The original identifiers are historical correspondences and do not require the original objects to remain accessible. The noreply policy was checked on the candidate; PR references and caches may retain earlier history and emails.

## Five preserved milestones

The dates below are Git author and committer dates, identical to each other for each milestone and preserved in the candidate. They are expressed in UTC to avoid day differences between time zones.

1. **Delegation and separate verification, 2026-06-26T10:32:02Z.** In the preserved root, `global/rules/workflow/agent-routing.md`, section `Delegation Thresholds`, assigns bounded explorations to a subagent when the workflow spans several files. The rule `Verification runs in fresh context` separates verification from implementation. This evidences general delegation instructions and independent verification, without yet equating them with the later three-front research contract. Original `d066234b87d101354091e54e09fd48a626753911` → [curated commit d91404d](https://github.com/JhonHawk/tricell-hive/commit/d91404deed85475e3d990a739b5e5297fac912df) · [milestone file](https://github.com/JhonHawk/tricell-hive/blob/d91404deed85475e3d990a739b5e5297fac912df/global/rules/workflow/agent-routing.md).

2. **Delegated external lookups, 2026-08-16T05:15:25Z.** In `global/rules/workflow/agent-routing.md`, section `Delegation Gates`, an explicit threshold appears for delegating several external reads or several queries needed to establish a fact. It is a documented extension of how research and evidence gathering are split. Original `24663660af78ff22b87778468b029390c6a5133d` → [curated commit dd0b788](https://github.com/JhonHawk/tricell-hive/commit/dd0b78801e56df083f8f53f24566030b2d9af050) · [milestone file](https://github.com/JhonHawk/tricell-hive/blob/dd0b78801e56df083f8f53f24566030b2d9af050/global/rules/workflow/agent-routing.md). That date corresponds to August 15 in America/Mexico_City.

3. **A dedicated research activity, 2026-09-20T20:16:21Z.** `global/skills/flow-research/SKILL.md` is added, covering investigation of code and documents, cross-checking of claims, contradictions, and limits. Selecting it does not authorize implementation or publication. Original `0ebb5b242182b5a7b09e720c3f897eac0c301c0f` → [curated commit 03bf6af](https://github.com/JhonHawk/tricell-hive/commit/03bf6aff5f6b3176adf890e3acf4c819a8b7a619) · [milestone file](https://github.com/JhonHawk/tricell-hive/blob/03bf6aff5f6b3176adf890e3acf4c819a8b7a619/global/skills/flow-research/SKILL.md).

4. **Three fronts and critical review, 2026-09-21T19:29:34Z.** `AGENTS.md`, section `Complete research`, requires reviewing Hive's previous implementation, external evidence, and reference sources. It delegates the three fronts to independent subagents in parallel when capacity allows. The main thread cross-checks sources, challenges assumptions, resolves contradictions, and synthesizes, without accepting a conclusion merely because another agent asserted it. The curation preserves that contract and generalizes the location of private sources. Original `5a51c8f8788560f977757887d8f19aa1d6412539` → [curated commit ceab31e](https://github.com/JhonHawk/tricell-hive/commit/ceab31eb5095997d1ee5e37c5e30c13e66425401) · [milestone file](https://github.com/JhonHawk/tricell-hive/blob/ceab31eb5095997d1ee5e37c5e30c13e66425401/AGENTS.md).

   ```text
   Hive implementation ----\
   External evidence -------+--> Main thread: cross-check and synthesis
   Reference sources -----/
   (independent subagents, in parallel when capacity allows)
   ```

5. **An explicit research role, 2026-09-29T07:12:07Z.** `content/agents/review/hive-research.md` defines a role for read-only delegated research over code or contradictory sources, with evidence-backed synthesis. The preserved role documents the intent and the contract; its existence does not prove that every host loads or runs it correctly. Original `9b110408cce6555a8b68e37e601d8d54f2e4507f` → [curated commit 3815430](https://github.com/JhonHawk/tricell-hive/commit/3815430646939d18aec21bdc92d077fc7d743de7) · [milestone file](https://github.com/JhonHawk/tricell-hive/blob/3815430646939d18aec21bdc92d077fc7d743de7/content/agents/review/hive-research.md).

## What the external dates corroborate

The GitHub API for the PRs linked below was queried on 2026-10-03T02:12:22Z. Git dates are modifiable metadata; so are they when GitHub returns them. A PR's creation date dates its container, without showing which commits it contained at the time. A commit that is an ancestor of the merge recorded by GitHub has later corroboration of presence, without showing the date it was conceived.

- [PR2](https://github.com/JhonHawk/tricell-hive/pull/2), created 2026-08-13T07:27:49Z and merged 2026-08-13T07:38:46Z: the root d066234 is an ancestor of the merge. It is not an exact member of this PR's change list; it is not claimed that PR2 introduced it.
- [PR3](https://github.com/JhonHawk/tricell-hive/pull/3), created 2026-08-17T09:31:24Z and merged 2026-08-17T11:36:33Z: the external-lookups milestone is an ancestor of the merge, without exact membership in its changes.
- [PR48](https://github.com/JhonHawk/tricell-hive/pull/48), created 2026-09-29T06:11:39Z and merged 2026-09-29T06:11:58Z: flow-research and the three fronts are ancestors of the merge, without exact membership in its changes.
- [PR51](https://github.com/JhonHawk/tricell-hive/pull/51), created 2026-09-29T07:14:42Z and merged 2026-09-29T07:22:50Z: the hive-research commit is an exact member of the queried list and an ancestor of the merge.

The relationship between [PR1](https://github.com/JhonHawk/tricell-hive/pull/1), merged on June 26, and the current root was not established in the bounded exploration. It is not used as corroboration of that root, nor is it claimed that the relationship does not exist. The first corroboration observed for the root among the verifiable merges is that of PR2, in August. The PRs and their relationships were verified through the API with authentication before the transition. Their public access depends on the repository's visibility. The files of the curated milestones were checked both against local objects and through the authenticated API of the private repository.

## Attributions and reproduction

Preserving Hive's evolution includes acknowledging real adaptations. The resources adapted from diagram-design have a [full MIT notice](../../../content/skills/flow-report/references/license-diagram-design.md). The curated history also preserves the Superpowers notice associated with the historical adapter at [global/hooks/flow-session-context/LICENSE.superpowers](https://github.com/JhonHawk/tricell-hive/blob/05b9979a31917d43743cddad3170c74548bb0c66/global/hooks/flow-session-context/LICENSE.superpowers), in its applicable snapshots. These credits are not removed to support a claim of originality.

To cross-check a milestone, from a copy of the repository with the curated history, replace `<curated-commit>` and `<repository-path>` with the milestone from this timeline:

```sh
git show <curated-commit>:<repository-path>
git show -s --format='%H%n%aI%n%cI' <curated-commit>
```

This document was prepared on 2026-10-02 for a later documentation incorporation. It did not exist in the 934 earlier commits and was not inserted retroactively into them. The documentation was incorporated through a new commit, expressly authorized and carrying its real creation date. Adding it afterwards does not show that the document existed at the dates of the milestones. The evidence supports preserved rules and specific historical relationships; it does not substitute for behavior tests, cryptographic evidence of dates, or a comparative study of priority.
