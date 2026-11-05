# Refinement

Guidance for improving a built screen. Nothing here is judged as a defect by the review. Work within the [precedence](../SKILL.md#precedence).

## Simplify

- Keep one primary action per region.
- Keep only the states the screen can actually reach.
- Remove decoration that carries no information: an element is removable when deleting it changes no label, state or action.

## Harden

Walk each of these and make sure the screen has a defined state: errors, missing permissions, empty data, the [extremes](criteria.md#extremes) and interruptions (navigation away, lost connection, a repeated submit).

## Polish

Walk the whole path of the main task, then classify every deviation from the [design record](design-record.md) as one of:

- **Missing token:** the value repeats and the record has no token for it. Propose adding the token to the record.
- **Stray case:** a token exists and the screen does not use it. Change the screen to the token.
- **Does not fit the concept:** the value contradicts the record's product or mode entry. Return it as a pending decision.
- **Local defect:** a one-off fault in this screen. Fix it.

A change to a token or to a library component still follows the pending-decision rule of the precedence.
