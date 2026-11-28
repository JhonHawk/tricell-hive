# Adversarial review by another provider

Read when the user asks for a review by another provider's agent. Read it only when the user explicitly asks for that review of the candidate. The user's pick of an agent authorizes the spend of another provider's quota for this work item and no other. It runs in addition to the code review the user chose, never in its place, under the delivery mode in force.

- At most one run on the candidate, and one more after the round of fixes that followed it. The answer is never reused for another work item.
- In unattended work, run it only when an amount was authorized beforehand for this spend, and stay within it.
- Send the other provider only the scope's diff and the intent: no secret value, account identifier, ticket key or path of a secret location.

## Run

1. List each other provider's command-line agent that is installed (`command -v`) and signed in, such as Codex, Claude Code, Cursor Agent, Grok or Antigravity, and offer each as an option for the user to pick. When none is available, report the option as unavailable, say why, and run only the chosen review. Install and sign in to nothing on your own initiative.
2. Take the scope: the candidate's diff against its base, including uncommitted changes when nothing is committed, and state the intent in one sentence from the request, the commits and the plan (ask when it is unclear).
3. Start the chosen agent in its non-interactive, read-only mode, in a temporary copy of the tree so repros leave the candidate untouched, with the scope, the intent and this brief:
   > Break this change. Prove each finding with a run against the real code: a failing test, a repro script or command output that shows the wrong behavior. Report each finding with a title, the exact command, the output it printed and the output the intent requires. Leave out anything you did not run, including style, naming and risks that need a future change to matter. No findings is a valid result.
4. Rerun every reported command yourself. Keep a finding only when the output shows the failure and the failure breaks the intent; drop the rest and merge findings that share a root cause.
5. Report without applying a fix: the intent, each kept finding with its command, its output and how it breaks the intent, and the count dropped with the reason ("no verified findings" when none survive). Kept findings join the review's findings and follow [review findings](review-findings.md) and [review evidence](review-evidence.md).
