# Promotion path

Read before a push, merge or branch on an environment branch after the base. It applies to a change that reaches a branch after the base in the environment chain.

- Branches after the base in the environment chain, such as `qa` and `production`, receive changes only through the declared promotion path, in order: never a direct commit or push, a cherry-pick, a hotfix straight into them, or a skipped environment. A production defect that cannot wait follows [urgent fixes](../../flow-deploy/references/urgent-fixes.md). Any deviation is its own question naming the branch, the commits and why the path cannot serve; no delivery mode or unattended delegation grants it, and without an answer the fix stays on the promotion path and the deviation is reported as pending.
