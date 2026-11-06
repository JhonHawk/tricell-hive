## MODIFIED Requirements

### Requirement: Resolved native identity and current work
The manager SHALL resolve the producer identity of a semantic command, when no explicit identity is given, from the first available source: a host environment variable verified against the host's official documentation, or against its published source code and an observed run of the installed version, then a process-to-session map written by the host's session hooks and found through the process ancestry. A host process SHALL be recognized by its command name, or by its script when it runs under an interpreter. When more than one verified variable is present, the manager SHALL use the variable of the host whose process is the nearest host ancestor, only when the host of every other variable is found further up the ancestry, and identity is ambiguous otherwise; when that host has no variable among those present, the inherited variables SHALL be ignored and that host's session map SHALL apply. A map entry SHALL be used only when it names exactly one live session for that ancestor and is newer than the host's last session change for that process; otherwise identity is unresolved. A subagent of Claude Code, Codex or Grok SHALL NOT be written to the map, and its events SHALL NOT be reported under another native session ID; a subagent's start MAY be reported as subagent activity of the parent session. A resolved identity is the host's actual native session ID; with no source, or an ambiguous one, the command SHALL be rejected as missing identity, and the host's conditional identity reference SHALL say how to pass it explicitly. The manager SHALL keep the current work of each (host, native session) so semantic commands may omit the work ID while exactly one work is current, and SHALL require the work ID otherwise. Opening a work SHALL register its project root when it is not registered.

#### Scenario: Claude Code without flags
- WHEN a semantic command runs in a Claude Code shell that exposes the session ID and omits identity flags
- THEN it publishes as that native session.

#### Scenario: Two sessions under one process
- WHEN the session map holds two live sessions for the nearest host ancestor
- THEN the command fails as missing identity and writes nothing.

#### Scenario: No identity source
- WHEN no flag, verified variable or session map applies
- THEN the command fails as missing identity and writes nothing.

#### Scenario: Grok subagent in the parent's process
- WHEN a Grok session starts a subagent whose own session submits a prompt and ends in the same process
- THEN the map keeps the parent as its only live session, a semantic command in the parent's shell publishes as the parent, and the viewer receives the subagent's start as activity of the parent.

#### Scenario: Codex thread in a shared daemon
- WHEN a semantic command runs in the shell of a Codex thread while another thread of the same app-server process submitted a later prompt
- THEN it publishes as the thread named by `CODEX_SESSION_ID`.

#### Scenario: Host started from another host's shell
- WHEN two verified variables are present and the nearest host ancestor is Pi running under `node`
- THEN the command publishes as the Pi session.

#### Scenario: Host without a variable under another host's shell
- WHEN a Grok session started from a Codex shell inherits `CODEX_SESSION_ID` and its map names one live session
- THEN the command publishes as that Grok session.
