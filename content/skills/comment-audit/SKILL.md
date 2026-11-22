---
name: comment-audit
description: Audit the comments of a repository or an area and remove them in batches, filing a ticket for each comment that excuses a workaround and moving each decision a comment records to the project's decision record, then leave a lint that rejects new comments while the rest are cleaned up, estimate the lines the cleanup sheds, and generate the project's own skill to repeat it, in .agents/skills with a link in .claude/skills. Use when the user asks to audit, delete, ban or ratchet the comments of a codebase, or to set up a no-comments lint; not for the comments one change adds or edits, which the no-comments review leaf handles.
---

# Comment audit

Remove the comments a codebase does not need, keep what they recorded where it belongs, and stop new ones with a lint that only gets stricter. The [no comments](../hive-core/review/no-comments.md) leaf is the criterion for each comment's category: read it before step 4. The audit adds to the normal flows and replaces none of their steps: steps 1–4 are research and change no repository file; step 5 adds to [flow-plan](../flow-plan/SKILL.md); steps 6–9 add to [flow-build](../flow-build/SKILL.md); [flow-verify](../flow-verify/SKILL.md) and flow-close follow. When you route from one flow to the next, name the audit in the command's argument, such as `/flow-plan comment-audit <area>`, so this skill keeps applying.

## Scope

1. **Name the scope.** The scope is the whole codebase or the area the user names, such as a package or an app. List the languages in it and the generated paths to leave out, such as API clients, bundles and declaration files a tool writes. One scope is one work item; a large scope may land in several pull requests, since the lint in step 8 accepts a partial cleanup.
2. **Pick the inventory tool.** For JavaScript and TypeScript, copy the template in [ratchet](references/ratchet.md) to the project's CI scripts folder and use its `--inventory`. For another language, write the port that reference describes. The template needs `@babel/parser`: when the project lacks it, adding it is part of the plan, after the package check that [reuse before writing](../hive-core/plan/reuse-before-writing.md) rung 5 sets.
3. **Record a baseline** at a pinned commit, from the inventory's JSON lines:
   - the number of comments, of lines that hold only comments (`fullLine: true`), and of lines that mix code and a comment, counted apart;
   - the comments marked `directive`, which stay and are not counted as removable. Before step 4, read the comments the inventory did not mark and add to `ALLOWED` the project's own tool directives, such as a build tool's magic comments;
   - the estimate of lines the cleanup sheds: the `fullLine` comments' lines plus the blank lines their removal leaves, with the mixed lines reported apart, since removing them deletes no line.

## Ledger

4. **Mark every comment.** With at most 200 comments in scope, mark them in the main thread. Otherwise give each lane, such as one app or package, to one `hive-research` child, briefed as [delegation](../hive-core/delegation/delegation.md) sets, with this section, the no-comments leaf, and a ledger file in the host's temporary location as its output destination. The child reads each comment with the code around it and its history (`git log -S`), writes one row per comment as `category · file:line · evidence`, and returns its counts per category. Mechanical classes, such as banners like `// ======` or labels like `// Arrange`, are one row each with the exact pattern that found them, so a later batch can repeat it. The categories, with their no-comments counterpart:
   - **Tautology** (Kill): it repeats what the code, a name or a type already says, or marks a section the structure shows.
   - **Stale** (Kill): it describes code or behavior the current source or its history contradicts, or records a past change.
   - **Excuse** (new): it justifies a shortcut, such as a silent catch or a fallback, so the shortcut looks acceptable. Name the shortcut and its consequence.
   - **Decision record** (Keep): it states a non-obvious decision or a constraint from outside the code, confirmed as step 4 of the leaf sets. A Keep that fails that confirmation is Stale.
   - **Constraint** (Constraint): it tells the next editor what not to do; propose the check the leaf's step 5 sets.
   - **Directive**: a tool reads it; it stays. A lint or type-checker suppression among them goes to [suppression review](../hive-core/review/suppression-review.md).

   A row without evidence stays unmarked and the report names it.

## Plan

5. Plan as flow-plan sets, adding: the batches, each one a lane or a mechanical class small enough to review; the decision record each Decision record goes to, which is the project's existing record when it has one, otherwise `<support home>/docs/decisions/` as [layout](../project-layout/references/layout.md) sets, read and indexed as it says, proposed to the user as a decision; the tracker where the excuses go, as one parent ticket with a child per shortcut through [create-issue](../create-issue/SKILL.md) and its breakdown; and where the lint runs in the project's lint script and CI. Ask before any write to the tracker, as create-issue sets.

## Cutover

Steps 6–9 add to flow-build.

6. **Record before removing.** For each batch, in this order:
   - create the excuses' tickets as the approved set;
   - write each Decision record's entry, with the paths it concerns, and each Constraint's approved check;
   - remove the comments and the blank lines their removal leaves, changing nothing else;
   - over each excuse's shortcut, put the workaround marker the [no comments](../hive-core/review/no-comments.md) leaf defines, naming its ticket. An excuse with no shortcut under it, such as a to-do, gets no marker once its ticket exists.
7. **Prove only comments changed.** Run `node <script> --same-code --base <the commit before the batch>`, before the batch's own commit; a nonzero exit names the file and the first token that changed: revert that change. Run the project's lint, type check and the tests of the files the batch touched. When a script, not the host's patch tool, edited the files, say in the report that the host's per-turn diff view does not show them, and give `git diff --stat` for the batch.
8. **Wire the lint.** Add the script to the project's lint command and to its CI, with an explicit `--base`: the pull request's base, or the commit before a push, with enough history to reach it. [Ratchet](references/ratchet.md) sets the exact wiring and the checks it needs. The base in Git is the allowance: each file keeps the comments it had until a batch removes them, and a removed comment cannot come back.
9. **Generate the project's skill** as [project skill](references/project-skill.md) sets, so the next batch runs without this audit: its inventory command, categories, mechanical patterns, ledger location, decision record, tracker and checks, registered for every host the project uses.

## Report

10. Add to flow-close's final report, against the baseline:
    - comments before and after, by category, and the lines shed, with the mixed lines apart;
    - the tickets filed, with links, and the decision-record entries written, with paths;
    - the comments left, by lane, and the rows left unmarked for lack of evidence;
    - the lint's command and the CI job that runs it, and the generated skill's path;
    - the line estimate for removing every removable comment, from step 3.

    Never set or report a deletion target: a comment that records a fact the code cannot hold stays recorded somewhere.
