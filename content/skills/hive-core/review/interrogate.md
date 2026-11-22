# Adversarial review by another provider

Read when the user asks for a review by another provider's agent. Read it only when the user explicitly asks for that review of the candidate; never offer, suggest or prepare it otherwise. The user's pick of an agent authorizes the spend of another provider's quota for this work item and no other. It runs in addition to the code review the user chose, never in its place, and under the delivery mode in force like every other review of the candidate.

## Limits

- At most one run on the candidate, and one more after the round of fixes that followed it. The answer is never reused for another work item.
- In unattended work, run it only when an amount was authorized beforehand for this spend, and stay within it.
- Send the other provider only the scope's diff and the intent. Include no secret value, no account identifier, no ticket key and no path of a secret location.

## Availability

The reviewer is a command-line agent of a provider other than the one you run in, such as Codex, Claude Code, Cursor Agent, Grok or Antigravity. Once the user asks for this review, list each one that is installed (`command -v`) and signed in, checked with that agent's own status or `--help` output, and offer each as an option for the user to pick. Run the one chosen. When none is available, report the option as unavailable, say why, and run only the chosen review. Never install or sign in on your own initiative.

## Run

1. Take the scope: the candidate's diff against its base, including uncommitted changes when nothing is committed. State the intent in one sentence from the request, the commits and the plan; ask when it is unclear.
2. Start the agent in its non-interactive, read-only mode, in a temporary copy of the tree so repros leave the candidate untouched. Give it the scope, the intent and this brief:
   > Break this change. Prove each finding with a run against the real code: a failing test, a repro script or command output that shows the wrong behavior. Report each finding with a title, the exact command, the output it printed and the output the intent requires. Leave out anything you did not run, including style, naming and risks that need a future change to matter. No findings is a valid result.
3. Rerun every reported command yourself. Keep a finding only when the output shows the failure and the failure breaks the intent. Drop the rest, and merge findings that share a root cause.
4. Report; apply no fix here. Kept findings join the review's findings and follow [review findings](review-findings.md) and [review evidence](review-evidence.md).

Report the intent, each kept finding with its command, its output and how it breaks the intent, and the count dropped with the reason. Write "no verified findings" when none survive.
