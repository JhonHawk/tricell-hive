## MODIFIED Requirements

### Requirement: Store retention and deletion
The server SHALL purge copies older than the configured age, 14 days by default, except copies whose logical session is pinned. Explicit deletion through a session member's authenticated same-origin management route SHALL delete all currently stored copies with the same exact project key, declared host and session ID, including archived and source-excluded members. Deleting SHALL never modify repository files. Per-copy tombstones SHALL be persisted before removing copy directories and SHALL prevent unchanged deleted content from returning; new or modified source content MAY be ingested again. Every API read SHALL require the server token, and copied files SHALL be served only under a per-copy capability derived from it. Partial deletion SHALL be reported as a failure with remaining-copy counts and a retry path, never as complete success. Restart recovery SHALL reconstruct actual surviving copies and complete pending logical-record cleanup when no member remains.

#### Scenario: Read without token
- **WHEN** a request for the index, a session or a copied file lacks the token or matching capability
- **THEN** the server answers 403 and returns no data.

#### Scenario: Delete without token or same origin
- **WHEN** deletion lacks the server token or carries a foreign Origin or Host
- **THEN** the server answers 403 and all copies remain.

#### Scenario: Delete a logical session
- **WHEN** the user confirms deletion of a session containing several tickets and archived members
- **THEN** every current member is deleted from the viewer, its source is untouched, and new or modified source content can appear later.

#### Scenario: Pin a session before a continuation arrives
- **WHEN** a pinned session receives a new walkthrough and its members later exceed the normal retention age
- **THEN** all its members remain protected until the session is unpinned.

#### Scenario: Partial storage failure during deletion
- **WHEN** a member cannot be deleted after other members were removed
- **THEN** the operation reports failure, publishes the surviving members and allows retry without resurrecting successfully deleted unchanged content.

#### Scenario: Pending record cleanup without restart
- **WHEN** the last copy was removed but logical-record cleanup failed
- **THEN** the periodic scan retries cleanup and any new ingestion for that identity waits until cleanup succeeds, without inheriting stale name or preference values; the UI distinguishes pending cleanup from a completed delete and lets the user check recovery without presenting a read as a repair.

This scenario was approved under D10-B (final cleanup recovery and implementation).

### Requirement: Hierarchical React viewer navigation
The React shell SHALL list logical sessions directly under projects, with visible-session counts and compact session title, ticket summary, walkthrough count and date presentation, without a ticket navigation level or connector branches. Projects SHALL initially open only when they contain a visible explicitly in-progress walkthrough, with manual expansion taking precedence. Sidebar visibility, project expansion, archive filter and selection by physical member key SHALL survive refresh in sessionStorage within the same tab and origin. Searching a member ticket SHALL find its session and allow navigation to that walkthrough without hiding its siblings. The sidebar brand header SHALL show a route icon. The session content SHALL stack accessible walkthroughs chronologically with separators, individual headings and informational work status. Session management actions SHALL appear once in the fixed session toolbar and SHALL target the entire logical session, independently of the selected member or filter. Lifecycle SHALL remain separate from source provenance and SHALL NOT be editable in the viewer. Explicit work states SHALL be presented as in progress or closed; absent status SHALL be presented as unreported, never inferred as conversation termination. Archive filtering SHALL operate on logical sessions. The shell SHALL retain keyboard and mobile navigation and SHALL consume the Go API and shared walkthrough renderer. Simultaneous document controls, anchors, ARIA references, printing and cleanup SHALL remain scoped to their owning walkthrough.

#### Scenario: Common session controls
- **WHEN** a session contains walkthroughs from several tickets
- **THEN** a single toolbar controls its archive, pin and delete state and no walkthrough repeats those management controls.

#### Scenario: Different lifecycle and provenance
- **WHEN** a walkthrough is closed while its source classification is unknown
- **THEN** its work indicator remains closed and its source uncertainty remains independently visible, without asserting termination of the host conversation.

#### Scenario: Refresh after manual navigation
- **WHEN** the user changes project expansion, sidebar visibility, archive filter or walkthrough selection and refreshes the same tab
- **THEN** the shell restores those choices instead of replacing them with initial defaults.

#### Scenario: Two walkthroughs with identical flow indexes
- **WHEN** two walkthroughs with their own capture tabs and carousel are mounted together
- **THEN** IDs and ARIA references remain unique, each interaction controls its own document, and unmounting retires its listeners.

### Requirement: Reversible copy archiving
The server SHALL support authenticated, same-origin POST and DELETE requests to `/api/sessions/<key>/archive`, where the key locates a stored member and the operation archives or restores its entire logical session, including source-excluded members, with 204 on success. Archive state SHALL belong to the exact project/host/session identity, default to false for a new identity, and appear as effective `archived` state in the index for all normal and excluded members. It SHALL survive restarts, index reconstruction and source-content replacement, independently of pinning, ingestion time, TTL and manifest work status. New members SHALL inherit their logical session's archive state. Archiving SHALL NOT edit source files or exempt an unpinned session's copies from normal TTL deletion. Archived sessions SHALL be hidden by default; the existing accessible icon-only toggle beside Projects SHALL include or hide them. The shell SHALL offer brief feedback with Undo for a successful archive action, preserving that action's identity across selection changes, instead of a persistent archived-copy banner in the document area. Source warnings and actual errors SHALL remain visible and distinct from archive feedback.

#### Scenario: Archive and restore a pinned session
- **WHEN** the user archives and later restores a pinned session
- **THEN** its whole group disappears from and returns to the default list, remains pinned throughout, and its source manifests and work states remain unchanged.

#### Scenario: Archived unpinned session expires
- **WHEN** an archived, unpinned copy exceeds the configured retention age
- **THEN** normal TTL purging removes that copy.

#### Scenario: Continuation of an archived session
- **WHEN** another ticket's walkthrough arrives with the same exact project, host and session ID
- **THEN** it remains archived with that session instead of making the session reappear in the active list.

#### Scenario: Undo after selection changes
- **WHEN** archiving hides the selected session and the user activates Undo while another session is selected
- **THEN** the archived session is restored as a whole and the newly selected session is not modified.

## ADDED Requirements

### Requirement: Fixed session toolbar
The viewer SHALL keep the selected session's title and management actions visible outside the scrolling document canvas. Deliveries and chronological walkthroughs SHALL scroll within the canvas without covering the toolbar. Navigation to a walkthrough SHALL leave its heading visible. All actions SHALL remain keyboard-accessible and reachable at narrow widths without page-level horizontal overflow. Valid session names SHALL NOT be truncated or split inside an unbroken token: bounded local horizontal scroll regions in the heading and deletion confirmation SHALL make overflowing names fully reachable by keyboard with visible focus, while ordinary spaced names reflow by words. Print output SHALL omit management controls and retain scoped walkthrough content.

#### Scenario: Long session
- **WHEN** the user scrolls from the beginning to the end of a multi-ticket session
- **THEN** the toolbar remains at a stable position and the walkthrough content scrolls below it.

#### Scenario: Narrow viewport
- **WHEN** the viewport is 390 CSS pixels wide
- **THEN** the title and every management action remain reachable without horizontal page scrolling or obstructing the entire content viewport.


#### Scenario: Read an unbroken session name
- **WHEN** a valid name without spaces exceeds the width of the fixed heading or deletion confirmation
- **THEN** a keyboard user can focus its local horizontal region, read either end without splitting the token, and tab onward to the actions, without horizontal page overflow.

### Requirement: Local session name and preferences
The viewer SHALL allow the user to edit one local name per exact project/host/session identity, using a prefilled single-field dialog without asking again for known ticket, project or date values. The name SHALL be trimmed, contain at most 200 Unicode code points, reject control characters and be rendered as text. An empty name or reset action SHALL remove the override and recover the existing derived title. The name SHALL appear in the session toolbar and sidebar, participate in search alongside individual walkthrough titles and tickets, and survive reload, restart and index reconstruction. Naming SHALL NOT modify identity, source manifests or individual walkthrough titles. The server SHALL persist name, archive and pin preferences in a versioned logical-session record independent of the reconstructible index. Corrupt preferences SHALL produce an explicit error rather than silently allowing destructive retention under default flags.

#### Scenario: Save a session name
- **WHEN** the user saves a valid name and reloads or the viewer rebuilds its index
- **THEN** the same session retains that name and all member source titles remain unchanged.

#### Scenario: Cancel or failed save
- **WHEN** the user cancels editing or saving fails
- **THEN** cancellation writes nothing, and a failed save preserves the entered name with an inline retryable error.

#### Scenario: Reset the local name
- **WHEN** the user resets the session name
- **THEN** the stored override is removed and the derived title appears without renaming any walkthrough.

#### Scenario: Find the edited name
- **WHEN** the user renames a session, reloads and searches for the saved name
- **THEN** the complete session is found with all its accessible siblings, while individual ticket and title searches still work.

### Requirement: Explicit session mutation scope
Session archive, pin, title and deletion writes SHALL require `X-Hive-Action-Scope: session` in addition to existing authentication and same-origin checks. An older client's write without that scope marker SHALL return 409 with `reload_required` and SHALL NOT change any copy or logical preference. The updated client SHALL show a reload action and SHALL NOT automatically replay the rejected write. Reads and CLI fingerprint confirmation SHALL remain unchanged.

#### Scenario: Older tab after runtime refresh
- **WHEN** an already open copy-oriented frontend sends its former delete or preference request to the updated server
- **THEN** it receives a reload-required rejection without affecting any member of the session.

### Requirement: Import existing copy preferences
Before using logical-session preferences, the viewer SHALL import each legacy group with archive enabled only when every member was archived and pin enabled when any member was pinned. It SHALL preserve physical keys, resources, source identity and ingestion times, retire legacy per-copy preference fields as the authority, and resume an interrupted import without replacing an already persisted logical preference. A mixed archive group SHALL remain active. Import and recovery SHALL operate locally without touching repositories. Migration instructions SHALL describe coordinated backup and rollback with an older binary.

#### Scenario: Mixed historical archive
- **WHEN** an existing session contains archived and active copies
- **THEN** its imported state is active and all accessible members are visible together.

#### Scenario: Fully archived historical session
- **WHEN** all historical members were archived
- **THEN** the imported session remains archived.

#### Scenario: Interrupted import
- **WHEN** a logical record was written before removing all legacy fields and the process stops
- **THEN** reopening completes the import without recalculating or losing the saved logical preference.
