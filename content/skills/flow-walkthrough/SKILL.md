---
name: flow-walkthrough
description: Walk a built user interface for a person to validate it. Show the first rendered view of a new or rearranged layout during a build, capture the changed screens and relevant states, and deliver the walkthrough as a message and viewer manifest, with a standalone HTML fallback when the binary is absent. Use at the human validation stop of flow-verify, for the first rendered view during a build, and when the user asks for a walkthrough or a demo of a feature already built.
---

# Walk a user interface for human validation

1. Serve the candidate locally with realistic data and an account that can reach the changed screens.
2. Capture each changed screen, with the `agent-browser` skill when it is installed, otherwise the host's browser tool, at desktop width and at 390 and 320 pixels, in light and dark when the project has both themes. Save the images in the work folder `_support/workspace/<date>-<work>/images/`.
3. Write `walkthrough.json` in the work folder as [the walkthrough manifest](references/walkthrough-manifest.md) sets, with `kind` `first_view` for the first rendered view during a build and `final` otherwise. Run `python3 <skill-dir>/scripts/render_walkthrough.py <work folder> --session-status in_progress`; when it fails, fix the manifest and run it again. When `command -v hive` finds the binary, run `hive view register --dir <work folder>`; otherwise the publisher writes `walkthrough.html` beside the manifest.
4. Send the user one message with the URL to open, the steps to try on each screen, what they should see, and the absolute path the publisher returned.
5. Ask whether they accept the result, with two options and corrections written in the free-text answer:
   - **Open the viewer**, only when `hive view register` succeeded: run `hive view open` and ask the same question again. When it prints that the viewer is not running, say the user can start it with `hive view start`, and ask again.
   - **Approve**.
6. For each correction: apply it, recapture the affected screens, update the manifest, run the publisher again, and ask again.
