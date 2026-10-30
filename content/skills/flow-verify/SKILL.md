---
name: flow-verify
description: "Verify a built candidate before it is accepted: affected tests, runtime checks, the UI gate, the full suite and the dedicated review, at the level the delivery, review and UI answers and the request set. Use when a candidate exists and needs evidence; not for writing code, deploying, or checking a deployed environment."
---

# Verify a built candidate

1. Run the full local suite and exercise the changed behavior through its real path.
2. For a UI change, have one `hive-review-ux` child review the changed screens against the project's UI guide. Fix an introduced defect before asking the user to validate, or offer its fix as an option; never recommend approving it as is.
3. Run one code review with the mechanism the user chose. Open the cited lines of each finding yourself; fix the confirmed P0 and P1 ones. Review again only when a fix changes a contract.
4. Run a security review with `hive-review-security` only when the user asks for one.
5. When the delivery stops for validation, show the result with `flow-walkthrough` and wait. After the user accepts a correction, run only the code review of that fix and its affected tests.
6. Integrate as the delivery says: push, open the pull request, wait for CI on a completion signal instead of polling, and merge. Then run `flow-close`.
