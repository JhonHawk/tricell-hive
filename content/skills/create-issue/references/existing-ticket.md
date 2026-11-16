# Existing ticket

Read at [create-issue](../SKILL.md) step 1, when the user asks to validate, review or complete a ticket that already exists. It replaces that skill's steps 2 to 8. A flow that only judges whether a ticket is ready, as [flow-research](../../flow-research/SKILL.md) does for a chosen ticket, reads the ready list of [ticket types](ticket-types.md) and not this reference.

1. **Read the ticket.** Read it whole with its comments, and each document it links as far as the ticket depends on it. With the code at hand, compare what the ticket states with the code: what already exists, what the code contradicts, and what data, relation or permission is missing. These are technical findings, and you state them on your own.
2. **Run the ready list.** Apply the ready list of [ticket types](ticket-types.md) and note, for each check that fails, what is missing.
3. **Check authorship and permission.** Before any write, compare the ticket's author with the account authenticated in the tracker, and read what that account may write there, as [trackers](trackers.md) sets.
4. **Pick the output.** Take the first case that holds:
   - **The account cannot write to the tracker.** Hand the report to the user in the message, written so they can send it: what is missing, the technical findings and the questions.
   - **The account is the author, or the user states the ticket is theirs.** Show the change, with each section as it is and as it would be, and edit the ticket after the user confirms. A ticket that stays blocked, because a business rule is undecided or it lacks a fact nobody in the session can supply, is marked as [trackers](trackers.md) sets.
   - **Otherwise the account is not the author.** Do not edit the ticket, whatever the permission allows. Write one comment with what is missing for each failed check, the technical findings, and each open question with a recommended answer. Show the comment and post it after the user confirms.

   A ticket that is too large, as the ready list's check 8 sets, gets its split proposed in that output; it is never split here. When the user owns it, offer the breakdown as the next step.
5. **Never decide for the owner.** An undecided business rule always goes to the ticket's owner as a question, in the comment or in the report. Never fill it in, and never rewrite a rule the owner wrote.
6. **Report.** Say whether the ticket is ready, what each failed check lacks, and what was posted or edited, with its link. Then end as the skill's step 9 sets.
