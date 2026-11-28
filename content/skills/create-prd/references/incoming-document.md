# Incoming document

Read at [create-prd](../SKILL.md) step 1, when the user hands over a requirement document that this skill did not write, whatever its format. It replaces that skill's steps 3 and 4; its steps 5 to 9 then run on the result.

1. **Read it whole,** with its attachments and each document it cites as far as a rule depends on it. When a part cannot be read, such as a scanned page without text, say which and ask for its text; never guess.
2. **Never edit the original.** The result is a new document, written where [documents](documents.md) sets. Never edit, comment on, rename or move the original, in a file or in a document tool. The new document names the original under its title: its title, its author, its date or version, and its link or path.
3. **Carry each rule over as its author wrote it.** Each behavior the original states becomes one rule entry of [PRD format](prd-format.md), and its source is the section, heading or page of the original where it stands. Keep the author's words where they state one behavior, and split a sentence that states two into two entries. Never correct, complete or drop a rule of the author. Give each entry its state:
   - `decided` when the original states the behavior plainly.
   - `to decide` when the sentence can be read two ways, when it contradicts another rule of the same document, or when the author marked it as pending; keep the author's words and add the question.

   Use the priority the original gives a rule; when it gives none, write the one you propose, marked as proposed in the entry's heading, and ask the author to confirm the priorities in one question.

   A sentence that states only a screen layout, a technical solution or a date is not carried over as a rule: name it in the report, as the skill's step 4 sets. A behavior the original does not state is never added as a rule, not even as an `assumption`: it becomes an open question.

   Put what the original says about the problem, the goals and the metrics in their sections, and leave an empty section empty, with an open question about it.
4. **Give each rule its verdict.** With the code at hand, compare each rule with it; without the code, every verdict is `not checked`.
5. **Start in `iterating`.** A restructured document is never approved on arrival, whatever its original says about itself: a label such as "ready for development" is a claim that the checks of the skill's step 5 test.
6. **Write the questionnaire for the author.** Write one question for each `to decide` rule, each empty section, and each check of the ready list that only the author can settle. Each question gives the rule's code with its place in the original, what is unclear in one sentence, the options, a recommended answer with its reason, and what each answer changes. Put first the questions whose answer changes the most rules and merge those that ask the same thing, in words the author can read, with no repository path or identifier. The questionnaire lists every question, since it travels to someone outside the session: it goes in the document's open questions and whole in the report. When the user is the author, ask them in this session instead, in rounds as the skill's step 3 sets, and leave in the document the ones still open.

Then continue at the skill's step 5. In this use, a check that a rule of the author fails is never fixed by rewording the rule: the rule becomes `to decide` and the failure becomes a question of the questionnaire.
