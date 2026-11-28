# Draw the frames

Read at step 5 of [create-design](../SKILL.md), before the first drawing. This file is the only home of three rules: the width of each frame, what the stage may change in the Figma file, and how the main thread draws.

## Width

Draw each frame at the width its platform sets:

- **An app that runs only on mobile** is drawn once, at 440 px, whatever width the file's existing frames use, with no second frame. A native app runs only on mobile, and so does an app whose shell in the code and in the sibling frames is a phone shell with no desktop layout, even when it has a viewport meta tag or breakpoints.
- **A web app** is drawn at one desktop width, the one the screen's existing frames in that file use, otherwise 1440 px. When the screen is served on mobile, as [UI verification](../../hive-core/ui/ui-verification.md) reads it, it also gets 440 px frames in the [mobile phase](approval.md#mobile-phase), after the user approves its web frames: one for the screen and one for each of its states whose layout changes on mobile, such as a dialog that becomes a drawer or a table that becomes a list of cards. A state that changes only its text, values or messages gets none.

Read the platform from what step 3 gathers. With neither code nor sibling frames, the platform the request or the ticket names decides; with none, draw it as a web app served on mobile and say so in step 6's message. The main width of a screen is 440 px for an app that runs only on mobile and its desktop width for a web app.

## What the stage may change in the file

- Create frames and sections on the app's page, so the work's sections continue that page's own. The app's page is the first that applies: the page holding the most frames of the screen's siblings, the frames step 3 measured; the page holding the most of that app's other screens; a page this work creates, named after the app, never after the work or its ticket. A work whose flows belong to two apps, such as a mobile app and a web portal, draws each app's flow sections on that app's page, and its Assets and Moodboard sections on the page of its first flow.
- Create local components and variables when the frames need them.
- Change or delete only what this work created.
- Never change or delete an existing variable, style, component or page, nor anything of a shared library: return each such change to the user as a decision, as the precedence of [UI criteria](ui-criteria.md#precedence) sets. Adding this work's sections to the app's page does not change that page; renaming, reordering or deleting it, or editing a node on it this work did not create, does.

## How the frames are drawn

The main thread draws the frames, never a child, since a child starts without the conventions and corrections this session holds. When the connection runs scripts in the file, draw by script: each write defines once the pieces its frames use, such as a container on the tool's automatic layout, a button, a field, a table row and a badge, bound to the file's variables, with the app shell cloned from an existing screen, and composes the frames from them. Each piece comes from the bound library or the siblings' components, as [design library](design-library.md#draw-with-it) «Draw with it» sets. Correct a frame by changing a piece or its data and drawing that frame again, never by moving or resizing its nodes one by one: separate fixes drift from each other and from the siblings.
