---
name: hive-rule-card-unanswered
description: "Hive rule leaf: read when a question card comes back unanswered, closed, dismissed or only partly answered. Sets what to write, what to ask again and how to resume."
disable-model-invocation: true
---

# A card comes back unanswered

Read when a waiting question card returns without a full answer.

- When a waiting card comes back unanswered and it carried the round alone, follow the host's required response first; otherwise act by who left it unanswered. The answers the card did return stand, and you ask none of that round's remaining cards before the steps below.
  - **The card closed by itself:** the tool returns an empty answer or a timeout while the turn continues, and its result holds no rejection, dismissal or decline, as a Codex card does after 120 seconds. Write the round in the text then: reply in one message that writes every decision and finding of the round still open as [decision format](../hive-rule-decision-format/SKILL.md) sets, those of the cards not yet asked included, a decision with its title line, its `Situation:` line, its option lines and its `I recommend` line, a finding with its title and option lines, and end with a waiting line saying that "ok" applies every starred option and that a reply such as `D2 B` changes one. The answer goes in the chat.
  - **The user left it unanswered:** the tool returns a rejection, a dismissal or a decline, also when it reports one as an error and also on a host that returns a decline after its own timeout; or the host ends the turn while the card is open, with or without a result; or the card returns a request to talk, answers to only some of its questions, or free text that adds, removes or replaces a decision or an option of the round. Never write the round out in the text on your own. Read what the user said, in the card or in their next message when the host ended the turn, and take the first of these that fits:
    1. It answers decisions of the round by their ID, or is a plain acceptance such as "ok": apply those answers, a plain acceptance only to the decisions the user was shown, and ask the decisions left open again through the card.
    2. It says the cancel was a mistake, or asks for the questions again: ask the same round again through the card, with the same IDs, titles, situations and options.
    3. It says the work is going in the wrong direction, changes what the decisions rest on, or adds, removes or replaces a decision or an option: drop the open decisions it touches, rework them with what the user said, and ask the new round through the card with new IDs.
    4. It tells you to drop the questions: drop them and do not ask them again, except one a later step cannot proceed without, which you then ask again with that reason.
    5. It changes the subject, asks a question or asks for an investigation that leaves the decisions as they were: answer it, and end that reply with one line that names the decisions still unanswered and how to get them back, such as `D4 and D5 are still open; say "ask again" for the card, or tell me what to change`, and ask them again through the card when the user says so, or asks you to continue work that cannot go on without their answer.

    When the user has said nothing about the open decisions yet, as after a decline while the turn continues, a card answered in part or a request to talk that names no subject, reply with that one line alone and wait.

    A decision asked again keeps its ID, since it is the same decision and not a reuse of its number. Before a card that returns, write one line saying which decisions return; a reworked round is a new round and follows the first of these points.
  - Keep asking later rounds through the card: a card that came back unanswered is still a waiting tool, not one that returns at once or that fails.
