# Provider API

`tooling/providers` owns the small typed catalog for Hive's optional
capabilities (Engram, Context7, pi-subagents, agent-browser,
codex-question-card, CodeGraph, codex-wait-floor). It does not
manage Hive content, run anything, or persist onboarding journals;
`tooling/management` owns the parent journal and consent.

Engram, Context7, pi-subagents and CodeGraph are manual-only: `providers.Catalog()` lists,
for each capability, the hosts it is relevant to, the official documentation
to follow, and the reason. agent-browser and codex-question-card have
executable recipes, and codex-wait-floor has an entry with no step. agent-browser has an install step
(`brew install agent-browser` or `npm install -g agent-browser`) and a Chrome
step (`agent-browser install`). codex-question-card has one step,
`codex features enable default_mode_request_user_input`, which uses Codex's own
writer to show its question card in Default mode, and the manual step
`codex-question-card-manual` stands in when `codex` is not on `PATH`.
codex-wait-floor plans no external step: choosing it makes Hive write
`min_wait_timeout_ms` and `default_wait_timeout_ms` under
`[features.multi_agent_v2]` in Codex's `config.toml` itself, with a backup and
an undo through `hive plan remove`; the minimum applies to Codex's v2
multi-agent runtime (v1 unverified). Its manual step is
`codex-wait-floor-manual`.

```go
for _, p := range providers.Catalog() {
    // p.ID, p.Hosts, p.Source, p.Reason — everything the CLI's capability
    // list and summary render. Nothing here is executed.
}

step := providers.Step{ID: "context7-manual", Provider: providers.Context7,
    Status: providers.Manual, ManualReason: p.Reason}
payload, _ := json.Marshal(step)

// ValidateStep admits a Manual step for a catalog provider and, for
// agent-browser and codex-question-card only, the Planned steps whose argv matches the recipe bound
// to their step ID. It rejects any other command, a command on a manual step,
// and unknown fields.
if err := providers.ValidateStep(payload); err != nil { return err }
```

`Step` is the provider-side journal unit persisted inside the parent
onboarding journal's payload (`management.ExternalStep.Payload`). `Method` and
`Argv` are `omitempty` and exist only on `Planned` steps, so manual payloads
written earlier keep validating. The CLI's `providerRunner`
(`tooling/cli/agent_browser_runner.go`) is the `management.ExternalRunner`:
manual steps are confirmed back unchanged, and the executable steps run
without a shell, in their own process group, with an empty stdin, a 15-minute
limit and stored output capped at 64 KiB. Reconcile settles each agent-browser
step from its stored result and the machine's state, and a step with no stored
result (an interrupted run) is `failed`, never replayed.
