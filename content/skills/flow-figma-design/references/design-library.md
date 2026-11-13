# Design library

Read at step 3 of [flow-figma-design](../SKILL.md), before step 4, whatever the app's `DESIGN.md` declares. This file is the only home of five rules: which records of the sibling screens to read, which Figma library the frames are drawn with, the inventory written before the first drawing, how each piece is drawn, and how the file and the library reach the app's `DESIGN.md`. [Design record](../../hive-ui-criteria/references/design-record.md) defines the two fields.

## Sibling records

Search for the file's location together with each sibling's frame node, since node identifiers repeat across files, a filtered read, in the project's change folders, archived ones included, and in the `examples.md` files of its work folders under `_support/workspace/`. Read the record of the example a sibling was approved from when one is found: its corrections are decisions too.

## Which library binds

Bind one Figma library at step 3: the one that holds the app's components and variables. A library of another design tool does not count. Take the first of these that gives one:

1. **The conversation.** The library the user named or shared in this conversation as the place of the app's components, by its name or by a link. For a link to a Figma file other than the one step 3 resolved, apply the first of these that fits: the user presents it as a library, a kit, a design system or the source of the components, so it is that library; the user links it as the screen to replicate or as a reference to follow, so the link does not make it one; the user states no purpose, so it is that library only when the connection lists it among the libraries the resolved file has enabled. When the user gave more than one library, take none of them and no other source yet: ask the user once, as an open question in the text, which holds the app's components, and bind the one they name.
2. **The design record.** The library the app's `DESIGN.md` declares in its `**Design library:**` field.
3. **Detection.** With neither, gather the libraries these three sources show:
   - the libraries whose components the sibling frames instance or whose variables they bind, read from the siblings' node trees: an instance whose main component, or a binding whose variable, lives outside the file comes from a library;
   - the library that the example record of a sibling names, in the records «Sibling records» finds;
   - the libraries the file has enabled, as the connection lists them; skip this source when the connection cannot list them.

   When they show exactly one library, bind it. When they show several, bind the one the sibling frames instance components from. When that leaves none or more than one, ask the user once, as an open question in the text, which of the libraries found holds the app's components, and bind the one they name. When the sources show no library, or the user names none, bind none.

Binding none changes only where new pieces come from: the frames still reuse what the siblings use, as «Draw with it» sets.

In step 4's message, name the bound library and which of the three origins gave it, or say that none was found. When the file does not have the bound library, adding it to the file is one of step 4's steps.

## Inventory

Before the first drawing, write the inventory in `examples.md` in the work's folder, where step 5 writes the record. When that file does not exist, create it with the tool, the file's location, the bound library and its origin. When it exists, as when the stage runs again for this work, keep every line it holds and add the inventory lines it lacks, with one exception: when the bound library is not the one the file names, replace that line and write every `searched:` line again. Step 5 adds the frames, the status lines and the other lines when the frames are drawn.

- `uses: <piece> · <component or variable> · <its library, or local> · <sibling node>`: one line per distinct component the sibling frames instance and per variable they bind, read from their node trees, whatever its origin: any library, or the file itself.
- `searched: <piece> · <terms> · found <component>`, or `· not found`: one line per piece the work's screens need, except a piece whose `uses:` line already names the bound library, such as a table header, a table cell, a card, a badge, a date picker, tabs or a chart. Search the bound library through the connection with two terms before writing `not found`: the piece's kind, and the name the app's code gives its component. With no bound library, write no `searched:` lines.

## Draw with it

Each piece and variable comes from the first of these that gives it, as an instance of the component or a binding to the variable:

1. the bound library, when the piece's `uses:` line names that library or its `searched:` line reads `found`;
2. the component or variable its `uses:` line names, from another library or from the file itself, whether or not a library is bound.

Define a piece or a variable in the file only when neither gives it, and record it as a `candidate: <piece> · <node>` line, which step 6's message lists. When the file still cannot use the bound library after step 4's steps, draw with your own pieces and say so in step 6's message.

## Record it

Ask in step 6's round, once per work, one decision that covers each of these that holds, whether to record it in the app's `DESIGN.md`:

- it has no `**Design file:**` field, or one that names a Figma file other than the one step 3 resolved;
- a library is bound, and it has no `**Design library:**` field, or one that names a Figma library other than the bound one;
- a file was resolved or a library bound, and the app has no `DESIGN.md`, or it has no `## Hive` section.

A field that names a file or a library of another design tool stays as it is: ask nothing about it. Show in the decision each value exactly as it will be written, beside the value it replaces when the field already holds one. In the same round, when that `DESIGN.md` has a `## Hive` section and no delivery answer is in force for its repository, ask the delivery question as the global guidance's «Delivery» sets for a mechanical change, worded to hold under either answer.

On a no, write `record: <the fields asked> · declined <date>` in `examples.md`, so the work does not ask again.

On a yes, when that `DESIGN.md` has a `## Hive` section, in the turn of step 7:

1. Write in that section only the fields the decision showed, `**Design file:**`, `**Design library:**` or both, in the format [design record](../../hive-ui-criteria/references/design-record.md) sets: add a field that is missing, replace the value of a field the decision showed as replaced, and change no other field or line. This is the one edit this stage makes in the repository.
2. Show the change in the message.
3. Deliver that edit at once, as [git-workflow](../../git-workflow/SKILL.md) sets under the delivery answer, as a mechanical change of its own, apart from any plan or build that follows: an edit left uncommitted is not carried to the branch a later stage works on. That delivery is the edit's own: it changes neither how the design work closes nor the state of its ticket, and after it do not run the close that git-workflow's last step sets. The stage goes on, and the work closes where step 8 or «A request for the design itself» sets; by then the edit is delivered, so a request whose deliverable is the design leaves no repository change to deliver. The user's yes, given with the values shown, is the validation an `interactive` delivery stops for. Under `hold`, leave the edit in the working tree and name it in the report.
4. Write `record: <the fields written> · written <date>` in `examples.md`.

On a yes, when the app has no `DESIGN.md`, or it has no `## Hive` section, write nothing in the repository: this stage does not create the record. Write `record: <the fields asked> · pending` in `examples.md`, and name it as pending in the report. A plan that follows writes the record with those fields, as [UI decisions](../../flow-plan/SKILL.md#ui-decisions) sets.

While [unattended delegation](../../unattended-delegation/SKILL.md) is active, ask none of the questions of this file: bind the library the conversation, the design record or detection gives without asking, bind none when the conversation gives more than one or detection leaves none or more than one, do not write `DESIGN.md`, and report as unresolved a file or library that `DESIGN.md` does not declare or declares differently.
