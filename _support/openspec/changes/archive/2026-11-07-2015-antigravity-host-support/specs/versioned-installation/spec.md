## ADDED Requirements

### Requirement: Antigravity user-scope adapter
The manager SHALL support Antigravity (the `agy` CLI and the Antigravity App, which share `~/.gemini/config`) as a user-scope host only, resolving its home to `~/.gemini/config` under the selected home and refusing `--scope project` with the unsupported-scope error. It SHALL expose each Hive skill as a relative symbolic link `<AntigravityHome>/skills/<skill>` to the shared store `~/.agents/skills/<skill>`, registered as a consumer of that shared resource, and SHALL leave every link, rule, agent and hook it does not own unchanged. Hive SHALL NOT write `~/.gemini/GEMINI.md`.

#### Scenario: Project scope requested
- **WHEN** a plan selects Antigravity with `--scope project`
- **THEN** it fails with the unsupported-scope error and proposes no writes.

#### Scenario: Foreign skill link present
- **WHEN** `<AntigravityHome>/skills/` already holds another tool's link and Hive installs, updates or removes Antigravity
- **THEN** that link keeps its target and Hive adds or removes only its own skill links.

## MODIFIED Requirements

### Requirement: Per-role model overrides
The manager SHALL let the user override the model and effort of one agent role on one registered user-scope host, through `hive models set` and `hive models reset` and through the Models view. Overrides SHALL be stored in Hive's state, SHALL NOT change release identifiers or release snapshots, and SHALL be applied by every later installation plan for that host, including updates and returns to a retained release. An override SHALL take precedence over the release's profile and the role's own effort. Applying or removing an override SHALL rewrite only that role's agent file on that host and SHALL keep the host's installation receipt consistent with the written files. Removing Hive from a host SHALL drop its overrides. An override SHALL be refused, without writing, when its host is not registered, its role is not in the host's installed release, its effort is set on a host that accepts none, is not one of `low`, `medium`, `high`, `xhigh`, `max`, and `ultra`, or is `ultra` on Pi, whose thinking levels end at `max`, its model is empty, longer than 200 characters, or contains a character outside letters, digits, and `._:/@+=[]-`, or applying it would change anything on the host other than agent files. The text commands SHALL accept any model the validation allows, whether or not the host lists it. Roles SHALL be grouped by the folder of their source (`content/agents/<group>/`); setting a group SHALL write an override with only the given parts to every role of the group, replacing each role's own override, and SHALL name before confirmation the roles that lose a part; resetting a group SHALL remove the overrides of its roles. On OpenCode, a new model without a chosen effort SHALL keep the profile's effort, including a variant embedded in the profile's model, without storing that effort in the override. The Models view SHALL offer, for the model, a filterable list made of `release default`, the effective model, the host's models, and a free-text entry. The host's models SHALL come from its listing command (`codex debug models`, `opencode models`, `pi --list-models`, `grok models`, `cursor-agent models`) or, for Claude Code, from the fixed aliases `fable`, `opus`, `sonnet`, and `haiku`, and for Antigravity, from the fixed agent model tiers `inherit`, `flash_lite`, `flash`, and `pro`; the command SHALL run only when an edit panel of that host is first opened in a session, as the binary found on `PATH`, without a shell, in a neutral working directory, detached from the terminal, with stdin from the null device, with a time limit after which its process group is terminated, and with a capped output whose excess is an error. With an explicit synthetic home no host command SHALL run. Internal host model caches SHALL NOT be read.

#### Scenario: Group set replaces own overrides
- **WHEN** a role has its own override and the user sets a model for its group
- **THEN** the confirmation names that role, and after applying every role of the group carries only the group's parts.

#### Scenario: Model list unavailable
- **WHEN** a host's listing command fails, times out, or exceeds the output cap
- **THEN** the list shows the reason and still offers the effective model and the free-text entry, and nothing is written.

#### Scenario: Override survives an update
- **WHEN** a role has a model override and a new release is installed with `hive update`
- **THEN** that role's agent file still carries the overridden model and effort, and the host's installation is reported as verified.

#### Scenario: Override removed
- **WHEN** the user resets a role's override
- **THEN** the role's agent file becomes byte for byte the file an installation without overrides writes.

#### Scenario: Unsupported effort
- **WHEN** the user sets an effort for a role on Grok, Cursor or Antigravity, or the effort `ultra` on Pi
- **THEN** the manager refuses it, names the reason, and changes nothing.
