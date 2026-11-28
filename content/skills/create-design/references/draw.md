# Draw the frames

Read at step 5 of [create-design](../SKILL.md), before the first drawing. This file is the only home of three rules: the width of each frame, what the stage may change in the Figma file, and how the main thread draws.

## Width

Draw each frame at the width its platform sets:

- **An app that runs only on mobile** is drawn once, at 440 px, whatever width the file's existing frames use, with no second frame. A native app runs only on mobile, and so does an app whose shell in the code and in the sibling frames is a phone shell with no desktop layout, even when it has a viewport meta tag or breakpoints.
- **A web app** is drawn at one desktop width, the one the screen's existing frames in that file use, otherwise 1440 px. When the screen is served on mobile, as [UI verification](../../hive-core/ui/ui-verification.md) reads it, it also gets a 440 px frame.

Read the platform from what step 3 gathers. With neither code nor sibling frames, the platform the request or the ticket names decides; with none, draw it as a web app with its 440 px frame and say so in step 6's message. The main width of a screen is 440 px for an app that runs only on mobile and its desktop width for a web app.

## What the stage may change in the file

- Create frames and sections on the work's page: the page of the file that holds the screen's siblings, the frames step 3 measured, so the work's sections continue that page's own. Only when no page of the file holds that app's screens, create one named after the app, never after the work or its ticket.
- Create local components and variables when the frames need them.
- Change or delete only what this work created.
- Never change or delete an existing variable, style, component or page, nor anything of a shared library: return each such change to the user as a decision, as the precedence of [UI criteria](ui-criteria.md#precedence) sets.

## How the frames are drawn

The main thread draws the frames, never a child, since a child starts without the conventions and corrections this session holds. When the connection runs scripts in the file, draw by script: each write defines once the pieces its frames use, such as a container on the tool's automatic layout, a button, a field, a table row and a badge, bound to the file's variables, with the app shell cloned from an existing screen, and composes the frames from them. Each piece comes from the bound library or the siblings' components, as [design library](design-library.md#draw-with-it) «Draw with it» sets. Correct a frame by changing a piece or its data and drawing that frame again, never by moving or resizing its nodes one by one: separate fixes drift from each other and from the siblings.
