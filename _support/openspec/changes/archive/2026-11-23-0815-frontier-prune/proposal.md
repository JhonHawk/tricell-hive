# frontier-prune — Write Hive for frontier models and prune what they do not need

| Field | Value |
| --- | --- |
| Status | Completed · PR #598, merge 013a9725 |
| Tracker | none |
| Git | tricell-hive, base `development` at `34395b92`, proposed branch `refactor/frontier-prune` |
| Playbook | refactor |

## Decisions

2026-11-22
- **Writing target** — Which models is the distributed guidance written for? Claude Opus 5.5 and OpenAI gpt-6.2-sol, no longer the small models (Haiku, flash-class, Sonnet as a target).
- **D1 (free text)** — What happens to the role profiles on Sonnet 5.5 and gpt-6.1-terra? They stay: the user holds that wording that works for Opus/Sol works one tier down (Sonnet, Terra) and one up (Fable 5.2, Astra 6.2). Research found no vendor statement of cross-tier transfer, and agentic benchmarks place Sonnet and Terra close to Opus and Sol.
- **D2 (free text)** — Whose failures count toward a correction? The writing target, one tier up and one tier down; Haiku and Luna are discarded.
- **D3 B** — How is the prune done? In one batch, after changing «Measurement» so that the writing-target vendors' current prompting guidance together with the user's decision is evidence to relax or remove a rule.
- **D6 A** — What happens with Grok 4.7, the Grok `verifier`? It stays on the floor and in its profile, although independent benchmarks place it below Sonnet 5.5.
- **D8 A** — What happens with the six roles whose body is generic advice? Their bodies are trimmed and their names, profiles and boundary footer stay, so no installed role is renamed.
- **assumption (D4)** — Grok 4.7's variant `grok-4.7-build-fast` counts with Grok 4.7 under D6 A; another main-session model joins the floor only by the user's decision, never a Haiku, Luna or flash-class model. Only current model versions count: `gpt-6-astra`, `gpt-6.1-sol` and `grok-4.6` leave the floor; `grok-4.7-build-fast` stays, since it is the user's current Grok main-session model and normalization does not map it to `grok-4.7`. Discarded: keeping old versions. Observations already counted on them stay in their watch items. **Replaces:** the first D4 wording, which also removed `grok-4.7-build-fast` (guidance plan review P2-6).
- **assumption (D5)** — Codex's `focused` profile (gpt-6.1-terra) moves from `medium` to `high`. **Retired:** both roles on `focused` set their own effort (`hive-write-docs` medium, `hive-read-state` low), which replaces the profile's, so the change would alter no role (server plan review P2-1); the profile stays `medium`.
- **assumption** — The principle leaves stay: pstack keeps its principle leaves too and the user adopted them (skill-leaves D7). Discarded: deleting five of them.
- **assumption** — The «Candidates» section of `close-report` stays: it is a Hive feature, not scaffolding. Discarded: removing it.
- **assumption** — The `create-report` HTML skeletons stay; only its diagram references are trimmed. Discarded: keeping the baseline and one variant.
- **assumption** — Seven playbooks that only order stages are deleted and become rows of `hive-mode`'s table; the `Playbook` field stays. Discarded: removing the field.
- **D10 A** — The build missed the first size criteria (712,928 bytes of Markdown, footers at 71–78%) without cutting Hive procedures; what now? The criteria follow what was reached (AC7 ≤ 715,000, AC8 footers ≤ 80%): vendor guidance supports removing what the model already does, not Hive's own procedures.
- **assumption** — The flash-class pilot code in `tests/pilot/` stays; only its tie to the floor goes. Discarded: deleting the pilot tooling.
- **assumption** — The stage-packet entry keeps its missing-packet, end-line and missing-heading fallbacks in one shorter sentence: they cover deployment state and silent cuts, and the `--audit-session` packet criterion reads the end line. Discarded: removing them. **Replaces:** the first wording, which kept only the cut clause (guidance plan review P2-3).
- **assumption** — `review-evidence` and `review-findings` stay separate leaves: their moments differ (a reviewer confirming a finding, the main thread after verify's reviews), and the index requires one trigger per leaf. Discarded: merging them (guidance plan review P2-2).
- **assumption** — Flow steps keep the links to the leaves they read at that moment (`stored-data`, `security-boundaries`, `defect-fix` and the like), as the global «Rule leaves» section and `instruction-resources.md` set. Discarded: deleting them because the global index carries the same trigger (guidance plan review P3-11).
- **assumption** — The «Reporting» paragraph of the six remaining playbooks goes, reversing the 2026-11-21 addition: it restates the close-report row `flow-close` already picks, and its enforcer row in `rule-enforcers.md` goes with it. Discarded: keeping it.
- **assumption** — The pilot runtime code stays; the pilot tests that read the real floor document (`tests/pilot/qualify_run_test.go`) change with it. **Replaces:** "the pilot tooling is out of scope" (both plan reviews, P1).

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> Evaluación de reglas. Cuales de las reglas que tenemos y skills son realmente necesarias? Por ejemplo, cuales son innecesarias porque el modelo deporsi lo haria de esa manera y seria redundante, tautologias. Es posible hacer una poda? Validando acortar reglas innecesariamente redactadas de forma larga. Vamos a ampliar nuestro piso, ls reglas ahora van a ser orientadas para Opus 5.5, Sol 6.2 y de ahi para arriba, ya no orientemos hacia los modelos pequeños.

> Esto claro que incluye research sobre cómo crear y mantener skills, si esta sesión tiene una auditor de skills ejecutemoslos aun sin hacer cambios

No ticket.

## Objective

Today the guidance is written for the role-profile models, Sonnet and Terra included, and the support floor counts failures on Haiku, flash-class and old versions; `content/` holds 803,383 bytes of Markdown, with leaf preambles, playbooks and role bodies that restate other homes or generic engineering practice. When done, the writing target is Opus 5.5 and gpt-6.2-sol, the floor is that target with one tier up and down plus Grok 4.7, «Measurement» accepts vendor guidance with the user's decision as evidence, and the Markdown of `content/` is at least a fifth smaller with every invariant and every non-obvious rule kept.

## Scope and acceptance

In scope: `content/` Markdown (global core, `hive-core` leaves, every skill, every role), the floor document, `integrations/agent-profiles.json`, the repository `AGENTS.md` «Guidance structure» and «Measurement», the `model-qualification` spec, the content tests that pin moved or cut text, the harness-audit rule catalog, and the public pages that state the floor (`README.md`, `llms.txt`, `site/`). Out of scope: scripts and their behavior, the HTML report skeletons, the pilot runtime code (its tests that read the floor document change), renaming or merging roles, and any change to invariants (authorization, secrets, irreversible effects, preservation).

- AC1 — The floor document marks Opus 5.5 (`opus`) and gpt-6.2-sol as the only writing targets; its floor holds them, Fable 5.2, gpt-6.2-astra, Sonnet 5.5 (`sonnet`), gpt-6.1-terra, Grok 4.7 and `grok-4.7-build-fast`, and no Haiku, Luna, flash-class or superseded version (`gpt-6-astra`, `gpt-6.1-sol`, `grok-4.6`); the floor test requires every profile model to have a row and the writing-target rows to be exactly those «Two sets» lists. **Replaces:** the first AC1, which also moved Codex `focused` to `high` and removed `grok-4.7-build-fast`. · false today because `model-floor.md` marks `sonnet`, `gpt-6.1-terra` and `grok-4.7` as writing targets and lists `claude-haiku-5-5`, three flash models and three superseded versions, and no Fable or Astra row · Observed in: `_support/docs/harness-engineering/model-floor.md` and `go test ./tests/content -run TestModelFloor`
- AC2 — The repository `AGENTS.md` «Measurement» accepts, as evidence to relax or remove a rule, the user's decision resting on the current prompting guidance of a writing-target model's vendor that states the behavior as a default or the rule as scaffolding to remove; «Guidance structure» no longer says a vendor guide cannot justify a removal, and its observable-criterion rule no longer cites smaller models as its reason. · false today because `AGENTS.md:22` says "A frontier model's prompting guide or default behavior alone does not justify removing a rule" and line 21 says "since literal and smaller models act on what they can check" · Observed in: `AGENTS.md` and `_support/openspec/changes/frontier-prune/specs/model-qualification/spec.md`
- AC3 — The Markdown under `content/` totals at most 635,000 bytes. **Replaces:** the first AC3 (630,000), after the plan review kept the flow-step leaf links and the two review leaves. · false today because `find content -name '*.md' -exec cat {} + | wc -c` prints 803383 · Observed in: `content/` at the candidate **Retired:** replaced by AC7 after the build (D10 A).
- AC7 — The Markdown under `content/` totals at most 715,000 bytes. **Replaces:** AC3; the build cut every restatement, tautology, default and scaffolding passage and reached 712,928 bytes, and the rest is Hive procedure the user chose to keep (D10 A). · false today because the base prints 803383 · Observed in: `content/` at the candidate
- AC4 — The playbooks `investigate`, `bugfix`, `feature`, `hotfix`, `infra-ci`, `pr-review-deploy` and `deliverable` no longer exist as files, and `hive-mode`'s table still names all 13 kinds with what is specific to each. · false today because `content/skills/hive-mode/playbooks/` holds 13 files · Observed in: `content/skills/hive-mode/`
- AC5 — In each of `hive-build-backend`, `hive-build-data`, `hive-build-infra`, `hive-write-tests`, `hive-tune-performance` and `hive-design-architecture`, the body between the frontmatter and the boundary footer is at most 900 bytes, with name and profiles unchanged, and each of the 21 roles' boundary footer, from «Stay within the parent» to the end, is at most 65% of its base size. **Replaces:** the first AC5 (1,800 bytes per whole file), which the frontmatter and footer alone nearly filled (server plan review P1-2), and its fixed footer caps, since the footers differ by role (870 to 2,164 bytes). · false today because the bodies of `hive-build-backend`, `hive-build-data`, `hive-build-infra` and `hive-design-architecture` measure 1,266, 988, 1,407 and 1,574 bytes and every footer is at its base size and the footer is 1,883 bytes in each implementing role · Observed in: `content/agents/` **Retired:** replaced by AC8 after the build (D10 A).
- AC8 — In each of the six generic roles the body between the frontmatter and the boundary footer is at most 900 bytes, with name and profiles unchanged, and each of the 21 roles' boundary footer is at most 80% of its base size. **Replaces:** AC5; the footers reached 71–78% keeping every boundary they state (D10 A). · false today because four of those bodies exceed 900 bytes and every footer is at its base size · Observed in: `content/agents/`
- AC6 — `README.md`, `llms.txt` and the site's home page name Opus 5.5 and gpt-6.2-sol and describe the writing target as Opus 5.5 and gpt-6.2-sol with a floor one tier up and down, and no longer present Haiku or flash-class models as floor candidates. · false today because `README.md:50-52` and `site/src/content/docs/index.mdx:39-41` name the Haiku and flash-class candidates · Observed in: `README.md`, `llms.txt`, `site/src/content/docs/index.mdx`

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- Audits on 2026-11-22 at `32bef8c8` (five read-only children: `hive-core` leaves, flow skills, other skills and roles, vendor guidance, pstack and earlier Hive), a `hive-review-harness` audit, and a tier-transfer research; findings are condensed in `design.md`. `ec51a107` merged the `comment-audit` change since, touching `hive-core/build/defect-fix.md`, `review/no-comments.md`, `review/suppression-review.md` and adding `content/skills/comment-audit/`: the build reads the current files, not the audit's line numbers.
- `quick_validate.py` (Claude Code and Codex copies, PyYAML 6.0.3) on 18 skills at `32bef8c8`: 16 valid; `hive-core` and `hive-typescript` fail on their documented Claude Code keys.
- `codex debug prompt-input` from the repository root: 58 KB of AGENTS chain, 22 KB of skill list, 11 KB of role list per session.
- `34395b92` (PR #597, `comment-audit` license-header fixes) changed only `comment-audit/references/ratchet.md` in `content/`; T6's base already includes it.
- Markdown bytes at `ec51a107`: `content/` 802,485; `hive-core` 207,512; flow skills and `hive-mode` 192,644; other skills 268,129 plus `comment-audit` 16,248; `content/agents` 87,283; `global.md` 25,909.
- `go vet ./... && go test ./... && python3 -m unittest discover -s tests/skills -p '*_test.py'` · vet clean, 23 Go packages ok, 412 Python tests OK, 296 s · 2026-11-22 (at `ec51a107`)
- `go test ./tests/content -run 'TestModelFloor' -v` · ok, 2 tests · 2026-11-22
- `go test ./tests/content -run 'TestGlobalGuidance|TestIndexTriggers'` · ok · 2026-11-22
- `wc -c` of the six generic roles · backend 3454, data 3163, infra 3600, write-tests 2797, tune-performance 3089, design-architecture 3980 · 2026-11-22
- Role body and footer bytes (`python3 -I -c "import sys; [print(len(t.partition('\\n# ')[2].partition('Stay within the parent')[0].encode()), len(t.partition('Stay within the parent')[1].encode()+t.partition('Stay within the parent')[2].encode()), f) for f in sys.argv[1:] for t in [open(f).read()]]" content/agents/*/*.md`) · bodies: backend 1266, data 988, infra 1407, write-tests 602, tune-performance 918, design-architecture 1574; footers: 1883 in 11 implementing roles, 2164 owner, 1625 verify-change and review-ux, 1422 verify-task, 870 in 6 observing roles · 2026-11-22
- Byte bases per task scope (`find <scope> -name '*.md' -exec cat {} + | wc -c`) · T3 102715, T4 104199, T5 192644, T6 372558 · 2026-11-22
- `rg -n -i 'haiku|flash-class' README.md llms.txt site/src/content/docs/index.mdx` · `llms.txt:73`, `README.md:52` · 2026-11-22
- `rg -c 'Opus 5.5|gpt-6.2-sol' README.md llms.txt site/src/content/docs/index.mdx` · no match · 2026-11-22
- `rg -c 'gpt-6-astra|gpt-6.1-sol|grok-4.6|claude-haiku|luna' _support/docs/harness-engineering/model-floor.md` · 6 · 2026-11-22
- `rg -c 'claude-fable-5-2|gpt-6.2-astra' _support/docs/harness-engineering/model-floor.md` · no match · 2026-11-22
- `rg -n 'playbooks/(investigate|bugfix|feature|hotfix|infra-ci|pr-review-deploy|deliverable)\.md' content site _support/docs` · 2 lines · 2026-11-22
- `go test ./tests/pilot -run 'Qualify'` · ok, 28 s · 2026-11-22
- `rg -c` of the 13 kind rows in `content/skills/hive-mode/SKILL.md` · 13 · 2026-11-22
- `rg -n "playbooks/(…)\.md|13 playbooks" content site _support/docs README.md llms.txt` · 6 lines · 2026-11-22
- Fable 5.2 and gpt-6.2-astra model IDs and aliases · current vendor documentation · not measured, T1 · 2026-11-22
- `rg -n "alone does not justify removing a rule|literal and smaller models" AGENTS.md` · lines 21 and 22 · 2026-11-22
- `ls content/skills/hive-mode/playbooks/` · 13 files · 2026-11-22
- `find content -name '*.md' -exec cat {} + | wc -c` · 803383 (at `34395b92`) · 2026-11-22
- `python3 -I <flow-plan-dir>/scripts/check_change.py _support/openspec/changes/frontier-prune` · exit 0 · 2026-11-22

## Effects

None: the tasks take no effect on an external system.

## Delivery

2026-11-22: `hold` — no commit and no publication; the verified working tree is the result, and the change folder stays on disk until the user picks another mode. Code review: Native, Claude Code `/code-review`. No security review: the change touches no security boundary (invariant sentences keep every condition). `hold` authorizes no commit, push, pull request, merge or local installation refresh.

## Review status and progress

2026-11-22 — Plan written.
2026-11-22 — Server reviewer's second reread: one P3 (a footer rewritten without its opening words would pass the check) fixed in T6 and the design.
2026-11-22 — Server reviewer's reread: T7 gets a byte target so the targets sum to AC3; the «Writing target» bullet carries only model values; `qualify_run_test.go` takes a fixed floor document (its OpenCode cases have no row left); AC5 gets a measuring command and per-role footer bases.
2026-11-22 — Base moved to `34395b92` (PR #597, `comment-audit` only). Guidance reviewer's reread: P2 (the tier rule for main-session models had no test and contradicted the `grok-4.7-build-fast` row) fixed by naming the rule as the user's decision, never Haiku, Luna or flash-class; P3 (`README.md:31-33` names 13 playbooks) added to T8 with a wider check.
2026-11-22 — Plan review by two `hive-review-plan` children (server: 2 P1, 6 P2; guidance: 1 P1, 9 P2, 3 P3). Applied: the pilot floor test enters T1; AC1 drops the Terra effort (D5 retired) and keeps `grok-4.7-build-fast`; AC3 to 635,000; AC5 measures the body and the footer; the spec delta carries the restatement clause, the `--audit-session` wording and three scenarios forward; the packet fallbacks, the two review leaves, the flow-step leaf links, the CI-wait home in `merge-gate`, `global.md`'s «Inspect relevant state», the `project-verification-format` safety sentences and HA-IF-12/16 stay; T8 owns the public and architecture text about playbooks; per-task byte targets sum to AC3; verified-context lines added.
