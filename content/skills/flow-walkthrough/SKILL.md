---
name: flow-walkthrough
description: Walk a built user interface for a person to validate it. Show the first rendered view of a new or rearranged layout during a build, capture the changed screens and relevant states, and deliver the walkthrough as a message and viewer manifest, with a standalone HTML fallback when the binary is absent. Use at the human validation stop of flow-verify, for the first rendered view during a build, and when the user asks for a walkthrough or a demo of a feature already built.
---

# Walk a user interface for human validation

1. Serve the candidate locally with realistic data and an account that can reach the changed screens.
2. Capture each changed screen at desktop width and at 390 and 320 pixels, in light and dark when the project has both themes. Save the images in `_support/workspace/<date>-<work>/images/`.
3. Send the user one message with the URL to open, the steps to try on each screen, what they should see, and the captures.
4. Ask whether they accept the result or want corrections. For each correction: apply it, recapture the affected screens, and ask again.
