# Secrets

## When to read

Read before retrieving a secret, documenting one by reference, writing a full value at the user's request, or reading output that may contain one, and when a secret value is exposed. The global guidance keeps the invariants, which hold in every mode, including unattended delegation. Reading this reference does not invoke a workflow.

## Handle a secret

- Never print a value or a fragment of one, including through your own redaction code. Keep a secret out of command arguments, URLs, subagent prompts, and messages to external services, where process lists, shell history, logs, or other parties retain it.
- When reading output that may contain secrets, such as logs, transcripts, or environment dumps, select only the fields you need rather than masking what you recognize; a pattern that misses one format leaks it. To find or read a secret in a file, never use a tool that returns matching lines or file contents, such as `grep` or `rg` without `-l` or `-c`, `cat`, or `head`, since the value reaches the transcript; list key names only, and pass the value only through stdin, a file, or a variable scoped to the one command that needs it.
- Prefer the established access path: existing local keys, the project's secret store, or an already-authenticated provider tool. Retrieve each authorized secret once per session into the session copy. Use 1Password only when the user explicitly asks; its installation or an unlocked vault grants nothing, and no other manager substitutes for that request. Never export a private key to work around missing access.
- Document secrets by reference: templates use placeholders such as `$DB_PASSWORD`; reference documents mask all but what identifies the resource, such as `AKIA****TBQX`, and name where the full value lives. Before writing a full value at the user's request, state the risk and offer these forms.

## When a value is exposed

Never repeat the value, and verify where it reached before judging the exposure; treat an unconfirmed scanner match as unconfirmed, not as a leak.

- Only this session's or a child's output or transcript, with no file, commit, or external service: stop using the route that exposed it and continue the work. Do not mention the exposure or rotation unless the user raises credentials.
- A durable or shared place, such as a tracked file, commit, pushed branch, pull request, tracker, or third-party log or service: tell the user at once which secret and where, stop only the action that would spread it further, and recommend revoking and rotating it; deleting the local copy does not contain it.
- Rotating real credentials needs the user's authorization.
