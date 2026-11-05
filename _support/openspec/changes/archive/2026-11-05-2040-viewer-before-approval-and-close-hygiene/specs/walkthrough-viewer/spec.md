## MODIFIED Requirements
### Requirement: Viewer-first walkthrough publication
Substantive publication SHALL use only the compatible v3 API and confirmed readback. A missing executable or an incompatible viewer SHALL preserve inputs and report publication unavailable without starting a daemon or producing any other format; a stopped viewer SHALL do the same, except that semantic open, stage and evidence declarations SHALL be queued as «Queued semantic publication while the viewer is unavailable» sets. Previously generated HTML pages and user-selected artifact units SHALL not be deleted or rewritten by the removal of the renderer. Research SHALL NOT publish to the viewer; a work's viewer session SHALL start at its first plan or build flow that was explicitly invoked (its command typed, its route accepted, or research continuing into build at the user's request), created with a summary of about one paragraph on what the work is about. The agent SHALL NOT open the viewer on its own: at each stop where the viewer holds something for the user to review (a plan at its approval stop, a UI walkthrough, including the build's first rendered view) it SHALL offer to open it inside what that stop already asks, until the user accepts once in the conversation. When the stop asks through the host's question tool, the offer SHALL be one more option of the stop's own approval, acceptance or validation question, never a separate question, a question with its own decision ID or its own question-tool call; choosing it SHALL open the work once and ask that same question again, unchanged, with the same ID and without the open option, keeping the answers to the round's other questions. The open option SHALL NOT be the first option and SHALL NOT be starred. Otherwise, or when the question already holds the most options the tool accepts, the offer SHALL be a reply line of the stop's waiting line or next-step block, and using it SHALL open the work while the stop keeps waiting for its answer. After acceptance the agent SHALL NOT offer or open it again unless the user asks.

#### Scenario: Binary present, server incompatible
- **WHEN** publication finds a v1 server
- **THEN** it requests the coordinated v3 update and preserves inputs without HTML substitution

#### Scenario: User opens the viewer before deciding
- **WHEN** a validation stop asks its approval question through the host's question tool and the user picks the open-viewer option
- **THEN** the work opens in the viewer once
- **AND** the same approval question is asked again with the same ID and without the open-viewer option, and the round's other answers stand

#### Scenario: User declines
- **WHEN** the user answers a review stop's question without choosing the open-viewer option
- **THEN** the viewer stays closed and the offer is repeated at the next review stop

#### Scenario: Research report
- **WHEN** a research report ends, with or without a route to plan or build
- **THEN** nothing is published to the viewer and no open-viewer offer is made
- **AND** the first plan or build flow of that work creates its viewer session with a one-paragraph summary
