# Incoming document

Read at [create-prd](../SKILL.md) step 1, when the user hands over a requirement document that this skill did not write, whatever its format. It replaces that skill's steps 3 and 4; its steps 5 to 9 then run on the result.

1. **Read it whole.** Read the document whole, with its attachments and each document it cites as far as a rule depends on it. When a part cannot be read, such as a scanned page without text, say which part and ask for its text; never guess what it says.
2. **Never edit the original.** The result is a new document, written where [documents](documents.md) sets. Never edit, comment on, rename or move the original, in a file or in a document tool. The new document names the original under its title: its title, its author, its date or version, and its link or path.
3. **Carry each rule over as its author wrote it.** Each behavior the original states becomes one rule line of [PRD format](prd-format.md), and its source is the section, heading or page of the original where it stands. Keep the author's words where they state one behavior, and split a sentence that states two into two lines. Never correct, complete or drop a rule of the author. Give each line its state:
   - `decided` when the original states the behavior plainly.
   - `to decide` when the sentence can be read two ways, when it contradicts another rule of the same document, or when the author marked it as pending; keep the author's words and add the question.

   A behavior the original does not state is never added as a rule, not even as an `assumption`: it becomes an open question. Put what the original says about the problem, the goals and the metrics in their sections, and leave an empty section empty, with an open question about it.
4. **Give each rule its verdict.** With the code at hand, compare each rule with it; without the code, every verdict is `not checked`.
5. **Start in `iterating`.** A restructured document is never approved on arrival, whatever its original says about itself: a label such as "ready for development" is a claim that the checks of the skill's step 5 test.
6. **Write the questionnaire for the author.** Write one question for each `to decide` rule, each empty section, and each check of the ready list that only the author can settle. Each question gives the rule's code with its place in the original, what is unclear in one sentence, the options when there are any, a recommended answer with its reason, and what each answer changes. Put first the questions whose answer changes the most rules, and merge the ones that ask the same thing. Write it in words the author can read, with no path or identifier of the repository. The questionnaire lists every question, since it travels to someone outside the session: it goes in the document's open questions, and whole in the report, so the user can send it. When the user is the author, ask the first three in this session instead, as the skill's step 3 sets, and leave the rest in the document.

Then continue at the skill's step 5.
