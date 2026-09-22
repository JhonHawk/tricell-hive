# Delivery decisions during planning

Before declaring an implementation plan ready, resolve the Git delivery mode for its named repositories and target branches. Reuse an explicit choice still applicable to this work; a repository default or remembered preference only informs the recommendation. Ask in the session language, mark one option Recommended and explain the actual effects. This is a planning question, not a prerequisite imposed on research or all editing tasks.

| Mode | Delivery contract to present |
| --- | --- |
| `interactive` | Work branch, authorized local commits and verification, then stop before publication for human validation. Present whether the subsequent authorized path ends at an open PR or includes merge after the stated checks and review. Human validation releases that agreed boundary; it does not add unlisted effects. Recommend for user-judged screens and flows. |
| `automatic` | Work branch, commits, push, PR to the named base, agreed checks/review and merge when its conditions pass. Name any deployment triggered by the destination before asking. Recommend only when this full delivery is appropriate; required human acceptance still stands. |
| `direct-base` | Verified commits and push directly to the named base, without a PR, only where project policy permits. Disclose triggered deployment. |
| `hold` | Implement and verify in the working tree; no commit or publication. Present the diff and leave delivery pending. |

Offer these modes with their concrete effects, marking an unavailable route with the reason rather than suggesting bypassing project policy. An explicit narrower instruction such as commit only or PR without merge takes precedence; do not force it into a broader mode. An unanswered question leaves the proposed delivery unapproved. A mode authorizes only its expressly selected delivery effects for otherwise authorized work, not implementation, production deployment or unspecified external effects by implication.

When dedicated implementation code review is relevant, resolve its mechanism and boundary in the same planning exchange: the current host's native review, an available external service, or human review. Discover availability without launching the review. State candidate, timing, relevant effects, and whether execution is authorized. Do not prescribe a provider absent from the project, treat an installed tool as consent, or confuse this with read-only plan feedback. Record pending execution consent explicitly if the user is only approving the design.

Record the selected mode, repositories/base, stop/resume conditions, review mechanism and remaining permissions in the existing plan. Avoid another questionnaire or a separate policy file per task. Changing the authorized target or effects requires resolving only that changed decision.
