---
name: hive-ui-criteria
description: Detect a project's component library, its installed version and its customization layer, keep them in a verified DESIGN.md, and give design, build and review work observable UI criteria that never override the project's design system. Use when designing, building, refining or reviewing a web or product screen, with or without a component library, and when creating or refreshing the project's DESIGN.md; not for terminal interfaces.
---

# UI criteria

Criteria for screens, ranked below the project's own design system. Follow the steps in order; load a reference only when the table at step 5 names your situation.

1. **Find the design record.** Look for `DESIGN.md` at the root of the app that owns the UI, next to its package manifest; in a monorepo, the app whose screens the change touches. When it exists, read it first, then check it for staleness as [design record](references/design-record.md#staleness) sets.
2. **Detect when the record is missing or stale.** Follow [design record](references/design-record.md#detection): library, installed version, customization layer, library tools. Write or refresh the record only as [design record](references/design-record.md#writing) «Writing» sets; until then, use what you detected without saving it.
3. **Apply the precedence below** to every design choice.
4. **Pick a mode for each area of the app** as [criteria](references/criteria.md#screen-modes) sets; its screens share it. Default to operate in an app with a component library.
5. **Load what the task needs:**

   | Situation | Read |
   | --- | --- |
   | Create or change a screen, or review one | [criteria](references/criteria.md) |
   | Refine a built screen | [refinement](references/refinement.md) |
   | Create or refresh `DESIGN.md`, or detect the library | [design record](references/design-record.md) |

6. **Report** as the output contract sets.

## Precedence

This section is the only home of the rule. When two sources disagree, the earlier one wins:

1. an accessibility requirement the project declares (in `DESIGN.md` or the project guide) or the ticket states;
2. the project's recorded decisions (the `## Hive` «Decisions» list of the record);
3. the customization layer (theme or preset, CSS variables, wrapper components, copied component code);
4. the library's defaults;
5. this skill's defaults: the guide values in [criteria](references/criteria.md) that are not accessibility floors, such as the 44 px touch target.

Consequences:

- A library default that is below a skill default holds. A 36 px library button is not a deviation from the 44 px guide value.
- The accessibility floors in criteria sit outside levels 2 to 5: a recorded decision, the customization layer or a library default never excuses a floor failure in what the change builds or customizes.
- An element the change reuses without modifying it (an unmodified library default, or a component the project customized earlier) that fails a floor is not a defect of the change: list it among the defects already there, with its reason.
- With a declared accessibility requirement (level 1), such an element is still listed among the defects already there, but with its priority and marked as failing that requirement, so it reaches the user as a finding; the change's verdict still covers only what it introduced.
- Any change, whether a fix or a design choice, that alters a token, the size, radius, spacing or motion of a library component, or edits an installed component, returns to the user as a pending decision unless the brief or a recorded decision already names it. Do not apply it on your own.

## Output contract

Return:

- the decisions made and the source that justifies each, by precedence level;
- what you detected, with the file path of each item;
- the checks you actually ran, and the ones you could not run, with the reason;
- the defects already there, and each pending decision.

Never state a library API or token you did not read in that version's documentation, as [design record](references/design-record.md#detection) step 6 sets.
