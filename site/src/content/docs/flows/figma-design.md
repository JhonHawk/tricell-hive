---
title: Figma design
description: An optional stage before the plan that draws a new screen in your Figma file, following your app's existing screens, has it reviewed, lets you approve or correct it, and records the approved example.
---

Figma design settles what a screen looks like before the work is planned. The agent draws one example of each screen as frames in your Figma file, following the screens your app already has, a reviewer agent checks it before you see it, you approve or correct it, and the approved frame becomes the reference that [Plan](/flows/plan/), [Build](/flows/build/) and [Verify](/flows/verify/) compare against. The stage is optional: nothing forces you through it.

## When to use

[Research](/flows/research/) recommends this stage in its next-step block when the host has a connection that can write to Figma and the work brings a new screen, even one whose layout your existing screens already set and even when the change would otherwise go straight to Build, or changes an existing screen by adding, removing or moving a region whose composition nothing fixes. That holds after assessing a ticket too. A change that keeps the screen's layout, such as a button, a copy, a field or a column, gets no recommendation, and neither does work that runs [unattended](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/unattended-delegation/SKILL.md). You can also type the command yourself, or ask for a design as the deliverable, such as replicating a screen in Figma.

Without a Figma connection there is no stage: the work goes from Research to [Plan](/flows/plan/), which draws its sketches in text.

Example requests:

```text wrap
Design the new billing summary screen in Figma before we plan it.
Replicate the checkout screen in our Figma file.
```

## What it asks you

Typing the command, or choosing the stage in the next-step block, is your instruction for the agent to write to the Figma file what this work creates. A write that spends money or paid quota still gets its own question.

- The Figma file, only when the project's `DESIGN.md` does not declare it and you have not named it. The agent asks once, in the text; without a file, the work goes on to Plan.
- Which Figma library holds your app's components, only when you have named or shared none, `DESIGN.md` declares none, and the agent finds several without your existing screens pointing to one of them. A library you share in the conversation, even as a plain link to its file, counts without being declared.
- Before the first drawing, a list of the steps only you can take in Figma, such as publishing a library the frames need, or a statement that there are none, and the library the agent will draw with and where it found it.
- One decision to approve or correct the completed set of screens. All applicable 440 px mobile frames are ready before the question. The agent names its recommended example for each screen with alternatives; you can name a different one with your corrections. Corrections name the affected screens, leave the others approved, and return only the affected screens and their mobile frames for approval.
- Once per work, in the same round, whether to record the Figma file and the library in the app's `DESIGN.md` when they are not declared there, with the values shown as they will be written. When you have not yet said in this session how changes to that repository are delivered, the same round asks it.
- Once per convention, in the same round, whether to record in `DESIGN.md` a convention your correction set for more screens than this one, such as the row actions of every list, so later screens follow it.
- In the same round, each business rule the frames decide that nothing settled yet, such as whether a rejection needs a comment or who may approve, as a decision of its own: approving the set does not approve its rules. What your request, the ticket, research, the decisions recorded in `DESIGN.md` or the existing screens and their code already settle is not asked, and a minor rule that changes neither scope, a contract, stored data, cost, reversibility nor who may act is listed as an assumption you can reply to change. The reviewer agent looks for these rules too, without seeing the drawer's list.

## What it produces

- Frames on a page of your Figma file named after the work, one per screen unless you asked for alternatives for a screen you named, drawn with the project's design system, real labels and data of realistic volume. The main agent draws them itself, by script from shared pieces on the tool's automatic layout with your app's shell, so spacing stays the same across screens. The agent draws with a Figma library when you named or shared one, your app's `DESIGN.md` declares one, or your existing screens already use one: before drawing it lists the components those screens use and searches the library for each piece the new screens need, so those pieces are instances of its components, and the record lists any piece the library lacks as a candidate to add to it. When the work has more than one screen, the agent draws one first, has it reviewed, fixes it and gives you its location without stopping, then draws the rest from its pieces. A reviewer agent compares the frames with your request, your existing screens, that library and your recorded decisions before you see them, and the message names anything it left open. The agent changes or deletes only what this work created, and returns any change to an existing style, variable, component or shared library to you as a decision.
- A rule card beside each frame whose screen has business rules that the frame's own text does not say: a separate annotation, outside the frame and not part of the interface, with one row per rule and where it comes from (the specification, your decision, an assumption or still to decide). The agent draws the cards after it reconciles the reviewer's first pass, and the reviewer's second pass covers them; when you answer a rule, its row is updated in the same turn. Details are in the [rule cards reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-figma-design/references/rule-cards.md).
- When you approve an example and the work names a ticket, the business rules the work adopted, which the ticket does not already state, are written to that ticket without asking: a section in its description, which leaves the rest of the text as it was, and a comment saying what was added or cancelled. A rule that a later decision replaces is marked cancelled with its reason, never deleted. The same write happens when the [build](/flows/build/) starts and at the [close](/flows/close/); the [business rules reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/references/business-rules.md) holds the record and the write.
- A record of the approved example, kept with the work's files under `_support/workspace/` and copied into the plan: the file's location, the node of each frame, a local image of it, and its status. The stage publishes nothing to the [viewer](/viewer/).
- When you say yes to recording the Figma file and the library, those two fields in the app's `DESIGN.md`, written in that turn, shown to you and delivered at once the way you chose. It is the only repository file the stage edits; when the app has no `DESIGN.md`, nothing is written and the entry stays pending. Details are in the [design library reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-figma-design/references/design-library.md).

## How it ends

When a rule would add a capability you did not ask for, change data or settings other screens use, or change a permission, the agent stops drawing the screens it shapes, asks the approval of the others, and ends with the command to start [Research](/flows/research/) for those rules; when the stage runs again, it picks up its record and draws only the stopped screens.

Once every screen is approved, the agent ends with the command to start [Plan](/flows/plan/), naming the approved example; it offers [Build](/flows/build/) instead only when Research had offered Build too, the approved screens add no region you had not asked for, and you said yes to no `DESIGN.md` entry, and no rule is still waiting for your answer. If you said yes to recording a convention in `DESIGN.md`, or the file and the library could not be written because the app has no `DESIGN.md`, the plan includes the task that writes it, delivered with the change. When the design itself was the deliverable, the work is not reported as done before you approve the example, and it then runs [Close](/flows/close/), which names each `DESIGN.md` entry still pending and closes it as an operational task when the stage edited no `DESIGN.md`.

## What it does not do

- It does not draw in text or offer examples without a Figma connection.
- It does not draw alternatives unless you ask for them, and then only for the screen you name.
- It does not edit the repository beyond those two `DESIGN.md` fields, or change existing variables, styles, components or pages.
- It does not plan or build the screen.

## Related flows

- [Research](/flows/research/) recommends the stage, and continues into it in the same turn when your request's deliverable is the design.
- [Plan](/flows/plan/) follows and takes the approved example as the screen's reference.
- [Build](/flows/build/) and [Verify](/flows/verify/) compare the built screen against that example.

## FAQ

### Do I need Figma to use Hive?
No. Without a connection that can write to Figma the stage does not run, and Plan draws text sketches instead.

### Can I skip the stage?
Yes. Reply that you want to go straight to Plan; the recommendation is one reply among others.

### Where is the exact rule?
Source: [content/skills/flow-figma-design/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-figma-design/SKILL.md)
