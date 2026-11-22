# Documents

Read at [create-prd](../SKILL.md) step 2, to find where the document lives, and before the first write at its step 8.

## Where the document lives

Take the first that holds:

1. The document already exists: it stays where it is, in its tool or its file.
2. The session can write to a shared document tool, such as Notion, Google Docs or Claude Docs, through a connection or a CLI: the document lives there. When it can write to more than one, ask which as a decision, the one the project guidance prefers first.
3. Otherwise, the document is a Markdown file, as «Markdown mode» sets.

## Before the first write

- Read what the account may do in the tool: create a document, edit it, copy it and comment on it. Name before the write each operation it will not be able to do.
- Use the destination the user named. When they named none, ask where the document goes, as a decision whose recommended option is a private draft where the tool has one. The user's request for the document and that answer are the instruction for the writes this reference sets, the frozen copies included; write nothing else in the tool.
- Never write into, rename or move a document this work did not create, unless the user points at that document for it.

## Operations by tool

Perform create (Markdown content under the parent the user named, or a private draft), edit of the live version, copy, comment, lock and link through the tool's own connection, and read its current help for the exact call. Known limits: Notion's duplicate takes no title, so rename the copy; Claude Docs has no copy operation, so read or export the document and create another from that content; no connection locks a copy, and Google Docs edit, comment and lock only through the API (not exercised in a real session). For another tool, say in the report which operation it lacks.

## Freezing a version

Freeze a version when it reaches `approved by business` or `ready for development`:

1. Copy the document inside the same tool. Title the copy with the document's title, its version, its state, its date and the words "do not edit", in the document's language, such as `Returns · v2 · approved by business · 2026-11-16 · do not edit`.
2. Read the copy back and compare its last rule entry with the live document's. A tool that copies in the background can return before the copy is complete: read it again, up to three times. When it is still incomplete, create the copy from the content you read instead.
3. Write the copy's link in that version's entry of the live document's version history.
4. Never write to a frozen copy again, not even to fix a typing error: a correction is a new version of the live document. None of these tools lets the connection lock the copy; say in the report how the user can lock it or restrict it to read-only in that tool, and never change a document's sharing or permissions yourself.

The live document goes on as the next version.

## Markdown mode

- **Inside a repository**, the document is `_support/docs/requirements/<module>/prd.md` in the repository's support home, which the global guidance's «Support folder» sets, with `<module>` in English kebab-case, unless the project guidance names another place for requirement documents. That file is added to the repository, which is the user's choice: name its path and ask before the first write, unless the user already asked for the file there. **Outside a repository**, it is `<module>-prd.md` in the working folder.
- Write the file without creating a branch or committing. Report its path and that it is not committed. Committing it is a delivery: it waits for the user to ask, and then the global guidance's «Delivery» and [git-workflow](../../git-workflow/SKILL.md) govern it.
- **Freezing** copies the file beside itself as `prd-v<N>.md`, or `<module>-prd-v<N>.md` outside a repository, with the same title line as a frozen copy in a tool. That file is never edited again.
- **PDF.** Each time the document goes to its owner, and at each freeze, render its PDF beside it: `python3 -I <create-prd-dir>/scripts/render_pdf.py <file.md>` ([script](../scripts/render_pdf.py)). It prints with a browser already installed and installs nothing: exit 0 prints the PDF's path; exit 2 means it could not read the file; exit 3 or 4 wrote no PDF and left a printable page beside the file, whose location and printing the user is told as the script's message says. Never install a browser or a converter for this. The PDF is derived and needs no versioning.
