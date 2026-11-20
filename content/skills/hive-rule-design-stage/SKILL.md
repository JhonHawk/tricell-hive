---
name: hive-rule-design-stage
description: "Hive rule leaf: read when routing a screen change toward a Figma design, or at the file and example steps of the stage. Sets the connection check, the offer, the file and one example per screen."
disable-model-invocation: true
---

# Design stage: offer, file and examples

Read when a route ends in a screen to design, and at steps 1, 3 and 5 of [create-design](../create-design/SKILL.md).

- **Connection.** Check yourself that the host has a connection that can write to Figma; never ask which tool to use.
- **When to offer the stage.** For a change the user asked for or a ticket they picked, when the connection exists and the project's Figma file is known, end with the next-step block that recommends `create-design`, with going straight to `flow-plan`, or to `flow-build` for a small understood change, as other replies, and name that stage in the scope line, when either holds:
  - what gets built includes a new screen, even when sibling screens fix its composition and the change would otherwise continue into `flow-build`;
  - an existing screen adds, removes or moves a region, the route would end at `flow-plan`, and neither the request nor an existing screen of the project fixes its composition.
- **When the file is known.** The user named it in the conversation, the app's `DESIGN.md` declares it in its `**Design file:**` field, or the project's persistent memory records it; a file of another design tool does not count. Name that file in the block, put its link in the command's argument, and say that the stage creates a page with frames in it.
- **Resolving the file in the stage.** Use the file the user or the offering block named in the conversation, otherwise the one the app's `DESIGN.md` declares, as [design record](../create-design/references/design-record.md) sets. With neither, ask the user for the Figma file once, as an open question in the text, and say that without one the work goes to the plan with text sketches; when they give none, end as step 1 of the stage sets for a missing connection. When the file comes from `DESIGN.md` and neither the block that offered the stage nor the user named it, name it, say that you will create a page with frames in it, and ask the user to confirm it once before the first write.
- **When not to mention the stage.** For a change that keeps the screen's layout, such as a button, a copy, a field or a column; on a host without the connection; when no Figma file is known; and while [unattended delegation](../unattended-delegation/SKILL.md) is active: the plan then settles the UI as [UI decisions](../hive-rule-ui-plan/SKILL.md) sets. Never ask a question that offers proposals for a screen: the next-step block is the offer.
- **How many examples.** Draw one example per screen, following the screen's siblings and the decisions recorded for the app. Draw two or three examples only for a screen the user names when asking for alternatives, and then name each screen's recommended example when asking the choice.
