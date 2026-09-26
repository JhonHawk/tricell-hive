# Provider API

`tooling/providers` owns the small typed catalog for Hive's optional
capabilities (Engram, Context7, pi-subagents). It does not manage Hive
content, install or configure anything, or persist onboarding journals;
`tooling/management` owns the parent journal and consent.

No capability's host/platform combination has passed its own native
install/host-integration validation gate yet. Until one does, Hive never
executes an installer for it: `providers.Catalog()` lists, for each
capability, the hosts it is relevant to, the official documentation to
follow, and the reason it is manual-only in this build.

```go
for _, p := range providers.Catalog() {
    // p.ID, p.Hosts, p.Source, p.Reason — everything the CLI's capability
    // list and summary render (A3/A4). Nothing here is executed.
}

step := providers.Step{ID: "context7-manual", Provider: providers.Context7,
    Status: providers.Manual, ManualReason: p.Reason}
payload, _ := json.Marshal(step)

// ValidateStep admits only a Manual step for a known catalog provider.
if err := providers.ValidateStep(payload); err != nil { return err }
```

`Step` is the provider-side journal unit persisted inside the parent
onboarding journal's payload (`management.ExternalStep.Payload`). Its
`Status` is always `providers.Manual` in this build; the CLI's
`manualProviderRunner` (`tooling/cli/provider_adapter.go`) is the
`management.ExternalRunner` that plans and journals it — Execute performs no
process, and Reconcile (including after a recovery) always confirms the same
Manual status rather than replaying anything.

When a capability's own native gate is later validated for a given
host/platform, its recipe (detection, versioned install, structured
execution, reconciliation, and any safe revert) is added back here alongside
a new non-Manual `Status`; until then, `Reason` and `Source` are the only
things this package tells the operator.
