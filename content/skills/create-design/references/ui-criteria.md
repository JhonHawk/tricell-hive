# UI criteria

Criteria for screens, ranked below the project's own design system. Follow the steps in order.

1. **Find the design record.** Look for `DESIGN.md` at the root of the app that owns the UI, next to its package manifest; in a monorepo, the app whose screens the change touches. When it exists, read it first, then check it for staleness as [design record](design-record.md#staleness) sets.
2. **Detect when the record is missing or stale.** Follow [design record](design-record.md#detection): library, installed version, customization layer, library tools. Write or refresh the record only as [design record](design-record.md#writing) «Writing» sets; until then, use what you detected without saving it.
3. **Apply the precedence below** to every design choice.
4. **Pick a mode for each area of the app** as [criteria](criteria.md#screen-modes) sets; its screens share it. Default to operate in an app with a component library.
5. **Load what the task needs:**

   | Situation | Read |
   | --- | --- |
   | Create or change a screen, or review one | [criteria](criteria.md) |
   | Create or refresh `DESIGN.md`, or detect the library | [design record](design-record.md) |

6. **Report** as the output contract sets.

## Precedence

This section is the only home of the rule. When two sources disagree, the earlier one wins:

1. an accessibility requirement the project declares (in `DESIGN.md` or the project guide) or the ticket states;
2. the project's recorded decisions (the `## Hive` «Decisions» list of the record);
3. the customization layer (theme or preset, CSS variables, wrapper components, copied component code);
4. the library's defaults;
5. the defaults of these criteria: the guide values in [criteria](criteria.md) that are not accessibility floors, such as the 44 px touch target.

Consequences:

- A library default that is below a default of these criteria holds. A 36 px library button is not a deviation from the 44 px guide value.
- The accessibility floors in criteria sit outside levels 2 to 5: a recorded decision, the customization layer or a library default never excuses a floor failure in what the change builds or customizes, meaning its own elements and compositions and each value it sets or overrides. Using a library component as the library publishes it is not building it.
- A library default is a value inside a library component used as the library publishes it: a token value, a variant or a style that neither the project's customization layer ([design record](design-record.md#detection) step 3) nor the change sets or overrides. Show that from those files before recording it; a value you cannot show unchanged counts as customized. A library token the change applies to its own element, or pairs with a surface the library does not pair it with, is something the change builds. Without a declared accessibility requirement (level 1), a check that measures a library default itself, such as its contrast, color, size, spacing or radius, records `not applicable: library default` with the measured value, also when the value misses a WCAG threshold, when an automated audit flags it, and when the change is the first to use that component. That result is neither a finding nor a defect already there: it takes no priority and no suggested fix, it is not asked to the user, and it appears only among the checks the output contract lists. Layout integrity, states, keyboard order, a missing focus indicator and accessible names stay applicable.
- A component the project customized earlier, which the change reuses without modifying it and which fails a floor, is not a defect of the change: list it among the defects already there, with its reason.
- With a declared accessibility requirement (level 1), a library default or an earlier customization that fails it is listed among the defects already there, also when the change is the first to use it, with its priority and marked as failing that requirement, so it reaches the user as a finding; the change's verdict still covers only what it introduced.
- Any change, whether a fix or a design choice, that alters a token, the size, radius, spacing or motion of a library component, or edits an installed component, returns to the user as a pending decision unless the brief or a recorded decision already names it. Do not apply it on your own.

## Output contract

Return:

- the decisions made and the source that justifies each, by precedence level;
- what you detected, with the file path of each item;
- the checks you actually ran, and the ones you could not run, with the reason;
- the defects already there, and each pending decision.

Never state a library API or token you did not read in that version's documentation, as [design record](design-record.md#detection) step 6 sets.
