# tgrep workspace server (`tgw`)

**Date:** 2026-09-07
**Status:** installed on this machine, under evaluation — scoped to cross-repo sweeps and concurrent fan-outs
**Rule:** `global/rules/tools/code-search.md` (workspace row); consumer: `sdd-explore`
**Upstream:** [microsoft/tgrep](https://github.com/microsoft/tgrep) v1.0.4 — trigram-indexed, ripgrep-compatible grep with a client/server mode

## Why this scope and no wider

Measured on this machine (M4 Max, ripgrep 15.2, median of 7 runs, output compared line by line):

| Target | Files | Single search, rg → tgrep | 8 concurrent searches, rg → tgrep |
|---|---:|---|---|
| ark-monorepo | 1.8K | 45 ms → 4-7 ms | 301 ms → 15 ms |
| sample-project-monorepo | 1.5K | 38 ms → 4-7 ms | — |
| `~/Development/projects` (54 repos) | 18.7K | 500 ms → 6-12 ms (rare literal), 60-120 ms (hot literal) | 3.6 s → 19 ms |

- Inside one of our monorepos the gain is ~40 ms per search: not worth a daemon per repo.
- `rg` saturates every core per query, so N concurrent searches cost N× serial. A subagent fan-out over the workspace is where the index pays: seconds → milliseconds.
- Patterns under 3 characters get no benefit (no trigram). Hot patterns are bounded by match delivery, not by the search.
- Output parity: identical across 33 cells; the only divergence was a PDF rg scans as text and tgrep skips by extension.
- **Without a running server the on-disk index is stale** — a new file returns 0 hits where rg returns 1. That is why the wrapper never queries the index unless the server process is alive.

## What is installed (machine-local, not deployed by the hive)

| Piece | Path | Notes |
|---|---|---|
| Binary | `~/.local/bin/tgrep` | pre-built `aarch64-apple-darwin` release, checksum verified against `checksums.txt` |
| Wrapper | `~/.local/bin/tgw` | `cd ~/Development/projects && tgrep --index-path <idx> "$@"`; falls back to `rg` when the server is down; appends one line per call to the usage log |
| Index | `~/.cache/tgrep/workspace/` | 205 MB, rebuilt in ~2 s; regenerable, never inside a repo |
| Server | `~/Library/LaunchAgents/com.jmartinez.tgrep-workspace.plist` | `tgrep serve ~/Development/projects`, `KeepAlive`, `RunAtLoad`; log `~/.cache/tgrep/workspace-serve.log` |
| Usage log | `~/.cache/tgrep/tgw-usage.log` | `timestamp  server|rg-fallback  ms  exit  args` — the evidence for the expansion decision |
| Global gitignore | `~/.config/git/ignore` gets `.tgrep/` | guards against a stray `tgrep index .` inside a repo |

The wrapper honours `TGW_ROOT`, `TGW_INDEX`, `TGW_LOG` for ad-hoc overrides. Paths in the output are relative to the workspace root (`sample-workspace/ark/ark-monorepo/...`); a subpath argument narrows the search (`tgw -l Foo sample-workspace/ark`).

Operate the server:

```bash
launchctl print gui/$(id -u)/com.jmartinez.tgrep-workspace | grep -E 'state|pid'
launchctl kickstart -k gui/$(id -u)/com.jmartinez.tgrep-workspace   # restart
tgrep status ~/Development/projects --index-path ~/.cache/tgrep/workspace
```

Resource envelope observed: 60 MB RSS idle, up to ~460 MB after heavy querying (50K-entry file cache); the watcher registers one FSEvents stream on macOS; a full stale check on start takes ~0.4 s for 18.7K files and a reconciliation walk runs about once an hour.

## Not wired into the harness search tools

Claude Code's Grep tool has no documented way to swap its bundled ripgrep (verified against `code.claude.com/docs/en/env-vars` and `tools-reference`, 2026-09-07). On this machine the agents search through `rg` in Bash anyway, so `tgw` reaches them only as an explicit command named by the rule — never as an `rg` shim, which would carry the staleness risk to every search.

## Re-measured 2026-09-11 — selectivity decides, not scope

The usage log showed real sweeps at 3.7 s, 14.2 s and 38.2 s. Re-measured on this machine
(median of 5, `tgw` vs `rg`, output line counts compared):

| Pattern | Scope | tgw | rg | |
|---|---|---:|---:|---|
| rare literal (`verifyEdgeAccessToken`) | 1 repo | 0.035 s | 0.039 s | tie |
| rare literal | workspace | 0.033 s | 0.496 s | **tgw 15× better** |
| common literal (`organizationId`, 2,817 matches) | 1 repo | 0.073 s | 0.044 s | 1.7× worse |
| short term (`auth`, 1,901 matches) | 1 repo | 0.640 s | 0.040 s | **16× worse** |
| 5-term alternation | 1 repo | 1.212 s | 0.046 s | 26× worse |
| 5-term alternation | workspace | 1.272 s | 0.645 s | 2.0× worse |
| 3-term alternation, all rare | 1 repo | 0.030 s | 0.040 s | tgw better |

**The variable is the pattern's trigram selectivity, not the scope and not the alternation
count.** `auth` has FEWER matches than `organizationId` and costs 9× more, because it is a
substring of `author`/`authentication`/`oauth`/`authorize`: the index returns a huge candidate
set that must then be verified. An alternation of rare terms is fast; a single short term is
slow. This is the earlier "hot patterns are bounded by match delivery" bullet, stated in terms
of what an agent can decide from the pattern before running it.

Consequence for `code-search.md`: the workspace row now carries the selectivity condition, and
the vocabulary sweep is routed to `rg` at every scope — a broad sweep uses exactly the short,
common terms that lose here, cross-repo included.

Parity re-checked at workspace scope: the only divergence remains PDFs that `rg` scans as text
and tgrep skips by extension. Source-code parity is exact.

## Expansion criteria — revisit with the usage log

Expand (per-repo servers, or `tgw` as the default for single-repo sweeps) only when the log shows one of:

- `server` rows dominate and a meaningful share of calls carry ≥100 ms of rg-equivalent cost (cross-repo or fan-out), i.e. the workspace path is actually used, not just available;
- a monorepo grows past ~20K text files (rg single search ≥300 ms) — then a per-repo server earns its keep;
- `rg-fallback` rows are rare (the server stays up) — a flaky daemon is a reason to shrink, not expand.

Shrink or remove if the fallback rate is high, the log stays empty after a few weeks of agent sessions, or a parity divergence shows up in real use. Rebenchmark after any tgrep upgrade (`BENCHMARKS.md` upstream changes the margin per release).

## Uninstall

```bash
launchctl bootout gui/$(id -u)/com.jmartinez.tgrep-workspace
rm ~/Library/LaunchAgents/com.jmartinez.tgrep-workspace.plist
rm -rf ~/.cache/tgrep ~/.local/bin/tgw ~/.local/bin/tgrep
```

Then drop the workspace row from `code-search.md` and the `tgw` bullet from `sdd-explore.md`.
