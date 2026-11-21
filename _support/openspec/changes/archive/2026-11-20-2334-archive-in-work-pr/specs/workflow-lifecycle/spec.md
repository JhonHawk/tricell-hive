## ADDED Requirements

### Requirement: Archives are delivered in a hygiene batch
When `Specs` is in the work's repository, the close of an integrated work SHALL NOT archive its change folder nor commit the support material the user chose to commit, archive or promote; it SHALL record them in persistent memory, one record per work under the topic `archive-pending/<id>`, where the id is the change-id, else the ticket key or a short slug, with the folder, the work's pull request and merge commit and each support action, SHALL discard with `git restore` the folder's uncommitted edits left after the merge, naming those that touch more than its checkpoint, and its report and footer SHALL name how many archives are pending and the hygiene reply that delivers them. Deleting support material SHALL stay at the close. When `Specs` is in another repository, the close SHALL keep archiving there as that repository's `Delivery` sets. A hygiene request SHALL treat as pending each change folder still under `<Specs>/changes/` on the base that has a record, or whose work the user confirms integrated, skipping one that an open hygiene pull request already archives; SHALL archive them on one branch from the base in the order each reached the base; SHALL apply the recorded support actions on that branch; and SHALL deliver it once under the delivery mode in force for that repository, asking only the delivery question when none is in force, leaving everything pending under `hold` and pushing to the base under `direct-base`. It SHALL remove from each record what its delivery carried once that delivery reaches the base, soft-delete a record left with nothing, and leave pending each folder whose delta changes a requirement that a folder left pending adds or changes. A folder without a record SHALL stay pending under unattended delegation.

#### Scenario: A close leaves its archive pending
- **GIVEN** a work integrated into the base with a change folder and a support record the user chose to commit
- **WHEN** the close runs
- **THEN** no branch or pull request is opened for them, the record `archive-pending/<change-id>` lists both, and the report names the pending count and the hygiene reply

#### Scenario: A hygiene request delivers two pending archives
- **GIVEN** two change folders still under `<Specs>/changes/` on the base, the second written while the first was pending
- **WHEN** the user asks for hygiene
- **THEN** one branch archives the first and then the second, one delivery carries them, and their records are emptied once the archive reaches the base

### Requirement: Plans read pending deltas
Before planning a change to a capability, the plan SHALL read the deltas of the change folders still under `<Specs>/changes/` that change that capability, since the capability spec does not hold them until a hygiene batch archives them, and SHALL write its own delta against the spec with those deltas applied.

#### Scenario: A plan on a capability with a pending delta
- **GIVEN** a pending change folder whose delta modifies a requirement of a capability
- **WHEN** a new plan changes the same requirement
- **THEN** its delta starts from the requirement as the pending delta leaves it

## MODIFIED Requirements

### Requirement: Archived change order within a day

Archiving a change folder SHALL move it to `<Specs>/changes/archive/YYYY-MM-DD-HHMM-<change-id>/`, with the UTC date and 24-hour time of the archive, so that two changes archived the same day keep their order. The spec-delta check SHALL require the delta of the latest folder of each requirement, ordering same-day folders by that time when both carry it, and SHALL accept either delta only when their order is unknown: on the same day when a folder has no time or both carry the same time, or when a folder without a time, which carries the local date of the older convention, and a timed one are at most a day apart. A requirement that one delta lists more than once SHALL be reported while that folder holds the latest delta of the requirement, and its blocks SHALL take part in no comparison.

#### Scenario: Two changes archived the same day

- **GIVEN** `2026-11-04-1403-a` and `2026-11-04-2003-b` both modify one requirement
- **WHEN** the spec matches the delta of `2026-11-04-1403-a`
- **THEN** the check reports that the spec differs from the delta of `2026-11-04-2003-b`

#### Scenario: Same-day folder without a time

- **GIVEN** `2026-11-04-a` and `2026-11-04-b`, without a time, both modify one requirement
- **WHEN** the spec matches either delta
- **THEN** the check reports nothing for that requirement

#### Scenario: Superseded duplicate

- **GIVEN** a folder lists a requirement twice and a later folder modifies it
- **WHEN** the spec matches the later delta
- **THEN** the check reports nothing for that requirement

### Requirement: Archive confirms unchecked tasks
Before merging a change folder's deltas into the specs and moving it to the archive, the hygiene batch SHALL run the coherence check with `--archive` and fix each coherence line it prints; a delivery-unit box whose predicate reads `n/a: <reason>` SHALL count as resolved; when the check lists unchecked task boxes, the batch SHALL show their count and ask whether to archive anyway before moving the folder, and a batch under unattended delegation SHALL leave the folder unarchived and pending.

#### Scenario: Unchecked boxes remain at archive
- **GIVEN** a pending change folder whose `tasks.md` holds unchecked boxes
- **WHEN** the hygiene batch runs the check with `--archive`
- **THEN** it shows the number of unchecked boxes and asks before archiving, and the folder moves only on the user's answer

#### Scenario: An unattended close with unchecked boxes
- **GIVEN** a hygiene batch running under unattended delegation and a folder with unchecked boxes
- **WHEN** the check lists them
- **THEN** the folder stays unarchived and pending

#### Scenario: A delivery-unit box that does not apply
- **GIVEN** a `## Delivery units` box `- [ ] perf: n/a: no hot path changes`
- **WHEN** the hygiene batch runs the check with `--archive`
- **THEN** that box is not counted as unchecked

### Requirement: Deltas merge deterministically at archive
The hygiene batch SHALL preview the merge of a change folder's deltas with `apply_deltas.py` before applying it, and SHALL apply it with `--apply`, which merges every delta into its capability spec, marks `proposal.md` completed and moves the folder to the archive. The merge SHALL refuse and write nothing when a MODIFIED requirement is missing from the spec, an ADDED requirement is already present with other text, a delta lists a requirement twice, a delta modifies a new capability, a delta has no operations, a spec has no `## Requirements` section, or the archive target exists. A scenario that a MODIFIED requirement drops SHALL be listed in the preview and named in the batch's report, and SHALL NOT block the merge.

#### Scenario: A modified requirement the spec lacks
- **GIVEN** a delta that modifies a requirement the capability spec does not contain
- **WHEN** the hygiene batch runs the merge script
- **THEN** it refuses with exit 1 and the specs and the change folder stay unchanged

#### Scenario: Preview before applying
- **GIVEN** a pending change folder with valid deltas
- **WHEN** the hygiene batch runs the script without `--apply`
- **THEN** it prints each operation per capability and the archive target, and writes nothing

#### Scenario: A modified requirement drops a scenario
- **GIVEN** a delta whose MODIFIED requirement omits a scenario the spec has
- **WHEN** the hygiene batch previews and applies the merge
- **THEN** the preview lists the dropped scenario, the merge proceeds, and the batch's report names it
