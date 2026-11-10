---
title: Figma design
description: An optional stage before the plan that draws proposals for a new screen in your Figma file, lets you pick or correct one, and records the approved example.
---

Figma design settles what a screen looks like before the work is planned. The agent draws two or three proposals as frames in your Figma file, you pick or correct one, and the approved frame becomes the reference that [Plan](/flows/plan/), [Build](/flows/build/) and [Verify](/flows/verify/) compare against. The stage is optional: nothing forces you through it.

## When to use

[Research](/flows/research/) recommends this stage in its next-step block when the work brings a new screen, or one that adds, removes or moves a region, nothing fixes its composition, and the host has a connection that can write to Figma. That holds after assessing a ticket too. A change that keeps the screen's layout, such as a button, a copy, a field or a column, gets no recommendation, and neither does work that runs [unattended](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/unattended-delegation/SKILL.md). You can also type the command yourself, or ask for a design as the deliverable, such as replicating a screen in Figma.

Without a Figma connection there is no stage: the work goes from Research to [Plan](/flows/plan/), which draws its sketches in text.

Example requests:

```text wrap
Design the new billing summary screen in Figma before we plan it.
Replicate the checkout screen in our Figma file.
```

## What it asks you

Typing the command, or choosing the stage in the next-step block, is your instruction for the agent to write to the Figma file what this work creates. A write that spends money or paid quota still gets its own question.

- The Figma file, only when the project's `DESIGN.md` does not declare it and you have not named it. The agent asks once, in the text; without a file, the work goes on to Plan.
- Before the first drawing, a list of the steps only you can take in Figma, such as publishing a library the frames need, or a statement that there are none.
- Which example to keep, as a decision with one option per proposal, or approve and correct when there is only one. When the screen is served on mobile, you also approve a 440 px frame of the chosen example.
- Once per work, in the same round, whether to record the Figma file in the app's `DESIGN.md` when it is not declared there.

## What it produces

- Frames on a page of your Figma file named after the work, drawn with the project's design system, real labels and data of realistic volume. The agent changes or deletes only what this work created, and returns any change to an existing style, variable, component or shared library to you as a decision.
- A record of the approved example, kept with the work's files under `_support/workspace/` and copied into the plan: the file's location, the node of each frame, a local image of it, and its status. The stage publishes nothing to the [viewer](/viewer/) and edits no repository file.

## How it ends

The agent ends with the command to start [Plan](/flows/plan/), naming the approved example. If you said yes to recording the file in `DESIGN.md`, the plan includes the task that writes it, delivered with the change. When the design itself was the deliverable, the work is not reported as done before you approve the example, and it then runs [Close](/flows/close/), which closes it as an operational task and names the `DESIGN.md` entry as pending.

## What it does not do

- It does not draw in text or offer proposals without a Figma connection.
- It does not edit the repository, including `DESIGN.md`, or change existing variables, styles, components or pages.
- It does not plan or build the screen.

## Related flows

- [Research](/flows/research/) recommends the stage and takes the request whose deliverable is the design.
- [Plan](/flows/plan/) follows and takes the approved example as the screen's reference.
- [Build](/flows/build/) and [Verify](/flows/verify/) compare the built screen against that example.

## FAQ

### Do I need Figma to use Hive?
No. Without a connection that can write to Figma the stage does not run, and Plan draws text sketches instead.

### Can I skip the stage?
Yes. Reply that you want to go straight to Plan; the recommendation is one reply among others.

### Where is the exact rule?
Source: [content/skills/flow-figma-design/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-figma-design/SKILL.md)
