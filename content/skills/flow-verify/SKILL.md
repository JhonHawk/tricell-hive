---
name: flow-verify
description: "Verify a built candidate before it is accepted: affected tests, runtime checks, the UI gate, the full suite and the dedicated review, at the level the delivery, review and UI answers and the request set. Use when a candidate exists and needs evidence; not for writing code, deploying, or checking a deployed environment."
---

# Verify a built candidate

Enter with a built candidate: commits, or the working tree under `hold`. Without one, say what is missing and return the work to `flow-build`.

1. Run the full local suite and exercise the changed behavior through its real path. Open each changed screen as the user would: one that shows a sign-in page, an error or an empty list is a setup failure to fix, not evidence. Serve the final handoff with the project's production-style build or preview command where the stack has a build step.
2. For a UI change, have one `hive-review-ux` child review the changed screens against the project's UI guide. Fix an introduced defect before asking the user to validate, or offer its fix as an option; never recommend approving it as is.
3. Run one code review with the mechanism the user chose. Open the cited lines of each finding yourself; fix the confirmed P0 and P1 ones. Review again only when a fix changes a contract.
4. Run a security review with `hive-review-security` only when the user asks for one.
5. When the delivery stops for validation, show the result with `flow-walkthrough` and wait. After the user accepts a correction, run only the code review of that fix and its affected tests.
6. Before each push, run the local checks: lint, typecheck and tests of each package the change touches and of the packages that depend on them. Rerun them after edits that change the candidate. Name the jobs that only CI can run, such as integration against a database or end-to-end suites, as CI-only in the report.
7. Integrate as the delivery says: push, open the pull request, wait for CI on a completion signal instead of polling, and merge. Then run `flow-close`.
