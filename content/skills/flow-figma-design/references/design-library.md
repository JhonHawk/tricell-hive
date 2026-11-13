# Design library

Read at step 3 of [flow-figma-design](../SKILL.md), before step 4, whatever the app's `DESIGN.md` declares. This file is the only home of five rules: which records of the sibling screens to read, which Figma library the frames are drawn with, the inventory written before the first drawing, how each piece is drawn, and how the file and the library reach the app's `DESIGN.md`. [Design record](../../hive-ui-criteria/references/design-record.md) defines the two fields.

## Sibling records

Search for the file's location together with each sibling's frame node, since node identifiers repeat across files, a filtered read, in the project's change folders, archived ones included, and in the `examples.md` files of its work folders under `_support/workspace/`. Read the record of the example a sibling was approved from when one is found: its corrections are decisions too.

## Which library binds

Bind one Figma library at step 3. A library of another design tool does not count. Take the first of these that gives one:

1. **The conversation.** The library the user named or shared in this conversation, by its name or by a link. A link to a Figma file other than the one step 3 resolved, in which the connection finds components, is that library, whether or not the user called it one.
2. **The design record.** The library the app's `DESIGN.md` declares in its `**Design library:**` field.
3. **Detection.** With neither, gather the libraries these three sources show:
   - the libraries whose components the sibling frames instance or whose variables they bind, read from the siblings' node trees: an instance whose main component, or a binding whose variable, lives outside the file comes from a library;
   - the library that the example record of a sibling names, in the records «Sibling records» finds;
   - the libraries the file has enabled, as the connection lists them; skip this source when the connection cannot list them.

   When they show exactly one library, bind it. When they show several, bind the one the sibling frames instance. When that leaves none or more than one, ask the user once, as an open question in the text, which of the named libraries holds the app's components, and bind the one they name. When the sources show no library, or the user names none, bind none: the frames follow the siblings and step 5 defines the pieces they lack.

In step 4's message, name the bound library and which of the three origins gave it, or say that none was found. When the file does not have the bound library, adding it to the file is one of step 4's steps.

## Inventory

Before the first drawing, create `examples.md` in the work's folder, where step 5 writes the record, with the tool, the file's location, the bound library and its origin, and these lines. Step 5 adds the frames, the status lines and the other lines when the frames are drawn.

- `uses: <piece> · <library component or variable> · <sibling node>`: one line per distinct component the sibling frames instance and per variable they bind, read from their node trees.
- `searched: <piece> · <terms> · found <component>`, or `· not found`: one line per piece the work's screens need that no `uses:` line gives, such as a table header, a table cell, a card, a badge, a date picker, tabs or a chart. Search the bound library through the connection with two terms before writing `not found`: the piece's kind, and the name the app's code gives its component.

With no bound library, write only the `uses:` lines, for the components local to the file that the siblings use.

## Draw with it

Each piece and variable the bound library holds comes from it, as an instance of its component or a binding to its variable: every piece with a `uses:` line or a `searched:` line that reads `found`. Define a piece or a variable in the file only when its `searched:` line reads `not found`, or when no library is bound, and record it as a `candidate: <piece> · <node>` line, which step 6's message lists. When the file still cannot use the bound library after step 4's steps, draw with your own pieces and say so in step 6's message.

## Record it

When the app's `DESIGN.md` does not declare the file step 3 resolved, or does not declare the bound library, ask in step 6's round, once per work, one decision that covers what is missing: whether to record it there. Show in that decision each value exactly as it will be written. In the same round, when no delivery answer is in force for the repository that holds that `DESIGN.md`, ask the delivery question as the global guidance's «Delivery» sets for a mechanical change, worded to hold under either answer.

On a yes, in the turn of step 7:

1. Write the `**Design file:**` and `**Design library:**` fields in the `## Hive` section of that `DESIGN.md`, in the format [design record](../../hive-ui-criteria/references/design-record.md) sets, adding only the field that is missing and changing no other line. This is the one edit this stage makes in the repository.
2. Show the change in the message.
3. Deliver that edit at once, as [git-workflow](../../git-workflow/SKILL.md) sets under the delivery answer, as a mechanical change of its own, apart from any plan or build that follows: an edit left uncommitted is not carried to the branch a later stage works on. The user's yes, given with the values shown, is the validation an `interactive` delivery stops for. Under `hold`, leave the edit in the working tree and name it in the report.

Write `record: design file and library · written <date>` in `examples.md`, or `· declined <date>` on a no, so the work does not ask again.

When the app has no `DESIGN.md`, or it has no `## Hive` section, write nothing: this stage does not create the record. Write `record: design file and library · pending`, and name it as pending in the report. A plan that follows writes the record with both fields, as [UI decisions](../../flow-plan/SKILL.md#ui-decisions) sets.

While [unattended delegation](../../unattended-delegation/SKILL.md) is active, ask neither question of this file: bind the library the conversation, the design record or detection gives without asking, bind none when detection leaves none or more than one, do not write `DESIGN.md`, and report an undeclared file or library as unresolved.
