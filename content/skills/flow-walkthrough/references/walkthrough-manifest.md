# Walkthrough manifest

Read before writing `walkthrough.json` for a walkthrough or first view. The publisher validates every rule below and names the field that fails; fix the manifest and run it again rather than editing generated output.

## Where it goes

Write one `walkthrough.json` per work folder, beside its `images/` folder, as [artifact layout and retention](../../workspace-conventions/references/artifact-layout.md) places them. The first-view manifest and the final one are the same file: the final one replaces it. Then run the publisher as the skill's Deliver section sets and use the output target below.

## Publication target

The publisher resolves whether the `hive` executable is available on PATH once per invocation. With the binary available, it validates the manifest and referenced resources, atomically applies an explicit status update when requested, and returns `walkthrough.json`. It does not read page assets, build HTML or create or replace `walkthrough.html`; an existing page stays unchanged and is not the current output. Register the work folder through the skill's delivery step so the viewer renders the JSON.

Without the binary, it generates the standalone `walkthrough.html` fallback beside the JSON. Keep its existing preparation, concurrent-write checks and recovery: validation or preparation failure changes neither output, and a page publication failure restores the manifest when no other writer replaced it. Re-run from the authoritative JSON after interruption; this is not a cross-file crash transaction.

Server state, an older runtime, registration failure or failed confirmation never enables HTML when the executable is present. Do not add an override to bypass that condition. Missing-binary fallback does not require installing or starting the viewer. Keep previous evidence; generating a different output does not authorize deleting it. Retention of either primary output and its resources follows [artifact layout and retention](../../workspace-conventions/references/artifact-layout.md).

## Fields

Keys are English and exactly as shown; unknown keys fail. Text values are in the session language.

| Field | Required | Value |
| --- | --- | --- |
| `schema_version` | yes | `1` |
| `lang` | yes | `es` or `en`: the language of the page's fixed labels. Use the session language; any other language uses `en` labels. |
| `kind` | yes | `first_view` or `final` |
| `title` | yes | Ticket and one-line description, such as `gh-67 — sidebar navigation` |
| `ticket` | no | The change ID without date, such as `gh-67-flow-walkthrough-skill`; omit when there is no ticket or change |
| `session` | yes | `{"id": …, "host": …, "status": …}`; see below |
| `candidate` | yes | `{"branch": …, "commit": …, "base": …}`; write `"commit": "working tree"` before a commit exists |
| `delivery` | no | `[{"kind", "label", "url"}]`: verified PR or commit links; see below |
| `created_at` | yes | ISO 8601 time with offset |
| `result` | no | One sentence: the verdict and what the person must still do |
| `checks` | no | `[{"name", "status", "evidence"}]`, one per verified criterion plus the UX verdict and the tests, build and lint that ran; `status` is `pass`, `fail`, `warn`, `info` or `not_run` |
| `flows` | yes | At least one; see below |
| `action_required` | no | List of the decisions or checks the person still owes |
| `environment` | no | `[{"label", "value"}]`: URLs and account names; never a password, token or cookie |
| `provenance` | no | List of lines: when and from what the page was produced |

**`session.id`.** Use the host's session ID when the host documents one for the shell, such as `CLAUDE_CODE_SESSION_ID` or `PI_SESSION_ID`. Otherwise write `<host>-<UTC time of your first manifest in this session as YYYYMMDDThhmmssZ>` and reuse it for the rest of the session. `session.host` is the host's name, such as `claude-code` or `codex`.

**`session.status`.** Optional; exactly `in_progress` or `closed`. Active producers write `in_progress`, including when explicitly resuming the same work. Existing manifests without it remain valid and the viewer reports `unknown`; do not retrofit historical records or infer status from `kind`, age, Git classification or the host process. `closed` means this work passed through flow-close (completed, abandoned or paused), not that the person approved the interface or the host conversation ended. Before closing or deleting this work's review artifacts, follow [Close the current walkthrough](../../flow-close/SKILL.md#close-the-current-walkthrough).

The publisher's `--session-status in_progress|closed` explicitly updates only this manifest and publishes the output selected above. Without that option it validates without changing the manifest, including an absent status. `--confirm-viewer` requires `--session-status closed` and checks the existing viewer through `hive view status --dir <work folder>`. Exit `0` means publication succeeded and, when requested, the current closed copy was confirmed; `1` means validation or publication failed; `2` means closed was written but confirmation failed and the source evidence unit must be retained. Replacements and fallback recovery follow the publication target above. No raw CLI diagnostics are relayed.

**`delivery`.** Optional array; each object has exactly `kind`, `label` and `url`. `kind` is `pull_request` or `commit`; `label` is non-empty text (for example `tricell-hive #123` or `abc1234`), never HTML. `url` is an absolute HTTPS URL with an ASCII host (letters, digits, `.`, `-` or `_`, including punycode) or a valid IPv6 literal in brackets, valid percent escapes and an optional numeric port in `0..65535`. Scheme casing is immaterial. Credentials, control characters, literal whitespace, backslashes and empty ports are rejected. Unicode hosts are rejected without IDNA normalization; labels and URL paths may contain Unicode. There is no provider allowlist or remote status lookup; a linked PR can still be open and its label does not assert integration. Never put secrets in URLs or labels. Keep `candidate` unchanged; a local hash or `working tree` is not a delivery URL.

Add already verified deliveries when producing the walkthrough. At close, update only the current work's manifest with its known PR/commit URL before rendering, closing and confirming it; preserve existing links that still describe that work. Without a known delivery, omit the field. Do not infer URLs, rewrite historical manifests, or scan other work folders in the session.

The schema remains version `1`; manifests without `delivery` remain valid. Older strict readers reject the new optional field. Update the manager binary before deploying the content that writes it, and explicitly restart an already running viewer before expecting it to read the new contract. Grouping requires no migration or backfill. When produced, the standalone fallback embeds validated delivery metadata; link rendering uses the shared renderer's independent URL validation.

**Each flow** has:

- `id`: lowercase letters, digits and hyphens, at most 64, unique in the manifest.
- `title`: ticket and what the person can now do.
- `steps`: the steps from a named starting point; omit in `first_view`.
- `captures`: `[{"path", "caption", "alt"}]`; `path` under `images/`, `.png`, `.jpg`, `.jpeg` or `.webp`.
  `viewport` is optional: `desktop` (default) or `mobile`. A flow with both shows a Desktop/Mobile switch above its capture strip instead of two flows; use it for the same point seen at both sizes.
- `recording`: omit in `first_view`; otherwise one of
  - `{"status": "none"}` when the flow is not recorded;
  - `{"status": "recorded", "path": "images/<flow>.mp4", "duration_seconds": …, "candidate": …}`;
  - `{"status": "not_recorded", "reason": …}`;
  - `{"status": "stale", "path": …, "candidate": …}`.

Each capture needs `path` and `caption`; `alt` defaults to the caption. Limits: a manifest of 1 MiB, a capture of 20 MiB, a recording of 200 MiB, 50 flows, and 100 captures per flow.

**Paths** are relative to the work folder, use only letters, digits, `.`, `_` and `-` in each segment, never start with `/` or contain `..`, and name a file that exists and is not a symbolic link.

## Example

```json
{
  "schema_version": 1,
  "lang": "en",
  "kind": "final",
  "title": "gh-67 — sidebar navigation",
  "ticket": "gh-67-flow-walkthrough-skill",
  "session": {"id": "codex-20261018T223000Z", "host": "codex", "status": "in_progress"},
  "candidate": {"branch": "feat/gh-67-sidebar", "commit": "1a2b3c4d", "base": "development"},
  "delivery": [{"kind": "pull_request", "label": "example #123", "url": "https://example.com/repository/pull/123"}],
  "created_at": "2026-10-18T16:30:00-06:00",
  "result": "All checks pass; confirm the sidebar spacing matches the demo.",
  "checks": [
    {"name": "AC1 — the sidebar lists every section", "status": "pass", "evidence": "rendered at 1440 and 390 px"},
    {"name": "Unit tests", "status": "pass", "evidence": "pnpm test, 42 passed"}
  ],
  "flows": [
    {
      "id": "sidebar-navigation",
      "title": "gh-67 - open a section from the sidebar",
      "steps": ["Open /docs signed in as the test account.", "Select Skills > flow-plan.", "Expected: the page opens and the item shows as active."],
      "captures": [
        {"path": "images/sidebar-desktop.png", "caption": "Step 2 — active item", "alt": "Sidebar with flow-plan highlighted"}
      ],
      "recording": {"status": "none"}
    }
  ],
  "action_required": ["Confirm the spacing or describe the change."],
  "environment": [{"label": "App", "value": "http://localhost:4321"}, {"label": "Account", "value": "test.user@example.com"}],
  "provenance": ["Generated 2026-10-18 from the verification results."]
}
```
