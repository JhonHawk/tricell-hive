# Remove narrating comments

Read before a review, for narrating comments. The review is that of a candidate, over the files the candidate changed. A comment that says what the next line already says costs the reader a line and goes stale when the code changes.

Scope: the comments the candidate adds or edits; untouched comments stay, so the diff stays reviewable.

1. Sort each comment in scope:
   - **Kill:** it narrates the code (`// loop over users`), restates a name or a type, restates a rule the code, its tests or its configuration in the repository already express, even when it states only part of the rule, marks a section the structure already shows, holds commented-out code, or records a past change (git holds history). A restated rule is a Kill only when the evidence names the line, test or setting that expresses it. A part the code does not express, a detail the code contradicts included, falls outside that Kill: cut the comment down to that part and sort it on its own, as this step sets.
   - **Keep:** it states a fact the code cannot: why a surprising choice was made, a protocol or vendor quirk with its source, a link to the issue or spec that forces the shape, a warning about a side effect invisible at the call site, or a license header.
   - **Constraint:** it tells the next editor what not to do (`do not remove`, `do not reorder`, `talk to X before changing`).
2. Delete every Kill; when a comment exists because a name is unclear, rename the symbol instead.
3. A lint or type-checker suppression (`eslint-disable`, `@ts-ignore`, `# noqa`, `#[allow]`) is a defect, not a comment: read [suppression review](suppression-review.md).
4. Test each Keep with the symbol it protects or its history (`git log -S`). Delete a Keep whose reason is gone or cannot be confirmed; when deleting is ambiguous between a lost fact and a stale note, keep it and report it open.
5. For each Constraint, propose the cheapest check that would enforce it in scope (a type, an assertion, a test or a lint rule) and ask the user before adding it; in unattended work, add it only when the user authorized that kind of change beforehand. Once the check is written, delete the comment. Otherwise keep the comment, report the constraint as unenforced and say what out-of-scope work would enforce it.
6. Change no application logic: revert any part of a deletion that changes more than comment lines.

**A project that bans comments.** When the project's lint script or CI runs a comment check, such as the ratchet of [comment audit](../../audit-comments/SKILL.md), a comment that steps 4 and 5 would keep is moved instead: write it as an entry, with the paths it concerns, in the decision record the check's error message or the project's comment skill names, and delete the comment (step 6 does not revert that entry); with no record named, keep the comment and report it open. A directive a tool reads, a license header and the workaround marker stay. The marker is one line: the language's line-comment token, then `workaround: ` and the ticket's key or URL, or, while no ticket exists, the record entry as `<path>#<anchor>`; inside JSX it is `{/* workaround: <ticket> */}`. Report a marker whose target does not exist as a defect.

Report the number of comments deleted, each restated-rule Kill with the line, test or setting it names, the Keeps with their reason, the entries moved with their paths, the checks proposed and the constraints left unenforced. The corrections join the candidate the review then reads.
