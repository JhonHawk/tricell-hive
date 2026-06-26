# in-vivo test report — template

The **versioned record of an in-vivo verification run**: which acceptance criteria were
walked against the running app, what passed, what bugs the walk caught, and where the raw
evidence lived. It is the durable counterpart to the ephemeral screenshots/PDFs — those
expire in `_support/evidence/`; this report is what survives.

## Why this exists (the versioning split)

In-vivo runs produce two kinds of artifact, and they are versioned differently:

| Artifact | Nature | Home | Versioned? |
|---|---|---|---|
| Screenshots, PDFs, traces, HAR dumps | Raw, bulky, ephemeral | `<repo>/_support/evidence/` | **No** — gitignored, purged at close |
| **This report** (AC walk, results, bugs found) | Synthesized, durable, text-only | `<project>-specs/evidence/<epic-id>/` | **Yes** — committed in the specs repo |

The rule the report enforces: **a green in-vivo run that leaves no versioned trace is not
done.** Raw evidence answers "prove it happened"; this report answers "what was verified"
for anyone reading the epic six months later, when the screenshots are long gone.

**The specs repo never versions images.** Binaries bloat git history permanently (no
effective delta, never reclaimed once committed) and the report's text already records
what was verified. A screenshot worth showing the client goes to the delivery channel
(release notes, email) — never into the versioned repo. The report references raw evidence
by path/description, never embeds or commits it.

## When it is written

By the in-vivo gate of `/flow-build` (the `verify` gate) and by `/flow-deploy` post-deploy
verification — every run that walks Gherkin ACs against a live app. One report per
in-vivo run, named `in-vivo-<task-or-epic>-YYYY-MM-DD.md` under
`<project>-specs/evidence/<epic-id>/`. The raw screenshots/PDFs in `_support/evidence/` are
purged at close — never promoted into the versioned repo. A screenshot the client needs
goes to the delivery channel (release notes, email), not here.

## Template

```markdown
# in-vivo report — <epic-id> · <task-id(s)>

| | |
|---|---|
| Epic / tasks | <EPIC-ID> — <one-line scope> · <TASK-IDs> |
| Run date | YYYY-MM-DD |
| Environment | local (backend :<port> + <app> :<port>) / qa / <url> |
| Branch / merge | <branch> → <integration branch> via PR #<n> (CI <green/ref>) |
| Verdict | PASS / PASS WITH CONCERNS / FAIL |

## Acceptance criteria walked

Each row is a Gherkin AC exercised against the running app — not a unit assertion.

| Step (AC) | Expected | Result |
|---|---|---|
| <Given/When/Then in one line> | <expected outcome> | ✓ <observed> / ✗ <what failed> |

## Bugs caught in-vivo

Defects the automated suite did NOT catch — the in-vivo gate's reason to exist. Omit the
section only if there were none (and say so in one line).

| # | Symptom | Root cause | Fix (commit/PR) |
|---|---|---|---|
| 1 | <user-visible symptom> | <why> | <sha / PR #> |

## Evidence

- Raw (ephemeral, not versioned, purged at close): `<repo>/_support/evidence/<run>/` — <N screenshots, M PDFs>
- Delivery-worthy (if any) sent via <release notes / email> — described here, not committed

## Notes / unverified

<Anything left unverified and why (disabled flag, external-only dependency, missing
credential), per the in-vivo carve-out. "Nothing unverified" is a valid line.>
```

## Conventions

- **Walk ACs, not unit assertions.** Each row is a Gherkin criterion exercised through the
  real UI/integration path. If a row could be a unit test, it does not belong here.
- **The bugs section is the point.** An in-vivo report with an empty bugs section across a
  whole epic is a signal the walk was shallow, not proof the code is clean — note what was
  exercised so the emptiness is trustworthy.
- **Reference raw evidence by path; never paste bulky artifacts.** The report stays
  readable and small; the raw stays in `_support/` until purged.
- **One source of truth for the verdict.** The report's verdict is what the ledger and the
  tracker cite — they link to it, they do not restate the table.
