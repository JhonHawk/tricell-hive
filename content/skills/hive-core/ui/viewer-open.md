# Open the viewer

Read at a stop where the user reviews captures. It applies at each such stop, such as a UI walkthrough or a build's first rendered view.

Never open the viewer on your own, and never ask about it on its own. At each such stop, offer it inside what the stop already asks:

- **In the card.** Add one option, `<letter> - Open the viewer`, to the stop's own acceptance or validation question, as that decision's last option (never first, never with `(★)`, with no effect after its name and without a decision ID of its own) and, where [asking in text](../messages/questions.md) writes the round in the text, as its last option line. When the question already holds the most options the tool accepts, use the reply line below.
- **When the user picks it.** Run `hive view open` once for the work (`--session-id ID` when several sessions are open), then ask the same question again through the question tool with the same ID, title, situation and options except this one, keeping the round's other answers, after a one-line message such as "The viewer is open; D7 again". This changes four rules of [decision format](../messages/questions.md) and [asking a round](../messages/questions.md):
  - an ID is never reused;
  - the text before a round's first card holds the context line, the report and the waiting line;
  - nothing follows the round's last card but the reply to an unanswered card;
  - every option carries its effect.

  Ask the round's remaining cards first, unless one comes back unanswered, then open the viewer and ask the question again.
- **Without a card, or when the round was already asked,** add *open the viewer* as one more reply, in the waiting line or as a lettered reply of the next-step block; when the user uses it, run `hive view open` once while the stop keeps waiting.
- **After that,** once the user opened it, do not offer or reopen it in this conversation unless asked; when the user answers without opening it, offer it again at the next such stop.

## Open signed in

At a stop that [Sign in for the user](walkthrough.md#sign-in-for-the-user) sends here, the stop's question also gets `<letter> - Open signed in`, offered and handled as the viewer's option above, with these differences: it goes just before `Open the viewer`; when the user picks it, open the window as that section sets, then ask the question again keeping `Open the viewer` while it has not been used; and it is offered again at each such stop whose window is not open, while an open window is brought to the new stop's start instead.
