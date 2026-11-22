# Remove narrating comments

Read before a review, for narrating comments. The review is that of a candidate, over the files the candidate changed. A comment that says what the next line already says costs the reader a line and goes stale when the code changes.

Scope: the comments the candidate adds or edits. Comments the candidate leaves untouched are out of scope, so the diff stays reviewable.

1. List each comment in scope and sort it:
   - **Kill:** it narrates the code (`// loop over users`), restates a name or a type, marks a section the structure already shows, holds commented-out code, or records a past change (`// removed the old retry`). Git holds history.
   - **Keep:** it states a fact the code cannot: why a surprising choice was made, a protocol or vendor quirk with its source, a link to the issue or the spec that forces the shape, a warning about a side effect that is not visible at the call site, or a license header.
   - **Constraint:** it tells the next editor what not to do (`do not remove`, `do not reorder`, `talk to X before changing`).
2. Delete every Kill. When a comment exists because a name is unclear, rename the symbol instead of keeping the comment.
3. Treat a lint or type-checker suppression (`eslint-disable`, `@ts-ignore`, `# noqa`, `#[allow]`) as a defect, not a comment: read [suppression review](suppression-review.md) for it.
4. Test a Keep before you keep it. Trace the symbol it protects, or read its history with `git log -S`, to confirm the reason still holds. A Keep whose reason is gone or cannot be confirmed is deleted. When deleting is ambiguous between a lost fact and a stale note, keep it and report it open.
5. For each Constraint, propose the cheapest check that would enforce it in scope: a type, a runtime assertion, a test, or a lint rule. Ask the user before you add the check; in unattended work, add it only when the user authorized that kind of change beforehand. When the check is approved and written, delete the comment. Otherwise keep the comment, report the constraint as unenforced, and say what out-of-scope work would enforce it.
6. Change no application logic. When a deletion changes the diff by more than comment lines, revert that part.

**A project that bans comments.** When the project's lint script or CI runs a comment check, such as the ratchet of [comment audit](../../comment-audit/SKILL.md), a comment steps 4 and 5 would keep is moved instead: write it as an entry, with the paths it concerns, in the decision record the check's error message or the project's comment skill names, and delete the comment; step 6 does not revert that entry. With no record named, keep the comment and report it open. A directive a tool reads, a license header and the workaround marker stay. The workaround marker is one line: the language's line-comment token, then `workaround: ` and the ticket's key or URL, or, while no ticket exists, the record entry as `<path>#<anchor>`; inside JSX it is `{/* workaround: <ticket> */}`. It names the ticket once one exists. Confirm the marker's target exists, and report a missing one as a defect.

Report the number of comments deleted, the Keeps with their reason, the entries moved to the decision record with their paths, the checks proposed, and the constraints left unenforced. The corrections join the same candidate the review then reads.
