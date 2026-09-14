# Code-search tool evaluations — the rubric and the rejected tools

**Living document.** Append a section per tool evaluated; never date-stamp the file.
Companion: `tgrep-workspace-server.md` is the operations doc for the one tool we KEPT.
Routing rule these feed: `global/rules/tools/code-search.md`.

This exists because zvec-grep was evaluated twice — 2026-09-02 and 2026-09-11 — and the
second evaluation did not know about the first. A verdict that lives only in memory gets
re-litigated.

## The rubric

Five axes. The first three are the same for every tool; the fourth is the one that has to be
adapted per tool class, and adapting it is where the judgment goes.

1. **Scale and cost** — files/symbols indexed, index time, RSS, output size. Median of 5 runs,
   never a single probe.
2. **Multiple targets, one of them a control** — 4-6 repos spanning the stacks in play, plus
   one deliberately outside the dominant stack. `optional reference project` (Go) is the standing control in a
   TypeScript-dominant portfolio: it separates "this tool is bad" from "this tool is bad for us".
3. **Resource envelope and residue** — what it installs, where, and what it leaves behind.
   Verify the uninstall, do not assume it.
4. **Correctness — the axis that changes by tool class:**
   - *Literal search* (tgrep): exact output parity against `rg`, cell by cell.
   - *Structural/graph* (ripwire): resolved relations vs fabricated ones — for every claimed
     edge, does an import or package relationship support it?
   - *Semantic retrieval* (zvec-grep): rank of the correct answer for questions whose answer
     was established INDEPENDENTLY first, with `rg` + `Read`, before scoring.
5. **Verdict with revisit criteria** — what specifically would change the answer. A verdict
   with no revisit condition is an opinion.

### The baseline is what we actually do

Every vendor benchmark compares against a straw man — "reading whole files", "grepping and
opening files repeatedly". Our baseline is `rg` + targeted `Read`, and it is much cheaper than
what the vendor measured against. Re-measure against it or the numbers mean nothing.

### Verify the harness before believing a bad result

Both evaluations below produced a false verdict on the first pass, from measurement errors in
the harness rather than defects in the tool:

- ripwire looked like it was missing call sites — it was `--uses` paginating at 100, and an
  `rg '\bNAME\('` ground truth matching language builtins.
- zvec-grep scored 2/10 — four questions had run against repos with no index, and three had an
  empty ground-truth set because the definition regex missed class methods.

A result worse than plausible is a hypothesis about the harness first, about the tool second.

## ripwire — rejected 2026-09-11

[redhat-et/ripwire](https://github.com/redhat-et/ripwire) v0.6.0 — C++23, tree-sitter, builds a
ranked call graph. Evaluated across 5 TS repos + optional reference project (Go).

**Install hazard:** the official `curl | bash` installer, and the release tarball, carry 17
skills plus hooks that land in `~/.claude/skills/`. That violates the hub's "never modify
`~/.claude/` autonomously". Install the binary alone if this is ever revisited.

**Scale is not the problem** — index 0.03-0.08 s, 21-166 MB RSS, orient output 21-37 KB.

**The graph is.** Edge density per symbol: Go 1.90 against 0.17-0.55 across every TS repo;
sample-project produced 17,870 symbols and 3,097 edges with 51% of attempted calls lost. On our
stack the call graph comes out nearly empty.

| verb | result |
|---|---|
| `--uses=SYM` | correct — 0 real misses, 0 false positives on every distinctive symbol; 196/196 at high cardinality. More precise than `rg`: drops comments, string literals, the definition |
| `--callers=SYM`, domain names | clean — 0 unsupported edges on 22 of 36 probes |
| `--callers=SYM`, names colliding with builtins | fabricated — `toLowerCase` 28/29, `digest` 35/35, `push` 32/33, `select` 17/17, Go's `append` 18/19 |
| `--situ` | blind to config/build-graph changes: 4 `tsconfig.json` edits returned blast radius 0, tests 0 |

The mechanism is one thing: resolution by NAME, without types. Any in-repo definition sharing a
name with a builtin captures every call to that builtin. **Go is not immune** — the failure is
name collision, not the language; Go simply has fewer collisions among domain symbols.

**Token cost inverts the sales pitch.** A fixed 3,955 B legend rides every response, so `--uses`
costs 1.8×-8.5× MORE than `rg` below ~200 call sites and only breaks even at 196.

**Verdict: rejected.** The only correct verb is also the most expensive one in the cardinality
range we actually work in, and the graph — where the differential value would come from — is not
trustworthy on TypeScript.

**Revisit if:** the resolver gains type information (an LSP or tsc-backed mode), or the fixed
legend becomes suppressible. A Rust or Go-dominant project would deserve its own measurement —
the domain-symbol results there were clean.

**Credit where due:** it publishes its own losses (`counts_floor`, `declined`, `ambiguous`,
`unresolved`, `external`) more honestly than most. Disclosure does not make a fabricated edge
usable, but it is why the failure was diagnosable in an afternoon.

## zvec-grep — rejected 2026-09-11 (and informally 2026-09-02)

`@zvec/zvec-grep` (`zg`) 0.2.1 — local vector + FTS retrieval over a per-repo index. 0.2.2 exists
but pnpm's 7-day cooldown blocks it; the cooldown was NOT bypassed. OSV clean on 0.2.1.

**First evaluated 2026-09-02** against `jbcontext`, informally, pre-rubric — jbcontext won on
actionability. Discarded, never reached the rules, and left 70 MB of indexes plus an uncommitted
`.gitignore` line behind for nine days.

**Method:** 10 natural-language questions across 4 TS repos, none naming the target symbol, each
with its correct file established by hand before scoring.

| | |
|---|---|
| recall@1 · @5 · @10 | 3/10 · **5/10** · 6/10 |
| query latency (median) | 787 ms |
| index | 35 MB/repo, 2.6-7.4 s (907 files → 11,566 entities) |
| output vs an `rg` vocabulary sweep | **1,579 B vs 23,706 B** — and the sweep returns 154-1,080 FILES to triage against 6-10 ranked hits |

**Triage cost is where it genuinely wins**, and by a lot. A 400-file sweep is not an answer, it
is another task.

**No predictor of failure, which is what sinks it.** Two hypotheses tested and both refuted:
docstring density (`hashContactPhone` has 6 lines of adjacent prose and misses; `getDoctorById`
has 1 and lands at #2) and entity size (misses return 5-28 line entities, not giant classes).
The failures return topically plausible but wrong neighbours — "duration string parsed into
milliseconds" returns Prometheus constants and `prisma.service.ts`, never `parse-duration.ts`.
It matches domain vocabulary, not the functional relationship.

If you cannot anticipate which questions land, every result needs verifying — so it can never be
a sole source. That is consistent with the existing rule that a search hit is a pointer and never
a verdict, but it caps what the tool buys.

**Verdict: rejected** at 6/10 without a failure predictor.

**Revisit if:** a remote embedding model is acceptable. The evaluation is anchored to
`local/potion-code-16m-v2` — 16M parameters, 256 dimensions, a small model — and `zg config model
set` accepts remote endpoints. That decision is not only about recall: it means an API key and
sending source code to a third party.

**Install hazard:** `zg install` wires MCP integrations into every agent on the machine, same
class of problem as ripwire's installer. It was not run in either evaluation.
