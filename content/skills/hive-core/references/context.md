# Context at phase boundaries

## When to read

Read before ending the turn at one of the points below. It sets one line that the message may add; reading it does not invoke a flow or grant any effect.

## Where it applies

Only these turns, each of which ends to wait for the user's answer at a phase boundary:

- the build route question when [plan handoff](../../flow-plan/references/plan-handoff.md) returns a ready plan;
- the user's validation stop in [flow-verify](../../flow-verify/SKILL.md#validation-by-the-user);
- the close or pause question while a flow is still in progress, in [flow-close](../../flow-close/SKILL.md).

It applies nowhere else: not to design, delivery or review rounds, not to the route question of [flow-research](../../flow-research/SKILL.md), whose findings live only in the conversation, and not to any other turn inside a flow.

## Suggest compaction

**Condition:** at one of those turns, the context used exceeds about half of the model's window, as the host shows it or as you estimate from the session.

**Actions, in order:**

1. Confirm that the state needed to resume is on disk: the progress and next step of the change record, or, without a retained plan, what this message already summarizes. Write no persistent memory for it; the memory cadence of the global guidance still governs.
2. Add one line to the message, before the pause line, that suggests the host's compaction capability, naming its command as an example, such as `/compact`, or `/summarize` in Cursor CLI, and says that compacting now, while the prompt cache is warm, costs less than after a long wait. When a native question card that waits for the answer carries the question, the line also says how: dismiss the card, run the command, then send any message or the answer itself; you then ask again in text, as the dismissed-card rule of [questions](questions.md) sets.

The line is not a question: it has no options, blocks nothing and does not replace the pause line. It never splits the flow or the session: the work continues in the same session after the compaction, and the close question's own rule for recommending a new session is unchanged. Only the user can run the command, since no host gives the agent a tool to compact its own session. Add the line at most once per turn. After a compaction, the reread rule of the global guidance applies.
