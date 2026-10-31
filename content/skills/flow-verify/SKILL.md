---
name: flow-verify
description: "Verify a built candidate before it is accepted: the full suite, the in-vivo walk of each acceptance criterion with its captures, the user's validation alongside the UX, code and requested security reviews, one round of fixes, and the integration. Use when a candidate exists and needs evidence; not for writing code, deploying, or checking a deployed environment."
---

# Verify a built candidate

Enter with a built candidate: commits, or the working tree under `hold`. Without one, say what is missing and return the work to `flow-build`.

1. Run the full local suite, serve the candidate with the project's production-style build or preview command where the stack has a build step, and exercise the changed behavior through its real path. Check each `AC<n>` by its observation. Open each changed screen as the user would: one that shows a sign-in page, an error or an empty list is a setup failure to fix, not evidence. During this walk, capture each changed screen as `flow-walkthrough` step 2 sets; those are the walkthrough's captures.
2. In the same step, start the reviews and show the result:
   - launch the code review with the mechanism the user chose; for a UI change, one `hive-review-ux` child, briefed with the captures so it measures and interacts instead of capturing again; and `hive-review-security` only when the user asks for it. `hive-verify-change` runs only when the user asks for it;
   - when the delivery stops for validation, show the result while they run, with `flow-walkthrough` for a UI change, saying that the reviews are still running and that approval is subject to their findings.
3. When the user's answer and the reviews are in, fix in one round: the user's corrections, the confirmed P0 and P1 findings, opening each finding's cited lines yourself, and every UI defect the change introduced; never recommend approving one as is. Recapture only the screens a fix changed, and run only the code review of the fixes and their affected tests; review again in full only when a fix changes a contract.
4. When that round changed something visible, show only those screens with `flow-walkthrough` and ask again; otherwise continue.
5. Before the push, when the candidate changed since `flow-build` ran its checks, rerun lint, typecheck, build and the tests of the packages those changes touch and of the packages that depend on them. Name the jobs that only CI can run, such as integration against a database or end-to-end suites, as CI-only in the report.
6. Integrate as the delivery says: push, open the pull request, wait for CI on a completion signal instead of polling, and merge. Then run `flow-close`.
