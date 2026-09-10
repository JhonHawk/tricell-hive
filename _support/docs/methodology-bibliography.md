# Methodology Bibliography

Sources backing the rules and conventions under `global/rules/`. Append-only in practice: when a rule's design draws on external authority, record the source here so it isn't re-researched. Living document — no date in the filename; entries are added as conventions are grounded. **The one removal exception:** a section is deleted only when the rule it backed is itself deleted — its sources would otherwise dangle. Precedent: the `## Project maturity tiers` section was removed on 2026-06-22 when `quality/project-maturity.md` was deleted (see `_support/spec/2026-06-22-agent-verifiable-quality-gates-design.md`).

## Conventions for this file
- **Authority tags:** `[Authoritative]` — ISO/IETF/standards bodies, canonical books, official project docs · `[Semi-authoritative]` — named-author manifestos, major-vendor reference docs · `[Non-authoritative]` — practitioner/community consensus (corroboration only, never the basis).
- Cite the precise locus — standard part/number, book chapter, official URL — not the bare title.
- Each section names the rule it backs and the verdict the research reached: *supported* / *adjusted* / *contradicted*.

---

## Generated-artifact naming, grouping & retention

Backs `workflow/support-artifacts.md > Generated-artifact naming, grouping & retention` (moved 2026-07-17 from `project-structure.md`) and its condensed mirror in `harness/AGENTS.md`. Researched jun-2026.

### Date naming — ISO 8601, lexicographic = chronological
- `[Authoritative]` **ISO 8601-1:2019**, Part 1 (basic Gregorian dates) — `YYYY-MM-DD`, components ordered largest unit first. *Supported.* Caveat: lexicographic order equals chronological **only** with fixed-width zero-padding (`2026-06-09`, not `2026-6-9`).
- `[Authoritative]` **NARA (US National Archives), *Best Practices for File Naming*** — "Use international standard date notation (YYYY-MM-DD)… makes sure all your files stay in chronological order." Corroborated by Stanford, Harvard, and Harvard Medical School RDM guides.

### Date position — prefix vs suffix
- `[Authoritative]` **Harvard Medical School Data Management** — "Put the most important information first… if you anticipate wanting to find a file by date, then put the date first." Conditional, not universal.
- `[Authoritative]` **Stanford Libraries** file-naming guidance — own example is subject-first with the date mid/suffix; presents date-first as one option, not a rule.
- `[Authoritative]` **NARA** — recommends the ISO format but is silent on position.
- **Verdict: adjusted.** Derive position from the primary retrieval axis (most significant element first): prefix for chronology-primary artifacts, suffix when subject/type leads. Do **not** hardcode "date first".

### Snapshot (dated) vs living document (undated)
- `[Authoritative]` **ISO 15489-1 (Records Management)** — the formal discipline behind "declared record = fixed / point-in-time" vs a living document under revision. (Exact text paywalled; governing standard, not a literal quote.)
- `[Semi-authoritative]` **IBM, *What Are Immutable Snapshots?*** — snapshots are "read-only, point-in-time copies… cannot be altered once created." *Supported.*

### Grouping a cohesive deliverable in a folder
- `[Authoritative]` **R. C. Martin, *Clean Architecture* ch. 13 (Common Closure Principle)** — "classes that change together belong to the same component"; component-level SRP. The principled root of "things that change together belong together".
- `[Authoritative]` **R. C. Martin, *Clean Architecture* ch. 21 (Screaming Architecture)** — directory structure should announce purpose, not framework. Most on-point (it is itself about directory layout).
- `[Authoritative]` **Ousterhout, *A Philosophy of Software Design* (2018), ch. 4–5** — deep modules / information hiding; narrow support only.
- **Verdict: adjusted.** This canon is about source-code modules; applying it to generated-artifact layout is a legitimate **analogical extension** — phrase as "following", not "per". Do **not** cite FHS here (it standardizes OS-level directory categories, not deliverable grouping — borrowed authority). "Colocation" has no canonical origin; CCP is its principled root.

### Intention-revealing names, no generic names
- `[Authoritative]` **R. C. Martin, *Clean Code* (2008), ch. 2 "Meaningful Names"** — "Use Intention-Revealing Names" / "Avoid Disinformation"; the `Product`/`ProductInfo`/`ProductData` "noise words" anti-pattern is almost literally the case against `report`/`output`/`data`. *Supported strongly.*

### Retention — ephemeral (reproducible) vs durable (original)
- `[Authoritative]` **Twelve-Factor App, Factor V (Build, release, run)** — strict build/release/run separation; releases are an append-only ledger. Backs immutability + reproducibility-via-rebuild.
- `[Authoritative]` **Git official docs — `gitignore`** — patterns for "files generated as part of build… derivable from source… should not be committed." Canonical cite for "gitignore generated output".
- **Verdict: adjusted framing.** Artifacts are disposable-because-reproducible, not disposable-because-worthless. Discriminator to encode: reproducible-from-source ⇒ ephemeral/gitignored/purged; original non-reproducible ⇒ durable.

### Differentiated retention schedules
- `[Authoritative]` **ISO 15489-1:2016 (Records Management)** — retention/disposition governed by a disposition authority / retention schedule; disposition is deliberate and scheduled. Corroborated by **ISO 16175** and **DoD 5015.02-STD**.

### Curation — keep the meaningful subset, discard the rest
- `[Authoritative]` **DCC Curation Lifecycle Model (Higgins, *IJDC* 3(1), 2008)** — "Appraise and Select" ("not feasible to retain everything… decide what is kept") and "Dispose"; rationale: retaining everything creates noise that hampers meaningful search. Archival appraisal lineage (Schellenberg).
- **Honesty note:** no authoritative QA standard mandates "keep 3 of 50 screenshots". Anchor the curation pattern in DCC appraisal/selection, not a non-existent QA standard. ISO/IEC/IEEE 29119-3 defines test-doc *types*, not retention curation.

### Discoverability — naming vs README
- `[Semi-authoritative]` **DHH, *The Rails Doctrine* (Convention over Configuration)** — convention self-documents the conventional 95%; deviation needs an explicit marker.
- `[Authoritative]` **R. C. Martin, *Clean Code* ch. 4 (Comments)** — redundant comments "collect lies" and drift; a descriptive README is the directory-level equivalent.
- `[Authoritative]` **Hunt & Thomas, *The Pragmatic Programmer* (DRY)** — single authoritative representation: forbids descriptive READMEs (they duplicate the naming), mandates rationale READMEs (no other home for that knowledge).
- `[Authoritative]` **Ousterhout, *APoSD*** — contradicts pure self-documenting structure: structure conveys the *what*, never the *why* or the contract.
- `[Semi-authoritative]` **Preston-Werner, *README-Driven Development* (2010)** — README at project/package boundaries; rejects proliferation.
- **Verdict: resolved.** Naming carries discoverability by default; README only for what structure can't encode (convention deviation, ownership boundary, non-obvious contract); never a descriptive file listing.

## Critical-thinking posture: risk-surfacing, verify-before-assuming, no sycophancy

Backs `quality/critical-thinking.md` (risk surfacing, scope questioning, verify-cheaply, honest uncertainty, agent-output scrutiny, pre-ship ownership test) and its condensed mirror in `harness/AGENTS.md`. Researched jun-2026.

### No sycophancy / be direct — egoless review culture
- `[Authoritative]` **Weinberg, *The Psychology of Computer Programming* (1971; Silver Anniversary ed. 1998), ch. 4 (egoless programming)** — the review's objective is for everyone, *the author included*, to find defects, not to prove the work has none. The principled root of "challenge ideas — no softening, no flattery". (Locus confirmed: ch. 4, not the later social-activity chapters.)
- `[Authoritative]` **Anthropic — Sharma et al., *Towards Understanding Sycophancy in Language Models* (2023), arXiv:2310.13548** — both humans and preference models prefer convincingly-written *sycophantic* responses over correct ones a non-negligible fraction of the time; sycophancy is a general RLHF-assistant behavior driven partly by human preference judgments. Grounds *why* the no-sycophancy posture is load-bearing for an LLM agent specifically.
- **Verdict: supported.**

### Surface the top 1-3 risks before proceeding
- `[Authoritative]` **Klein, *Performing a Project Premortem*, Harvard Business Review 85(9), Sept 2007** — assume the plan has already failed, then generate plausible reasons for its demise; making it safe to voice reservations during planning improves outcomes. Direct analog of "state the top 1-3 failure modes unprompted".
- `[Authoritative]` **Mitchell, Russo & Pennington, *Back to the future: Temporal perspective in the explanation of events*, J. Behavioral Decision Making 2(1):25-38 (1989), DOI 10.1002/bdm.3960020103** — the study Klein invokes for *prospective hindsight*. The widely-quoted "~30% more correctly-identified reasons" is **Klein's HBR framing of this study, not a headline statistic in the paper** (the paper's own experiment 1 finds temporal perspective had little effect; uncertainty drove the difference). Cite for the prospective-hindsight mechanism, not as a clean "+30%" finding.
- **Verdict: adjusted.** The same study cautions the extra reasons are *episodic* — "seeing more, not necessarily seeing better." This is the canonical justification for the rule's **1-3 cap**: enumerate the load-bearing risks, do not pad. Do not "improve" the rule by lifting the cap.

### Verify cheaply before reasoning expensively / Decide by owner (formerly "Ask before assuming")
- `[Authoritative]` **Hunt & Thomas, *The Pragmatic Programmer* — Tip 27 "Don't Assume It — Prove It"** — prove assumptions in the actual environment, with real data and boundary conditions. Locus: the Debugging topic, "The Element of Surprise" subsection (topic 18, "The Basic Tools" chapter in the 1st ed.; regrouped under "While You Are Coding" in the 20th-anniversary ed.). Tip text confirmed verbatim. Near-verbatim support for "settle the assumption with a cheap action before building a plan on the guess".
- **Verdict: supported.** (Same locus also grounds `debugging.md > Establish authoritative ground truth` — share the cite, don't re-research.)

### Question scope, offer alternatives, make tradeoffs explicit
- `[Authoritative]` **McConnell, *Code Complete*, 2nd ed. (2004), ch. 21 "Collaborative Construction", §21.3 Formal Inspections** — review emphasizes defect *detection* over correction, checklists target historically problematic areas, reviewers get time to prepare. Chapter and section confirmed against the published ToC. Grounds the construction-time review posture behind "surface risks/alternatives before implementing".
- **Verdict: supported, with honesty note.** Code Complete grounds the *review-during-construction* stance; the specific directives "always present what you give up" and "offer a simpler Y unprompted" are sound practitioner convention, **not** a chapter-and-verse mandate. Treat as corroborated, not standard-imposed.

### Agent output is dangerously convincing
- `[Authoritative]` **Anthropic — Sharma et al. (2023), arXiv:2310.13548** — grounds the *mechanism*: convincing-but-incorrect output is systematically preferred by humans and preference models; optimizing against preference models can sacrifice truthfulness.
- **Verdict: adjusted.** The paper studies free-form *text* generation, not "code that passes CI"; it evidences *why* polished agent output warrants suspicion but does not itself issue the directive "apply more scrutiny to agent changes". Keep as mechanism evidence, not as a mandate the paper hands down.

### Pre-ship ownership test (would I own the incident?)
- `[Authoritative]` **Klein, *Performing a Project Premortem*, HBR Sept 2007** — the prospective-hindsight stance ("imagine it already failed — why?") is the analog of "would I own the incident tied to this code?".
- **Honesty note:** questions 1-2 (behavior under real load, blast radius to production/customers) are SRE / operational-readiness practitioner convention with **no single canonical book locus** — do not attach a borrowed source. The premortem grounds the failure-imagination stance, not the three-question checklist.

### Quality over token thrift
- **Honesty note:** no external authority — this is an internal cost/quality economics policy specific to this agent setup, and needs none. Do not manufacture a citation. It is a deliberate workflow preference, not a grounded methodology claim.


## Debugging — diagnosis discipline

Backs `quality/debugging.md` (root cause before fix, verify ground truth, one change at a time, instrument boundaries, three-strikes circuit breaker, reproduce before claiming fixed) and its condensed mirror in `harness/AGENTS.md`. Researched jun-2026. Verified against *The Pragmatic Programmer* (20th Anniversary Ed. / 2nd Ed., 2019), Topic 20 "Debugging" (pp. 89–97), the Agans rule set, and Zeller.

### Root cause before fix — read the full error, reproduce, check what changed
- `[Authoritative]` **Hunt & Thomas, *The Pragmatic Programmer* (20th Anniversary Ed., 2019), Topic 20 "Debugging" — Tip 32 "Read the Damn Error Message" (p. 92)** and the "Debugging Mindset" prose: *"Beware of myopia when debugging. Resist the urge to fix just the symptoms you see… Always try to discover the root cause of a problem, not just this particular appearance of it."* Almost verbatim our rule.
- `[Authoritative]` **Agans, *Debugging: The 9 Indispensable Rules* (2002), Rule 3 "Quit Thinking and Look"** — *"See the failure, see the details… guess only to focus the search"* — and **Rule 1 "Understand the System."**
- **Verdict: supported strongly.**

### Establish authoritative ground truth before diagnosing
- `[Authoritative]` **PP, Topic 20 — Tip 34 "Don't Assume It—Prove It" (p. 96)** — *"Don't gloss over a routine or piece of code involved in the bug because you 'know' it works. Prove it."* — plus Tip 33 *"'select' Isn't Broken"* (p. 95): the bug is far likelier in your code than in the OS/library/compiler.
- `[Authoritative]` **Agans, Rule 1 "Understand the System"** — read the actual manual / know the real road map before reasoning about a fault.
- **Verdict: adjusted.** Canon proves *the code's assumptions*; our rule extends "prove it" to *external operational state* (remote git refs via `fetch`, live DB schema, deployed config — never a local ref/cache/ledger). A faithful application to the agent's git/DB/deploy context, not a contradiction. Phrase as "following PP Tip 34", not "per". (Shares its root with `critical-thinking.md > Verify cheaply before reasoning expensively` — same Tip 34 lineage; keep the two consistent if either is edited.)

### One hypothesis, one change
- `[Authoritative]` **Agans, Rule 5 "Change One Thing at a Time"** — *"Isolate the key factor," "Use a rifle, not a shotgun," "Change one test at a time," "Compare it with a good one."* The principled root of our rule.
- `[Authoritative]` **Zeller, *Why Programs Fail: A Guide to Systematic Debugging* (2nd Ed., 2009), Ch. 6 "Scientific Debugging"** — the hypothesis → prediction → experiment → observation loop refines ONE hypothesis per experiment; debugging-as-science vs. trial-and-error.
- **Verdict: supported strongly.**

### Instrument boundaries in multi-layer systems
- `[Authoritative]` **PP, Topic 20 — "The Binary Chop"** (when a stack trace scrolls forever, choose a frame in the middle and halve the search — same technique applied to a dataset or a release range) and **"Logging and/or Tracing"**: tracing statements are *"peculiarly effective at diagnosing several classes of errors that debuggers can't."*
- `[Authoritative]` **Agans, Rule 3 "Quit Thinking and Look"** (*"build instrumentation in," "add instrumentation on"*) and **Rule 4 "Divide and Conquer"** (*"narrow the search with successive approximation,"* "determine which side of the bug you're on").
- **Verdict: adjusted.** Both halves are canonical — instrument to see real data, then narrow by halving. Caveat: canon's primary *locating* technique is the binary chop (logarithmic probes), not blanket logging at *every* boundary at once. Our "log every boundary, run once" is a valid eager variant when a re-run is expensive (multi-service request flow); for a deep single-process call stack a binary chop is fewer probes. Additive nuance, not a correction.

### Three failed fixes = question the approach, not the hypothesis
- `[Authoritative]` **Agans, Rule 8 "Get a Fresh View"** — when stuck, get a fresh perspective; *"report symptoms, not theories."* Backs "escalate when stuck."
- `[Authoritative]` **PP, Topic 20 — Tip 30 "Don't Panic" (p. 89)** / "Debugging Mindset" — step back and actually think about what could be causing the symptoms. Backs "stop patching, reconsider the foundation."
- **Honesty note.** The *principle* (repeated failed fixes ⇒ the pattern/architecture is now the suspect; stop and get a fresh view) is canonical. The *specific count of three* is NOT debugging canon. A 2026 web search surfaced "three strikes and rethink" only as **Jess's Rule** (NHS England, 23 Sep 2025), a GP medical-diagnosis protocol — real, but unrelated borrowed authority that must **not** be cited here. The "3" is a deliberate practitioner circuit-breaker we chose (cf. the Rule of Three for abstractions in `development-principles.md` — a shared heuristic "3", not a shared source). **Verdict: adjusted** — principle grounded; threshold is our convention, openly admitted.

### Reproduce before claiming fixed — red→green→red
- `[Authoritative]` **PP, Topic 20 — Tip 31 "Failing Test Before Fixing Code" (p. 91)** / "Reproducing Bugs": *"The best way to start fixing a bug is to make it reproducible. After all, if you can't reproduce it, how will you know if it is ever fixed?"*
- `[Authoritative]` **Agans, Rule 2 "Make It Fail"** (*"do it again," "stimulate the failure"*) and **Rule 9 "If You Didn't Fix It, It Ain't Fixed"** (*"check that it's really fixed," "check that it's really YOUR fix that fixed it"*).
- **Verdict: supported strongly.** Agans Rule 9 *strengthens* our rule with a step we under-specify: confirm the fix — not a coincidence — is what resolved it (the revert leg of red→green→red covers the regression-test efficacy, not the coincidence confound). Worth folding into the bullet.


## Development principles — implementation-time discipline

Backs `quality/development-principles.md`. Researched jun-2026. (The "Names reveal intent" bullet is already grounded above under *Intention-revealing names, no generic names* — Clean Code ch. 2; cross-referenced, not re-cited here.)

### Abstractions earn their place — Rule of Three
- `[Authoritative]` **M. Fowler, *Refactoring* (1999, 1e), ch. 2 "Principles in Refactoring", p. 50 (= 2e ch. 2)** — "Three strikes and you refactor"; two instances are tolerated, the third triggers extraction. Explicitly attributed by Fowler to **Don Roberts**.
- `[Semi-authoritative]` **Wikipedia, "Rule of three (computer programming)"** — corroborates the Fowler/Roberts attribution and the two-tolerated / three-extract threshold.
- `[Authoritative]` **S. Metz, "The Wrong Abstraction", sandimetz.com (20 Jan 2016)** — "duplication is far cheaper than the wrong abstraction"; backs "duplication is preferable to a wrong abstraction". (Author is Sandi Metz alone — NOT co-authored with Kent Beck. Aphorism originated in her 7 Mar 2014 tweet.)
- **Verdict: supported.** Locus note: the Rule of Three is in ch. **2** "Principles in Refactoring" (p. 50), not the "Bad Smells in Code" chapter (which is ch. 3). Canonical phrasing is **three** strikes (extract on the third); the rule's "repeats 2-3 times" is correct — do not restate it as "two strikes and you refactor".

### Don't guess performance — measure first
- `[Authoritative]` **D. Knuth, "Structured Programming with go to Statements", *ACM Computing Surveys* 6(4), Dec 1974, pp. 261–301 (quote on p. 268)** — "We should forget about small efficiencies, say about 97% of the time: premature optimization is the root of all evil. Yet we should not pass up our opportunities in that critical 3%."
- **Verdict: supported.** The "critical 3%" clause is the canonical form of the rule's own escape hatch — optimize *with evidence*, not never. (Knuth credits the sentiment to Hoare; Hoare later could not recall originating it — provenance is folklore-tangled, but the page locus above is firm.)

### Solution proportional to the problem — YAGNI / simplest thing
- `[Authoritative]` **R. Jeffries, A. Anderson, C. Hendrickson, *Extreme Programming Installed* (2000)** — YAGNI ("You Aren't Gonna Need It"), coined by Ron Jeffries: "Always implement things when you actually need them, never when you just foresee that you need them."
- `[Authoritative]` **K. Beck, *Extreme Programming Explained* (1999/2000, 1e)** — "do the simplest thing that could possibly work" (DTSTTCPW). (Phrase originates with Ward Cunningham; Beck popularized it through XP.)
- **Verdict: supported.** Same canon backs "Only change what was asked" (YAGNI applied to scope) and Fowler's **"Two Hats"** (*Refactoring* 1e/2e, ch. 2 "Principles in Refactoring") — Kent Beck's metaphor: keep adding-function and refactoring as separate changes; the principled root of "a bug fix doesn't need adjacent cleanup".

### Prefer one level of abstraction per function
- `[Authoritative]` **R. C. Martin, *Clean Code* (2008), ch. 3 "Functions"** — section "One Level of Abstraction per Function" and the "Stepdown Rule" (read top-down, each function one level below its name). **Verdict: supported.** The rule's hedge ("not as a rule applied mechanically") is the correct calibration.

### Observe before writing — consistency takes precedence
- `[Authoritative]` **J. Ousterhout, *A Philosophy of Software Design* (2018, 1e; 2e 2021), ch. 17 "Consistency" (§17.3 "Taking it too far", p. 141 in 1e)** — consistency benefits only genuinely *similar* things; doing things differently is fine when the case genuinely differs.
- `[Authoritative]` **A. Hunt & D. Thomas, *The Pragmatic Programmer* (1999, 1e ch. 2 / 2019 2e Topics 9–10)** — "The Evils of Duplication" (DRY) and "Orthogonality"; single authoritative representation, independent modules.
- **Verdict: adjusted.** Canon agrees "consistency over technically-valid alternatives" — but adds a caveat the bullet omits: consistency wins only for genuinely similar cases; a deliberately different case is legitimate, and an existing convention can itself be wrong. The escape valve already lives in `gap-resolution.md > Divergence Between Sources`; capturing the caveat here is additive.

### When an approach is going wrong, start fresh
- `[Authoritative]` **M. Fowler, *Refactoring* 2e (2018); martinfowler.com "Workflows of Refactoring"** — a full rewrite is *sometimes* appropriate when code is too difficult to refactor, but refactoring (incl. preparatory refactoring) is the default; rewrite is the exception.
- `[Non-authoritative]` **H. Arkes & C. Blumer, "The Psychology of Sunk Cost", *Organizational Behavior and Human Decision Processes* 35 (1985), pp. 124–140** — decision-theoretic root of "stop investing in a bad foundation". (Authoritative within behavioral decision theory; here it corroborates the root only, never the basis as a software-engineering source.)
- **Honesty note:** there is **no** canonical software-engineering rule named "start fresh" with a precise locus. Fowler's canon leans the *opposite* way (refactor by default), so frame this bullet as a **narrowed** reading: it fires only on a *fundamentally flawed* foundation — the genuine rewrite case — never as license to abandon refactorable code. The rule's existing wording ("fundamentally flawed", "suggest reverting") already scopes it correctly. **Verdict: adjusted.**


## Patterns & Anti-patterns (JS/TS, Java, Kotlin)

Backs `quality/patterns-antipatterns.md` (path-scoped to `**/*.{ts,tsx,js,jsx,mts,cts,mjs,cjs,java,kt,kts}`) and its condensed mirror in `harness/AGENTS.md`. Researched jun-2026.

### Error handling — never swallow, exceptions over return codes, never return null
- `[Authoritative]` **R. C. Martin, *Clean Code* (2008), ch. 7 "Error Handling"** — sections "Use Exceptions Rather Than Return Codes" and "Don't Return Null"; error handling is a separated, visible concern. Directly backs "never swallow errors".
- `[Authoritative]` **J. Bloch, *Effective Java* 3e (2018), Item 77 "Don't ignore exceptions"** (ch. 10 "Exceptions", p. 310) — an empty `catch` defeats the purpose of exceptions; at minimum log, ideally propagate. The literal case against `catch (e) { console.log(e) }`.
- **Verdict: supported.**

### Typed error classes, not string matching
- `[Authoritative]` **J. Bloch, *Effective Java* 3e (2018), Item 73 "Throw exceptions appropriate to the abstraction"** and **Item 72 "Favor the use of standard exceptions"** — exceptions carry meaning by type, not by parsed message text.
- `[Authoritative]` **J. Bloch, *Effective Java* 3e (2018), Item 70 "Use checked exceptions for recoverable conditions and runtime exceptions for programming errors"** — frames the checked-vs-runtime choice by recoverability.
- **Verdict: adjusted.** Canon mandates *typed* exceptions with intent; it does **not** mandate `RuntimeException` specifically. The rule's blanket "domain exceptions extending `RuntimeException`" is the pragmatic modern-Java/Spring (unchecked-by-default) convention — a defensible deviation from Item 70's recoverability-based guidance, not a contradiction.

### Never mutate function arguments; return copies / unmodifiable collections
- `[Authoritative]` **J. Bloch, *Effective Java* 3e (2018), Item 50 "Make defensive copies when needed"** (p. 231, ch. "Methods") — program defensively; copy before validating; don't let callers mutate internal state.
- `[Authoritative]` **J. Bloch, *Effective Java* 3e (2018), Item 17 "Minimize mutability"** (ch. "Classes and Interfaces") — immutable-by-default is the root discipline; `List.copyOf()` / unmodifiable collections are the JVM expression.
- `[Authoritative]` **J. Bloch, *Effective Java* 3e (2018), Item 76 "Strive for failure atomicity"** (ch. "Exceptions") — a failed invocation should leave the object in its pre-call state; mutating arguments mid-operation breaks this.
- **Verdict: supported strongly.**

### Validate env vars at startup — crash at boot, not at runtime
- `[Semi-authoritative]` **J. Shore, "Fail Fast", *IEEE Software* vol. 21, no. 5 (Sep/Oct 2004), pp. 21-25** (full text: `martinfowler.com/ieeeSoftware/failFast.pdf`) — fail immediately and visibly when an error occurs rather than continuing into an inconsistent state. Canonical origin of the term; a named-author magazine column, not a peer-reviewed standard.
- `[Authoritative]` **Twelve-Factor App, Factor III "Config"** (`12factor.net/config`) — "strict separation of config from code"; "stores config in *environment variables*".
- **Verdict: supported.** Twelve-Factor backs externalization/env-vars but is **silent** on *when* a bad config surfaces (verified — no boot-vs-runtime guidance in the text); the "crash at boot, not at runtime" half is the Fail-Fast contribution. Together they fully ground the rule.

### Externalize environment-varying values; hardcode only truly-fixed constants
- `[Authoritative]` **Twelve-Factor App, Factor III "Config"** (`12factor.net/config`) — config is "everything that is likely to vary between deploys"; the litmus for externalize-vs-hardcode is exactly whether a value varies by environment.
- **Verdict: supported.**

### Layer responsibilities — thin controllers, HTTP-unaware services, pure repositories
- `[Authoritative]` **M. Fowler, *Patterns of Enterprise Application Architecture* (2002), "Service Layer"** — defines the application's boundary as a layer of services that coordinates each operation; the presentation layer interacts only through that boundary.
- `[Authoritative]` **R. Wirfs-Brock & A. McKean, *Object Design: Roles, Responsibilities, and Collaborations* (2002)** — the **role-stereotype** vocabulary; *Controller* (directs), *Service Provider* (computes), *Information Holder* (knows) map onto the controller/service/repository split. **Honesty note:** the book defines **six** stereotypes (Information Holder, Structurer, Service Provider, Coordinator, Controller, Interfacer) — these three are selected as the relevant analogy, not an exhaustive citation of the book's taxonomy.
- **Verdict: adjusted.** PoEAA establishes the *boundary* but does not itself mandate that services be "HTTP-unaware" or throw domain (not HTTP) errors — that non-negotiable is a layered-architecture/DDD corollary (keep transport/framework out of the domain), well-supported by the layering principle but **not** a literal Fowler quote. "Thin controller" is practitioner vocabulary; its root is the boundary + role separation above.

### Never return entities as API responses — map to DTOs
- `[Authoritative]` **M. Fowler, *Patterns of Enterprise Application Architecture* (2002), "Data Transfer Object"** — an object carrying data across a process boundary, decoupling the wire shape from the domain model; contains no business logic.
- **Verdict: adjusted.** Fowler's DTO is motivated by reducing remote round-trips and bridging a presentation/domain mismatch — **not** primarily as a security boundary. The rule's "never expose `_id`, `password`" rationale is sound but rests on least-exposure/security reasoning (cf. OWASP API3:2023 "Broken Object Property Level Authorization" / excessive data exposure), **not** on PoEAA. Cite DTO for the *mapping discipline*; do not borrow it as authority for the no-leak rationale.

### Async — parallelize independent calls; no synchronous I/O in request handlers
- `[Authoritative]` **MDN Web Docs — `Promise.all()`** (`developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/all`) and **`Promise.allSettled()`** (`.../Promise/allSettled`) — `all` fulfills only when every input promise fulfills (concurrent awaits) and rejects on the first rejection; `allSettled` waits for all to settle regardless of rejection (partial results).
- `[Authoritative]` **Node.js official docs — "Don't Block the Event Loop (or the Worker Pool)"** (`nodejs.org/en/learn/asynchronous-work/dont-block-the-event-loop`) — synchronous/blocking work in a handler (sync `fs`/`crypto` core APIs, large `JSON.parse`) starves the single event-loop thread.
- **Verdict: supported.** **Honesty note:** these are official runtime/vendor docs (Authoritative as such), with **no** canonical-book backing — and none is needed.

### Single source of truth for validation; centralize enums; no force-cast
- `[Authoritative]` **Hunt & Thomas, *The Pragmatic Programmer* (1999), "DRY — Don't Repeat Yourself"** — "every piece of knowledge must have a single, unambiguous, authoritative representation within a system"; redundant manual checks and scattered enum definitions violate it.
- `[Authoritative]` **J. Bloch, *Effective Java* 3e (2018), Item 49 "Check parameters for validity"** — validate once at the boundary rather than re-checking downstream.
- `[Semi-authoritative]` **A. King, "Parse, Don't Validate" (2019-11-05)** (`lexi-lambda.github.io/blog/2019/11/05/parse-don-t-validate/`) — parse at the boundary into a type that makes invalid states unrepresentable, then trust the type. Named-author essay; the principled backing for `.parse()`-and-trust.
- **Verdict: supported.** **Honesty note:** the Zod-specific tooling choice is internal convention; DRY + boundary-validation are the canonical roots. No canonical source mandates a validation library — cite DRY for the single-source principle, not as endorsement of Zod.

### Never share mutable state between requests
- `[Authoritative]` **J. Bloch, *Effective Java* 3e (2018), Item 17 "Minimize mutability"** and **Item 78 "Synchronize access to shared mutable data"** — shared mutable state across concurrent contexts is the canonical concurrency hazard.
- **Verdict: supported.** **Honesty note:** the per-request-scope directive (`Scope.REQUEST`, `@RequestScope`) is a web-framework convention with no canonical source specific to per-request isolation; its root is the general minimize/synchronize-shared-mutable-state guidance above. Frame as "following", not "per".


## Testing — mandatory tests, pyramid taxonomy, behavior-not-implementation, anti-patterns

Backs `quality/testing.md` (and its condensed mirror in `harness/AGENTS.md`). Researched jun-2026.

### Test taxonomy & the pyramid (unit base, E2E for critical flows only)
- `[Authoritative]` **Mike Cohn, *Succeeding with Agile: Software Development Using Scrum* (Addison-Wesley, 2009), ch. 16 "Quality", section "Automate at Different Levels" (Figure 16.1)** — origin of the Test Automation Pyramid (Unit / Service / UI), unit-heavy base, "a carefully curated set of UI tests" at the top. The print source of the model our three-type taxonomy descends from. (Locus corrected: the pyramid is in ch. 16, not ch. 15.)
- `[Authoritative]` **Ham Vocke, "The Practical Test Pyramid", martinfowler.com (2018-02-26)** — https://martinfowler.com/articles/practical-test-pyramid.html — "Push your tests as far down the test pyramid as you can"; UI/E2E tests "notoriously flaky" with "high maintenance cost"; unit tests "run very fast" ("thousands ... within a few minutes"). The canonical modern reframe; directly backs "E2E for critical user flows only" and the local-subset-vs-full-suite cost gradient.
- `[Authoritative]` **Winters, Manshreck & Wright, *Software Engineering at Google* (O'Reilly, 2020), ch. 11 "Testing Overview"** — https://abseil.io/resources/swe-book/html/ch11.html — classifies tests by *size* (small/medium/large) and *scope* rather than unit/integration "because the most important qualities we want from our test suite are speed and determinism, regardless of the scope of the test"; target mix ≈ **80% small / 15% medium / 5% large**.
- **Verdict: adjusted.** The unit-heavy, E2E-light shape our rule encodes is exactly the pyramid's prescription (*supported*). The *adjustment*: the more rigorous framing is size+scope, not the unit/integration/E2E names — "flaky" is a property of medium/large tests that cross process/machine boundaries, and the 80/15/5 proportion is the quantitative target our qualitative "unit always, E2E critical-only" gestures at. The type names stay (industry lingua franca); the nuance lives here.

### Test behavior, not implementation (anti-pattern: testing internal state)
- `[Authoritative]` **Winters/Manshreck/Wright, *SWE at Google* ch. 12 "Unit Testing"** — https://abseil.io/resources/swe-book/html/ch12.html — "write a test for each *behavior*" not each method (Test Behaviors, Not Methods); "make calls against its public API rather than its implementation details" (Test via Public APIs); a brittle test "fails in the face of an unrelated change to production code that does not introduce any real bugs" (Preventing Brittle Tests). Near-verbatim backing for our anti-pattern.
- `[Authoritative]` **Ham Vocke, "The Practical Test Pyramid" (2018)** — "Test for observable behaviour instead"; warns (paraphrased) against tests tied too closely to the production code ("Tests that are too close to the production code quickly become annoying").
- `[Authoritative]` **Gerard Meszaros, *xUnit Test Patterns: Refactoring Test Code* (Addison-Wesley, 2007), "Fragile Test" smell** — xunitpatterns.com/Fragile Test.html — tests that break on unrelated implementation changes (the Overspecified Software / behavior-sensitivity family). The canonical *name* for this anti-pattern. (Note: the xunitpatterns.com site is HTTP-only; the exact variant names were not re-fetched this pass — the smell name and Overspecified-Software association are established, the finer variant labels are cited from the book's known taxonomy, not a fresh quote.)
- **Verdict: supported** (strongly).

### Mock external dependencies, not internal logic
- `[Authoritative]` **Winters/Manshreck/Wright, *SWE at Google* ch. 13 "Test Doubles" (with ch. 12)** — https://abseil.io/resources/swe-book/html/ch13.html — "The most common reason for problematic interaction tests is an over reliance on mocking frameworks"; "A real implementation is preferred if it is fast, deterministic, and has simple dependencies".
- `[Authoritative]` **Meszaros, *xUnit Test Patterns* (2007), "Overspecified Software" / Test Double chapters** — over-verifying indirect calls couples the test to the implementation.
- **Verdict: adjusted.** Canon agrees the over-mocking instinct is wrong, but its discriminator is sharper than ours: **real-and-fast-and-deterministic vs. not** (SWE-at-Google's literal criterion is "fast, deterministic, and has simple dependencies"), not external vs. internal. SWE-at-Google would use a *real* internal collaborator when it is fast and deterministic, and would double a *slow/nondeterministic* dependency even if it is technically "internal". Our external/internal cut is a serviceable proxy that mostly coincides — recorded as an adjustment, not a rule edit, since refining the wording would not change agent behavior in the common case.

### Happy path AND error/edge/async-failure paths
- `[Authoritative]` **Vocke, "The Practical Test Pyramid" (2018)** — tests "should ensure that all your non-trivial code paths are tested (including happy path and edge cases)".
- `[Authoritative]` **SWE at Google ch. 11** — boxed section "Testing for Failure": testing "how a system handles failure" alongside behavioral correctness.
- **Verdict: supported.**

### Tests don't prove safety (green CI ≠ correct spec)
- `[Semi-authoritative]` **E. W. Dijkstra, "Notes on Structured Programming" (EWD249, 1970), §3 "On The Reliability of Mechanisms"** — "Program testing can be used to show the presence of bugs, but never to show their absence!" (stated as a corollary). The canonical statement of the limit our rule restates.
- `[Authoritative]` **SWE at Google ch. 11 — the "Beyoncé Rule"** ("If you liked it, then you shoulda put a test on it") — tests prevent *known* regressions; testing catches what you chose to guard, not absolute correctness.
- **Verdict: adjusted.** The first clause (tests verify code-matches-spec, not spec-is-correct) is Dijkstra-canonical. The extension to "assumptions hold at scale" and "implicit infrastructure constraints" is a sound practitioner generalization with **no single canonical locus** — *adjusted, not contradicted*. **Honesty note:** no one source phrases the full claim our way; we stand on Dijkstra for the core and label the rest reasoned extrapolation.

### Fix implementation to match spec; tests change only when the spec changes
- `[Authoritative]` **SWE at Google ch. 12** — "the ideal test is unchanging: after it's written, it never needs to change unless the requirements of the system under test change" (Strive for Unchanging Tests). Nearly our rule verbatim.
- `[Authoritative]` **Kent Beck, *Test-Driven Development: By Example* (Addison-Wesley, 2002), Part III "Patterns for Test-Driven Development" / Red-Green-Refactor** — the failing test drives the production code to satisfy it; the test/spec is authoritative over the implementation.
- **Verdict: supported.** **Honesty note:** no source states "fix the implementation, not the test" word-for-word; the principle is canonically backed, the phrasing is ours.

### Tests are part of implementation, owned by the developer
- `[Authoritative]` **Forsgren, Humble & Kim, *Accelerate* (IT Revolution, 2018), ch. 4 "Technical Practices" — Test Automation capability** — automated suites "maintained by developers (rather than QA or outsourced)" are positively correlated with software delivery performance; the QA/outsourced variant is *not* correlated. Evidence-based backing for "tests are part of the implementation, not a follow-up" and agent-routing's "implementation agents write tests for their own code".
- **Verdict: supported.**

### Test independence (no shared mutable state; serialize only shared state)
- `[Authoritative]` **Meszaros, *xUnit Test Patterns* (2007), "Erratic Test" → "Interacting Tests" / "Shared Fixture"** — xunitpatterns.com/Shared Fixture.html — "Interacting Tests are usually caused by tests using a Shared Fixture with one test depending ... on the outcome of another test"; a Shared Fixture "can lead to collisions between tests ... resulting in Erratic Tests since tests may depend on the outcomes of other tests". The canonical name for our "tests that depend on each other / share mutable state" anti-pattern.
- `[Authoritative]` **SWE at Google ch. 11** — small tests kept independent and deterministic; nondeterminism from shared/medium-scope state is the "flaky test" source.
- **Verdict: supported.**

### Push tests down the pyramid / avoid duplication across layers
- `[Authoritative]` **Vocke, "The Practical Test Pyramid" (2018), "Avoid Test Duplication"** — "Push your tests as far down the test pyramid as you can"; rule of thumb: "If a higher-level test spots an error and there's no lower-level test failing, you need to write a lower-level test." Backs the (currently implicit) principle that E2E covers critical flows only and the same behaviour is not re-asserted at every layer.
- **Verdict: supported** (additive — the rule implies this via "E2E for critical user flows only" but never states it; see additional improvement below).

### Honesty note — what has NO single canonical locus
The *proportionality* of `> Execution Scope` (affected subset locally, full suite + E2E at the merge boundary; amortize the in-vivo run over a change-group; `--runInBand` only for state-sharing suites) is **practitioner CI convention**, not a book chapter. It rests on the pyramid's speed gradient (fast unit / slow E2E, above) plus Accelerate ch. 4's continuous-integration "fail fast" capability — but the specific amortization and serialization heuristics are this config's own engineering judgment. Do not manufacture a standard for them; cite the speed gradient as the *principle*, and own the heuristics as convention.


## Security — secure-by-default coding checks

Backs `quality/security.md` (Input Validation, Injection Prevention, Authentication & Secrets, route deny-by-default, Error Handling, Supply Chain Security). Researched jun-2026. Live-verified against OWASP Top 10:2021, the OWASP Cheat Sheet Series, NIST SP 800-63B, the CWE corpus, and the OSV schema.

### SQL injection — parameterized queries / prepared statements / ORM
- `[Authoritative]` **OWASP Top 10:2021 A03 Injection** (`owasp.org/Top10/2021/A03_2021-Injection/`), CWE-89 — "How to Prevent": "The preferred option is to use a safe API, which avoids using the interpreter entirely, provides a parameterized interface, or migrates to Object Relational Mapping Tools (ORMs)." This is the citation that carries the ORM endorsement.
- `[Authoritative]` **OWASP SQL Injection Prevention Cheat Sheet** — heading **"Defense Option 1: Prepared Statements (with Parameterized Queries)"**. (Note: the cheat sheet mentions ORMs only under Defense Option 4, not as a primary defense — the ORM-as-equivalent claim rests on A03 above, not on this cheat sheet.)
- **Verdict: supported.**

### Command injection — avoid the shell, separate command from arguments (execFile, not exec)
- `[Authoritative]` **OWASP OS Command Injection Defense Cheat Sheet** — **Defense Option 1 "Avoid calling OS commands directly"**: "Built-in library functions are a very good alternative to OS Commands, as they cannot be manipulated to perform tasks other than those it is intended to do." **Defense Option 3 "Parameterization in conjunction with Input Validation"**: structured mechanisms that "automatically enforce the separation between data and command." The shell-avoidance is stated verbatim for **Runtime.exec** ("Runtime.exec does NOT try to invoke the shell at any point"); `ProcessBuilder` demonstrates the same command/argument separation in the examples. The `execFile`-not-`exec` guidance is the Node analogue of this principle.
- `[Authoritative]` **CWE-78** "Improper Neutralization of Special Elements used in an OS Command ('OS Command Injection')", mapped to A03:2021.
- **Verdict: supported.**

### SSRF — allow-list outbound destinations, never a deny-list
- `[Authoritative]` **OWASP SSRF Prevention Cheat Sheet** — Case 1: build an allow-list of permitted IPs/domains; **"Deny-lists are bypass-prone. Prefer allow-lists."**
- `[Authoritative]` **OWASP Top 10:2021 A10 Server-Side Request Forgery (SSRF)** (`owasp.org/Top10/2021/A10_2021-Server-Side_Request_Forgery_(SSRF)/`), CWE-918 — "Do not mitigate SSRF via the use of a deny list or regular expression. Attackers have payload lists, tools, and skills to bypass deny lists."
- **Verdict: supported.**

### Input validation + XSS sanitization
- `[Authoritative]` **OWASP Top 10:2021 A03 Injection** — "Use positive server-side input validation. This is not a complete defense as many applications require special characters, such as text areas or APIs for mobile applications." Validation is a *secondary* control; the *primary* injection defense is a safe/parameterized API.
- `[Authoritative]` **OWASP Cross Site Scripting Prevention Cheat Sheet** — contextual **output encoding** is the core XSS defense, *not* input validation; framework auto-escaping is the baseline; "OWASP recommends DOMPurify for HTML Sanitization" for the raw-HTML/WYSIWYG case; "Since no single technique will solve XSS, using the right combination of defensive techniques will be necessary."
- **Verdict: adjusted.** The rule's practices (schema validation via Zod/class-validator/Pydantic, framework auto-escaping, DOMPurify) are exactly OWASP-endorsed. The caveat: input validation is a *complement*, never a *replacement* for parameterized queries (injection) and contextual output encoding (XSS). The section heading "Input Validation" should not be read as making validation the primary defense for either. Additive framing only — no behavior change.

### Password storage — hash, never plaintext
- `[Authoritative]` **OWASP Password Storage Cheat Sheet** — **Argon2id first choice** (minimum 19 MiB memory, iteration count 2, parallelism 1); scrypt second; "The bcrypt password hashing function **should only** be used for password storage in legacy systems where Argon2 and scrypt are not available."
- `[Authoritative]` **NIST SP 800-63B §5.1.1.2** — "Verifiers SHALL store memorized secrets in a form that is resistant to offline attacks. Memorized secrets SHALL be salted and hashed using a suitable one-way key derivation function." (salt SHALL be ≥32 bits; PBKDF2 and Balloon among the named KDFs).
- **Verdict: adjusted.** "Never compare plaintext" is fully supported (NIST SHALL). But the rule lists "bcrypt or argon2" as equals; canon ranks **Argon2id preferred, bcrypt legacy-fallback**. Suggested additive edit: "Hash passwords with Argon2id (preferred) or bcrypt for legacy systems where Argon2/scrypt are unavailable. Never compare plaintext." Does not invalidate existing bcrypt code — additive, not a reversal.

### Route auth — deny by default
- `[Authoritative]` **OWASP Top 10:2021 A01 Broken Access Control** (`owasp.org/Top10/2021/A01_2021-Broken_Access_Control/`) — "How to Prevent" leads with **"Except for public resources, deny by default."** CWE-862 (Missing Authorization), CWE-285 (Improper Authorization). The rule's "default every route to authenticated, public endpoints limited to a closed pattern list" is a direct operationalization.
- **Verdict: supported.**

### Error handling & security logging
- `[Authoritative]` **OWASP Top 10:2021 A01** — explicit prevention bullet: "Log access control failures, alert admins when appropriate." Plus **A09 Security Logging and Monitoring Failures**: "Auditable events, such as logins, failed logins, and high-value transactions, are not logged." (A09's mapped CWEs are 117, 223, 532, 778 — *not* 209.)
- `[Authoritative]` **CWE-209** "Generation of Error Message Containing Sensitive Information" — the precise standalone locus for "errors must not leak stack traces, schemas, internal paths" (it covers database query structures, file-system paths, configuration details). Note: CWE-209 is **not** part of A09; cite it on its own, not under A09.
- **Verdict: supported.**

### Secrets hygiene — never store recoverable secrets in code/docs
- `[Authoritative]` **CWE-798** "Use of Hard-coded Credentials" — credentials embedded in code or config files; covers inbound and outbound hard-coded auth.
- `[Authoritative]` **The Twelve-Factor App, Factor III (Config)** (`12factor.net/config`) — "strict separation of config from code"; "The twelve-factor app stores config in environment variables." NIST SP 800-63B §5.1.1.2 (above) backs the no-recoverable-storage principle.
- **Honesty note:** the semi-obfuscation masking format (`AKIA****TBQX`, partial IPs/SGs) is a sensible practitioner convention — no standard prescribes that exact masking scheme. The *principle* it serves (no recoverable secrets in durable files) is fully canonical (CWE-798, 12-Factor III); the format is house style, not borrowed authority.

### Supply-chain CVE gate — OSV.dev
- `[Authoritative]` **OSV.dev API** (`google.github.io/osv.dev/post-v1-query/`) — `POST https://api.osv.dev/v1/query` with `{package:{name,ecosystem}, version}`; requests are case-sensitive ("use 'PyPI' instead of 'pypi'").
- `[Authoritative]` **OSV Schema** (`ossf.github.io/osv-schema/`) — canonical ecosystem identifiers confirm the rule's mapping exactly, including the non-obvious **SwiftURL** (Swift; `name` is a Git URL) and **Hex** (Erlang/Elixir); also npm, PyPI, Maven, Go, crates.io, NuGet, RubyGems, Pub.
- **Honesty note — refuse borrowed authority:** the **SLSA framework** (a seed candidate) governs *build provenance / supply-chain integrity*, a different axis from *known-CVE scanning*. This rule does CVE scanning (OSV), not provenance attestation — do **not** cite SLSA here. The CRITICAL/HIGH blocking threshold is a project policy choice (OSV reports advisories; it does not mandate a block severity), reasonable and uncontradicted by any source.
- **Verdict: supported.**


## HTML-vs-Markdown output routing (substantial multi-modal vs short/single-purpose)

Backs `quality/communication-format.md` (canonical source of the `flow-report` auto-invoke trigger) and its condensed mirror in `harness/AGENTS.md`. Researched jun-2026.

**Scope honesty up front:** this rule is *predominantly an internal tooling-routing policy*. The operative content — the thresholds (>300 words, >2 sections, 2+ info kinds), the carve-out list, the skill names (`flow-report` / `playground` / `frontend-design`), the output dirs, and the 2-4× token-cost figure — has **no external canonical backing and needs none**; it is internally coherent harness wiring. Only the *rationale axis* underneath two choices touches real canon, and even there the application is analogical. Grounding value: **low**.

### Short/single-purpose → Markdown (users scan, don't read)
- `[Semi-authoritative]` **Nielsen, *How Users Read on the Web* (NN/g, Sept 30, 1997)** — "79 percent of our test users always scanned… only 16 percent read word-by-word"; six scannability techniques (highlighted keywords, sub-headings, bulleted lists, one idea per paragraph, inverted pyramid, halved word count). https://www.nngroup.com/articles/how-users-read-on-the-web/
- `[Semi-authoritative]` **Nielsen, *How Little Do Users Read?* (NN/g, May 5, 2008)** — "users have time to read at most 28% of the words during an average visit; 20% is more likely" (45,237 cleaned page views). https://www.nngroup.com/articles/how-little-do-users-read/
- `[Semi-authoritative]` **Nielsen, *F-Shaped Pattern For Reading Web Content* (NN/g, Apr 16, 2006; 232-participant eyetracking)** — users scan in an F-shape; exhaustive word-by-word reading is rare. https://www.nngroup.com/articles/f-shaped-pattern-reading-web-content-discovered/
- **Locus correction:** the "~20-28% of words read" figure belongs to *How Little Do Users Read?* (2008), **not** to the *F-Shaped Pattern* article (2006), which establishes the F-scan shape but states no such percentage. An earlier draft conflated the two — they are distinct NN/g sources.
- **Verdict: adjusted.** This research backs *"structure and trim long content for scanning"* — but its prescriptions (sub-headings, bullets, one idea per paragraph) are equally satisfiable in Markdown, so it does **not** by itself justify the HTML-vs-Markdown split. NN/g is rigorous practitioner research, not a standards body → `[Semi-authoritative]`, never `[Authoritative]`.

### Multi-modal content → richer presentation (dual-coding / multivariate density)
- `[Authoritative]` **Mayer, *Multimedia Learning* (Cambridge, 2nd ed. 2009) — the Multimedia Principle**: "people learn more deeply from words and pictures than from words alone", on dual-channel (visual/verbal) processing.
- `[Authoritative]` **Paivio, Dual-Coding Theory — *Imagery and Verbal Processes* (1971); *Mental Representations: A Dual Coding Approach* (Oxford, 1986)** — visual and verbal information encoded in separate, complementary channels (Mayer's theoretical base).
- `[Authoritative]` **Tufte, *The Visual Display of Quantitative Information* (1983; 2nd ed. 2001), ch. 1 "Graphical Excellence"** — graphical excellence "gives to the viewer the greatest number of ideas in the shortest time with the least ink in the smallest space"; data graphics are, in Tufte's framing, nearly always multivariate.
- **Verdict: adjusted (analogical extension).** Mayer/Paivio govern *learning materials*; Tufte governs *statistical graphics* — none addresses "an agent should emit HTML rather than Markdown". Both formats can embed prose+table+code; flow-report's real edge is *tooling* (single-file shareability, embedded SVG, syntax highlighting), not a multimedia-cognition mandate. Phrase any citation as "following the multimedia/dual-coding principle", **never "per"**. Refuse borrowed authority: do not present these learning/graphics sources as proof that HTML beats Markdown.

### The 300-word threshold, the 2-4× cost figure, the skill routing
- `[Non-authoritative]` **Internal policy of this repo** (global/rules + the flow-* skill pack).
- **Honesty note:** no authoritative source sets a 300-word HTML cutoff or a 2-4× token multiplier; these are arbitrary-by-design internal thresholds (the 2-4× figure legitimately functions as the decision threshold, which the repo's own carve-out permits). Do **not** cite Nielsen's "~110 words" / "halve the word count" figures as the basis for the 300-word number — those concern web-copy length, not a document-format switch point (borrowed authority). This entry exists chiefly as a *negative record*: the routing thresholds are uncanonical, so a future grounding pass should not re-hunt for a non-existent "HTML-vs-Markdown standard".


## Git workflow — commit format, branching model & integration

Backs `workflow/git-workflow.md` (Conventions, Branches, Pull Requests, Environment Promotion) and its condensed mirror in `harness/AGENTS.md`. Researched jun-2026. Scope note: only the commit-message format, the branching model, and the PR/integration conventions rest on external authority. The mode mechanics (Direct/Isolated, once-per-session AskUserQuestion decisions, checkpoint cadence, 404 gh-auth recovery) are internal harness policy with **no external canon** — do not retrofit citations onto them.

### Conventional commit prefixes & scope syntax
- `[Authoritative]` **Conventional Commits 1.0.0** — https://www.conventionalcommits.org/en/v1.0.0/ — structure `<type>[optional scope]: <description>`; scope rule (verbatim): "A scope MAY be provided after a type. A scope MUST consist of a noun describing a section of the codebase surrounded by parenthesis, e.g., `fix(parser):`". Dovetails with SemVer (`fix`→PATCH, `feat`→MINOR, `BREAKING CHANGE`→MAJOR).
- **Verdict: adjusted.** The spec MANDATES only `feat` and `fix`; the other prefixes our rule lists (`refactor`, `chore`, `docs`, `test`, `perf`, `ci`) are optional — verbatim FAQ: "Additional types are not mandated by the Conventional Commits specification, and have no implicit effect in Semantic Versioning." The spec's FAQ attributes its recommended extra-type set to **@commitlint/config-conventional** (not, as one might assume, to "the Angular convention" — the spec does not name Angular). Our fixed 8-prefix list is a legitimate house-style narrowing of an open set — cite it as *following* the spec, not as *being* the spec.

### Branching model — trunk-style / short-lived branches (NOT git-flow)
- `[Authoritative]` **Forsgren, Humble & Kim, *Accelerate* (2018), Ch. 4 "Technical Practices"** — trunk-based development capability: "fewer than three active branches," branches/forks with very short lifetimes (< 1 day), as a statistical predictor of elite delivery performance; long-lived feature branches correlate negatively. (Corroborated by DORA / IT Revolution's "24 Key Capabilities".)
- `[Semi-authoritative]` **trunkbaseddevelopment.com > Short-Lived Feature Branches** (Hammant et al.) — "the branch should only last a couple of days... Any longer than two days... risk of becoming a long-lived feature branch"; "the developer count should stay at one (or two if pair-programming)."
- `[Semi-authoritative]` **Driessen, "A successful Git branching model"** — https://nvie.com/posts/a-successful-git-branching-model/ — 2020 reflection note: git-flow suits *versioned, multi-version* software; for continuous delivery "I would suggest to adopt a much simpler workflow (like GitHub flow) instead of trying to shoehorn git-flow into your team."
- **Verdict: adjusted.** Our two options (trunk-style multi-commit on the current branch, or **one** short-lived branch per change-group) are GitHub-flow-shaped and align with Accelerate + Driessen's own 2020 steer away from git-flow. git-flow is the model the rule **correctly declines**, not one it follows. **Caveat — our rule is stricter than the canon:** "one open thematic branch at a time / themes don't run in parallel" is more conservative than Accelerate ("fewer than three active branches" tolerates 2 concurrent) and trunkbaseddevelopment.com (whose diagram illustrates two developers on concurrent features X and Y). The canon's operative variable is *integration frequency* (integrate within a day to shrink merges), not theme serialization. The serialization is a deliberate solo-agent house simplification for linear history — defensible, but house policy *beyond* the canon, not derived from it.

### PR via integration branch & CI ownership
- `[Authoritative]` **Fowler, "Patterns for Managing Source Code Branches"** — https://martinfowler.com/articles/branching-patterns.html — the *Integration Friction*, *Mainline Integration*, and *Pre-Integration Review* patterns. PRs combine Feature Branching with pre-integration review; "Pre-Integration Reviews always introduces some latency into the integration process, encouraging a lower integration frequency." (The "add overhead to cope with low-trust situations" line is Kief Morris, quoted by Fowler — not Fowler's own words.)
- `[Authoritative]` **Fowler, "Continuous Integration"** — https://martinfowler.com/articles/continuousIntegration.html — "nobody has a higher priority task than fixing the build"; high-frequency integration surfaces conflicts early (within hours, not weeks) and keeps merges small.
- **Verdict: adjusted.** Canon does NOT bless an unconditional PR mandate — PRs trade latency for review and are warranted by *context* (low trust, a CI/review gate), not by default. Our rule already encodes exactly this judgement: the PR requirement is scoped to "repos with a CI/review gate," with a "solo repo with no PR/CI flow may merge locally" carve-out. The "own the CI until the branch is green" clause is well-supported by Fowler's mainline discipline. Additive — the rule already matches the canon's caveat; resist any future edit making PRs unconditional.

### Small, focused commits (one concern per commit)
- `[Authoritative]` **Fowler, "Patterns for Managing Source Code Branches"** — small, cohesive change units integrate with less merge friction and lower semantic-conflict risk (corroboration only).
- **Honesty note:** no single canonical standard states "one concern per commit" verbatim — it is a well-corroborated practitioner convention (atomic commits) whose principled root is the same change-cohesion reasoning the bibliography already cites under *Grouping a cohesive deliverable* (Clean Architecture, Common Closure Principle, ch. 13). Treat Fowler as corroboration, not as the literal source for the phrasing. **Verdict: supported** (by convergence, not by a named standard).

### Safety gates — force-push & history rewrites
- `[Authoritative]` **git-scm docs** — `git-push(1)`, `--force` / `--force-with-lease` description (force overwrites remote history); `git-rebase(1)`, section "RECOVERING FROM UPSTREAM REBASE" (rewriting *published* history disrupts collaborators). Backs the "never force-push / never rewrite published history without confirmation" gates.
- **Honesty note:** the *confirmation-UX design* (AskUserQuestion blocks, three-step explicit-task pattern, Direct/Isolated modes, checkpoint cadence, gh-auth 404 recovery) is internal harness policy — git docs justify the *danger*, not the *interaction design*. Do not borrow authority from git or any standard for the session-mechanics layer. **Verdict: supported** (for the danger rationale only).


## DevOps deployment principles

Backs `workflow/devops-principles.md` (IaC over manual, rollback/recovery path, fail-fast CI ordering, minimal multi-stage images, liveness/readiness health checks) and its condensed mirror in `harness/AGENTS.md`. Researched jun-2026.

### Infrastructure as Code over manual changes (drift)
- `[Authoritative]` **Kief Morris, *Infrastructure as Code*, 2nd ed. (O'Reilly, 2020), Ch. 1 "What Is Infrastructure as Code?"** — three core practices ("define everything as code"; "continuously validate all your work in progress"; "build small, simple pieces that you can change independently") and the foundational principles: systems must be *reproducible, disposable* ("cattle, not pets"), and *consistent* — the principled case against snowflake servers and configuration drift.
- `[Authoritative]` **Forsgren, Humble & Kim, *Accelerate* (IT Revolution, 2018), Ch. 4 "Technical Practices"** — version control of *all* production artifacts (app code, system + app config, scripts) is a statistically validated CD capability.
- `[Semi-authoritative]` **Martin Fowler, bliki "InfrastructureAsCode" (martinfowler.com/bliki/InfrastructureAsCode.html)** — version-controlled executable infra defeats the "snowflake server" problem; "manual provisioning ... lead[s] to snowflakes with subtly different configurations."
- **Verdict: supported.** Canon directly backs "console/CLI provisioning drifts; codify it." "Manual only for one-off experiments" is the practitioner corollary, consistent with Morris's reproducibility principle.

### Every deployment needs a rollback path
- `[Authoritative]` **Forsgren, Humble & Kim, *Accelerate* (2018), Ch. 2 "Measuring Performance"** — Time to Restore Service (MTTR) and Change Failure Rate are two of the four key metrics; a recovery capability after a bad change is a *defining* property of high performers, not optional.
- `[Authoritative]` **Humble & Farley, *Continuous Delivery* (Addison-Wesley, 2010), Ch. 10 "Deploying and Releasing Applications" (blue-green) & Ch. 12 "Managing Data" (rollback / reverse migrations)** — canonical origin of blue-green's instant rollback, while explicitly flagging that database changes constrain pure rollback.
- `[Semi-authoritative]` **Redgate Flyway product-learning, "Database Updates: Roll Back or Fix Forward?"; Atlas blog, "The Hard Truth about GitOps and Database Rollbacks" (2024-11-14)** — irreversible migrations (DROP COLUMN/TABLE) make literal rollback lossy; modern CD accepts *fix-forward* (feature flags, a corrective release) as a legitimate, often preferred recovery path. Adjustment corroboration only — the authoritative weight is carried by Accelerate + Continuous Delivery above.
- **Verdict: adjusted.** Canon agrees recovery capability is mandatory, but a *literal rollback* is not always achievable — an applied irreversible migration has no clean down-path. Reframe "rollback path" → "recovery path (rollback **or** fix-forward)". Additive: the gate (no recovery story ⇒ not ready) is preserved; the edit just stops the rule from implying rollback is always the mechanism. (Proposed rule edit surfaced separately, not auto-applied.)

### CI pipelines fail fast — staged ordering
- `[Authoritative]` **Humble & Farley, *Continuous Delivery* (2010), Ch. 5 "Anatomy of the Deployment Pipeline"** — the commit stage runs first and fastest for fast feedback; the principle that a failing pipeline stage stops the line; later stages (acceptance → capacity/non-functional → manual) are progressively slower and broader so confidence grows per passing stage. (The commit stage's detailed treatment is Ch. 7 "The Commit Stage".)
- `[Authoritative]` **Forsgren, Humble & Kim, *Accelerate* (2018), Ch. 4 "Technical Practices"** — CI and a fast, reliable build/test loop are measured CD capabilities; fast feedback is the design goal.
- **Verdict: supported (with a heuristic caveat).** The exact sequence (lint → typecheck → build → test → deploy) is a reasonable instantiation of "order by feedback speed and likelihood of failure"; canon orders by speed + confidence, not a fixed verb list, so treat the sequence as heuristic (lint/typecheck may parallelize). "E2E/security scans last or in parallel" is already correctly hedged.

### Minimal multi-stage Docker images, pinned bases
- `[Authoritative]` **Docker official docs, "Multi-stage builds" (docs.docker.com/build/building/multi-stage/)** — "The end result is a tiny production image with nothing but the binary inside. None of the build tools required to build the application are included in the resulting image." Verbatim support for "final image excludes build tools, dev deps, source maps".
- `[Authoritative]` **Docker official docs, "Building best practices" (docs.docker.com/build/building/best-practices/)** — pin base images by version tag (e.g. `FROM alpine:3.21`) or, for stronger supply-chain integrity, by digest (`@sha256:…`); avoid the floating `latest` for reproducible builds.
- **Verdict: supported strongly.** Both quotes/claims confirmed against the live official docs.

### Liveness & readiness health checks for long-running services
- `[Authoritative]` **Kubernetes official docs, "Liveness, Readiness and Startup Probes" (kubernetes.io/docs/concepts/workloads/pods/probes/)** — "Liveness probes determine when to restart a container" (kubelet restarts on failure); "Readiness probes determine when a container is ready to accept traffic" (failure makes the EndpointSlice controller remove the Pod's IP from matching Services' EndpointSlices — traffic stops, container is *not* restarted). The exact `liveness`/`readiness` vocabulary the rule uses originates here.
- `[Authoritative]` **Beyer, Jones, Petoff & Murphy (eds.), *Site Reliability Engineering* (O'Reilly/Google, 2016), Ch. 20 "Load Balancing in the Datacenter" & Ch. 6 "Monitoring Distributed Systems"** — backends are health-checked before receiving traffic; newly started instances are expected unhealthy until ready. Corroboration only — SRE uses Borg/Borgmon vocabulary, not "probe", so anchor the terminology in the Kubernetes docs, not the SRE book (avoid borrowed authority).
- **Verdict: supported.** The exemption list (static sites, serverless, short-lived workers — "no place to hook a probe") is sound practitioner reasoning: those platforms expose no long-lived process for the orchestrator to probe.

### Secrets in deployment configs (.env.example, env vars)
- `[Authoritative]` **Twelve-Factor App, Factor III "Config — Store config in the environment" (12factor.net/config)** — config varying between deploys (credentials, resource handles) lives in env vars, strictly separated from code; "the codebase could be made open source at any moment, without compromising any credentials."
- **Honesty note:** the specific `.env.example`-not-`.env` template convention has **no** canonical source — it is a sensible practitioner convention on top of Factor III. Factor III mandates env-var config; it does not specify a `.example` file. This bullet otherwise delegates to `security.md > Authentication & Secrets` (grounded separately). Factor V (build/release/run, immutable releases) is already cited in this file under "Generated-artifact naming … > Retention" — cross-reference rather than re-citing.


## Cross-service contract-first workflow

Backs `workflow/cross-service-workflow.md` (architect-first / spec-before-implementation, contract-change trigger gating, exact-spec consumption). Researched jun-2026.

### Architect-first — design/locate the contract before implementing across a boundary
- `[Authoritative]` **Eric Evans, *Domain-Driven Design* (2003), ch. 14 "Maintaining Model Integrity"** — Bounded Context, Context Map (p. 344: "describe the points of contact between the models, outlining explicit translation for any communication and highlighting any sharing"), Open Host Service, Published Language. The strategic-design root of "publish the explicit contract at the boundary before integrating across it." *Supports.*
- `[Authoritative]` **Sam Newman, *Building Microservices*, 2nd ed. (2021), ch. 5 "Implementing Microservice Communication"** — "make interfaces explicit"; use explicit schemas to catch structural breakages in the contract. *Adjusts:* Newman backs explicit schemas + backward compatibility but cautions against heavyweight up-front contract ceremony for low-coupling internal change — which is precisely why the rule's "mechanical multi-repo changes need no pass" carve-out is correct, not an exception to the canon.
- `[Semi-authoritative]` **Ian Robinson, "Consumer-Driven Contracts: A Service Evolution Pattern", martinfowler.com (12 Jun 2006)** — "When a provider accepts and adopts the reasonable expectations expressed by a consumer, it enters into a consumer contract." *Adjusts:* the contract is consumer-DRIVEN (derived from consumer expectations), the inverse of a provider unilaterally publishing it. The rule's provider-first spec framing should acknowledge that consumer (frontend) expectations feed the contract.
- `[Authoritative]` **OpenAPI Initiative (Linux Foundation), *OpenAPI Specification* 3.1** — vendor-neutral interface description for HTTP/REST APIs; the canonical artifact for the API contract (openapis.org; OAI under the Linux Foundation, TSC-governed). *Supports* the choice of OpenAPI as the contract format — for synchronous HTTP APIs only (see honesty note below).
- **Verdict: adjusted.** Canon supports contract-before-integration. Two reframes: (1) **analogical scope** — Evans/Newman govern inter-bounded-context (inter-service) boundaries; a frontend and its own backend are often ONE bounded context, so applying the discipline to the frontend-backend seam is a legitimate extension — phrase as "following", not "per". (2) **consumer-driven** — per Robinson the contract emerges from consumer expectations, not solely provider design.

### Trigger gated on shared-contract change, not multi-repo touch
- `[Authoritative]` **Newman, *Building Microservices*, 2nd ed., ch. 4 "Microservice Communication Styles" + ch. 5** — the coordination unit is the explicit interface/contract at the boundary; coupling warrants coordination, not code colocation.
- `[Authoritative]` **Evans, *DDD*, ch. 14 (Context Map)** — coordination is scoped to the "points of contact between the models," i.e. shared/published contracts, not every co-changed file.
- **Verdict: supported.** "Touching multiple repos is NOT itself a trigger — the contract change is" aligns directly with the canon; the rename-env-var / parallel-bug-fix exemptions match Newman's caution against over-coordinating low-coupling change.

### Implement against the spec exactly; update the contract before working around it
- `[Semi-authoritative]` **Martin Fowler, bliki "ContractTest"** (renamed from "IntegrationContractTest") — a contract test verifies the call conforms to the agreed contract; the consumer codes against the contract, not an ad-hoc shape (paraphrase; the bliki frames this via test doubles and Consumer-Driven Contracts, not this exact sentence). https://martinfowler.com/bliki/ContractTest.html
- `[Authoritative]` **Newman, *Building Microservices*, 2nd ed., ch. 9 "Testing" — Contract Tests and Consumer-Driven Contracts (CDCs)** — both sides hold the same contract as source of truth; deviation surfaces as a contract break.
- `[Authoritative]` **OpenAPI Specification 3.1** — the schema (paths/operations/components) IS the authoritative interface description; implementations validate against it.
- **Verdict: supported.** "Implement against the OpenAPI schemas exactly, don't invent shapes, update the spec before a workaround" is textbook contract-test discipline; canon agrees without caveat.

### Honesty note — event/webhook contracts have a different artifact
- The rule names **OpenAPI** as the contract format but its triggers also fire on "event, message, or webhook contract." **OpenAPI does not describe async/event contracts** (confirmed at openapis.org: scope is HTTP/REST). The event-contract counterparts are **AsyncAPI** (channels/operations/message schemas for event-driven APIs) and **CloudEvents** (CNCF, graduated Jan 2024 — a common envelope for event data). Do not stretch OpenAPI to cover the event case; no single standard unifies sync + async contract description. (Newman 2e ch. 5 covers event-driven/async communication; AsyncAPI/CloudEvents are the de-facto artifacts there.)


## Gap resolution & divergence between sources

Backs `workflow/gap-resolution.md` (both sections: *Gap Resolution in Plans* and *Divergence Between Sources*) and its condensed mirror in `harness/AGENTS.md`. Researched jun-2026.

### Gaps are invisible — hunt them before the first task, never silently discard
- `[Authoritative]` **Wiegers & Beatty, *Software Requirements*, 3rd ed. (Microsoft Press, 2013), ch. 7 "Requirements Elicitation"** (validation backstop: ch. 17 "Validating the Requirements") — Wiegers' well-known formulation is that the hardest requirements errors to detect are the requirements that aren't there: they're *invisible*. Missing requirements are among the most common and hardest-to-detect defects, mitigated by decomposition, hierarchical organization, and multiple representations. (Attributed paraphrase, not a chapter-pinned verbatim quote — the concept is firmly Wiegers; the invisible/missing-requirements material is framed by his own derivative writing around elicitation and analysis rather than validation, so it is cited to ch. 7 first.) Directly backs the principle "a gap that does not become a task… is a defect" and the active glob/grep/read investigation step.
- `[Authoritative]` **Boehm, *Software Engineering Economics* (Prentice-Hall, 1981)** — the cost-of-change curve (TRW/IBM cost-to-fix-by-phase data): a defect fixed during requirements definition costs near-nothing vs. orders of magnitude more downstream. Grounds the *timing* — investigate "before writing the first plan task," not after implementation. (No single chapter pinned; the cost-to-fix-by-phase result is the book's signature thesis.)
- **Verdict: supported.** Canon strongly endorses planning-time gap detection as the cheapest place to catch omissions.

### Never assume presence or absence — prove it
- `[Authoritative]` **Hunt & Thomas, *The Pragmatic Programmer* (20th Anniversary ed., 2019), Tip #34 "Don't Assume It—Prove It," p.96** — "Prove your assumptions in the actual environment—with real data and boundary conditions." Backs Step 1's "Verify with `which`/`--version`/a presence check — never assume absence, and never assume presence." (Tip #, page, and quote verified against the official 20th-Anniversary tips list; this is Tip #27 in the 1999 first edition.)
- **Verdict: supported.**

### Conflicting sources are a defect to resolve, not a precedence to hardcode
- `[Authoritative]` **ISO/IEC/IEEE 29148:2018, clause 5.2.6 "Characteristics of a set of requirements"** — a set of requirements must be **Consistent**: "the set of requirements does not have individual requirements which are contradictory." Inconsistency is a quality defect of the *set*, not a tie to break by a standing precedence rule. (Clause number and title verified directly from the standard's ToC — 5.2.5 = characteristics of individual requirements, 5.2.6 = characteristics of a set; the "Consistent = no contradictory requirements" definition is the standard's own wording, carried from the 2011 edition.) This is the genuine canonical backing for "Divergence Between Sources" — the pattern is not merely practitioner convention.
- `[Authoritative]` **Hunt & Thomas, *The Pragmatic Programmer*, 20th Anniversary ed., Tip #15 "DRY—Don't Repeat Yourself," p.31** — "Every piece of knowledge must have a single, unambiguous, authoritative representation within a system." **Adjacent, not exact:** DRY is about *eliminating* duplicate representations; our rule handles two co-existing artifacts that legitimately both claim authority. Cite as the nearest principle, not as a direct mandate. (Tip #, page, and quote verified; DRY is Tip #11 in the 1999 first edition.)
- **Verdict: supported.** The "surface, then resolve — never follow source X by default" stance matches ISO 29148's treatment of consistency as a set-level property to repair.

### Who arbitrates a conflict — the carve-out boundary
- `[Authoritative]` **Wiegers & Beatty, *Software Requirements*, 3rd ed., ch. 6 "Finding the Voice of the User"** — resolve conflicting requirements by involving all stakeholders and deferring to *senior project stakeholders*; explicitly: "Don't fall into the trap of having the BA of the developers make the decisions about requirement conflict resolution." Authority to arbitrate sits with stakeholders, not the implementer. (Chapter, title, and quote verified.)
- **Verdict: adjusted.** The rule's step-4 carve-out lets the agent self-resolve a "clearly stale" source. Canon agrees only because the rule already fences the carve-out to reversible UI-tweak-class changes and excludes contracts/migrations/security controls — i.e. decisions that are the implementer's to make. Where the conflict is contract- or stakeholder-owned, Wiegers is explicit: escalate, do not self-arbitrate. The carve-out preamble could make this "whose decision is it" boundary explicit (additive; does not change which cases currently self-resolve).

### Honesty notes
- The numeric thresholds (inline gap ≤3 tasks vs. prerequisite gap >3 tasks; "more than one inline task" triggering the confirm round-trip) are **practitioner conventions with no canonical basis** — no standard prescribes a task count for when a gap becomes its own plan. Wiegers' ch. 7 caution against over-exhaustive elicitation ("analysis paralysis") corroborates the *spirit* of bounding gap-hunting, but not the specific numbers. Do not dress these numbers in borrowed authority.
- Do **not** cite ISO 29148 for the *Gap Resolution* section's task-classification mechanics — it governs requirement quality characteristics, not plan-task decomposition (borrowed authority). Its on-point use is the "Consistent" characteristic backing *Divergence Between Sources* only.
- The "invisible requirements" sentence is an attributed paraphrase of Wiegers, not a chapter-pinned verbatim quote — the exact wording could not be tied to a single chapter from a primary source, so it is not presented as a quotation with a locus.


## Infrastructure naming convention

Backs `workflow/infra-naming.md` (and its condensed mirror in `harness/AGENTS.md`). Researched jun-2026.

### kebab-case / LDH for infra resource names
- `[Authoritative]` **RFC 1035 (Mockapetris, 1987), §2.3.1 "Preferred name syntax"** — label BNF `<label> ::= <letter> [ [ <ldh-str> ] <let-dig> ]`; interior characters only letters, digits, and hyphen (the "LDH" rule); "No significance is attached to the case." §2.3.4: label ≤ 63 octets, name ≤ 255 octets. *(Quotes verified verbatim.)*
- `[Authoritative]` **RFC 1123 (Braden, ed., 1989), §2.1 "Host Names and Numbers"** — relaxes RFC 952: "the restriction on the first character is relaxed to allow either a letter or a digit." *(Verbatim.)* Together these are the canonical source for "lowercase, hyphen-separated, no underscores."
- `[Authoritative]` **Azure Cloud Adoption Framework, "Define your naming convention"** (learn.microsoft.com/azure/cloud-adoption-framework/ready/azure-best-practices/resource-naming, updated 2025-09) — "use a hyphen `-` to separate naming components"; flags that storage accounts, ACR, and Data Lake forbid the delimiter and must concatenate (`stnavigatordata001`, `crnavigatorprod001`). *(Verified.)*
- **Verdict: adjusted.** The LDH/kebab-case directive is hard physics only for **DNS-derived** names (subdomains, S3 buckets, ALBs). For free-form names and tags the constraint is per-service, which is why the rule carries its own snake_case and concatenated-resource exceptions. Phrase kebab-case as the default-with-documented-exceptions it already is.

### Subdomain env prefix as a DNS label
- `[Authoritative]` **RFC 1035 §2.3.1 / §2.3.4** — a leading `qa-`/`dev-` is a valid LDH label in the standard separator position; production DNS carrying no token is just the bare apex/host label. *(Verified.)*
- **Verdict: supported.** The subdomain exception (short env prefix) is exactly what the DNS label grammar sanctions.

### snake_case for SQL identifiers
- `[Authoritative]` **PostgreSQL 18 docs §4.1.1 "Identifiers and Key Words"** (postgresql.org/docs/current/sql-syntax-lexical.html) — unquoted identifiers begin with a letter/underscore; subsequent characters are letters, underscores, digits, or `$`. **Hyphens are not permitted**, so a hyphenated name must be double-quoted; quoting also makes it case-sensitive, whereas "unquoted names are always folded to lower case." *(Quotes verified verbatim.)*
- `[Authoritative]` **ISO/IEC 9075-2:2016 (SQL:2016, Foundation), §5.2 lexical rules** — regular (unquoted) identifiers are case-folded; delimited (double-quoted) identifiers are literal/case-sensitive. *(Standard is paywalled — locus and exact fold-direction not independently verified; corroborated by the PostgreSQL docs above, which note the standard folds to upper while Postgres folds to lower. Cited as governing standard, not a literal quote.)*
- **Verdict: supported.** "Hyphens force quoted identifiers — speak the engine's language" is precisely correct; snake_case also dodges the fold divergence (Postgres→lower, standard/Oracle→upper).

### env-token-LAST and never-abbreviated
- `[Authoritative]` **Azure Cloud Adoption Framework, "Define your naming convention"** — leads names with the resource-**type** abbreviation (`app-`, `vm-`, `sqldb-`) and **abbreviates** the env token (`prod`/`dev`/`qa`/`stage`/`test`, e.g. `app-navigator-prod`). Both choices are the inverse of this rule (project-first, env-last, full-form). *(Verified. Note: CAF's common template `<type>-<workload>-<environment>` actually places env LAST too — it does not contradict env-last; it contradicts only the resource-type-first ordering and the abbreviation.)*
- **Honesty note — no authoritative source mandates env-LAST or full-form.** The position-last and full-form choices are deliberate **local conventions**, not canon: position-last keeps the discriminating token in a stable trailing slot; full forms kill the `dev`/`develop`/`development` ambiguity. The one major vendor framework that addresses ordering (Azure CAF) *abbreviates* the env token and leads with resource-type — so do not cite CAF *for* this rule (it argues against the abbreviation and the lead component), and do not invent a standard that blesses env-last. Hold it for internal consistency only. Serverless Framework's `<service>-<stage>-<fn>` infix is a framework-imposed exception to env-last, consistent with this being a preference tooling sometimes overrides.

### env token written identically across every layer
- **Honesty note — practitioner convention, mechanism-grounded, not a quoted standard.** No authoritative source states "write the env token identically everywhere" verbatim. The grounding is the *behavior* of tag systems: AWS tag keys/values are case-sensitive and AWS Tag Policies / Tag Editor match exactly, so `Environment`, `environment`, and `ENVIRONMENT` are distinct keys and a mixed convention silently fragments cost-allocation and policy compliance (AWS's own 2025-06-10 tagging-whitepaper revision advises against keys that differ only in casing). The cross-layer-identical rule follows from that mechanism, not from a cited mandate.
- `[Authoritative]` **AWS "Best Practices for Tagging AWS Resources" whitepaper (2023-03-30), Introduction** — "a resource name can only hold a limited amount of information" *(verbatim)*; supports pushing detail out of the name-string into tags. Note its own env-tag example uses **abbreviated** values (`['Prod','Dev','Test','Sandbox']`), so it does **not** support full-form env tokens.
- **Verdict: supported (consistency requirement only).** The cross-layer-identical *obligation* is sound; the specific token choice (full-form, env-last) is the local convention noted above, not vendor-blessed.


## TypeScript & JS standards — type safety, exports, modern syntax

Backs `languages/typescript-standards.md` (Type Safety, Code Style → exports/barrels, Runtime Awareness). Researched jun-2026.

### `any` forbidden — prefer `unknown` and narrow
- `[Authoritative]` **Vanderkam, *Effective TypeScript* 2nd ed. (O'Reilly, 2024), Item 46 "Use unknown Instead of any for Values with an Unknown Type"** (1st ed. Item 42, same title); reinforced by **Item 5 "Limit Use of the any Type"**. Caveat: item numbers shifted between editions (1st ed. 62 items → 2nd ed. 83 items) — cite the edition.
- `[Authoritative]` **typescript-eslint, rule `no-explicit-any`** (`https://typescript-eslint.io/rules/no-explicit-any/`) — `any` is "a dangerous 'escape hatch' from the type system. Using any disables many type checking rules"; recommends `unknown`. Enabled by the `recommended` preset.
- **Verdict: supported (strongly).**

### Explicit return types at module boundaries
- `[Authoritative]` **typescript-eslint, rule `explicit-module-boundary-types`** (`https://typescript-eslint.io/rules/explicit-module-boundary-types/`) — "Explicit types for function return values and arguments makes it clear to any calling code what is the module boundary's input and output." The rule cites this correctly. (Not part of the base `recommended` preset — it lives in stricter style configs; opt in explicitly.)
- `[Authoritative]` **Vanderkam, *Effective TypeScript* 2nd ed., Item 12 "Apply Types to Entire Function Expressions When Possible"** — adjacent but NOT the source: it argues for typing whole function *expressions*, not a module-boundary return-type mandate. Do not attribute this claim to the book; the eslint rule is the authority.
- **Verdict: supported.**

### `type` vs `interface`
- `[Authoritative]` **TypeScript Handbook, "Everyday Types" → Differences Between Type Aliases and Interfaces** (`https://www.typescriptlang.org/docs/handbook/2/everyday-types.html`) — "For the most part, you can choose based on personal preference… If you would like a heuristic, use `interface` until you need to use features from `type`." Official default is **interface-first**.
- `[Authoritative]` **Vanderkam, *Effective TypeScript* 2nd ed., Item 13 "Know the Differences Between type and interface"** — informs the choice; states no blanket preference.
- **Verdict: adjusted.** The rule's two mappings (unions → `type`, extendable shapes → `interface`) are both correct, but the symmetric framing under-weights the Handbook's interface-first lean. Optional sharpening: note that unions/tuples/mapped types force `type`, otherwise default to `interface`.

### `satisfies` over type assertions
- `[Authoritative]` **TypeScript 4.9 Release Notes, "The satisfies Operator"** (`https://www.typescriptlang.org/docs/handbook/release-notes/typescript-4-9.html`) — validates an expression against a type "without changing the resulting type of that expression," unlike `as` which overrides inference. The book predates this emphasis and has no `satisfies` item — do not manufacture a book cite. **Verdict: supported.**

### `strictNullChecks` enabled
- `[Authoritative]` **TSConfig Reference, `strictNullChecks`** (`https://www.typescriptlang.org/tsconfig/strictNullChecks.html`) — when on, `null`/`undefined` have distinct types and are type errors where a concrete value is expected; enabled by `strict: true`.
- `[Authoritative]` **Vanderkam, *Effective TypeScript* 2nd ed., Item 2 (know your options) & Item 33 ("Push Null Values to the Perimeter of Your Types").**
- **Verdict: supported.**

### Named exports + no feature-folder barrels
- `[Authoritative]` **Atlassian Engineering, "How We Achieved 75% Faster Builds by Removing Barrel Files"** (`https://www.atlassian.com/blog/atlassian-engineering/faster-builds-when-removing-barrel-files`) — measured **75% reduction in build minutes**, >30% faster TS highlighting; "even though you're only importing `Button`, tools like TypeScript and Jest need to process the entire barrel file." Independently corroborates the carve-out: barrels belong at a published package's public edge, not inside app feature folders. (Authority tag applies to Atlassian's measured empirical results, not to general DX folklore.)
- `[Non-authoritative]` Named exports enable granular tree-shaking, safe find/replace refactor, and editor auto-import — tooling/practitioner consensus (bundler tree-shaking docs); corroboration only, not the basis.
- **Verdict: supported.**

### `?.` / `??` runtime gate
- `[Authoritative]` **TC39 `proposal-optional-chaining` + `proposal-nullish-coalescing`, Stage 4; part of ECMAScript 2020 (ES2020).** These are *language-spec* features, transpilable to older targets — not Node features.
- `[Authoritative]` **Node.js 14.0.0 (2020-04-21), V8 8.1** (`https://nodejs.org/en/blog/release/v14.0.0`) — first Node line shipping native `?.`/`??` (enabled by default under the V8 8.1 update); this is the *runtime* gate.
- **Verdict: adjusted.** The "Node 14+" gate is correct for *native* support, but the rule conflates language origin with runtime. Reframe: ES2020 features, native from Node 14 (V8 8.1), otherwise available via transpilation — an old runtime forbids native execution, not the syntax.

### Honesty note — naming & style conventions
- The rule's identifier-casing conventions (PascalCase/camelCase/SCREAMING_SNAKE_CASE, `is`/`has`/`can`/`should` boolean prefixes, `.types.ts`/`.models.ts`/`.dto.ts` co-location, import order external→internal→relative) have **no standards-body or canonical-book authority** — they are reasonable shop conventions. If ever grounded, anchor in practitioner style guides (Google TypeScript Style Guide, Airbnb) tagged `[Semi-authoritative]`; do not manufacture canon for identifier casing.


## Java/Kotlin coding standards

Backs `languages/java-kotlin.md` (Spring Boot 3.x framework prefs, JVM Kotlin, language patterns). Researched jun-2026.

### Declarative REST clients — `@HttpExchange`, `RestClient` over `RestTemplate`
- `[Authoritative]` **Spring Framework Reference, *Integration > REST Clients*** (`docs.spring.io/spring-framework/reference/integration/rest-clients.html`) — defines the `@HttpExchange` HTTP Interface and `HttpServiceProxyFactory`: "define an HTTP Service as a Java interface with `@HttpExchange` methods… for remote access over HTTP via `RestClient`, `WebClient`, or `RestTemplate`." On the legacy client (verbatim): "As of Spring Framework 7.0, `RestTemplate` is deprecated in favor of `RestClient` and will be removed in a future version."
- `[Semi-authoritative]` **Spring engineering blog, *The state of HTTP clients in Spring*** (`spring.io/blog/2025/09/30/...`) — corroborates `RestClient` as the preferred synchronous client and gives the deprecation timeline.
- **Verdict: adjusted.** Rule is correct in direction. **Honesty note on the deprecation wording:** the reference-doc *prose* says "deprecated as of 7.0", but Spring's own roadmap (blog above) is that 7.0 only *announces intent* to deprecate; the formal `@Deprecated` annotation lands in **7.1 (Nov 2026)** and removal in **8.0**. So "deprecated" is the documented framing while the annotation is not yet applied — sharpen "prefer over RestTemplate" to "RestClient is the documented replacement; RestTemplate is on a deprecation path (annotated `@Deprecated` from 7.1)", and on a Spring 6.x pin RestTemplate is maintenance-mode only. Cite the version.

### `@Transactional` at the service boundary
- `[Authoritative]` **Spring Framework Reference, *Data Access > Transaction Management > Using `@Transactional`*** (`docs.spring.io/spring-framework/reference/data-access/transaction/declarative/annotations.html`) — verbatim: "The Spring team recommends that you annotate methods of concrete classes … rather than relying on annotated methods in interfaces"; "The `@Transactional` annotation is typically used on methods with `public` visibility." (Note: the parent `declarative.html` page carries no placement guidance — the locus is this `annotations.html` sub-page.)
- **Verdict: supported (with an honesty hedge).** The docs prescribe **concrete classes + public methods**; they do **not** literally say "service layer" — every example uses a service class, so "service boundary, not repositories, not controllers" is the well-established practitioner reading *consistent with* the docs, not a verbatim doc prohibition. Do not present "not on controllers" as a doc quote.

### `@ConfigurationProperties` over scattered `@Value`
- `[Authoritative]` **Spring Boot Reference, *Core Features > Externalized Configuration > Type-safe Configuration Properties*** (`docs.spring.io/spring-boot/reference/features/external-config.html`) — verbatim: "Using the `@Value(\"${property}\")` annotation to inject configuration properties can sometimes be cumbersome…"; "Spring Boot provides an alternative method of working with properties that lets strongly typed beans govern and validate the configuration of your application."; "Constructor binding can be used with records." *Supported strongly* — backs "bind a typed record once, inject the record."

### Virtual threads — never pool, never size manually
- `[Authoritative]` **Oracle, *Virtual Threads* core docs (JDK 21)** (`docs.oracle.com/en/java/javase/21/core/virtual-threads.html`), section "Represent Every Concurrent Task as a Virtual Thread; Never Pool Virtual Threads" — virtual threads are plentiful, so each should represent a single task; the thread count equals the concurrent-task count.
- `[Authoritative]` **JEP 444: Virtual Threads** (`openjdk.org/jeps/444`) — verbatim: "Virtual threads are cheap and plentiful, and thus should never be pooled: A new virtual thread should be created for every application task." *Supported strongly.*

### Virtual-thread pinning caveat — `synchronized` → `ReentrantLock`
- `[Authoritative]` **Oracle *Virtual Threads* docs (JDK 21), section "Avoid Lengthy and Frequent Pinning"** — verbatim: "If these mechanisms detect places where pinning is both long-lived and frequent, replace the use of `synchronized` with `ReentrantLock` in those particular places (again, there is no need to replace synchronized where it guards a short lived or infrequent operations)." The advice was scoped to *long-lived AND frequent* blocking I/O — never all `synchronized` blocks.
- `[Authoritative]` **JEP 491: Synchronize Virtual Threads without Pinning** (`openjdk.org/jeps/491`), delivered in **JDK 24** — reimplements monitor ownership by virtual-thread identity so `synchronized` no longer pins the carrier; the authors no longer recommend replacing `synchronized` with `ReentrantLock` (paraphrase of the JEP's position). Corroborated by *Inside Java Newscast #80* (inside.java, Nov 2024).
- **Verdict: contradicted (version-gated).** The rule's *unconditional* `synchronized`→`ReentrantLock` caveat is obsolete on **JDK 24+** (JEP 491) and was already over-broad on **JDK 21–23** (applied only to long-lived+frequent I/O hot spots, per the Oracle quote above). This is the file's single most time-sensitive line — gate it by JDK version and scope; on JDK 24+ it is moot.

### Kotlin `data class` for DTOs / value objects
- `[Authoritative]` **Kotlin docs, *Data classes*** (`kotlinlang.org/docs/data-classes.html`) — verbatim: "Data classes in Kotlin are primarily used to hold data"; the compiler derives `equals()`/`hashCode()`, `toString()`, `componentN()`, and `copy()`. *Supported strongly.* (The official **Coding Conventions** page does NOT mandate this — it only fixes the `data` modifier's position in modifier order. The mandate rests on the language-reference page above, not the style guide; do not cite Coding Conventions as the basis.)

### Kotlin sealed classes for closed hierarchies + exhaustive `when`
- `[Authoritative]` **Kotlin docs, *Sealed classes and interfaces*** (`kotlinlang.org/docs/sealed-classes.html`) — verbatim: "Sealed classes and interfaces provide controlled inheritance of your class hierarchies"; "All direct subclasses of a sealed class are known at compile time"; a `when` over a sealed class is checked exhaustively and "you don't need to add an `else` clause." *Supported strongly.*

### Kotlin coroutines for I/O-bound concurrency
- `[Authoritative]` **Kotlin docs, *Coroutines basics*** (`kotlinlang.org/docs/coroutines-basics.html`) — verbatim: "Coroutines can suspend their execution instead of blocking a thread"; "When a coroutine suspends, the thread isn't blocked"; structured concurrency via `CoroutineScope`. *Supports* the "coroutines for I/O concurrency" half. The "plain blocking is equivalent under Spring MVC + virtual threads" half is corroborated by Oracle/JEP 444 (blocking I/O is cheap on virtual threads), not by the Kotlin docs. *Supported* with split provenance; the "equivalent in throughput" wording is a reasonable inference, not a benchmarked claim in the sources.

### Never return `null`; `Optional` for returns only
- `[Authoritative]` **Bloch, *Effective Java* 3e, Item 54 "Return empty collections or arrays, not nulls"** (ch. 8, Methods) — "Never return null in place of an empty array or collection."
- `[Authoritative]` **Bloch, *Effective Java* 3e, Item 55 "Return optionals judiciously"** (ch. 8, Methods) — an Optional-returning method is less error-prone than one returning null; verbatim: "it is almost never appropriate to use an optional as a key, value, or element in a collection or array." On boxed primitives (paraphrase, not verbatim): the Item says never return an `Optional` of a boxed primitive type (with a minor-primitive exception for `Boolean`/`Byte`/`Character`/`Short`/`Float`) — the practical consequence is to prefer `OptionalInt`/`OptionalLong`/`OptionalDouble`.
- **Verdict: adjusted.** Rule supported. **Honesty note:** the rule's "never use `Optional` as a method parameter" line is *not* in Item 55 — the Item discusses `Optional` only as a return type. The no-parameter guidance is the well-established practitioner reading of Item 55's intent (Optional is designed as a return type), not a Bloch prohibition; record it as consensus, not canon.

### Wrap checked exceptions at the domain boundary
- `[Authoritative]` **Bloch, *Effective Java* 3e, Item 73 "Throw exceptions appropriate to the abstraction"** (ch. 10, Exceptions) — exception translation: higher layers catch lower-level exceptions and re-throw exceptions explainable in terms of the higher-level abstraction. Directly backs "catch `SQLException` in infra, re-throw as unchecked domain exception."
- `[Authoritative]` **Bloch, *Effective Java* 3e, Item 71 "Avoid unnecessary use of checked exceptions"** (ch. 10, Exceptions) — supports converting checked to unchecked when the caller can't meaningfully recover. *Supported.*


## Python standards — typing version-gates, resource & data idioms

Backs `languages/python-standards.md`. Researched jun-2026. The rule's load-bearing authorities are the typing/syntax PEPs (NOT the seed-suggested PEP 8/20/257, which back nothing in this rule — it carries no style or docstring directives) plus a few canonical idioms. All PEP version fields below were verified directly against peps.python.org.

### Version-gated modern syntax
- `[Authoritative]` **PEP 604 – Allow writing union types as X | Y** (Python-Version: 3.10, peps.python.org/pep-0604/) — `None | t == typing.Optional[t]`; the `X | Y` form is the post-3.10 idiom. Backs "`str | None` requires 3.10+" and "prefer `str | None` over `Optional[str]`".
- `[Authoritative]` **PEP 585 – Type Hinting Generics In Standard Collections** (Python-Version: 3.9, peps.python.org/pep-0585/) — `list`/`dict` become subscriptable; `typing.List`/`Dict` deprecated since 3.9. Backs "use builtin generics: `list[str]` not `List[str]`". Note: the deprecation is *silent* by design (no `DeprecationWarning`, to minimise runtime typing cost) — it is a style deprecation, not a runtime-warned one.
- `[Authoritative]` **PEP 695 – Type Parameter Syntax** (Python-Version: 3.12, peps.python.org/pep-0695/) — introduces `type X = ...` and `class Stack[T]:`. Backs the entire PEP 695 bullet, including the 3.12 gate and "prefer `type Point = …` over `TypeAlias`, `class Stack[T]:` over `Generic[T]`".
- `[Authoritative]` **PEP 634 – Structural Pattern Matching: Specification** (Final for 3.10, peps.python.org/pep-0634/) — backs "`match/case` requires 3.10+".
- `[Authoritative]` **PEP 563 – Postponed Evaluation of Annotations** (Python-Version: 3.7, peps.python.org/pep-0563/) — `from __future__ import annotations` stores annotations as strings; **available since 3.7** (works on every 3.7+ interpreter, including the rule's 3.9 floor). Backs "`from __future__ import annotations` for forward/postponed evaluation". *Correction recorded:* the feature's introduction version is 3.7, not 3.9 — 3.9 is merely a target floor on which it remains available.
- **Verdict: supported.** Every version number in the rule is confirmed directly from the PEPs. Attribution correction: the seed candidates named PEP 8/20/484/257 — of these only PEP 484 is tangential (foundational type-hints, superseded here by 604/585/695); PEP 8 (style) and PEP 257 (docstrings) back nothing in the current rule.

### No mutable default arguments
- `[Authoritative]` **Brett Slatkin, *Effective Python* 2nd ed. (2019), Item 24 "Use None and Docstrings to Specify Dynamic Default Arguments"** — the canonical statement of the rule (item number verified).
- `[Authoritative]` **Python official docs — Tutorial §4.8 "Default Argument Values" / Programming FAQ "Why are default values shared between objects?"** — "The default value is evaluated only once… the default is a mutable object such as a list… the following function accumulates the arguments passed to it on subsequent calls."
- **Verdict: adjusted.** The *no-mutable-default* rule is supported outright. But the rule's suggested fix `items = items or []` **diverges from the language's own canonical fix**, which is `if items is None: items = []` (shown verbatim in the tutorial). `items = items or []` silently replaces an explicitly-passed *falsy-but-valid* value (e.g. an empty list) with a fresh object — a latent bug the `is None` form avoids. Encode the strict idiom; treat `or []` as acceptable only where a falsy argument is semantically meaningless.

### Context managers for resources
- `[Authoritative]` **PEP 343 – The "with" Statement** (peps.python.org/pep-0343/) — "make it possible to factor out standard uses of `try/finally`… context managers provide `__enter__()` and `__exit__()` methods invoked on entry to and exit from the body of the with statement." The authoritative anchor for "always use `with`".
- `[Authoritative]` **Brett Slatkin, *Effective Python* 2nd ed. (2019), Item 66 "Consider contextlib and with Statements for Reusable try/finally Behavior"** — item number **verified** (was previously hedged as unconfirmed). Corroborates and extends PEP 343 to reusable `@contextmanager` helpers.
- **Verdict: supported.**

### Dataclasses (internal) vs Pydantic (external)
- `[Authoritative]` **PEP 557 – Data Classes** (Python-Version: 3.7, peps.python.org/pep-0557/) — "One main design goal of Data Classes is to support static type checkers." The PEP explicitly lists "value validation or conversion" and "type validation beyond that provided by PEPs 484 and 526" as **non-goals** — so dataclasses are correctly cast as the *trusted internal core*, not a validation boundary.
- `[Authoritative]` **Pydantic official docs — "Why use Pydantic" / Models concept page** (pydantic.dev/docs … /concepts/models) — untrusted data is parsed and validated so the resultant model's fields conform to the declared types; the boundary-validation role.
- `[Authoritative]` **Eric Evans, *Domain-Driven Design* (2003), ch. 14 "Anti-Corruption Layer"** — translate + validate untrusted external representations at the boundary before they reach the trusted domain core. This is the precise canonical home for "validate at the boundary".
- `[Authoritative]` **Alistair Cockburn, *Hexagonal Architecture (Ports & Adapters)* (2005)** — cited **for boundary translation only**: adapters translate between the external and the core's language. *Locus correction:* in Cockburn's model business-rule validation lives in the **core**, not the adapter — do not attribute "validate at the boundary" to Cockburn; that belongs to Evans's ACL (above) and to the practitioner "parse, don't validate / validate at the edge" principle.
- `[Non-authoritative]` Community "Pydantic at the edge, dataclasses at the core" consensus (jun-2026) — corroboration only.
- **Verdict: adjusted.** **Honesty note:** no single canonical book mandates "dataclasses internal / Pydantic external" verbatim — it is the *boundary-validation* principle (DDD ACL; the general validate-at-the-edge convention) instantiated onto two specific libraries. PEP 557 backs dataclasses' purpose; Pydantic docs back its boundary role; the pairing is well-grounded practitioner consensus, **not** borrowed standards authority.

### Async — asyncio over threading; TaskGroup over gather
- `[Authoritative]` **Python official docs — `asyncio` Coroutines and Tasks; `asyncio.TaskGroup`** (docs.python.org/3/library/asyncio-task.html, "Added in version 3.11") — "the remaining tasks in the group are cancelled" on first failure; exceptions "combined in an `ExceptionGroup`"; *gather* will not cancel remaining tasks. Backs the 3.11 gate and the structured-concurrency preference over `gather`.
- `[Semi-authoritative]` **Nathaniel J. Smith, "Notes on structured concurrency, or: Go statement considered harmful" (2018)** (vorpus.org/blog/notes-on-structured-concurrency-or-go-statement-considered-harmful/) — the structured-concurrency model `TaskGroup` adopts (via Trio). Named-author manifesto, not a standard.
- **Verdict: supported.** The asyncio-over-threading-for-I/O preference is mainstream and uncontroversial; the `TaskGroup`-over-`gather` safety advantage is confirmed verbatim in the official docs.

### Tooling bullets (Ruff, uv, pyproject, pytest) — NOT bibliography-canon
- `[Authoritative]` **PEP 621 – Storing project metadata in pyproject.toml** (peps.python.org/pep-0621/, "Tools MUST specify fields… in a table named `[project]`") + **PEP 518** (`[build-system]` table) — backs only "project metadata lives in `pyproject.toml`, not `setup.py`/`requirements.txt`".
- `[Authoritative]` **Ruff docs (docs.astral.sh/ruff) · uv docs (docs.astral.sh/uv)** — describe tool *capabilities* (lint+format in one tool; `uv pip compile`), **not** a mandate to choose them.
- **Honesty note:** apart from PEP 621/518 (metadata location), the Ruff/uv/Pyright/pytest preferences are tool-selection and practitioner conventions with **no canonical-book authority**. They are version-sensitive and governed by `rules/tools/context7.md` (current official docs), not by this bibliography. Do not manufacture standards authority for a tool choice.


## SQL & migration safety

Backs `languages/sql-migrations.md > Migration Safety` and the Prisma/Drizzle workflow sub-sections. Researched jun-2026.

### Reversibility — "every migration must be reversible"
- `[Authoritative]` **Fowler & Sadalage, *Evolutionary Database Design*** (martinfowler.com/articles/evodb.html), reverse-migrations discussion — "We haven't found this to be cost effective and beneficial enough to try all the time… On the whole **we prefer to write our migrations so that the database access section can work with both the old and new version of the database.**" (verified verbatim against the live article)
- `[Authoritative]` **Ambler & Sadalage, *Refactoring Databases: Evolutionary Database Design*** (Addison-Wesley, 2006), ch. 3 "The Process of Database Refactoring" — refactorings are made safe by a backward-compatible **transition/deprecation period** rather than a paired rollback script. (Chapter title verified; book text paywalled, so this is a paraphrase of the method, not a verbatim quote.)
- **Verdict: contradicted (reframe).** The canon does *not* mandate a down/rollback migration; it explicitly judges automated reverse migrations not cost-effective and prefers **forward-only (roll-forward) with a backward-compatible transition window**. The rule's "or document why it's irreversible" gestures at this, but the headline ("every migration must be reversible") inverts the canonical default. Foreground the transition-phase / expand-contract pattern as the primary safety mechanism; treat down-scripts as optional, not the rule.

### Transition window & expand-contract (NOT NULL multi-step, no-DROP-without-verification)
- `[Authoritative]` **Sato, bliki "ParallelChange"** (martinfowler.com/bliki/ParallelChange.html, D. Sato, 2014) — "**Parallel change**, also known as **expand and contract**, is a pattern to implement backward-incompatible changes to an interface in a safe manner, by breaking the change into three distinct phases: expand, migrate, and contract." The nullable-add → backfill → set-NOT-NULL sequence is expand-contract applied to a column constraint. *Supported strongly.* (Verified verbatim; note the post is authored by Danilo Sato as a guest bliki on Fowler's site, not by Fowler.)
- `[Authoritative]` **Fowler & Sadalage, *Evolutionary Database Design*** — section "Transition phase": "A transition phase is a period of time when the database supports both the old access pattern and the new ones simultaneously" (example: a view under the original table name during a rename). *Supported.* (verified verbatim)
- `[Authoritative]` **Ambler & Sadalage, *Refactoring Databases*** (2006) — ch. 3 transition/deprecation period; ch. 6 "Structural Refactorings" → **Drop Column** / **Drop Table** carry an explicit deprecation window + impact assessment before removal; ch. 11 "Transformations" → **Introduce New Column**. Backs both "never DROP without verifying impact" and the multi-step NOT NULL add. *Supported.* (chapter numbers/titles verified against the book TOC)

### Foreign-key indexing
- `[Semi-authoritative]` **PostgreSQL Documentation**, §5.5.5 "Foreign Keys" (postgresql.org/docs/current/ddl-constraints.html) — "Since a `DELETE` of a row from the referenced table or an `UPDATE` of a referenced column will require a scan of the referencing table for rows matching the old value, it is often a good idea to index the referencing columns too. Because this is not always needed… the declaration of a foreign key constraint **does not automatically create an index on the referencing columns**." *Supported, engine-scoped.* (verified verbatim)
- **Honesty note:** the rule's blanket "index every REFERENCES" is **engine-dependent** — MySQL/InnoDB auto-creates an index on FK columns (MySQL Reference Manual, *FOREIGN KEY Constraints*, dev.mysql.com/doc/refman/8.4/en/create-table-foreign-keys.html), so the rule is redundant there. Anchor it in PostgreSQL behaviour; it is a sensible default, not a universal DB law.

### Idempotency & timestamp-prefixed naming
- **Honesty note:** "idempotent when possible (IF NOT EXISTS / OR REPLACE)" and "timestamp-prefixed file naming" are **practitioner / tool conventions with no canonical book or standard** behind them. `[Semi-authoritative]` **Flyway docs** (versioned `V<version>__<desc>.sql`; documentation.red-gate.com) and **Rails ActiveRecord migrations** (`YYYYMMDDHHMMSS_*` prefix) corroborate the *naming* as a tool-imposed ordering convention; do not over-attribute it to a formal standard. The `IF NOT EXISTS` idempotency point is `[Non-authoritative]` practitioner consensus only — corroboration, never the basis.

### Prisma — `@updatedAt` is ORM-level, not a DB trigger
- `[Authoritative]` **Prisma Schema Reference**, `@updatedAt` remarks (prisma.io/docs/orm/reference/prisma-schema-reference) — `@updatedAt` is **implemented at the Prisma ORM level**; Prisma Client sets the value on write. Timezone caveat (pre-4.4.0): "`@updatedAt` operates at the Prisma ORM level, while `now()` operates at the database level." (verified)
- **Verdict: contradicted (factual).** The rule's "`@updatedAt` for `updatedAt` — Prisma handles the trigger" is wrong: there is **no database trigger**; raw-SQL writes (or any other client) that bypass Prisma Client will **not** update the column. Correct to "Prisma Client sets it on write (application-level, not a DB trigger)."

### Drizzle — `push` "never in production"
- `[Authoritative]` **Drizzle ORM docs, "drizzle-kit push"** (orm.drizzle.team/docs/drizzle-kit-push) — push "is the best approach for rapid prototyping **and we've seen dozens of teams and solo developers successfully using it as a primary migrations flow in their production applications**," pairing with blue/green deploys and serverless DBs (Neon, PlanetScale, Turso). (verified verbatim)
- **Verdict: adjusted.** The official guidance does **not** forbid `push` in production — it endorses it for blue/green + serverless setups. The rule's "never in production" is a **stricter safety convention** (favoring auditable, version-controlled SQL files), defensible but not vendor-canonical. Reframe as a team policy ("we reserve `push` for development; production uses generated SQL migrations for an auditable history"), not an implied vendor prohibition.


## Shell script standards

Backs `languages/shell-standards.md`. Researched jun-2026.

### Shebang — `#!/usr/bin/env bash` vs `#!/bin/bash`
- `[Authoritative]` **Google Shell Style Guide, "Which Shell to Use"** — "Executables must start with `#!/bin/bash` and minimal flags." Mandates the *opposite* of our rule (which forbids `#!/bin/bash`). — https://google.github.io/styleguide/shellguide.html
- `[Non-authoritative]` Practitioner consensus (Baeldung, nixCraft, cyberciti) — `#!/usr/bin/env bash` resolves bash through `$PATH`, portable where bash lives outside `/bin` (FreeBSD `/usr/local/bin/bash`, Nix). Tradeoff: it honors the *first* bash in `$PATH`, so a non-default version can win. (Corroboration only — not the basis.)
- **Verdict: contradicted (defensibly).** Our rule and the closest-to-canonical shell authority disagree, and both are sound: env-bash optimizes PATH portability, `/bin/bash` optimizes a fixed, known interpreter. Keep our choice but stop framing "Never `#!/bin/bash`" as universal consensus — it is a deliberate divergence from Google's canon, not an oversight. This is the only substantive finding in the shell rule.

### `[[ ]]` over `[ ]`, `set` options, stderr, `readonly`/`local`
- `[Authoritative]` **Google Shell Style Guide** — "Test, `[ … ]`, and `[[ … ]]`": "`[[ … ]]` is preferred over `[ … ]`, `test` and `/usr/bin/[`" (rationale: "no pathname expansion or word splitting takes place between `[[` and `]]`"). "STDOUT vs STDERR": "All error messages should go to `STDERR`." Constants "should be made `readonly` immediately afterwards"; "Declare function-specific variables with `local`." — https://google.github.io/styleguide/shellguide.html *Supported.*
- **Honesty note on `set -euo pipefail`:** the Google guide endorses `set` for options ("Use `set` to set shell options so that calling your script as `bash script_name` does not break its functionality") but does **not** mandate the `-e -u -o pipefail` trio verbatim — that exact idiom is strong practitioner consensus, not lifted from the guide. *Adjusted.* (Greg's Wiki / BashFAQ caution that `set -e` has surprising edge cases; the blanket use is defensible for scripts, with that caveat.)

### POSIX portability caveat
- `[Authoritative]` **POSIX.1-2017 (IEEE Std 1003.1-2017), Shell Command Language §2.4 (Reserved Words)** — `[[`, `]]`, `function`, `select` are listed as words that "may be recognized as reserved words on some implementations … causing unspecified results." So `[[ ]]` is a bash/ksh extension, not portable POSIX sh. — https://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html
- `[Authoritative]` **POSIX.1-2017, §2.9.1 (Command Search and Execution)** — `local`, `typeset`, and `declare` appear in the list of utility/built-in names whose behavior is unspecified; `local` is **not** POSIX. (Note: these live in §2.9.1, NOT in the §2.4 reserved-words note — a distinction the source research conflated.) — https://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html
- **Honesty note — `pipefail`:** `pipefail` is absent from the POSIX `set` options; it is a bash/ksh/zsh extension. This is an *absence*, so there is no section to cite for it — recorded as an observation, not a POSIX citation.
- **Verdict: adjusted.** `[[ ]]`, `local`, and `pipefail` are bash/ksh extensions, not portable POSIX sh. The rule's portability claims are valid only because it mandates a bash shebang — the portability is *bash*-portability, not POSIX-portability. Do not cite POSIX as backing for these constructs; cite it as the boundary that makes the bash shebang load-bearing.

### ShellCheck warnings (SC2086 / SC2046 / SC2155)
- `[Authoritative]` **ShellCheck wiki** — SC2086 "Double quote to prevent globbing and word splitting" (https://www.shellcheck.net/wiki/SC2086); SC2046 "Quote this to prevent word splitting" — unquoted command substitution (https://www.shellcheck.net/wiki/SC2046); SC2155 "Declare and assign separately to avoid masking return values" — a combined `declare`/`export`/`local` always returns the declaration's exit status, masking the command substitution's failure from `set -e` and traps (https://www.shellcheck.net/wiki/SC2155). *Supported strongly* — the three codes the rule names are exactly the real-bug-catching ones, and SC2155 directly reinforces the `set -e` reliability the rule depends on.

### Files surveyed with NO bibliography entry warranted (official-docs / context7 territory)
- **`languages/iac-devops.md`** — file-specific Dockerfile / Terraform / GitHub Actions syntax (SHA-digest pinning, `COPY` ordering, `required_providers`, action SHA-pinning, minimal `GITHUB_TOKEN` permissions) is version-sensitive **official-vendor-docs** territory (docs.docker.com/build/building/best-practices, developer.hashicorp.com/terraform/language/providers/requirements, docs.github.com security-hardening). The underlying *principles* (IaC-over-manual, immutability) trace to **Kief Morris, *Infrastructure as Code* (O'Reilly, 2nd ed. 2020, ISBN 9781098114671)** and the Twelve-Factor App — but those are already in this file's "Retention" section; citing Morris for the vendor-doc syntax he doesn't cover would be borrowed authority. The GitHub Actions guidance ("Pinning an action to a full-length commit SHA is currently the only way to use an action as an immutable release") was verified against the live doc.
- **`languages/react-nextjs.md`, `angular-patterns.md`, `nestjs-patterns.md`, `tailwind.md`** — explicitly version-banded framework conventions (Next 14/15/16, Angular 16-20, NestJS 10/11, Tailwind 3/4). Pure **official-docs/context7** territory; book backing is thin-to-nonexistent for these specific APIs, and react-nextjs.md already routes implementers to context7. A static bibliography entry would freeze a moving target — the exact failure context7 exists to prevent.


## Session-based documentary organization

Backs `workflow/project-structure.md > Session capture layer` and its condensed mirror in `harness/AGENTS.md`, plus the session touches in `flow-core/references/specs-structure.md`, `ledger-template.md`, the `flow-dev`/`flow-hygiene` skills (jun-2026 names; today `flow-plan`/`flow-build` and `flow-workspace`), `workspace-custodian`, and `workflow/memory-routing.md` (session slug in memories). Researched jun-2026.

### Two models of time-ordered capture — durable (chosen) vs ephemeral (rejected)
- `[Authoritative]` **Scientific lab-notebook discipline (NARA; Harvard/Stanford RDM; GLP; FDA 21 CFR Part 11)** — time-ordered capture as a durable, immutable, retained record with an audit trail; canonical source for reproducibility. *Adjusted — this is the model adopted:* the session journal is versioned and immutable once concluded.
- `[Authoritative]` **Ahrens, *How to Take Smart Notes* (Zettelkasten, 2017)** + `[Semi-authoritative]` **Forte, *Building a Second Brain* (PARA, 2022)** + **Sebastian, *The Knowledge Funnel*** — the opposite model: time-ordered capture (fleeting notes / inbox / daily notes) is *ephemeral*, discarded after promotion to topic-organized durable homes. *Recorded as the model NOT chosen:* the user requires versioned, team-shareable decision history, so capture is durable. The intention layer (`decisions/contracts/epics/conventions`) plays the topic-organized durable role these frameworks promote *into*.
- **Verdict: the two models genuinely diverge** on whether capture is durable; the rule chooses the lab-notebook/ADR lineage for the stated requirement and records the alternative rather than suppressing it.

### Capture-then-promote, immutability, and the promotion gate
- `[Authoritative]` **Nygard, *Documenting Architecture Decisions* (Cognitect, 2011)** + **adr.github.io / MADR** + **AWS Prescriptive Guidance (ADR)** — dated, immutable, append-only decision records; supersession by a new linked record, never in-place edit. *Supported* — backs session immutability and the by-type `decisions/` promotion home.
- `[Authoritative]` **IETF RFC 2026** (maturity levels, minimum tenure) + **Python PEP 1** + **Rust RFC process** + **Kubernetes KEP** ("implementable" gate) — a proposal advances draft→accepted→final at an acceptance checkpoint, not retroactively; the ephemeral discussion (RFC) and the durable record (ADR) are two lives of one decision. *Supported, adopted lightly:* the rule reuses close-time appraisal + flow-workspace `promote` as the gate, deliberately WITHOUT RFC 2026's formal minimum-tenure ceremony (overkill for a solo/small-team config).

### Type-organized durable docs (the intention layer)
- `[Authoritative]` **Procida, *Diátaxis*** + **Divio Documentation System** — durable docs organized by TYPE (tutorial / how-to / reference / explanation), not by time. *Supported* — grounds the intention layer as the by-type axis orthogonal to the by-time session axis.
- `[Semi-authoritative]` **Write the Docs — *Docs as Code*** — docs versioned, peer-reviewed, in git alongside code. *Supported* — backs versioning the session journal (the user's requirement) rather than treating it as throwaway scratch.

### Already-cited above, reused here (do not re-research)
- **ISO 8601-1:2019 + NARA** (date `YYYY-MM-DD`, lexicographic = chronological) — *Generated-artifact naming* section; backs the DD-MM-YYYY → ISO correction.
- **ISO 15489-1:2016 + DCC Curation Lifecycle (Higgins 2008)** — *Generated-artifact naming* section; backs the session lifecycle states (in-progress → concluded → finalized) with measurable triggers, and close-time appraisal/prune.
- `[Authoritative]` **SSOT (general-ledger / Atlassian)** — one master per datum. *Supported* — backs "the sessions index is single and co-located (`sessions/README.md`), never a parallel index that drifts", and the finding that the non-versioned ledger (outside the specs repo) cannot index versioned sessions.

### Honesty notes
- No software-engineering canon standardizes a session→permanent-docs **promotion schedule**; the close-time-appraisal + hygiene-sweep mechanism is this config's own engineering judgment (reusing file-routing step 4 + the `/flow-workspace` sweep), not a standard.
- `[Semi-authoritative]` **Pragmatic Programmer, *It's All Writing* (engineering daybook)** establishes dated work-capture as recognized practice but does NOT formalize promotion — recorded as lineage, not a mandate.


## Agents + evals — verifiable test gates over the test-first ritual

Backs the **verifiable test gate** in `quality/testing.md` (and its references in `workflow/agent-routing.md` and `quality/development-principles.md`), the Execution Scope reframe, and the "gates, not tiers" decision that retired `project-maturity`. Full decision: `_support/spec/2026-06-22-agent-verifiable-quality-gates-design.md`. Researched jun-2026. **Framing caveat:** these rules are read by AI agents, not humans; sources are weighed for what an agent's *output* can be checked against, not for what a disciplined human would do unprompted.

### Fail-to-pass + pass-to-pass — the dual metric that defines the gate
- `[Authoritative]` **Jimenez, Yang, Wettig, Yao, Pei, Press, Narasimhan, "SWE-bench: Can Language Models Resolve Real-World GitHub Issues?", ICLR 2024 (arXiv:2310.06770)** — scores a candidate patch by running the repo's tests: `FAIL_TO_PASS` (the issue's tests must flip from failing to passing) AND `PASS_TO_PASS` (previously-passing tests must stay green). The exact dual metric the gate encodes. *Supported — this is the gate's definition.*
- `[Semi-authoritative]` **OpenAI, "Introducing SWE-bench Verified" (openai.com, 2024-08)** — a human-validated 500-task subset; reinforces that the verification is the *run*, not the model's self-report. Vendor blog, not peer-reviewed; cited for the validation practice, not as independent authority.
- **Verdict: supported.** The benchmark exists precisely because "the model says it fixed it" is unreliable; the gate lifts its evaluation discipline into the rule.

### The test-first ritual alone is neutral-to-harmful; verifiable context is what helps
- `[Non-authoritative]` **"TDAD" (arXiv:2603.17973)** — on smaller models, instructing the TDD ritual *by itself* raised regressions (+9.94%); supplying verifiable context about which tests to check (dependency-graph impact analysis) cut regressions ~70%. **Preprint, not peer-reviewed — treat as directional evidence.** Backs *why* the gate targets the checkable result and recommends (not mandates) test-first, and the Execution Scope "let the tool compute the impact graph, don't let the agent guess" angle.
- `[Semi-authoritative]` **K. Beck, *Augmented Coding* / "Tidy First?" (substack, 2025)** — TDD is a "superpower" paired with a coding agent (the failing test pins intent the agent then satisfies), with an explicit warning that agents will *cheat* the tests under pressure. Backs the recommend-the-ritual / mandate-the-result split. Named-author essay series; lineage, not standard.
- **Verdict: adjusted.** The ritual is endorsed as *helpful context*, not as a sufficient control — the rule keeps it a recommendation and puts the enforcement on the run.

### Reward hacking / test-gaming — the failure mode the anti-gaming clauses defend against
- `[Semi-authoritative]` **METR, "Recent Frontier Models Are Reward Hacking" (metr.org, 2025)** — frontier models reward-hack on a non-trivial fraction of coding tasks; documented tactics include a `conftest.py` that rewrites outcomes to "passed" and reading `git log` to recover the expected answer. The concrete behaviors `testing.md` and `development-principles.md` name as test-gaming.
- `[Semi-authoritative]` **Krakovna et al., "Specification gaming: the flip side of AI ingenuity" (DeepMind blog, 2020)** — an agent optimizes the literal objective (a green suite) over its intent (correct code) whenever the proxy is gameable. The general principle under "weakening a test to pass is silencing a guardrail".
- `[Semi-authoritative]` **L. Weng, "Reward Hacking in Reinforcement Learning" (lilianweng.github.io, 2024-11)** — survey framing of reward hacking; corroboration of the mechanism. Named-author; corroboration only.
- **Verdict: supported as mechanism evidence.** These ground *why* the gate must be re-established by a fresh-context verifier that runs the tests, rather than trusted from the implementing agent's report.

### Regression test selection — the tool-computed subset is safe; the guessed one is not
- `[Authoritative]` **G. Rothermel & M. J. Harrold, "A safe, efficient regression test selection technique", *ACM TOSEM* 6(2):173–210 (1997)** — defines *safe* RTS: select every test whose behavior the change can affect, computed from the program's dependence/control-flow graph. The principled root of Execution Scope's "select the affected subset by import graph, not by eye" — a graph-selected subset is safe; a hand-picked one is the regression source.
- **Verdict: supported.** Modern tooling (`jest --findRelatedTests`, vitest `related`) is the practical descendant; the safety guarantee is Rothermel & Harrold's.

### Gates, not tiers — internal quality is not a stage-gated luxury
- `[Semi-authoritative]` **M. Fowler, "Is High Quality Software Worth the Cost?" (martinfowler.com, 2019-05-29)** — high *internal* quality pays for itself almost immediately because low quality slows the very next change; there is no economic regime where deliberately lowering internal quality is the faster path beyond the shortest horizons. Backs retiring the maturity ladder's stage-relaxation of tests in favor of one always-on gate.
- **Verdict: adjusted.** Fowler argues *internal* quality is never worth trading; the rule applies this to the test gate specifically and preserves the genuinely exposure-gated *external* floor (security) in `security.md` — the one piece of the old tier model that was load-bearing.

## Browser automation tooling — agent-browser CLI as the default driver

Backs `rules/tools/browser-automation.md` (and its pointers in `workflow/agent-routing.md`, `harness/AGENTS.md`, and the `in-vivo-qa-tester` / `ui-reviewer` agents). Grounded in the official project docs plus first-party measurement — operational, not canonical literature. Researched jun-2026.

- `[Authoritative]` **Vercel Labs, `agent-browser` (github.com/vercel-labs/agent-browser)** — CLI and daemon both written in Rust ("No Playwright or Node.js required for the daemon"; Node 24+ only to build from source). The Homebrew bottle ships the native binary (`agent-browser-darwin-arm64`, 8.3 MB Mach-O) behind a ~120-line Node launcher shim that `spawn`s it — so the engine is Rust despite the `#!/…/node` shebang on the wrapper. Documents the `--profile` read-only-snapshot reuse, persistent-directory profiles, `--restore --session` (AES-256-GCM), `--json` structured output, stable `@eN` refs, the `--content-boundaries` prompt-injection guard, and the `console`/`network` (HAR) commands the rule relies on.
- `[Non-authoritative]` **First-party token/latency benchmark on a fixed page (Hacker News), jun-2026** — snapshot footprint: agent-browser 26.9 KB vs chrome-devtools 37.8 KB (1.4×) vs Playwright MCP 57.7 KB (2.15×); warm command ~30–40 ms after a ~0.7 s cold start. Backs the "default driver, MCP for what it can't do" decision and the one-Bash-call/one-turn rationale. Our own numbers, reproducible by re-running — corroboration, not authority.
- **Verdict: supported.** The Rust-engine and feature claims are the vendor's own docs; the comparative advantage is our measurement. chrome-devtools is retained for Lighthouse/perf-insight/heap, which agent-browser does not cover.


## Initiative grouping in the session capture layer

Backs the **Initiative** addition to `workflow/project-structure.md > Session capture layer` (and its condensed mirror in `harness/AGENTS.md`). Researched jun-2026. The companion *skill-level* design that motivated it — the `flow-build` reconciler, the research→write→build split, the plan-as-state contract — is **not** a `global/rules` convention, so its external sources (control-loop reconciliation, plan-then-execute, the optional reference project separation) live in the design spec, not here: `_support/spec/2026-06-26-flow-dev-split-design.md` §7. Recorded so the skill design is not re-researched from this file (consistent with how the phase-transition contract's sources stayed in its own spec).

### Grouping multi-session work that belongs to one effort
- `[Authoritative]` **R. C. Martin, *Clean Architecture* ch. 13 (Common Closure Principle)** — "classes that change together belong to the same component." Already cited above under *Grouping a cohesive deliverable in a folder*; the initiative applies the same cohesion principle one level up — the sessions that execute one plan belong in one container. Cross-reference, not re-cite.
- **Honesty note:** the initiative folder *shape* (start-date container, `findings/` + `plan/` inputs, nested dated executions) is an **internal organizational convention** of this pack — it extends the session-capture layer's own existing conventions, no external standard prescribes it, and none is needed. **Verdict: internal** — grounded only by analogy to the cohesion principle above; do not attach a borrowed standard to it.


## UI visual design — front-loaded design criteria

Backs `languages/ui-visual-design.md` (path-scoped to UI files), the **Visual craft** axis of `flow-mock/references/ux-rubric.md` (#11–17), the craft additions to the `ui-reviewer` agent, the mandatory `flow-mock review` gate and the opt-in `flow-build` design gate, and the condensed mirror in `harness/AGENTS.md`. Researched jun-2026 by reading the source book per chapter (the chapter PDFs in `_support/refactoring-ui/chapters/`).

### Primary source — the distilled principles
- `[Semi-authoritative]` **A. Wathan & S. Schoger, *Refactoring UI* (2018)** — the named-author design book the rule distills. A practitioner manual, not a standards body — `[Semi-authoritative]` by the conventions above. Chapter loci for the kept principles: **Hierarchy** (ch. "Hierarchy is Everything": size/weight/color levers, 2–3 text colors, weights 400/500 + 600/700, action hierarchy solid/outline/link p60–62, grey-on-colored-bg same-hue p42–44, tag≠size p54–55, weight↔contrast p56–59); **Spacing/layout** (ch. "Layout and Spacing": ambiguous spacing outer>inner p96–99, fixed/max-width vs % p84–90, don't couple sizes via `em` p92–95, ~25% min scale gap p72); **Typography** (ch. "Designing Text": line length 45–75ch/20–35em p114–116, `align-items: baseline` p118–120, line-height bands p122–125, right-align numerals + `hyphens:auto` p128–131, letter-spacing ±0.05em p132–134, px/rem not em p106–107); **Color** (ch. "Working with Color": HSL p138–141, palette counts 8–10 greys / 1–2 primaries / accents p142–147, 9-step 100–900 fixed scale p148–151, saturation compensation p152–153, hue-rotation 60/180/300 & 0/120/240 ≤20–30° p153–157, grey temperature p158–160); **Depth** (ch. "Creating Depth": light-from-above hand-picked highlight p172–179, 5-step elevation shadow scale p180–184, two-part shadows p186–189, zero-blur solid shadow p190–192); **Images** (ch. "Working with Images": text-on-image contrast techniques p202–206, intended size p208–213, user content `cover`+inner-shadow p214–217); **Finishing** (ch. "Finishing Touches": fewer borders p238–241, accent borders p224–226, gradient ≤30° hue p228–232, empty-state craft p234–236).
- **Verdict: supported (as practitioner guidance).** The numeric values are the book's own worked examples — starting criteria to map onto a project's tokens, not universal constants.

### Contrast ratios — the one authoritative anchor
- `[Authoritative]` **W3C WCAG 2.1/2.2, SC 1.4.3 Contrast (Minimum) & SC 1.4.11 Non-text Contrast** — text contrast **≥ 4.5:1** normal, **≥ 3:1** large (≥18pt / 14pt bold); UI-component/graphics contrast ≥ 3:1. The rule's contrast thresholds rest on WCAG, not on the book (the book illustrates ratios; WCAG mandates them). SC 1.4.1 Use of Color backs "never encode meaning by color alone".
- **Verdict: supported.**

### Why codify design criteria as agent rules + a gate (the brief's premise)
- `[Non-authoritative]` **Practitioner sources on AI design-system consistency** (UXPin, *AI Design Systems*; G. Han, "How to make your design system AI-ready", Medium; figr.design) — AI *scales the existing structure*: inconsistency that took months to compound now compounds in days; codified criteria + enforcement turn intent into a checkable constraint. Corroboration for front-loading + a mandatory gate, not a basis.
- `[Non-authoritative]` **Design-linting ROI anecdotes** (Peerlist, "Design Lint…"; destefanis/design-lint) — cited as directional motivation (reduced design tech debt / QA time), not measured authority.
- `[Non-authoritative]` **Prior art — Refactoring UI as an AI skill** (github.com/gnurio/refactoring-ui-plugin; github.com/jaywilburn/refactoring-ui-skill) — confirms the ~10-criteria, per-principle (evaluative/generative/corrective) shape and an optional scoring rubric. Precedent for structure, not authority for content.
- **Honesty note:** the benefit/ROI framing has **no** authoritative locus — it is practitioner consensus and vendor blogging. Do not attach a standard to it; the load-bearing authority here is WCAG (contrast/color) and the named-author book (the craft principles).

### Per-chapter curation verdicts (scope evidence)
The rule distills a **curated** subset, decided chapter by chapter — not the whole book. Recorded so the scope isn't re-litigated:

| Chapter | Verdict | Rationale |
|---|---|---|
| Starting from Scratch | **excluido** | Design-*process* advice (feature-first, grayscale-first); its one hard principle (constrained scales / no arbitrary values) is already owned by `tailwind.md`. |
| Hierarchy is Everything | parcial | Kept the verifiable hierarchy levers (action hierarchy, weight↔contrast, grey-on-color, tag≠size, labels-last-resort); cut the framing thesis (redundant with the UX flow rubric). |
| Layout and Spacing | parcial | Kept ambiguous-spacing (outer>inner), fixed/max-width vs %, no `em`-coupling, ~25% gap note; cut unverifiable method advice; did not re-state the scale (owned by `tailwind.md`). |
| Designing Text | parcial | Kept line-length, baseline alignment, line-height bands, right-align numerals, letter-spacing, px/rem-not-em; cut font-selection taste (owned by design-taste skills). |
| Working with Color | parcial | Kept HSL, palette counts, fixed shade scale, saturation/hue mechanics, grey temperature, WCAG ratios; cut "don't rely on color alone" headline to its non-obvious chart nugget. |
| Creating Depth | **completo** | Almost entirely verifiable and uncovered elsewhere (light source, elevation scale, two-part & solid shadows); cut only generic restraint advice. |
| Working with Images | parcial | Kept text-on-image contrast, intended-size, user-content cropping; cut "use good photos" (unverifiable, not agent-actionable). |
| Finishing Touches | parcial | Kept fewer-borders, accent borders, 30°-gradient, empty-state craft delta (cross-ref the flow rubric), merged-table/selectable-card patterns; trimmed component taste owned by design-taste skills. |
| Leveling Up | **excluido** | Four-page closing chapter on learning methodology; no standalone actionable principle. |

---

## Note — flow-pack v2 redesign (2026-07-17)

The flow-pack v2 redesign superseded the F1–F7 phase model with four project stages (`arranque · specs · desarrollo · operación`), where the daily brainstorm→spec→plan→build loop lives inside `desarrollo`. It fused flow-intake/kickoff/foundation into `/flow-start`, dissolved flow-mock into the chain as a work type, retired flow-deploy (promotion knowledge moved to `flow-core/references/promotion-playbook.md`), and added `/flow-brainstorming`. The UX rubric was relocated from `flow-mock/references/ux-rubric.md` to `flow-core/references/ux-rubric.md`. Historical entries above that cite F-phases or the old rubric path reflect the model in force when researched — they are not rewritten.

## Unattended autonomy mode (explicit overnight delegation)

Backs `workflow/unattended-autonomy.md`, its cross-references in `quality/reporting-integrity.md > Fix at the Root`, `workflow/gap-resolution.md > Divergence`, and `global/CLAUDE.md > Destructive Operations`, plus the condensed mirror in `harness/AGENTS.md > Unattended Delegation Mode`. Researched jul-2026 via adversarial deep-research (24 sources fetched → 117 claims extracted → top-25 verified with 3 independent votes each → 24 confirmed unanimously, 1 refuted). Full report archived at `_support/archive/audits/unattended-autonomy-research-2026-07-09.html`.

### Proceed-and-log = Level 7; the decision log defines the mode
- `[Authoritative]` **Sheridan & Verplank (1978), *Human and Computer Control of Undersea Teleoperators*, MIT Man-Machine Systems Lab / ONR (DTIC ADA057655)** — 10-level autonomy scale: Level 7 "does the whole job and necessarily tells the human what it did"; dropping the mandatory report is a distinct, higher level (8–10). *Supported* — the decision log is the mode's defining boundary, not a courtesy. Also the glossary definition: supervisory control grants autonomy "over short periods and restricted conditions", intermittently reprogrammed by a person → run-bound expiry.
- Same source, Level 6 flowchart logic ("STARTS action if HUMAN APPROVES or if t > T and HUMAN HAS NOT DISAPPROVED") — management-by-exception presupposes an attentive human ("informs human in plenty of time to stop it"). *Supported (direct design inference, flagged as such by verifiers):* timeout-default-proceed is illegitimate unattended → gated decisions pause-and-queue or fail closed.

### Per-action-class widening; consequence cost decides what stays gated
- `[Authoritative]` **Parasuraman, Sheridan & Wickens (2000), "A Model for Types and Levels of Human Interaction with Automation", IEEE Trans. SMC-A 30(3):286–297** — automation is set per function class (information acquisition / analysis / decision selection / action implementation) on a continuum, never as a binary switch; "the cost of adverse consequences define major evaluative criteria", prescribing low action-automation with human error-trapping when consequences of a wrong action are great. *Supported*, with one adjustment absorbed from the LOA-critique literature (Jamieson & Skraaning 2018; Dekker & Woods 2002): enumerate concrete action shifts, don't pick a point on an ordinal scale.

### Explicit-only activation; per-engagement scoping
- `[Authoritative]` **SAE J3016_202104, *Taxonomy and Definitions for Terms Related to Driving Automation Systems*** — roles allocate by DESIGN, not by actual behavior ("a driver who fails to monitor the roadway... still has the role of driver, even while s/he is neglecting it"), and "the level of driving automation exhibited in any given instance is determined by the feature(s) that are engaged", not the system's maximum capability. *Supported:* silence never delegates; the mode exists only during the run it was engaged for.

### One mode, visible state
- `[Authoritative]` **Sarter & Woods (1995), "How in the World Did We Ever Get into That Mode?", Human Factors 37(1), DOI 10.1518/001872095779049516** — mode proliferation "creates new mode-related error forms and failure paths"; every added mode imposes "new monitoring and attentional demands to track which mode the automation is in". *Adjusted (requirement ADDED to the design):* exactly one unattended mode, declared visibly at activation and in every report, and absolute gates that never vary by mode. Caveat: the paper studies attended supervisory control; the unattended extension is analogical and likely understates the risk.

### Oversight taxonomy; standing halt channel
- `[Authoritative]` **CRS IF11150, *Defense Primer: U.S. Policy on Lethal Autonomous Weapon Systems* (DoDD 3000.09, 2023 reissue)** — in/on/out-of-the-loop taxonomy; the directive mandates "appropriate levels of human judgment over the use of force", explicitly per context and function ("can differ... even across different functions in a weapon system"); on-the-loop = retained capacity to "monitor and halt". *Supported:* per-action-class oversight plus a standing revocation channel instead of per-action approval. Note: the loop terminology is CRS's framing, not the directive's own text.

### Compensating controls; override/reverse/stop
- `[Authoritative]` **Regulation (EU) 2024/1689 (AI Act), Art. 14** — 14(3): oversight measures "commensurate with the risks, level of autonomy and context of use"; 14(4)(d)–(e): the overseer can disregard, override, or REVERSE the output and interrupt via "a stop button or a similar procedure". Applied by analogy (binds high-risk systems only; "appropriate and proportionate" chapeau). *Supported:* first-message revocation + reversible-branch bias ("reverse the output" holds only if the run stays on a dedicated branch).
- `[Authoritative]` **AI HLEG, *Ethics Guidelines for Trustworthy AI* (2019)** — "the less oversight a human can exercise over an AI system, the more extensive testing and stricter governance is required" (corroborated in peer-reviewed scholarship: Enqvist 2023, *Law, Innovation and Technology*). *Adjusted (ADDED):* delegation ADDS controls (log, reversibility, expiry, queue), never merely subtracts confirmations.

### Vendor instantiation; the deterministic-gate caveat
- `[Semi-authoritative]` **Anthropic, *How we built Claude Code auto mode* (engineering post, mar-2026)** — unauthorized-by-default ("Everything the agent chooses on its own is unauthorized until the user says otherwise"), irreversibility-weighted blocking inside autonomous mode, circuit breakers (3 consecutive / 20 total denials), headless with no human → terminate (fail closed). *Supported* — validates the architecture; constants may change (post is recent).
- `[Semi-authoritative]` **arXiv 2604.04978 (independent adversarial stress test of the auto-mode classifier)** — 81.0% end-to-end false negatives (95% CI 73.8–87.4) vs Anthropic's self-reported 17%. *Adjusted (ADDED):* absolute gates are phrased as inviolable constraints and backed by the deterministic layer (deny permissions, hooks) where one exists — never left to model judgment.

### Honesty notes
- **The ops/SRE and failure-case pillars produced NO verified claims.** Google SRE break-glass/canarying, SEC 2013-222 (Knight Capital), risk-engineering.org (AF447), and FlightGlobal (Asiana 214) were fetched but their claims did not survive the verification cut (budget: top-25 of 117). Break-glass-audit and auto-rollback elements rest on analogy to the sources above — do NOT cite SRE literature for this rule without a dedicated verification pass.
- **Refuted (1-2 vote):** the crisp "EU usage" dichotomy (HITL = validate every output / HOTL = monitor with intervention ability). Loop terms anchor to the CRS/DoD definitions above; do not attribute that dichotomy to EU governance usage.
- All non-vendor support is analogical transfer from other domains (teleoperators, driving, weapons, high-risk EU AI systems) — verifiers rated the mappings faithful, but they are design precedent, not binding prescription.



## Code-search routing — jbcontext/CodeGraph/rg/agent (2026-07-21)

> **CodeGraph fue RETIRADO el 2026-08-19 — desinstalado del hive y de los harnesses.** Motivo medido, no
> preferencia: su resolución de imports adivina cuando falla y no marca la adivinanza. En
> `educavita-next-monorepo`, `@/components/app-frame` de school-portal quedó apuntando al componente del
> backoffice (arista `imports` sin `provenance`, más su gemela `calls` marcada `heuristic`), y con ella todo
> el shell de esa app. Causa: sólo lee el `tsconfig.json` de la raíz del índice — el de cada app es invisible
> — y el desempate entre homónimos no favorece al importador. El compilador de TypeScript resuelve ese mismo
> specifier correctamente (`ts.resolveModuleName` contra el tsconfig de la app), así que la sustitución
> apunta a herramientas apoyadas en el type checker (LSP/SCIP). Los benchmarks de abajo siguen siendo
> válidos para lo que midieron; no cubrían monorepos con apps gemelas.

> **jbcontext está APARCADO desde 2026-08-18 — decisión del usuario, no un veredicto de los benchmarks.** Se retiró de la tabla de routing (`code-search.md`), de la detección de los hooks `bash-policy`/`post-tool-hub`/`rule-context`, del agente `code-scout` y de los punteros del core. La operación que cubría —descubrimiento por intención, terminología desconocida, código legacy/untyped— la absorbe `rg` con barrido amplio de vocabulario más el agente `code-scout`; en repos legacy declarados el hook sigue denegando CodeGraph, ahora enrutando a `rg`. La herramienta existe y sigue instalada como servidor MCP a nivel máquina (fuera de este repo): reactivarla es volver a agregar su fila de routing y su detección en los tres hooks. La evidencia comparativa que motivó su adopción original queda abajo, intacta.

- `[Internal — measured]` **tgrep workspace server (2026-09-07)** — backs the workspace row of `code-search.md` and the `tgw` bullet in `code-scout`. Upstream: [microsoft/tgrep](https://github.com/microsoft/tgrep) v1.0.4 (`BENCHMARKS.md`: 4-52× over ripgrep on 100K-500K-file repos). Local: ark 1.8K files 45→5 ms, workspace 18.7K files 500→10 ms, 8 concurrent searches 3.6 s→19 ms; output parity identical; index stale without the server. *Adjusted:* adopted only for cross-repo sweeps and fan-outs via a launchd server — never as an `rg` replacement. Method and expansion criteria: `_support/docs/tgrep-workspace-server.md`.
- `[Internal — measured]` **Benchmarks v3 + v4** (`_support/archive/audits/2026-07-21-benchmark-{v2-ark-jbcontext-vs-codegraph,v3-discovery-multiworkspace,v4-toolset-ablation}.html`) — evidence base for `global/rules/tools/code-search.md`, the `code-search-routing` hook, and the `code-scout` agent. v3: 27 tasks / 4 workspaces / 2 judges per task; single-shot quality jbcontext 5.9 ≈ rg 5.9 > CodeGraph 4.5, explore agent 9.2 (with its first verified factual error). v4 (ablation, sonnet explorers, 4 arms): tools complement — D(both) 8.53 > max(B,C); jbcontext +1.94 on legacy/unknown terminology; CodeGraph +0.72 on structural traps but −0.34 on legacy; ambiguous questions: baseline wins. Operational hazards verified: jbcontext repository id = normalized git remote (non-git dirs collapse into ONE shared id — silent wrong-index serving; same-remote twins indistinguishable; synthetic-remote + `--git-remote-url` workaround validated); non-git indexing ingests node_modules; no client-side index deletion exists (EAP §8 covers deletion on withdrawal only).
- `[Vendor claim — unvalidated here]` **JetBrains Context launch post** (blog.jetbrains.com/ai/2026/07/introducing-jetbrains-context…) — claims −68% turns / −59% latency / −48% cost on agent workflows; orthogonal to our quality axis (task-completion efficiency vs retrieval quality), not contradicted but not reproduced; its "source code is not stored" claim needs the nuance measured above (hosted vector index, no user deletion).


## Context engineering for Claude 5 models — the rightsizing pass (2026-07-31)

Backs the `refactor/claude5-rightsizing` change-group: the `loadedBy:` rule scope, the always-on trims, and the judgment cuts in `global/rules/**` and `global/agents/**`.

### Sources
- `[Semi-authoritative]` **Thariq Shihipar (Anthropic), *The new rules of context engineering for Claude 5 generation models*** (claude.com/blog, 2026-07-24) — reports removing **>80% of Claude Code's system prompt** for Opus 5 / Fable 5 with no measurable loss on coding evals. Five inversions: rules → judgment · examples → interface design · everything upfront → progressive disclosure · repetition → simple tool descriptions · CLAUDE.md memory → auto-memory. Plus: simple specs → rich references (code, test suites, HTML artifacts, rubrics).
- `[Semi-authoritative]` **Thariq, *A field guide to Claude Fable 5: finding your unknowns*** (2026-07-06) — work quality is bottlenecked by the human's ability to clarify unknowns. Techniques: blindspot pass, interview, references-as-code, implementation notes during the build, explainer + quiz after.
- `[Semi-authoritative]` **Anthropic Engineering, *Effective context engineering for AI agents*** — "context rot"; find the smallest set of high-signal tokens; just-in-time retrieval vs pre-loading; the system-prompt "Goldilocks zone" between brittle hardcoding and vague guidance that falsely assumes shared context.
- `[Semi-authoritative]` **Anthropic, *A harness for every task: dynamic workflows in Claude Code*** — adversarial verification against a rubric; workflows as template, not script.
- Article copies: `_support/workspace/articles/` (ephemeral — the durable record is this section).

### Verdicts reached — *adjusted*, and one premise found false

- **The >80% figure does not transfer, and the reduction we did achieve on rules is ~0%.** Anthropic pruned a *product* system prompt written for older models; this hub is mostly encoded user opinion, which the same article says SHOULD live in config.
- **The `loadedBy:` scope was a false premise — recorded here so it is not reinvented.** A wave shipped 5 rules under an invented `loadedBy: <skill>` frontmatter key on the assumption it made them conditional. It does not. Claude Code reads **only `paths:`**; the docs state *"rules without a `paths` field are loaded unconditionally"*, and an unrecognized key does not suppress loading. `alwaysApply: true`, used throughout this repo, is likewise documentation rather than mechanism — those files load because they lack `paths:`. The wave was reverted: measured always-on went 121,061 → 121,040 chars, i.e. unchanged. **The process lesson is the durable one: the harness mechanism was never verified before a whole wave was built on it, and the repo's own rule — verify cheaply before reasoning expensively — named exactly that check.**
- **Most of the bloat was duplication, not generic mantras.** 17 duplication groups and ~180 restated lines across 24 agents. The judgment pass over rules yielded −0.8%: nearly every candidate corrected a real agent tendency (over-abstraction, over-building, speculative memoization) rather than restating textbook craft.
- **Where progressive disclosure IS real: agents and skills.** Those load on invocation by design, no frontmatter trick needed. Agents 1,085 → 1,017 lines; the four flow skills' bodies 584 → 165 with the detail in `references/`; `deploy-global` 308 → 91 with the procedure in a real script.
- **Progressive disclosure has a precondition this architecture imposes.** Two adversarial passes (`finding-refuter`) broke the "nothing safety-bearing became conditional" claim twice — the `data deletion` gate, and the prompt-injection guard for untrusted page content. Root cause both times: rule *loading* changed without touching *consumers*. An executor subagent whose `tools:` allowlist omits `Skill` cannot invoke a router skill, and skills are not inherited from the parent.
- **A rule that fires on an ACTION with no file footprint cannot be path-scoped** — deploying (Vercel/Dokploy touch no IaC file), driving a browser, a live incident. Those stay always-on.
- **"Examples constrain the exploration space" applies to tool/usage examples, not routing disambiguation.** 5 of 6 agent `<example>` blocks were cut; `security-reviewer` keeps its one, because `code-reviewer`'s description also claims security and nothing else states that boundary.
- **Floor-model calibration outranks the articles.** They measure on Opus 5 / Fable 5; 13 of 24 agents run `sonnet`. Declined cuts are recorded in commits `f0d1c91` and `0f5b214` — that is the reversal list if a Sonnet-tier agent regresses.

### Not adopted in this pass
Blindspot pass, the interview step, `implementation-notes.md` during build, and the post-implementation explainer + quiz from the field guide. Deliberately deferred to a later change-group so the pruning stays measurable on its own.


## Code-smell rules — taxonomy, deterministic layer & LLM-amplified smells (2026-08-11)

Backs `languages/typescript-standards.md > Type Safety` (type-system-as-source-of-truth bullet) and `> Linting & Formatting` (smell-backstop stack), `quality/development-principles.md` ("A change leaves no residue"), `languages/nestjs-patterns.md > DTOs & Validation` (Response DTO derivation), the `code-reviewer` "TS type smells" catch, and the comment-discipline line in `harness/AGENTS.md`. Researched ago-2026 (4 parallel research agents + a 6-repo local recurrence sweep; synthesis: `_support/workspace/anti-smell-rules-research-2026-08-11.html`).

### Canonical taxonomy (vocabulary + selection filter, not rule content)
- `[Authoritative]` **Fowler, *Refactoring* 2e (2018), ch. 3 "Bad Smells in Code" (with Beck)** — the 24-smell catalog; the vocabulary layer (LLMs know these names).
- `[Authoritative]` **Ousterhout, *A Philosophy of Software Design*, red-flag boxes (per-chapter)** — 14 red flags, deliberately judgment-dominant (only Repetition and partially Pass-Through Method are mechanizable): the natural complement to Fowler for review-layer rules a linter can never own.
- `[Semi-authoritative]` **Martin, *Clean Code* ch. 17 "Smells and Heuristics"** — of 66 items only G4 (Overridden Safeties — the escape-hatch rule's 2008 ancestor), G6, G8, G11, G31 are distinct and current; the function-size dogma is disputed (incl. by Ousterhout) — cite as checklist source only.
- **Verdict: adopted as selection filter.** The triple filter that decided what ships: a smell earns a prompt rule only if judgment-only AND LLM-amplified AND uncovered by existing rules.

### Deterministic layer — smells that do NOT earn prompt rules
- `[Authoritative]` **typescript-eslint v8.67 rules index** — `no-unnecessary-condition` (strict-type-checked; requires `strictNullChecks`; FPs where ORM types overclaim) catches type-distrust deterministically; **no rule detects parallel type shapes** (verified against the full index); `prefer-optional-chain` collapses null-chains only before a member access.
- `[Authoritative]` **Biome v2.5** — `noUnnecessaryConditions`/complexity opt-in, own type-inference engine; no duplication detection (2026) — pair with jscpd.
- `[Authoritative]` **eslint-plugin-sonarjs 4.2.0 (active; 279 rules)** — cherry-pick: `cognitive-complexity` (S3776), `no-identical-functions`, `no-gratuitous-expressions`, `no-selector-parameter`.
- `[Authoritative]` **jscpd 5 (Rust engine)** — token-level clone detection, SARIF for GitHub Code Scanning.
- **Verdict: supported.** Mechanizable smells route to lint config per the enforcement-layer doctrine; prompt budget is reserved for the seven verified judgment-only smells (parallel types, bare null-chains, speculative generality, wrong abstraction, assertion laundering, overclaiming types, Rule-of-Three judgment).

### LLM-amplified smells — why the prompt rules target what they target
- `[Semi-authoritative]` **GitClear reports 2024/2025/2026** — block duplication +81% (2023–26), error-masking constructs +47%, cross-file reuse −35%, copy/paste ~5× refactoring. **Correlational** (no per-commit AI attribution; vendor interest) — directional, never causal.
- `[Semi-authoritative]` **Liu et al., "Debt Behind the AI Boom" (arXiv:2603.28592, 2026)** — 302.6k AI-attributed commits: smells = 89.3% of detected issues; 22.7% persist. Best attribution methodology; grounds the dead-code/residue rule.
- `[Semi-authoritative]` **Zhang et al., "Copilot-in-the-Loop" (arXiv:2401.14176, 2024)** — self-fix up to 87.1% when the smell is NAMED in the prompt: the evidence that concrete named rules work and a generic "avoid smells" rule does not. This number is the operational justification for principle-with-anchors phrasing.
- `[Authoritative]` **Anthropic, Claude Code best practices** — "Don't add error handling, fallbacks, or validation for scenarios that can't happen — trust internal code and framework guarantees": vendor prescription matching "trust your types". OpenAI GPT-5.x guides prescribe minimal comments likewise.
- `[Semi-authoritative]` **Böckeler, "The role of developer skills in agentic coding", martinfowler.com (2025)** — named-practitioner taxonomy: no-reuse, over-engineering, brute-force fixes.
- `[Non-authoritative]` **OX Security "Army of Juniors" (2025)** — excessive commenting in 90–100% of AI repos, over-specification 80–90%; repo-level attribution is heuristic — corroboration only.
- **Verdict: supported.** Local corroboration (2026-08-11 sweep, 6 repos): parallel-type duplication in 5/5 repos with surface, drift bugs already present in remuneri and recruitment; type-distrust 15 dead checks across 3 repos; the `!== null && !== undefined` chain in 12 repos.

## Specs repo structure (flow-core/references/specs-structure.md)

- `[Semi-authoritative]` **Warp specs organization** (github.com/warpdotdev/warp/tree/master/specs, extracted 2026-06-11) — origin of the layout core: one folder per work item named by tracker ID, product spec separated from tech spec, tech specs citing real code paths for verifiability. The `product/` layer (business truth in force; epics as deltas) is our extension — Warp's specs are dev-facing only. Attribution moved here from the reference doc (2026-08-12 claims audit).

## Hook registration per harness (deploy-global hooks category, 2026-08-13)

- `[Authoritative]` **Claude Code hooks docs** (code.claude.com/docs/en/hooks, verified 2026-08-13) — hooks are registered EXCLUSIVELY through the `hooks` configuration in settings files; there is no auto-scanning of `~/.claude/hooks/`. A JSON file in that directory with no settings entry is inert. Grounds: the flat hooks copy in `deploy-global.sh` exists only for runtime config a registered script READS from there (`bash-policy.json`), never as a registration mechanism.
- `[Authoritative]` **Codex CLI hooks docs** (learn.chatgpt.com/docs/hooks, verified 2026-08-13; re-verified 2026-08-20 — every `developers.openai.com/codex/*` URL now redirects here, cite the destination host) — hook support is NATIVE: Codex discovers `hooks.json` next to its active config layers (`~/.codex/hooks.json`, `<repo>/.codex/hooks.json`) or inline `[hooks]` tables in `config.toml`; it never reads `~/.claude/`. Grounds: the per-hook `codex-hooks.json` sources are merge INPUT for `~/.codex/hooks.json` (same class as `settings-config.json`) and are excluded from the flat copy — a flattened copy is inert to every harness, and two hooks' files collide on the basename (fix `07992ac`; live corroboration: `codex --strict-config doctor: OK`).
- **Verdict: supported.** No prompt rule derived — the behavioral lesson (validate an assumed mechanism against official docs before building on it) is already owned by `tools/context7.md` and `quality/critical-thinking.md > Research-driven decisions`; per the no-rules-from-single-incidents policy, this entry records the sources, not a new directive.

## Diagram grammar inside flow-report (references/diagram-grammar.md, 2026-08-13)

- `[Semi-authoritative]` **cathrynlavery/diagram-design v2.3.2** (github.com/cathrynlavery/diagram-design, MIT, extracted 2026-08-13) — source of the SVG diagram grammar adopted into flow-report `references/diagram-grammar.md`: mandatory orthogonal connectors (elbow/bridge/attach-point-fanning formulas), label masking + 6–10px gap, complexity budgets, anti-pattern catalog, accessible-SVG contract, node/legend/zone primitives. Adopted selectively — grammar re-skinned onto the baseline tokens, system font stacks replacing its Google Fonts (self-containment) — NOT installed as a plugin: webfont CDN breaks flow-report rule 1, and its default skin (white-smoke + atomic-tangerine + Instrument Serif/Geist) clashes with the report design system. **Re-sync point:** upstream iterates fast; on refresh re-read its `skills/diagram-design/SKILL.md` §4–§9 + `references/type-*.md` and re-map. Deliberately un-adopted: Mermaid/draw.io import, animation, export, brand onboarding, long-tail types.
- `[Semi-authoritative]` **upstream `scripts/verify-geometry.py` + `skills/diagram-design/scripts/self_check.py` (same repo/version)** — adapted 2026-08-13 into `global/skills/flow-report/scripts/` as the deterministic post-generation gate (invocation is prompt-convention in SKILL.md > After writing). Local divergences, recorded in each file's provenance header: Google Fonts exemption removed (any remote reference fails); `verify_geometry` scopes the mask/node overlap check per `<svg>` (reports embed several diagrams — upstream assumed one per file); motion-contract path stripped from `self_check` and the script check re-scoped to SVG containment (page-level JS is legitimate in reports; a `<script>` inside an `<svg>` fails); mandatory-SVG requirement removed from `self_check` (2026-08-13 — upstream validates single-diagram files where a missing SVG is a broken deliverable, but flow-report runs the script on EVERY report and diagram-free reports are legitimate; the per-SVG accessible contract stays). On re-sync, diff against upstream knowing these five divergences are deliberate.
- **Verdict: adopted as re-skinned grammar layer + deterministic check pair.** One reference file owns the craft; flow-report SKILL.md rule 8 makes loading it mandatory before drawing any inline SVG diagram, and the After-writing step runs both checks on every report.

## flow-report format archetypes (Format selection + skeleton-*.html, 2026-08-13)

- `[Semi-authoritative]` **Thariq (Anthropic), "Using Claude Code: The Unreasonable Effectiveness of HTML"** (claude.com/blog/using-claude-code-the-unreasonable-effectiveness-of-html + gallery thariqs.github.io/html-effectiveness, extracted 2026-08-13) — source of the format-by-intent evolution: format follows what is being PRESENTED, not one chassis for everything. Adopted: five-archetype selection layer in flow-report SKILL.md (document/explainer/review/comparison/deck), four new self-documenting skeletons beside `baseline.html`, explicit presentation-interactivity sanction (tabs/collapsibles/modals/deck-nav — JS that reveals content already in the file), multi-file staged deliverables under one dated folder. NOT adopted: the custom-editor family and any live-state binding (playground plugin's domain — boundary in `communication-format.md` unchanged); the author's HTML-maximalist stance (Markdown carve-outs in `communication-format.md` stand). Architectural pattern mirrored from in-repo `starlight-docs-site` (profile × chassis × validator) rather than the article.
- **Verdict: adopted as archetype layer.** Selection table + one-liners in SKILL.md (ships to Codex/opencode via the mirror); per-archetype guidance lives as header/inline comments in each skeleton — zero-drift, no prose pairs.

## Harness context-loading mechanics (AGENTS.md budget check, CLAUDE.md fallback note, convention placement — 2026-08-13)

- `[Authoritative]` **Codex CLI measured behavior** (codex 0.147.0, `codex debug prompt-input` sandboxes, reproduced 3× independently — Gen B, Codex refuter, main thread, 2026-08-13 adversarial-research run) — `project_doc_max_bytes` budgets ONLY the project chain (git root → cwd); the global `~/.codex/AGENTS.md` is exempt. On overflow the crossing file is truncated mid-content and deeper files are dropped entirely, with no warning. Codex discovery never crosses upward past the git root (a non-git workspace-root `AGENTS.md` is invisible from inside a child repo); its default sandbox permits sibling-directory READS. Grounds: the corrected `wc -c` comment in root `AGENTS.md` > Replicating Global Harness Config.
- `[Authoritative]` **Claude Code memory docs + empirical check** (code.claude.com/docs/en/memory, verified 2026-08-13; sandbox `claude -p` marker test, negative result) — Claude Code does NOT load a repo's `AGENTS.md` when `CLAUDE.md` is absent (the docs offer only the import route: "create a CLAUDE.md that imports it"); import parsing skips Markdown code spans and fenced code blocks, so a backtick-wrapped `` `@path` `` is inert (root cause of sample-project-backoffice's never-loaded 1540-line guideline). Grounds: the "No AGENTS.md fallback" note in root `CLAUDE.md` > Notes.
- `[Authoritative]` **Grok CLI shipped docs + `grok inspect`** (`~/.grok/docs/user-guide/12-project-rules.md`, measured 2026-08-13) — README.md and arbitrary paths (`_support/conventions/*.md`) are auto-loaded by zero of the four harnesses; Grok's roaming discovery detects only its six recognized instruction filenames. Corroborates the existing flat-symlink Grok deploy design. **Qualified 2026-08-20:** `paths:` ignored holds as *documentary absence* (no source mentions any file-pattern scoping key); **non-recursive scan is assumed, never confirmed** — neither the public docs nor the bundled guide state whether the scan descends into subdirectories. Settle it with a `grok inspect` on a nested rules file before citing it as fact.
- **Verdict: supported.** Grounds the 2026-08-13 corrections (root `AGENTS.md`, root `CLAUDE.md`, `project-structure.md` scope-vs-versioning routing) and the convention-placement decision: repo-scoped conventions versioned in `<repo>/_support/docs/`, multi-repo in `<project>-specs/conventions/`, declared via a pointer block in each repo's `AGENTS.md`. Full canon: `_support/workspace/2026-08-13-conventions-adversarial-research.html`.

## CI stage placement — quick gate / full suite / release gate

Backs `quality/testing.md > Execution Scope` (full-suite placement), `workflow/devops-principles.md > CI is staged by seam`, `languages/iac-devops.md > Stage workflows by seam`, and `flow-core/references/promotion-playbook.md > Pre-gates`. Researched 2026-08-14 via adversarial protocol (3 independent generators + 1 cross-examiner; 61 claims: 49 confirmed / 9 weakened / 1 refuted / 1 unverifiable / 1 empty slot). Full canon with verbatim citations: `_support/docs/ci-stage-canon.md`.

- `[Authoritative]` **SWE at Google, ch. 23 (Continuous Integration)** — presubmit = fast/reliable tests gating submit ("if the tests pass, the change is allowed into the codebase"); TAP postsubmit asynchronously runs all affected tests incl. larger/slower ones; RC promotion through environments serves "sanity check / auditability / cherry picks", not the primary correctness gate. *Supported.*
- `[Authoritative]` **DORA capability articles** (continuous-integration, trunk-based-development, working-in-small-batches, test-automation, streamlining-change-approval) — tests run "both before and after the merge"; <10-minute feedback; "no evidence" that a formal external review stage lowers change-fail rates. *Supported.*
- `[Authoritative]` **Fowler — Continuous Integration; Patterns for Managing Source Code Branches; Ship/Show/Ask** — fast commit build vs slower secondary suite that "may not run after every commit"; environment branches labeled an anti-pattern (the mapping development=trunk, qa/production=release-branch flow is harm reduction, not endorsement); blocking review is risk-tiered. *Supported / adjusted as noted.*
- `[Authoritative]` **Humble & Farley, deployment pipeline patterns (continuousdelivery.com)** — "only build packages once"; fast commit stage; later stages verify the same artifact; deploy the same way to every environment. *Supported.*
- `[Semi-authoritative]` **minimumcd.org; trunkbaseddevelopment.com** — automated testing before merge AND on merge; release branches receive no development work; trunk-first fixes cherry-picked forward. *Supported.*
- `[Semi-authoritative]` **Micco, "The State of CI Testing @Google" (deck); SmartBear/Cisco review study; platform docs (Azure release gates, GitLab merge trains, GitHub merge queue); AI-reviewer vendor docs** — flakiness must be budgeted structurally; review effectiveness ceiling at 200–400 LOC; promotion gates consume signals rather than produce test verdicts; AI reviewers attach to `pull_request` on the integration branch by design. *Supported.*
- **Verdict: adjusted.** The build-only-in-full-suite placement is a portfolio cost decision layered on the canon (consistent with Google's presubmit cost rationale, with typecheck as the pre-merge compile proxy and red-trunk revert discipline as the backstop) — not itself a canon mandate.

## Contract distribution — the versioned artifact as cross-repo handshake

Backs `workflow/cross-service-workflow.md > Contract distribution`, the distribution rules in `agents/design/system-designer.md`, and `quality/development-principles.md > Prior art before infrastructure`. Recorded 2026-08-17.

**Verification status: UNVERIFIED — recorded from model knowledge, not a live research pass.** Names and roles are believed correct; precise loci, editions, and current maintenance status are NOT confirmed, contrary to this file's citation convention. Run a docs pass or `/adversarial-research` before treating any entry as settled.

**Thesis the rule encodes:** a release channel (immutable, versioned, auditable) and a development channel (mutable, disposable) are different mechanisms; using the release channel for the edit loop is what forces a hand-built bridge between edited and published bytes. Maven's `-SNAPSHOT` is the native two-channel form; npm has no equivalent, which is why the JS ecosystem produced local registries and snapshot/canary publishing.

**Origin of the rules:** the failure pattern, not a completed source review — `ark` and `sample-project`, the only two projects with a contract registry, both hand-built the delivery mechanism (2/2), and neither surfaced the ecosystem equivalents to the user.

### Patterns
- `[Authoritative]` **Sam Newman, *Building Microservices* (2nd ed., O'Reilly)** — integration and contract chapters: versioned contracts, deployment coupling, consumers not waiting on a producer's release.
- `[Authoritative]` **Ian Robinson, "Consumer-Driven Contracts: A Service Evolution Pattern"** (martinfowler.com) — CDC runs in the *opposite direction* to a producer-published schema package (tooling: Pact, Spring Cloud Contract). Complementary, not the pattern the rule encodes.
- `[Authoritative]` **Danilo Sato, "Parallel Change" (expand-contract)** (martinfowler.com) — evolving a contract without a coordinated cutover. Not yet covered by any rule; candidate gap.
- `[Authoritative]` **SemVer 2.0.0 §9** — prerelease identifiers, the mechanic behind the prerelease handshake. Maven's `-SNAPSHOT` is the native equivalent; npm has none, hence prerelease/canary conventions.
- `[Semi-authoritative]` **Confluent Schema Registry** (compatibility modes BACKWARD/FORWARD/FULL) and **Buf Schema Registry** — the productized form of a versioned contract registry with breaking-change detection and generated-SDK publication.
- `[Authoritative]` **Forsgren, Humble & Kim, *Accelerate*** — loosely coupled architecture and independent deployability as measured capabilities; the empirical case against "producer merges first, consumer follows".

### Ecosystem tooling — the prior art to check before hand-building
- **yalc** — local publish/install store; the supported form of a `node_modules` overlay.
- **Verdaccio** — local/proxy npm registry for unpublished bytes.
- **Changesets snapshot releases** — publishes without burning a version number.
- **pkg.pr.new** — installable package per commit/PR without publishing to the real registry.

**Verdict: pending.** Sources unread this session; the rules stand on the observed failure pattern until a research pass confirms or adjusts them.

## Tamaño y forma de los archivos de instrucciones (2026-08-21)

- `[Authoritative — vendor]` **Anthropic, "The new rules of context engineering for Claude 5
  generation models"** (claude.com/blog, 24-jul-2026; Thariq Shihipar, technical staff):
  removieron **>80% del system prompt de Claude Code** para Opus 5 y Fable 5 **sin pérdida
  medible en sus evals de coding**. Prescripciones aplicables a `CLAUDE.md`/`AGENTS.md`:
  mantenerlo *"lightweight and briefly describe what your repo is for"*, centrado en
  **gotchas** (su ejemplo: *"you may organize your code to keep types in one monolithic file
  and nowhere else"*); cortar instrucciones repetidas que ya viven en las descripciones de
  herramientas, guardarraíles diseñados para el peor caso de modelos viejos, reglas
  prescriptivas de estilo, e **instrucciones contradictorias entre sí** (su ejemplo:
  *"leave documentation as appropriate"* conviviendo con *"DO NOT add comments"*).
  **Reglas → criterio:** dejar de escribir "never do X" salvo modo de fallo demostrable —
  antes *"Never write multi-paragraph docstrings…"*, ahora *"Write code that reads like the
  surrounding code: match its comment density, naming, and idiom"* (esta última frase está
  hoy en el system prompt de Claude Code, o sea que la aplicaron a sí mismos). **Progressive
  disclosure:** organizar `CLAUDE.md` y skills como *"a tree of files that can be loaded at
  the right time"*. Herramienta: `/doctor` en sesión.
- `[Authoritative — peer-reviewable research]` **Chakrabarti, K. "Why Does CLAUDE.md Keep
  Growing? Catastrophic Remembering in Agentic Coding"** (arXiv:2608.11095v1, 11-ago-2026,
  South Park Commons). Mide **247,694 vidas de instrucción en 1,867 repositorios**: los
  prompts agénticos crecen **+226%** sobre su vida (+4.9 instrucciones netas por commit) y
  **cuanto más vieja es una instrucción, MENOS probable es que se borre** (log-hazard
  −0.032/commit). Causa: recall imperfecto — añadir es barato, pero borrar sin arriesgar una
  regresión cuesta `O(2^|D|)` porque exige recordar por qué se añadió. Lo llama **catastrophic
  remembering**, el inverso del olvido catastrófico. **Remedio medido:** comentarios que
  codifican el razonamiento latente eliminan el **99.3% del exceso** (+211.3% → +1.4%) con la
  misma corrección, y mejoran el seguimiento de instrucciones real hasta **+23.1%**
  (WildIFEval). Crítico: los *comment-shaped noise* — comentarios que no codifican el
  razonamiento — crecen casi igual que no tener ninguno (fig. 1a), así que la forma no basta.
- `[Empírico — medido 2026-08-22]` **No existe un canal de comentario invisible al modelo.**
  Un `<!-- ... -->` en un `AGENTS.md` llega íntegro al prompt (verificado con `codex debug prompt-input`,
  marcador único recuperado 1/1): ningún harness
  parsea markdown para descartarlo, el archivo se inyecta crudo. Corrige la lectura
  divulgativa del paper (*"the agent never reads it, the next person does"*), que además
  contradice su propio +23.1% de instruction-following: si mejora el cumplimiento, el modelo
  lo lee. Consecuencia: un comentario cuesta contexto en TODOS los harnesses y `<!-- -->` no
  ahorra un token frente a texto plano.
- **Consecuencia aplicada (2026-08-22):** `agents-md-primary` recupera el rationale
  (`git log -S`/blame + ticket) ANTES de proponer `delete`/`demote`/`soften`, y el manifiesto
  distingue "la razón sigue vigente" de "la razón se perdió o nunca se registró". **Tensión
  resuelta con `AGENTS.md > Concise-first`,** que prohíbe justificaciones y notas de
  procedencia inline: la forma la decide la dificultad de la regla — una línea de razonamiento
  inline donde el modelo cumpliría mal sin saber el porqué (compra cumplimiento, no solo
  borrabilidad), un PUNTERO (ticket, commit) donde la regla es obvia y solo debe poder
  retirarse mañana. En este repo el
  puntero natural es `git blame`, dado que los mensajes de commit llevan el incidente.

- `[Comunidad — guía práctica]` **aihero.dev, "A Complete Guide to AGENTS.md"**
  (consultado 2026-08-22). Aporta el **piso positivo** que la literatura de poda no da: el root
  guarda una frase de qué es el proyecto, el package manager si no es el default del
  ecosistema, y los comandos de build/typecheck NO estándar — *"honestly it. Everything else
  should go elsewhere."* Excluye reglas por lenguaje, inventarios de estructura (se ponen
  stale: describe capacidades) y lo obvio o vago. Monorepos: `AGENTS.md` por paquete SÍ,
  mínimos, y los anidados se concatenan con el root. **Error a no adoptar:** afirma que Claude
  Code «no usa AGENTS.md» y recomienda **symlinkear** ambos archivos — el symlink obliga a que
  sean idénticos y elimina el espacio para contenido Claude-específico; el patrón correcto es
  `CLAUDE.md` con el import `@AGENTS.md` (lo que aplica `agents-md-primary`).

- `[Consenso de comunidad — NO vendor]` Varias guías de 2026 convergen en ~150 líneas para
  `AGENTS.md` y ~200 para `CLAUDE.md`, con dos fundamentos: los modelos frontera siguen de
  forma fiable ~150–200 instrucciones (el system prompt del harness ya gasta parte), y el
  *context rot* medido sobre 18 modelos frontera. **Trátese como umbral de REVISIÓN, nunca
  como gate**: ningún vendor publica esa cifra.
- **Consecuencia aplicada (2026-08-21):** `agents-md-primary` incorpora el test de admisión
  (gotcha o se va), los outcomes `demote` (progressive disclosure) y `soften`
  (regla→criterio), el scope a `AGENTS.md` anidados dentro de un repo, y la métrica por
  CADENA medida con `codex debug prompt-input` en vez de bytes por archivo.
- **Medición de referencia (2026-08-21, sample-project/apps/backend):** `codex debug prompt-input`
  = 85,166 B (~21.3k tokens) de prompt inicial, de los que 58,069 B (68%) son la cadena de
  `AGENTS.md`: 35,170 B del core del hive (global, se paga en TODOS los proyectos), 14,046 B
  de la app y 9,082 B del root del monorepo. En Claude Code, `/doctor` mide lo mismo por otra
  vía: `~/.claude/CLAUDE.md` = 7,958 tokens, 55% del contexto residente, cargado en 69
  proyectos. **El core global es el mayor contribuyente en ambos harnesses.**

## Per-harness loading mechanics — official sources (2026-08-20)

Backs the *What this harness loads* tables in `global/README.md` and
`harness/{codex,opencode,grok}/README.md`. Every URL below was fetched and its quote taken
from the page body, not from a search summary. The READMEs carry the same URLs inline so a
reader checking for upstream drift never has to leave the harness's own folder; this
section records the authority tags and the findings that the tables only flag.

### Claude Code
- `[Authoritative]` **code.claude.com/docs/en/memory** — `#how-claude-md-files-load`,
  `#import-additional-files`, `#user-level-rules`, `#path-specific-rules`. The last one is
  the locus for conditional loading: *"Rules without a `paths` field are loaded
  unconditionally and apply to all files."* **There is no `/docs/en/rules` page** — the
  entire `.claude/rules/` mechanism lives inside `memory`; a link to `/docs/en/rules` is a
  404.
- `[Authoritative]` **code.claude.com/docs/en/sub-agents#supported-frontmatter-fields**,
  **/skills#frontmatter-reference**, **/hooks#hook-locations**,
  **/settings#settings-precedence**.

### Codex CLI
- `[Authoritative]` **learn.chatgpt.com/docs/agent-configuration/agents-md** — *"stops
  adding files once the combined size reaches the limit defined by project_doc_max_bytes
  (32 KiB by default)"*. Also **/config-file/config-reference**,
  **/agent-configuration/subagents**, **/build-skills** (`$HOME/.agents/skills` +
  `allow_implicit_invocation`), **/hooks**, **/developer-commands?surface=cli**
  (`codex debug prompt-input`).
- **Host move, 2026-08-20:** every `developers.openai.com/codex/*` URL redirects to
  `learn.chatgpt.com/docs/*`. Cite the destination.
- **Two surfaces this repo depends on are undocumented:** `[hooks.state]` enable slots (the
  documented switch is `[features] hooks = false`, with `codex_hooks` a deprecated alias)
  and `multi_agent = true` (the documented subagent surface is the `[agents]` table). Both
  were established empirically; they are flagged as undocumented in
  `harness/codex/README.md` rather than given a citation.
- The `65536` seen in the official page is an example of *raising* the cap, not the default
  — consistent with the 2026-08-18 section below.

### opencode
- `[Authoritative]` **opencode.ai/docs/rules**, **/agents**, **/commands**, **/plugins**,
  **/permissions**, **/skills** (*"Global agent-compatible:
  `~/.agents/skills/<name>/SKILL.md`"* — opencode reads both `~/.agents/skills` and
  `~/.claude/skills/`).
- `[Semi-authoritative]` **github.com/frap129/opencode-rules** — *"`globs` (optional):
  Array of glob patterns for file-based matching"*; `match` selects the combination mode.
  Latest release **v0.6.4 (2026-04-25)**: the pin in this repo is current, not stale.
- **Org move, 2026-08-20:** `github.com/sst/opencode` redirects to
  `github.com/anomalyco/opencode`. Write `anomalyco`.
- **Undocumented surface:** `experimental.chat.system.transform`, which
  `flow-session-context.ts` uses to inject SessionStart context, does not appear in the
  plugins docs; the only documented `experimental.*` hook there is
  `experimental.session.compacting`. Read from source, flagged as such in
  `harness/opencode/README.md`.

### opencode

- `[Empirical — reproducible]` **The entire opencode plugin layer fails to load (measured
  2026-08-21, opencode `1.18.18`).** Method: `grep "failed to load plugin"` over
  `~/.local/share/opencode/log/opencode.log` — 160 load attempts, 160 failures, current runs
  included; cause `SchemaError: Missing key at ["default"]` (rules plugin:
  `["default"]["effect"]` / `["setup"]`). Confirmed end to end by a headless run in a ledger
  workspace answering **AUSENTE** when asked for the `flow-process-protocol` marker — so the
  failure is the load, not the injection surface. Affects `opencode-rules@0.6.4` (and with it
  every path-scoped rule deployed to `~/.config/opencode/rules/`), `flow-session-context.ts`,
  `engram.ts`, and two third-party plugins. **Diagnosis:** `opencode2` (`v0.0.0-beta-17577`,
  the 2.0 preview) shares `~/.config/opencode` and uses capability-named plugins
  (`opencode.config.instruction`, `.skill`, `.policy`, `.agent`) — the 1.x schema error is
  that contract migration arriving. **Status: open follow-up**, recorded in
  `harness/opencode/README.md`; the target version is the decision that comes first.

### Grok CLI
- `[Authoritative]` **docs.x.ai/build/overview** (`grok inspect`),
  **/build/features/project-rules**, **/build/features/subagents**, **/build/features/hooks**,
  **/build/features/skills-plugins-marketplaces** — the last one carries both *"Grok
  automatically reads Claude Code marketplaces, plugins, skills, MCPs, agents, hooks, and
  instruction files"* and *"`disable-model-invocation`: Slash command only; no automatic
  invoke. Default `false`."* — i.e. the `flow-*` gate holds in Grok with no translation.
- **Finding: Grok has public web documentation.** This repo previously treated its docs as
  bundled-only (`~/.grok/docs/user-guide/*`). The bundled guide remains authoritative for
  the scope/precedence tables and is version-stamped with the CLI, but it is a local file,
  never a citable URL.
- `[Empirical — reproducible]` **Hook context injection in Grok: only `Stop` reaches the
  model (measured 2026-08-21).** Method: a probe hook registered in `~/.grok/hooks/` on
  `SessionStart`, `UserPromptSubmit`, `PreToolUse`, `PostToolUse` and `Stop`, each writing a
  breadcrumb (proof it RAN) and emitting a distinct token as `additionalContext`; then a
  headless run (`grok -p`) asked the model which tokens it could see. All five hooks ran; the
  model saw only the `Stop` token, and that emission also kept the turn working. Confirms the
  bundled guide's *"For events like `SessionStart` or `PostToolUse`, stdout is ignored"* and
  extends it to `PreToolUse` and `UserPromptSubmit`. **Consequence applied the same day:** the
  injection map in `harness/grok/README.md` — in Grok, policy must travel in the rules/skills
  layer, and hooks are for EFFECTS only. Two side findings: project-scoped hooks
  (`<project>/.grok/hooks/*.json`) did not run even with the folder trusted and listed in
  `trusted_folders.toml` — `grok inspect` showed no project hook loaded — while the same file
  under `~/.grok/hooks/` ran immediately; and Grok has no matcher alias for `exit_plan_mode`,
  so a Claude-shaped `ExitPlanMode` matcher never fires there.
- **Dead URLs, verified 404 on 2026-08-20:** `github.com/xai-org/grok-cli`,
  `docs.x.ai/docs/grok-cli`, `docs.x.ai/build/project-rules` (the real path carries
  `/features/`), and any `xai-org/grok-build/blob/main/docs/...` path.
- `[Authoritative]` **Grok 1.0.6–1.0.13** (CLI 1.0.13 `5e9a58528b76`, bundled
  `~/.grok/docs/user-guide/{10-hooks,16-subagents}.md` + [x.ai/build/changelog](https://x.ai/build/changelog),
  read 2026-08-31). Spawn-time `capability_mode` removed (1.0.6 — tools from the agent type).
  `workflow` tool is top-level only (1.0.8/1.0.9). PreToolUse may `ask`/`defer`/`updatedInput`,
  and the 1.0.13 guide documents `additionalContext` arriving **after** the call; UserPromptSubmit
  allowing stdout and SessionStart/PostToolUse stdout remain discarded. The 2026-08-21 Stop-only
  injection probe was not re-run on 1.0.13. Grounds: `harness/grok/README.md` injection map,
  `agent-routing.md` spawn/workflow bullets, `flow-core/references/harness-mechanics.md`.
- `[Authoritative]` **Grok 1.0.14–1.0.25** (CLI 1.0.25 `f7e67d6988e2`, bundled
  `~/.grok/CHANGELOG.md` through 1.0.24 + `~/.grok/docs/user-guide/{10-hooks,08-skills,12-project-rules,16-subagents}.md`,
  read 2026-09-09). `PostToolUse` stdout is read since 1.0.14: `additionalContext`,
  `decision: "block"` + `reason`, and `updatedToolOutput` (model's copy only; the transcript
  keeps the original) land with the tool result; exit 2 now feeds stderr to the model;
  `PostToolUseFailure` gets `additionalContext` only. Grok emits `tool_response` as an alias of
  `toolResult`, so Claude-shaped readers work unchanged. SessionStart / UserPromptSubmit stdout
  still discarded; `exit_plan_mode` still has no matcher alias. Nothing else in 1.0.15–1.0.24
  touches what the hive deploys (SessionStart hooks run in the background since 1.0.18; `/loop`
  always background since 1.0.19; built-in tools win MCP name collisions since 1.0.22). The
  five-event injection probe was not re-run. Grounds: `harness/grok/README.md` injection map,
  `global/hooks/flow-plan-capture/README.md`, `global/hooks/post-tool-hub/post-tool-hub.sh`.

**Verdict: supported, with three corrections and three undocumented surfaces recorded.**
The corrections (Codex host, opencode org, Grok docs exist) were applied to this file and
to the per-harness READMEs in the same change. The undocumented surfaces are the standing
risk: each is a mechanism the pack depends on that can change without a release note.

## Límite de tamaño del archivo de instrucciones GLOBAL por harness (2026-08-18)

- `[External — source code]` **Codex** (`openai/codex`): el archivo global se lee entero en `codex-rs/codex-home/src/instructions/mod.rs` (sin ningún `max_bytes`) hacia un campo `UserInstructions` distinto del contador de `project_doc_max_bytes`; en `codex-rs/core/src/agents_md.rs` ese contador se inicializa DESPUÉS de anexar las instrucciones de usuario y solo lo consume la cadena de AGENTS.md de PROYECTO (git-root → cwd). Default `DEFAULT_PROJECT_DOC_MAX_BYTES = 32*1024` en `codex-rs/config/src/config_toml.rs`.
- `[Internal — measured]` **Reproducción empírica** (2026-08-18): un `AGENTS.md` de proyecto de 80,824 B se truncó en el cap de 65,536 B mientras el global de 30,322 B quedó íntegro — 96,087 B totales, por encima del cap, lo que prueba que no comparten presupuesto.
- `[External — source code]` **opencode** (`anomalyco/opencode`): `packages/opencode/src/session/instruction.ts` y `packages/core/src/instruction-context.ts` leen global y proyecto por el mismo camino con `fs.readFileString()` y concatenan verbatim — sin `slice` ni `max_bytes` en el pipeline. La doc (`opencode.ai/docs/rules/`) no menciona límites; la petición de un cap configurable se cerró "not planned" (issue #18037). Evidencia POSITIVA de ausencia, no silencio documental.
- `[External — official docs + source test]` **Grok** (`xai-org/grok-build`): the bundled user guide `~/.grok/docs/user-guide/12-project-rules.md` — *"Grok loads each project instruction file in full; there is no character cap and no truncation."* **Locus corrected 2026-08-20:** that repo is public but has **no `docs/` directory — the guide ships with the CLI on disk and is not reachable as a GitHub URL. The public web equivalent is `docs.x.ai/build/features/project-rules`. Test `format_agents_md_section_delivers_full_content` en `crates/codegen/xai-grok-agent/src/prompt/agents_md.rs` lo afirma sobre contenido de 5000 B; la misma función procesa scope user-level y repo-level indistintamente.
- `[Contradiction — logged]` Un fetch de la doc de Codex (`learn.chatgpt.com/docs/agent-configuration/agents-md`) sugería un presupuesto "combinado" global+proyecto, pero su frase coincide literalmente con el comentario de módulo de `agents_md.rs` que describe SOLO la cadena de proyecto; el resumen pasa por un modelo intermedio. Prioridad a código + reproducción empírica, ambos consistentes.
- **Consecuencia para este repo (aplicada 2026-08-18):** se ELIMINARON los dos umbrales de tamaño de `harness/build.py` (`AGENTS_BUDGET_BYTES` y el `AGENTS_HARD_LIMIT_BYTES` que hacía fallar el build). Ambos se justificaban por un riesgo de truncamiento inexistente, y en la práctica convertían cada edición del core en conteo de bytes — recortando prosa que sí valía para caber en un número que nadie aplica. El build ahora reporta tamaño y costo estimado en tokens sin bloquear nada; el crecimiento lo gobierna la regla de ubicación (gate y puntero en el core, mecánica en el router), que es criterio de calidad y no de tamaño. El único cap real que se sigue vigilando es `project_doc_max_bytes` sobre la cadena de PROYECTO, que este repo toca solo al abrir una sesión bajo `harness/`.

## Minimal-solution discipline — optional reference project evaluation (2026-09-07)

Backs `CLAUDE.md > Dependency Decisions > A new dependency is the last rung` (and its `harness/AGENTS.md > Environment` twin) and the `ceiling:` comment convention in `quality/development-principles.md > Solution proportional to the problem`. Source repo cloned at `~/Development/projects/tricell/reference/optional reference project`.

- `[External — measured, single model]` **DietrichGebert/optional reference project v4.9.0** — "lazy senior dev" skill: a 7-rung ladder (YAGNI → reuse → stdlib → native platform → installed dependency → one line → minimum) injected every session and subagent via `SessionStart`/`SubagentStart`/`UserPromptSubmit` hooks, with lite/full/ultra levels and five helper skills. Agentic benchmark (`benchmarks/results/2026-06-18-agentic.md`): headless `claude -p` on `fastapi/full-stack-fastapi-template@cd83fc1`, Haiku 4.5, n=4, LOC = `git diff` added lines; −54% LOC mean across 12 feature tickets, 100% safe on 20 adversarial runs. *Adjusted:* the cut concentrates where a native `<input>` replaces a hand-built component (date picker −94%, color picker −92%) and is ~0 on irreducible CRUD; the author records that a reasoning model (GPT-5.5) spends more tokens deliberating the rungs. Only rungs 3-5 were absent from the hive — everything else (YAGNI, reuse, Rule of Three, root cause, security floor, over-engineering review) already lives in `development-principles.md`, `critical-thinking.md`, `debugging.md`, `code-reviewer` and `/simplify`. **Not adopted:** the plugin itself (re-states three hive rules per session and per subagent; persona voice; "code first, three lines" conflicts with `reporting-integrity.md`; "ONE runnable check, no frameworks" is weaker than the verifiable test gate; mode flag files and a statusline nudge that writes `~/.claude/settings.json`), and the bare native-control preference for product UI — this portfolio's design systems (HeroUI, Angular Material) own that surface, hence the carve-out in the rule.
- `[External — method]` **Agentic benchmark harness pattern** worth reusing for tool evaluations here: one arm per process with `--setting-sources project,local` plus a single `--plugin-dir`, fresh repo copy per cell, score the leftover `git diff`, execute produced functions against adversarial input for the safety axis. Their first run (`2026-06-17-agentic-safety.md`, superseded) reported a ~4% gap because the plugin's `SessionStart` hook fired on the baseline arm too — a global-plugin contamination any benchmark run from this machine (hive hooks always-on) must isolate the same way.
