## ADDED Requirements

### Requirement: Change records travel in the work's pull request
When the work's delivery commits and `Specs` is in the work's repository, the archive of its change folder and the support material the user chose to commit, archive or promote SHALL be commits of the work's branch, made after verify's round of fixes and before its first push, and before the local CI evidence that push needs. Under `interactive` the support-material decision and the question to archive with unchecked tasks SHALL join verify's validation round; under `automatic` and `ready`, which do not stop before the push, the recommended action of each support item SHALL be applied and reported with how to revert it, and a folder with unchecked tasks SHALL stay unarchived for the close. Checkpoint lines written after the archive SHALL be committed with the next fix commit; those left after the last push SHALL NOT be committed, and the close SHALL discard them when they only touch the checkpoint of the archived folder. The archive SHALL mark the proposal completed with the branch name as its detail. The close SHALL archive on a new branch from the base only a change folder its work branch did not archive: when the work kept unchecked tasks and the user declined to archive, the delivery was `automatic` or `ready`, or the work ran under unattended delegation, when the work was integrated without its archive, or under `hold`; and SHALL commit there only support material decided after the push. When `Specs` is in another repository, the change folder SHALL be archived at the close and delivered as that repository's `Delivery` sets. Deleting support material SHALL stay at the close. A fix after the push that changes a requirement the archived folder merged SHALL edit the spec and that folder's delta together. When bringing in the base adds a later archived folder that changes a requirement this branch's archived folder also changes, the branch SHALL rename its folder to the current UTC time and rewrite its delta of that requirement to the resolved spec block, before the push or the merge.

#### Scenario: A finished change archives on its own branch
- **GIVEN** a change folder whose tasks are all checked and a delivery that commits
- **WHEN** verify reaches the push
- **THEN** the work's branch holds the merged deltas and the archived folder before the push, and the close opens no branch for the archive

#### Scenario: Support material decided before the push
- **GIVEN** a work with a curated session record the user chooses to commit
- **WHEN** verify reaches the push
- **THEN** the record is a commit of the work's branch, and the close asks only about deletions and material created after the push

#### Scenario: Unchecked tasks under a delivery that does not stop
- **GIVEN** a change folder with an unchecked task and the `ready` delivery
- **WHEN** verify reaches the push
- **THEN** the folder is not archived and nothing is asked, and the close archives it on a new branch from the base

#### Scenario: Checkpoint lines after the last push
- **GIVEN** an archived folder whose `tasks.md` gained checkpoint lines after the last push
- **WHEN** the close runs its cleanup
- **THEN** it discards those lines and opens no branch for them

#### Scenario: A work integrated without its archive
- **GIVEN** a change folder still under `changes/` after its pull request was merged
- **WHEN** the close runs
- **THEN** it archives the folder on a new branch from the base, delivered as the work's delivery mode sets

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
Before merging a change folder's deltas into the specs and moving it to the archive, the archive step, before the push of the work's branch or at the close for an exception, SHALL run the coherence check with `--archive` and fix each coherence line it prints; a delivery-unit box whose predicate reads `n/a: <reason>` SHALL count as resolved; when the check lists unchecked task boxes, the step SHALL show their count and ask whether to archive anyway before moving the folder, and under unattended delegation SHALL leave the folder unarchived and record it as pending.

#### Scenario: Unchecked boxes remain at archive
- **GIVEN** a change folder whose `tasks.md` holds unchecked boxes
- **WHEN** the archive step runs the check with `--archive`
- **THEN** it shows the number of unchecked boxes and asks before archiving, and the folder moves only on the user's answer

#### Scenario: An unattended close with unchecked boxes
- **GIVEN** an archive step running under unattended delegation and a folder with unchecked boxes
- **WHEN** the check lists them
- **THEN** the folder stays unarchived and is recorded as pending

#### Scenario: A delivery-unit box that does not apply
- **GIVEN** a `## Delivery units` box `- [ ] perf: n/a: no hot path changes`
- **WHEN** the archive step runs the check with `--archive`
- **THEN** that box is not counted as unchecked

### Requirement: Deltas merge deterministically at archive
The archive step SHALL preview the merge of a change folder's deltas with `apply_deltas.py` before applying it, and SHALL apply it with `--apply`, which merges every delta into its capability spec, marks `proposal.md` completed and moves the folder to the archive. The merge SHALL refuse and write nothing when a MODIFIED requirement is missing from the spec, an ADDED requirement is already present with other text, a delta lists a requirement twice, a delta modifies a new capability, a delta has no operations, a spec has no `## Requirements` section, or the archive target exists. A scenario that a MODIFIED requirement drops SHALL be listed in the preview and named in the work's final report, and SHALL NOT block the merge.

#### Scenario: A modified requirement the spec lacks
- **GIVEN** a delta that modifies a requirement the capability spec does not contain
- **WHEN** the archive step runs the merge script
- **THEN** it refuses with exit 1 and the specs and the change folder stay unchanged

#### Scenario: Preview before applying
- **GIVEN** a change folder with valid deltas
- **WHEN** the archive step runs the script without `--apply`
- **THEN** it prints each operation per capability and the archive target, and writes nothing

#### Scenario: A modified requirement drops a scenario
- **GIVEN** a delta whose MODIFIED requirement omits a scenario the spec has
- **WHEN** the archive step previews and applies the merge
- **THEN** the preview lists the dropped scenario, the merge proceeds, and the final report names it
