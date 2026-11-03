# Walkthrough viewer delta

## ADDED Requirements

### Requirement: Explicit presentation grouping independent of project authority

The viewer SHALL derive a read-only navigation group from an exact registered original source root's valid `.engram/config.json` `project_name`. Equal trimmed case-sensitive explicit names SHALL share a presentation group across different canonical project keys. Missing configuration SHALL retain canonical-project grouping. Invalid or unavailable configuration SHALL retain canonical-project grouping and expose a bounded recoverable diagnostic without file content. Configuration lookup SHALL NOT ascend, scan child repositories, invoke Engram or follow configuration symlinks. Reading and grouping SHALL NOT rewrite stored works or change identities, provenance, collaboration eligibility, memberships or receipts.

#### Scenario: Existing works in separate projects share a declared name
- WHEN two stored works have distinct project keys and their original registered roots declare the same valid name
- THEN navigation shows one named group with both works and the correct visible count
- AND their IDs, project keys, sources and join boundaries remain unchanged.

#### Scenario: Configuration changes after publication
- WHEN the explicit name changes, is removed or becomes invalid
- THEN the next completed index refresh changes only the presentation, using canonical fallback on absence or invalidity
- AND invalidity is reported without hiding the works or changing their stored bytes.

### Requirement: Compatible opt-in navigation projection

The authenticated index SHALL expose version-1 navigation metadata only when `navigation=grouped` is requested. Legacy requests SHALL preserve the existing response contract and schema/storage versions. The projection SHALL reference each work in that index exactly once, provide stable namespaced grouping keys and original-root origin metadata, and use at most one configuration observation per source root per index. The frontend SHALL accept a legacy index without projection and SHALL fall back to canonical groups with a diagnostic when projection metadata is invalid, without weakening session validation.

#### Scenario: Mixed client and server generations
- WHEN an old client requests the index from a new server
- THEN it receives the legacy shape.
- WHEN the new client receives a legacy index from an old server
- THEN it renders the original groups without losing works.

## MODIFIED Requirements

### Requirement: User-controlled walkthrough server

The manager SHALL provide `hive view start`, `stop`, `status`, `open`, `serve` and `register`. The server SHALL bind only to 127.0.0.1, run until the user stops it, and never be started by register, status or open. The manager SHALL retain existing Git-root and `_support/workspace` registration resolution and SHALL additionally accept the exact non-Git input directory as a root when it contains valid explicit project configuration. It SHALL NOT infer parent membership or enroll children. Repeated registration of that root SHALL be idempotent and preserve added_at. Invalid explicit configuration SHALL provide an actionable failure for this new registration path without writing. Existing registrations SHALL not be removed when configuration later becomes invalid or absent.

#### Scenario: Second start
- WHEN a live server is recorded and hive view start runs
- THEN it reports the running server without creating another process.

#### Scenario: Legacy workspace entry
- WHEN register names a folder inside <root>/_support/workspace and root is not Git
- THEN it records root and retains the existing project name fallback; explicit presentation grouping applies only when configured.

#### Scenario: Idempotent registration
- WHEN an already registered repository is registered again
- THEN the command exits successfully without rewriting the registry.

#### Scenario: Register an explicitly configured non-Git workspace directly
- WHEN register names its exact root with valid configuration
- THEN that root is registered and reported.
- WHEN the same root is registered again
- THEN no duplicate entry or added_at change occurs.

### Requirement: Hierarchical React viewer navigation

The shell SHALL use the requested navigation projection for a single flat list of works per group and show each explicitly grouped work's original source as secondary metadata. It SHALL preserve the existing components, title/ticket search, visible counts, archive and pin behavior, selection by work ID, keyboard/mobile navigation and accessible full labels. Regrouping SHALL preserve selection and keep its destination visible; expansion preferences SHALL retain a legacy fallback. Declared lifecycle SHALL remain separate from process liveness; local management actions SHALL occur once in the toolbar. There SHALL be no legacy physical-member selection or manifest-renderer branch in v3. Missing, loading, empty, error and unauthenticated states SHALL remain actionable and SHALL NOT create synthetic work or bypass authentication.

#### Scenario: Select and filter works from a shared group
- WHEN works from a workspace and a repository share an explicit group
- THEN each row identifies its origin and opens its own unchanged timeline without a nested repository group.
- WHEN a title/ticket filter or archive filter changes
- THEN the group's count matches its visible rows.

#### Scenario: Refresh navigation
- WHEN the user changes selection, expansion or filters and reloads
- THEN the same work and manual choices remain selected with its scroll context, applying the defined legacy preference fallback.
