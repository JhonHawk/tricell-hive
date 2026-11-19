---
name: hive-rule-question-text
description: "Hive rule leaf: read when no permitted waiting question tool carries the question (no tool, headless run, child, unattended work) or the card has no option descriptions. Sets the text round."
disable-model-invocation: true
---

# Asking in text

Read when the host or mode has no permitted waiting question tool, or the tool gives its options no description field.

- Ask in text when the host requires it, or where no permitted waiting tool carries the question: a host or mode without one, a tool that returns an error other than the user's dismissal, a non-interactive or headless run, a delegated child and unattended delegation. Unless the host requires another form, write each decision and finding of the round there as [decision format](../hive-rule-decision-format/SKILL.md) sets, and end with that same waiting line. A different form still states the alternatives, their consequences and the recommended option with its reason, in the form the host permits; a change of channel does not remove that information. Where the waiting tool gives its options no description field, write the round in the text that way and then ask it through the card as [asking a round](../hive-rule-question-channel/SKILL.md) sets, with each option's effect in its label; when that card comes back unanswered, add one line saying that the card closed and naming the decisions still open, whose answer goes in the chat, such as `The card closed; reply here to D15–D19`, without writing the round again.
