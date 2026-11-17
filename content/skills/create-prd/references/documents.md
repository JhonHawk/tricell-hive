# Documents

Read at [create-prd](../SKILL.md) step 2, to find where the document lives, and before the first write at its step 8. It names each tool's own operation; perform it through the tool's own connection or CLI, and read that tool's current help for the exact call, since its options change between versions.

## Where the document lives

Take the first that holds:

1. The document already exists: it stays where it is, in its tool or its file.
2. The session can write to a shared document tool, such as Notion, Google Docs or Claude Docs, through a connection or a CLI: the document lives there. When it can write to more than one, ask which as a decision, the one the project guidance prefers first.
3. Otherwise, the document is a Markdown file, as «Markdown mode» sets.

## Before the first write

- Read what the account may do in the tool: create a document, edit it, copy it and comment on it. Name before the write each operation it will not be able to do.
- Use the destination the user named. When they named none, ask where the document goes, as a decision whose recommended option is a private draft where the tool has one.
- Never write into, rename or move a document this work did not create, unless the user points at that document for it.

## Operations by tool

| Operation | Notion | Google Docs | Claude Docs |
| --- | --- | --- | --- |
| Create | A page with Markdown content under the parent the user named, or as a private draft | A file created from Markdown, which the tool converts to a document | A document from Markdown |
| Edit the live version | Update the page's content | Only through the API; where the connection cannot edit in place, each version is a new file | Update the document |
| Copy | Duplicate the page, then rename the copy, since the duplicate takes no title | Copy the file under a new name | No copy operation: read or export the document and create another from that content |
| Comment | On the page or on a selection | Only through the API | On the document |
| Lock | Not through the connection: by hand in the application | Only through the API, a read-only restriction | Not through the connection: view-only access by hand |
| Link | The link the create or duplicate call returns; when it returns none, read the page back | The identifier of the file created or copied | The link the create call returns |

The Google Docs column comes from its documentation and was not exercised in a real session: check the tool's current help before relying on a cell of it. For another document tool, use its own operations for these six, and say in the report which one it lacks.

## Freezing a version

Freeze a version when it reaches `approved by business` or `ready for development`:

1. Copy the document inside the same tool, as the table sets. Title the copy with the document's title, its version, its state, its date and the words "do not edit", in the document's language, such as `Returns · v2 · approved by business · 2026-11-16 · do not edit`.
2. Read the copy back and compare its last rule line with the live document's. A tool that copies in the background can return before the copy is complete: read it again, up to three times. When it is still incomplete, create the copy from the content you read instead.
3. Write the copy's link in that version's row of the live document's version history.
4. Never write to a frozen copy again, not even to fix a typing error: a correction is a new version of the live document. None of these tools lets the connection lock the copy; say in the report how the user can lock it or restrict it to read-only in that tool, and never change a document's sharing or permissions yourself.

The live document goes on as the next version.

## Markdown mode

- **Inside a repository**, the document is `_support/docs/requirements/<module>/prd.md` in the repository's support home, which the global guidance's «Support folder» sets, with `<module>` in English kebab-case. **Outside a repository**, it is `<module>-prd.md` in the working folder.
- The file is a support artifact, not a change to deliver: write it without creating a branch, committing or asking a delivery question. Report its path and that it is not committed, and commit it only when the user asks, as [git-workflow](../../git-workflow/SKILL.md) sets.
- **Freezing** copies the file beside itself as `prd-v<N>.md`, or `<module>-prd-v<N>.md` outside a repository, with the same title line as a frozen copy in a tool. That file is never edited again.
- **PDF.** Each time the document goes to its owner, and at each freeze, render its PDF beside it: `python3 -I <create-prd-dir>/scripts/render_pdf.py <file.md>` ([script](../scripts/render_pdf.py)). It prints with a browser already installed and installs nothing. When it exits 0, it prints the PDF's path. When it exits 3 or 4, no PDF was written and it left a printable page beside the file: tell the user where that page is and how to print it, as the script's message says. Never install a browser or a converter for this. The PDF is derived from the Markdown file and can be rendered again, so it needs no versioning.
