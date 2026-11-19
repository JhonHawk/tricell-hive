---
name: hive-rule-promotion-path
description: "Hive rule leaf: read before a branch, merge or push that touches an environment branch after the base, such as qa or production. Changes reach them only through the declared promotion path."
disable-model-invocation: true
---

# Promotion path

Read before a branch, merge or push that carries a change to a branch after the base in the environment chain.

- Branches after the base in the environment chain, such as `qa` and `production`, receive changes only through the declared promotion path, in order: never a direct commit or push, a cherry-pick, a hotfix straight into them, or a skipped environment. A production defect that cannot wait, as [urgent fixes](../flow-deploy/references/urgent-fixes.md) defines, follows that reference. Any deviation is its own question naming the branch, the commits and why the path cannot serve; no delivery mode or unattended delegation grants it, and without an answer the fix stays on the promotion path and the deviation is reported as pending.
