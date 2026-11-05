# Walkthrough viewer delta

## MODIFIED Requirements

### Requirement: Viewer-first walkthrough publication
Substantive publication SHALL use only the compatible v3 API and confirmed readback. A missing executable or a stopped/incompatible viewer SHALL preserve inputs and report publication unavailable without starting a daemon or producing any other format. Previously generated HTML pages and user-selected artifact units SHALL not be deleted or rewritten by the removal of the renderer. Research SHALL NOT publish to the viewer; a work's viewer session SHALL start at its first plan or build flow that was explicitly invoked (its command typed, its route accepted, or research continuing into build at the user's request), created with a summary of about one paragraph on what the work is about. The agent SHALL NOT open the viewer on its own: at each stop where the viewer holds something for the user to review (a plan at its approval stop, a UI walkthrough, including the build's first rendered view) it SHALL offer to open it inside what that stop already asks (one more question of the stop's round, or a reply line of its waiting line or next-step block), never as a question with its own decision ID or its own question-tool call, until the user accepts once in the conversation; after acceptance it SHALL open the work once and not ask again unless the user asks.

#### Scenario: Binary present, server incompatible
- WHEN publication finds a v1 server
- THEN it requests the coordinated v3 update and preserves inputs without HTML substitution.

#### Scenario: User declines
- WHEN the user declines to open the viewer at a review stop
- THEN the viewer stays closed and the offer is repeated at the next review stop.

#### Scenario: Research report
- WHEN a research report ends, with or without a route to plan or build
- THEN nothing is published to the viewer and no open-viewer offer is made
- AND the first plan or build flow of that work creates its viewer session with a one-paragraph summary.
