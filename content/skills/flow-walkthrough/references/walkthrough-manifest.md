# Walkthrough manifest

Read before writing `walkthrough.json` for a walkthrough or first-view page. The renderer validates every rule below and names the field that fails; fix the manifest and run it again rather than editing the generated page.

## Where it goes

Write one `walkthrough.json` per work folder, beside its `images/` folder, as [artifact layout and retention](../../workspace-conventions/references/artifact-layout.md) places them. The first-view manifest and the final one are the same file: the final one replaces it. Then run the renderer as the skill's Deliver section sets; it writes `walkthrough.html` beside the manifest.

## Fields

Keys are English and exactly as shown; unknown keys fail. Text values are in the session language.

| Field | Required | Value |
| --- | --- | --- |
| `schema_version` | yes | `1` |
| `lang` | yes | `es` or `en`: the language of the page's fixed labels. Use the session language; any other language uses `en` labels. |
| `kind` | yes | `first_view` or `final` |
| `title` | yes | Ticket and one-line description, such as `gh-67 — sidebar navigation` |
| `ticket` | no | The change ID without date, such as `gh-67-flow-walkthrough-skill`; omit when there is no ticket or change |
| `session` | yes | `{"id": …, "host": …}`; see below |
| `candidate` | yes | `{"branch": …, "commit": …, "base": …}`; write `"commit": "working tree"` before a commit exists |
| `created_at` | yes | ISO 8601 time with offset |
| `result` | no | One sentence: the verdict and what the person must still do |
| `checks` | no | `[{"name", "status", "evidence"}]`, one per verified criterion plus the UX verdict and the tests, build and lint that ran; `status` is `pass`, `fail`, `warn`, `info` or `not_run` |
| `flows` | yes | At least one; see below |
| `action_required` | no | List of the decisions or checks the person still owes |
| `environment` | no | `[{"label", "value"}]`: URLs and account names; never a password, token or cookie |
| `provenance` | no | List of lines: when and from what the page was produced |

**`session.id`.** Use the host's session ID when the host documents one for the shell, such as `CLAUDE_CODE_SESSION_ID` or `PI_SESSION_ID`. Otherwise write `<host>-<UTC time of your first manifest in this session as YYYYMMDDThhmmssZ>` and reuse it for the rest of the session. `session.host` is the host's name, such as `claude-code` or `codex`.

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
  "session": {"id": "codex-20261018T223000Z", "host": "codex"},
  "candidate": {"branch": "feat/gh-67-sidebar", "commit": "1a2b3c4d", "base": "development"},
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
