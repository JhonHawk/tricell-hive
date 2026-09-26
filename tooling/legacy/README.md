# Pinned legacy recognition

`catalog.json` embeds installed-file bytes, SHA-256, source paths and manifest
mappings for Hive `16e7d3357a3c41530d5e31460c3024872566f3c7` only. Regenerate it
with `python3 tooling/legacy/generate.py /path/to/that/clean/checkout`.
The generator reads source; it does not run the legacy installer. Pi root
placeholders are rendered using each inspected Pi root before comparison.

`Scan` and `ScanExcluding` are read-only. The latter accepts only paths already
verified by the transaction manager as currently owned rebuild resources. An
unrecognized file at a catalog destination is a conflict, except unrelated
personal text at the shared global instruction file. Unknown files below a
legacy skill directory block its replacement. No recursive arbitrary deletion
or adoption is supported. All declared shared consumers must be selected.

Hook transforms remove exact event/command registrations and preserve unrelated
JSON values. They reject duplicate keys and malformed JSON rather than lose
content. Only the hooks value is rewritten; all bytes outside it are preserved. The
owned Pi async switch is removed only when unchanged, preserving other bytes. External package,
provider, shell and search preferences introduced by the old installer remain
user preferences; the old manifests do not establish their exclusive ownership.

The bundled inverse Pi patch is applied only to the exact post-patch SHA-256;
its result must equal the recorded pristine SHA-256. Package files are restored,
never removed. `testdata` contains the three exact patched pi-subagents 0.67.0
source fixtures, derived from pristine bytes verified against the pinned hashes;
the upstream MIT license is retained alongside them. No user settings are fixtures.

Revisions outside the catalog, unknown old manifest entries, modified owned
files and unexpected links require manual resolution. Directory removals are
explicit edits and may be executed only after their validated children; the
transaction manager owns backup, mutation, verification and recovery.
