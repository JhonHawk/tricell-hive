# Inventario de reglas de las skills de flujo

Levantado el 2026-10-15 sobre `development` `fd352a02` por un hijo `hive-research` (sesión Claude Code 69abc19b) para `flow-research`, `flow-plan`, `flow-build` y `flow-close` (`SKILL.md`). 405 unidades: 371 reglas y 34 oraciones que solo mandan leer una referencia (listadas bajo cada tabla). Clases: B base de la reconstrucción (2026-09-20/21), M movimiento masivo, O sesión observada, D decisión del usuario, OD ambas, R referencia externa, U sin origen. Columnas: `dup` otros lugares con la misma obligación (G global.md, RPT reporting.md, DEL delegation.md, Q questions.md, VER verification.md, PF plan-format.md, PR plan-review.md, DD delivery-decisions.md, CR change-records.md, DC data-changes.md, AL artifact-layout.md, GW git-workflow); `C` T = tensión; `D` y = comportamiento por defecto según las guías de 2026 resumidas en `design.md`. Las citas de sesión vienen de los mensajes de commit. La clave de commits y los límites del método están en `design.md`.


#### research

|#|L|sec|rule (first 12 words)|cls|commit|dup|C|D|
|-|-|-|-|-|-|-|-|-|
|1|8|Intro|Turn uncertainty into a supported answer that the next decision can use.|B|7cc491b4 09-21|-|-|y|
|2|8|Intro|This procedure can stand alone or support an already authorized task; it|B|7cc491b4 09-21|-|-|-|
|6|14|Frame|Identify the question, the decision it informs, relevant constraints, and what evidence|B|7cc491b4 09-21|G:27|-|y|
|7|14|Frame|Use context already supplied.|B|7cc491b4 09-21|G:20,27|-|y|
|8|14|Frame|Ask only for missing intent that affects the conclusion; discover repository facts|B|7cc491b4 09-21|G:20,27; Q:9-11|-|-|
|9|16|Frame|Inspect the relevant entry points, callers, tests, configuration, and project guidance before|B|7cc491b4 09-21|plan:14|-|y|
|10|16|Frame|When the question concerns project state or delivery, also reconcile the relevant|B|25d7d742 09-21|backlog-report.md:3; G:19|-|-|
|11|16|Frame|When collecting the state of named tickets, pull requests, CI runs, or|OD|7b999437 10-13|-|-|-|
|12|16|Frame|Follow references that bear on the question rather than inventorying the entire|B|7cc491b4 09-21|-|-|y|
|13|16|Frame|Let uncertainty and consequences determine depth, not a fixed file or source|B|7cc491b4 09-21|-|-|y|
|14|18|Frame|Treat search hits as entry points to inspect, and state the searched|B|23c1c790 09-21|agents/review/hive-research.md:12|-|-|
|16|20|Frame|In the main thread, delegate under the shared delegation contract any question|U|847acc14 09-25|DEL:11; G:78|-|-|
|17|20|Frame|Keep a short, bounded lookup in the main thread; at the sixth|O|60b9a547 09-30 (1st clause 847acc14 unknown)|DEL:11; G:78|-|-|
|18|20|Frame|Before the first delegated lookup, tell the user in one sentence which|O|959427e1 09-27 (847acc14 09-25)|DEL:11; G:78; G:? (959427e1 announce)|-|-|
|19|20|Frame|Delegation is internal work and needs no separate approval.|U|2afca607 09-23|build:38; VER:61|-|-|
|20|20|Frame|Select the installed `hive-research` role rather than a host's built-in explorer; when|O|6c0e5055 09-30 (2afca607 09-23)|DEL:15-21; PR:30; G:80|-|-|
|21|20|Frame|In each research child's brief, state what is already known or settled,|OD|7b999437 10-13|-|-|-|
|22|20|Frame|The role's contract sets what the child returns.|OD|7b999437 10-13|DEL:67|-|y|
|23|20|Frame|Check the claims that decide the answer or the recommendation by opening|OD|7b999437 10-13|DEL:67; G:79|-|-|
|24|20|Frame|Send high-consequence premises that this check cannot settle to one `hive-refute-claim` child,|OD|7b999437 10-13|VER:41 (hive-refute-claim)|-|-|
|25|20|Frame|Resolve conflicting conclusions yourself; agreement among subagents is not proof.|OD|7b999437 10-13|DEL:67; PR:22|-|-|
|26|24|Evide|Distinguish inspected facts, source claims, inference, and recommendations.|OD|7b999437 10-13|G:60; agents/review/hive-research.md:12|-|-|
|27|24|Evide|A test's presence is not evidence that it passed; a documented capability|B|7cc491b4 09-21|G:62; build:149|-|-|
|28|25|Evide|For project status, distinguish declared state from observed evidence and unresolved discrepancies.|B|25d7d742 09-21|G:19; backlog-report.md|-|-|
|29|25|Evide|Report discrepancies without changing tickets, documentation, code, or delivery state unless those|B|25d7d742 09-21|G:28|-|-|
|30|26|Evide|For version-dependent external behavior, and before recommending how to configure or use|U|fa361feb 09-23|G:58|-|-|
|31|26|Evide|Use an available documentation skill or tool when applicable; record meaningful access|B|7cc491b4 09-21|G:58|-|-|
|32|27|Evide|Resolve conflicting sources against their dates, versions, scope, and direct observations.|B|7cc491b4 09-21|-|-|y|
|33|27|Evide|Neither backlog nor documentation takes automatic precedence.|B|13ab6b11 09-21|-|-|-|
|34|27|Evide|Missing delivery evidence means delivery is not verified, not that it did|B|13ab6b11 09-21|-|-|-|
|35|27|Evide|Preserve an unresolved contradiction with the evidence needed to settle it.|B|7cc491b4 09-21|-|-|-|
|36|28|Evide|Use a bounded check when reading alone cannot answer a consequential question|B|7cc491b4 09-21|G:28|-|-|
|37|28|Evide|Do not turn a read-only investigation into an implementation or deployment.|B|7cc491b4 09-21|G:28|-|-|
|40|34|Deliv|State the answer, supporting source locations or URLs, material alternatives and tradeoffs|B|7cc491b4 09-21 (ed 75fe4ce5(OD))|G:62|-|y|
|41|34|Deliv|Do not present a recommendation as an implemented result.|B|7cc491b4 09-21|G:28,62|-|y|
|42|36|Deliv|When the findings identify follow-up work that no ticket covers, ask whether|U|754d236a 09-22 (ed 04e2d405(O))|G:74|-|-|
|43|36|Deliv|A question the findings fully answer needs no ticket.|U|2afca607 09-23|research:53|-|-|
|44|36|Deliv|Infer the tracker from project guidance, existing issue links, or configured project|U|754d236a 09-22|PF:44|-|-|
|45|36|Deliv|Open the ticket only after the user agrees, check for duplicates first,|U|754d236a 09-22|G:31|-|-|
|46|38|Deliv|Keep findings and iterative debate in the conversation.|D|dc3252e8 09-22|G:88|-|-|
|47|38|Deliv|Apply the shared research-retention authorization rule before creating or updating a research|D|dc3252e8 09-22|G:88|-|-|
|48|38|Deliv|When proposing retention, name the benefit and intended document, then wait for|D|dc3252e8 09-22|G:88|-|-|
|49|38|Deliv|After authorization, resolve the destination through the global artifact-placement instructions and existing|B|25d7d742 09-21 (ed dc3252e8(D))|G:101-104; plan:59; build:48|-|-|
|50|38|Deliv|Keep the saved record concise: question, evidence, findings, decisions and unresolved items;|B|7cc491b4 09-21 (ed dc3252e8(D))|G:88|-|-|
|51|40|Deliv|For an investigation-only request, stop at the findings.|B|7cc491b4 09-21|G:23,28|-|-|
|52|40|Deliv|When a round of findings concerns work the user could run next,|O|e3e534e8 09-28 (ed 04e2d405(O))|RPT:77; G:23,74|-|-|
|53|40|Deliv|A question the findings fully answer, with no work to run, ends|O|04e2d405 10-11|G:23; research:74|-|-|
|54|40|Deliv|When the user asks to do the work, such as "can you|O|04e2d405 10-11|RPT:77; G:28|-|-|
|55|40|Deliv|Ask it once per round; ask it again only when the findings|O|04e2d405 10-11|Q:60; G:20|-|-|
|56|40|Deliv|Start the chosen flow without a new command.|O|e3e534e8 09-28|RPT:77 (identical)|-|-|
|57|40|Deliv|Within a broader authorized task, carry the findings into that task without|B|7cc491b4 09-21|G:28; G:31|-|-|
|58|40|Deliv|Preserve established records named either `<topic>.research.md` or legacy `<topic>-research.md`; do not rename|B|7cc491b4 09-21|plan:61; build:48|-|-|
|59|42|Deliv|When the next steps alternate investigation and implementation toward one measurable objective,|OD|d5164f71 10-11|G:31 (cites it); RPT:77; build:79|-|-|
|60|42|Deliv|A round is one investigation followed by one implementation and its measurement;|OD|d5164f71 10-11|-|-|-|
|61|42|Deliv|The option states the objective and how it is measured, the round|OD|d5164f71 10-11|-|-|-|
|62|42|Deliv|Every step must stay within the local editing and verification that implementation|OD|d5164f71 10-11|G:28|-|-|
|63|42|Deliv|Choosing the route authorizes its rounds without a new route question: the|OD|d5164f71 10-11|-|-|-|
|64|42|Deliv|Report each round in one line with its measurement and continue.|OD|d5164f71 10-11|RPT:91|-|-|
|65|42|Deliv|Stop the route when:|OD|d5164f71 10-11|-|-|-|
|66|44|Deliv|the objective is met;|OD|d5164f71 10-11|-|-|-|
|67|45|Deliv|the round limit is reached;|OD|d5164f71 10-11|-|-|-|
|68|46|Deliv|a round ends without a measurable improvement;|OD|d5164f71 10-11|-|-|-|
|69|47|Deliv|a step needs an effect the route does not cover and no|OD|d5164f71 10-11|G:28|-|-|
|70|48|Deliv|a round's change needs `flow-plan`;|OD|d5164f71 10-11|-|-|-|
|71|49|Deliv|a finding changes the objective or scope;|OD|d5164f71 10-11|-|-|-|
|72|50|Deliv|a decision is genuinely the user's.|OD|d5164f71 10-11|-|-|-|
|73|52|Deliv|When the route stops with changes in the working tree, run [flow-close]:|OD|d5164f71 10-11|build:79; close:6|-|-|
|74|52|Deliv|Without such changes, end with the route question.|OD|d5164f71 10-11|research:53|-|-|

Read-reference sentences (6): #3 L12 (d2673600 10-12); #4 L12 (d2673600 10-12); #5 L12 (d2673600 10-12); #15 L18 (23c1c790 09-21); #38 L32 (d2673600 10-12); #39 L32 (d2673600 10-12)

#### plan

|#|L|sec|rule (first 12 words)|cls|commit|dup|C|D|
|-|-|-|-|-|-|-|-|-|
|1|8|Intro|Produce enough shared understanding for an implementer to act without inventing product,|B|7cc491b4 09-21|-|-|y|
|2|8|Intro|Use the CLI's available planning capabilities and respect its active mode.|B|7cc491b4 09-21|-|-|y|
|3|8|Intro|A native task list is a working view, not proof of a|B|7cc491b4 09-21|build:65|-|-|
|7|14|Groun|Establish the intended outcome, scope, constraints, authorized effects, and observable acceptance criteria|B|7cc491b4 09-21|G:27|-|y|
|8|14|Groun|Inspect the relevant implementation, callers, tests, configuration, and project guidance, delegating exploration|B|7cc491b4 09-21 (ed 60b9a547(O))|research:16|-|y|
|9|14|Groun|Reuse sound research but recheck facts that may have changed.|B|7cc491b4 09-21|-|-|y|
|10|14|Groun|While delegated exploration runs, draft the parts of the plan that settled|U|591b85de 09-25|G:78; DEL:45|-|-|
|11|16|Groun|Resolve missing consequential choices with the user, offering meaningful alternatives and their|B|7cc491b4 09-21|G:20; Q:9|-|-|
|12|16|Groun|Investigate discoverable facts directly.|B|7cc491b4 09-21|G:20,27|-|y|
|13|16|Groun|Treat information that clarifies the requested outcome as an update to the|B|25d7d742 09-21|-|-|-|
|14|16|Groun|If a technical uncertainty can invalidate the approach, resolve it with authorized|B|7cc491b4 09-21|-|-|-|
|15|16|Groun|Routine local coding choices need not be frozen in advance.|B|7cc491b4 09-21|-|-|y|
|16|18|Groun|When a decision changes the design of the plan or of the|OD|75fe4ce5 10-08|Q:56; G:20|-|-|
|17|18|Groun|When an answer returns the work to research or invalidates tasks already|D|53fea4f1 10-11|Q:64; G:20 (cited as "global rule")|-|-|
|18|18|Groun|Ask a consequential design question before writing dependent tasks; group the others|B|f469db22 09-21 (ed 53fea4f1(D))|Q:19; G:20|-|-|
|20|18|Groun|Work implemented without a plan resolves it at the start of `flow-build`|O|e27f244f 09-23|build:19|-|-|
|21|18|Groun|If no design questions remain, briefly state which existing decisions or conventions|B|f469db22 09-21|-|-|-|
|22|20|Groun|Before proposing a new mechanism or abstraction, inspect relevant existing solutions and|B|25d7d742 09-21|build:33; G:76|-|y|
|23|20|Groun|Define only the contracts affected by the change: inputs, outputs, error behavior,|B|7cc491b4 09-21|-|-|-|
|24|20|Groun|Record why the selected approach fits the evidence and constraints.|B|7cc491b4 09-21|-|-|y|
|25|20|Groun|Mention meaningful rejected alternatives only when they explain a decision.|B|7cc491b4 09-21|-|-|y|
|26|22|Groun|When an interface changes, identify its actual producers and consumers, the versions|B|23c1c790 09-21 (ed 3db867ef(O))|build:57|-|-|
|27|22|Groun|Default to one coordinated change that updates every producer and consumer and|O|3db867ef 09-26|build:57-60; DC:7|-|-|
|28|22|Groun|Recommend a compatible or phased transition only for a need the evidence|O|3db867ef 09-26|DC:7|-|-|
|29|22|Groun|A consumer that no evidence shows exists, a deployment gap of minutes|O|3db867ef 09-26|-|-|-|
|30|22|Groun|Neither is a gate that can only pass by proving an absence,|O|c0069831 09-26|build:59|-|-|
|31|22|Groun|Whenever you offer a transition, state what would run in parallel, what|O|3db867ef 09-26|-|-|-|
|32|22|Groun|Use the project’s existing contract and distribution mechanism; a separate specs repository,|B|23c1c790 09-21|-|-|-|
|33|22|Groun|Define evidence for the affected integration and any partial rollout.|B|23c1c790 09-21|build:60|-|-|
|36|24|Groun|Apply only the relevant criteria and pass the reference paths to delegated|B|23c1c790 09-21|DEL:35; G:80|-|-|
|37|26|Groun|When replacing or migrating an existing system, state which legacy paths are|B|a73c60ac 09-21|-|-|-|
|38|26|Groun|Name the retirement condition for temporary retention and the maintenance and verification|B|a73c60ac 09-21|-|-|-|
|39|26|Groun|Reuse decisions already settled by the request; resolve only consequential ambiguity before|B|a73c60ac 09-21|G:20|-|y|
|40|26|Groun|Do not introduce parallel implementations or fallback layers merely as an assumed|B|a73c60ac 09-21|G:73|-|-|
|42|28|Groun|Resolve the affected application and pattern, applicable project conventions, reuse, and intentional|B|dd3e7db1 09-21|ui-planning.md:3|-|-|
|44|30|Groun|Apply project and provider constraints before its defaults, and record any migration|B|25d7d742 09-21|infra-naming.md|-|-|
|45|34|Verif|Group tasks into coherent results that can be checked independently.|B|7cc491b4 09-21|PF:16|-|y|
|46|34|Verif|Prefer usable end-to-end increments when the change crosses layers.|B|7cc491b4 09-21|-|-|y|
|47|34|Verif|Include dependencies, concrete source locations, the expected behavior, and the verification command|B|7cc491b4 09-21|PF:16,111|-|-|
|48|34|Verif|Look up actual project commands rather than inventing them.|B|7cc491b4 09-21|-|-|y|
|49|36|Verif|Select the verification approach and timing for each coherent result, including prerequisites|B|ceab31eb 09-21|build:16,18|-|-|
|51|36|Verif|For substantial gates, identify scope/candidate, mechanism, environment, relevant effects and expected evidence;|B|6bb6e431 09-21|VER:26|-|-|
|52|36|Verif|Include applicable local in-vivo/UI verification in build completion and offer deployed-environment walks|B|6bb6e431 09-21|build:71,123|-|-|
|53|36|Verif|Use the project's commands, environments, and delivery gates; name unavailable checks and|B|ceab31eb 09-21|-|-|-|
|54|36|Verif|Keep this in the existing task or conversation rather than creating a|B|ceab31eb 09-21|VER:110|-|-|
|55|36|Verif|Give each task that changes behavior a test approach under the build|D|99e51eab 09-26|build:44; PF:107|-|-|
|56|38|Verif|Use checkbox tasks.|B|7cc491b4 09-21|-|-|-|
|57|38|Verif|Keep setup and documentation with the result they support; split tasks when|B|7cc491b4 09-21|-|-|-|
|58|38|Verif|Do not require full implementation code in the plan or replace a|B|7cc491b4 09-21|-|-|-|
|59|40|Verif|Before calling the plan ready, check:|B|7cc491b4 09-21|-|-|-|
|60|42|Verif|Important claims are grounded in inspected sources; facts and assumptions are distinguishable.|B|7cc491b4 09-21|G:60|-|-|
|61|43|Verif|Decisions that could change the outcome, contracts, or compatibility are resolved; remaining|B|7cc491b4 09-21|-|-|-|
|62|44|Verif|Every acceptance criterion is numbered and falsifiable at the base commit, appears|O|d393eef0 09-26|PF:26|-|-|
|63|44|Verif|An orphan criterion, a task whose verification misses a criterion it closes,|O|d393eef0 09-26|PF:26; PR:20|-|-|
|65|45|Verif|A change to stored data records the affected-record count per environment, or|O|00363434 09-27|DC (data-changes.md) count rule|-|-|
|66|46|Verif|Task interfaces and dependencies agree; no task relies on an undefined contract.|B|7cc491b4 09-21|-|-|-|
|67|47|Verif|Each task states its execution, consistent with its dependencies and write ownership.|U|f414c62e 09-22|PF:105|-|-|
|68|48|Verif|A change with a rendered UI effect plans the independent `hive-review-ux` and|U|91e99e5c 09-22|PF:121; VER:61; build:36,71|-|-|
|69|48|Verif|The implementer's walk does not replace them.|U|91e99e5c 09-22|VER:61; build:36|-|-|
|70|49|Verif|An implementer with this plan and its linked sources can identify what|B|7cc491b4 09-21|plan:1|-|y|
|71|51|Verif|Assign each task's execution: the main thread or a delegated child, with|U|f414c62e 09-22|PF:105; build:34|-|-|
|72|51|Verif|Prefer delegation for units with settled interfaces and no overlapping writes; keep|U|f414c62e 09-22|build:34|-|-|
|73|51|Verif|For delegated work, identify separable results and their interfaces, dependencies, and ownership,|B|ceab31eb 09-21|DEL:11|-|-|
|74|51|Verif|Link the project conventions and technical references each unit needs, without freezing|B|ceab31eb 09-21|-|-|-|
|75|53|Verif|After saving an implementation plan, read [plan review] and dispatch bounded read-only|B|42dcb409 09-21|PR:3|-|-|
|76|53|Verif|Dispatch one reviewer per affected domain in parallel, even for a small|B|42dcb409 09-21|PR:9 (identical)|-|-|
|77|53|Verif|Reviewers return feedback to the orchestrator, who validates findings and alone revises|B|42dcb409 09-21|PR:36; G:67|-|-|
|78|53|Verif|Disclose unavailable delegation and unresolved coverage.|B|42dcb409 09-21|G:80|-|-|
|81|59|Hando|Keep a bounded, understood plan in the conversation unless retention helps.|B|7cc491b4 09-21|-|-|-|
|82|59|Hando|Retain a plan when requested, when work must resume across sessions or|B|7cc491b4 09-21|CR (change-records.md)|-|-|
|83|59|Hando|Before writing a retained artifact, resolve its destination through the applicable global|B|25d7d742 09-21|G:99-104; research:49; build:48|-|-|
|85|61|Hando|They define the handoff in the change folder, not a mandatory document|R|1fbec0a3 09-23|PF:3|-|-|
|86|61|Hando|Keep the readiness procedure above here; use the reference to make its|B|7cc491b4 09-21|-|-|-|
|87|61|Hando|Preserve an established `<topic>-plan.md` or `<topic>.plan.md` path without renaming it merely to|B|13ab6b11 09-21|build:48; research:58|-|-|
|88|63|Hando|If the active CLI mode prohibits writing the artifact, present its contents|B|7cc491b4 09-21|-|-|-|
|89|63|Hando|For a planning-only request, stop with the plan or unresolved decisions.|B|7cc491b4 09-21|G:28|-|-|
|90|63|Hando|When implementation is already authorized and the active mode permits it, follow|B|7cc491b4 09-21|G:28; Q:11|-|y|
|91|65|Hando|When returning a ready plan without implementing, summarize it in the conversation,|D|3109e62c 09-21 -> c332a814 10-11|-|-|-|
|92|65|Hando|Do not paste the whole plan.|D|c332a814 10-11|-|-|-|
|93|65|Hando|End with the build route question under the global choice rules, which|D|c332a814 10-11|RPT:69-71; G:23|T|-|
|94|65|Hando|Leave the saved change folder uncommitted: keep its status on disk while|O|7e8a627f 09-29|build:48; close:44; DD:22|-|-|
|95|65|Hando|When execution is already explicitly authorized, continue as permitted instead of asking.|B|3109e62c 09-21 (ed c332a814(D))|G:20; Q:11|-|y|

Read-reference sentences (13): #4 L12 (d2673600 10-12); #5 L12 (d2673600 10-12); #6 L12 (d2673600 10-12); #19 L18 (53fea4f1 10-11); #34 L24 (23c1c790 09-21); #35 L24 (23c1c790 09-21); #41 L28 (dd3e7db1 09-21); #43 L30 (25d7d742 09-21); #50 L36 (ceab31eb 09-21); #64 L44 (d393eef0 09-26); #79 L57 (d2673600 10-12); #80 L57 (d2673600 10-12); #84 L61 (7cc491b4 09-21)

#### build

|#|L|sec|rule (first 12 words)|cls|commit|dup|C|D|
|-|-|-|-|-|-|-|-|-|
|1|8|Intro|Use the requested outcome or supplied plan as the execution brief.|B|7cc491b4 09-21|G:27|-|y|
|2|8|Intro|For work the user asked for, this skill does not require a|OD|de83d56a 10-14|G:74; research desc|-|-|
|3|8|Intro|It does not grant effects beyond the user's instructions.|B|7cc491b4 09-21 (ed de83d56a(OD))|G:28|-|-|
|7|14|Recon|Read the relevant project guidance, current source and tests, and the supplied|B|7cc491b4 09-21|-|-|y|
|8|14|Recon|Inspect existing changes and completed work.|B|7cc491b4 09-21|-|-|y|
|9|14|Recon|Treat the record as a claim to reconcile with current evidence, not|B|7cc491b4 09-21|G:89|-|y|
|10|16|Recon|Identify the candidate, affected applications/services and environments, and the checks needed to|B|6bb6e431 09-21|VER:9|-|-|
|11|16|Recon|A supplied plan is optional; keep a small task’s criteria in the|B|6bb6e431 09-21|-|-|-|
|12|16|Recon|Identify the next unmet acceptance criterion and its dependencies.|B|7cc491b4 09-21|-|-|y|
|13|16|Recon|Preserve valid completed work.|B|7cc491b4 09-21|-|-|y|
|14|16|Recon|If the current state invalidates a decision that sets scope, a contract,|B|7cc491b4 09-21|Q:64; G:20|-|-|
|15|16|Recon|Resolve ordinary implementation details within scope; do not stop at every task|B|7cc491b4 09-21|RPT:91|-|-|
|16|18|Recon|Identify the available test runner, actual project check/build commands, relevant baseline failures,|B|ceab31eb 09-21|VER:83|-|y|
|18|18|Recon|Reconcile a supplied verification plan rather than silently weakening it.|B|ceab31eb 09-21|VER:3; G:68|-|-|
|19|20|Recon|When the work will end in a Git delivery and neither a|O|e27f244f 09-23 (ed 09119f5e(O))|DD; RPT:79; G:28|-|-|
|20|20|Recon|The selected mode then carries the later gates, including waiting for the|O|e27f244f 09-23|G:28; DD|-|-|
|21|20|Recon|When the work will not end in a Git delivery and is|OD|d5164f71 10-11|G:28; DD|-|-|
|23|22|Recon|These apply even without a retained plan or specialist assignment.|B|23c1c790 09-21|-|-|-|
|25|24|Recon|Reconcile the supplied view decisions; if no plan exists, resolve the same|B|dd3e7db1 09-21|ui-planning.md|-|-|
|26|26|Recon|Before installing a missing tool, check the project-local and available system installations|B|a73c60ac 09-21|-|-|y|
|27|26|Recon|Prefer the project’s existing setup.|B|a73c60ac 09-21|-|-|y|
|28|26|Recon|Distinguish task-local dependency setup from system-wide installs or shell/global configuration changes; the|B|a73c60ac 09-21|G:28|-|-|
|29|26|Recon|Verify the actual OS and shell before using platform-specific commands rather than|B|a73c60ac 09-21|-|-|y|
|32|32|Execu|Choose a coherent next result and implement it in the existing source|B|7cc491b4 09-21|-|-|y|
|33|32|Execu|Look for an existing solution or pattern before adding a mechanism or|B|25d7d742 09-21|plan:22; G:76|-|y|
|34|32|Execu|Confirm that reused tools or patterns meet this project's requirements.|B|25d7d742 09-21|-|-|y|
|35|32|Execu|Use the project's declared dependency manager and preserve its lockfile conventions.|B|25d7d742 09-21|-|-|y|
|36|32|Execu|Select technical references from the actual project stack and affected behavior.|B|ceab31eb 09-21|-|-|y|
|37|32|Execu|Keep dependent edits ordered.|B|ceab31eb 09-21|-|-|y|
|38|34|Execu|Delegation is internal work and needs no separate approval; prefer it.|U|f414c62e 09-22|research:19|-|-|
|39|34|Execu|Announce the split before the first edit under the shared delegation rules,|U|f414c62e 09-22|G:78; DEL:11|-|-|
|40|34|Execu|Delegate each unit whose interface is settled and whose writes do not|U|f414c62e 09-22 (ed 811c7d23(O))|plan:72; G:78|-|-|
|41|34|Execu|A change spanning several files with its tests, unless it is mechanical,|U|2d497161 09-22 (ed 0fee2d02(D))|G:78; G:72; DEL:11|-|-|
|42|34|Execu|For a plan task, the brief names the change folder path and|D|d084e4ae 10-08|DEL:27; G:80|-|-|
|43|34|Execu|Do not paraphrase the task or its criteria in the brief: it|D|d084e4ae 10-08|DEL:27|-|-|
|44|34|Execu|Without a retained plan, the brief quotes the user's request for the|D|d084e4ae 10-08|-|-|-|
|45|34|Execu|The main thread keeps coordination, integration, and user decisions.|U|2d497161 09-22 (ed d084e4ae(D))|G:79; DEL:29,67|-|-|
|46|34|Execu|Work outside the plan's tasks, such as fixes after review or human|U|2d497161 09-22|-|-|-|
|47|34|Execu|Corrections the user requests during a validation pause follow the correction-round rule|O|01074e8f 09-29|VER:69|-|-|
|48|36|Execu|For changes with a rendered UI effect, the independent UI review and|U|f414c62e 09-22|VER:61; plan:68|-|-|
|49|36|Execu|Propose other independent verification or specialist review when the acceptance criteria or|B|ceab31eb 09-21|-|-|-|
|50|36|Execu|Independent surface checks may run in parallel; keep one owner for shared|B|6bb6e431 09-21|VER:102|-|-|
|51|36|Execu|The implementer still verifies its own change.|B|ceab31eb 09-21|-|-|y|
|52|38|Execu|Before adding a dependency, assess whether an existing project dependency, standard library,|B|a73c60ac 09-21|plan:22; G:73|-|y|
|53|38|Execu|Choose based on compatibility, maintainability, and fit rather than a universal library|B|a73c60ac 09-21|-|-|y|
|54|38|Execu|Explain a new dependency’s purpose and any overlap; consolidating existing usages remains|B|a73c60ac 09-21|-|-|-|
|56|40|Execu|Preserve an existing name unless its authorized change includes the necessary impact|B|25d7d742 09-21|infra-naming.md|-|-|
|57|42|Execu|For a changed shared interface, reconcile the actual producers, consumers, coexisting versions,|B|23c1c790 09-21|plan:26|-|-|
|58|42|Execu|When a plan keeps a compatible transition, such as an old version|O|c0069831 09-26|plan:30|-|-|
|59|42|Execu|If it rests only on a consumer no evidence shows exists, or|O|c0069831 09-26|plan:29-30|-|-|
|60|42|Execu|When the transition stands, verify it, including partial rollout, using the existing|B|23c1c790 09-21 (ed c0069831(O))|plan:33|-|-|
|61|44|Execu|Use red-green-refactor TDD for new or changed behavior whose failure would matter,|D|99e51eab 09-26|plan:55; PF:107|-|-|
|62|44|Execu|Write the behavioral test, observe it fail for the intended reason, implement|B|ceab31eb 09-21 (ed 99e51eab(D))|-|-|-|
|63|44|Execu|A missing test for such behavior is a reason to write one,|B|ceab31eb 09-21 (ed 99e51eab(D))|-|-|-|
|64|44|Execu|A mechanical change, as the global guidance defines it, adds no new|D|99e51eab 09-26 (ed 0fee2d02(D))|G:72|-|-|
|65|44|Execu|When a change alters behavior of a kind listed for TDD, TDD|D|99e51eab 09-26|-|-|-|
|66|44|Execu|For behavior-preserving refactors, use existing tests or characterization.|B|ceab31eb 09-21 (ed 99e51eab(D))|-|-|-|
|67|44|Execu|Verify any other behavior change with a direct check, or by extending|D|99e51eab 09-26|-|-|-|
|68|44|Execu|A plan's declared test approach for a task governs that task unless|D|99e51eab 09-26|PF:107|-|-|
|69|44|Execu|If no usable test infrastructure exists, choose an available validator or direct|B|ceab31eb 09-21|-|-|-|
|70|46|Execu|Preserve implementation that already exists before RED.|B|ceab31eb 09-21|-|-|-|
|71|46|Execu|Verify the regression check against an isolated pre-change baseline or controlled reproduction|B|ceab31eb 09-21|-|-|-|
|72|46|Execu|Report fail-to-pass only when that comparison was observed; otherwise disclose the missing|B|ceab31eb 09-21|-|-|-|
|73|46|Execu|A dependency/setup error is not behavioral RED.|B|ceab31eb 09-21|-|-|-|
|74|46|Execu|Review what assertions actually prove; do not weaken a valid check to|B|25d7d742 09-21|-|-|-|
|75|46|Execu|Do not edit or remove a test that the user or the|D|99e51eab 09-26|G:73|-|-|
|76|46|Execu|Add no test for a case the change cannot produce.|D|99e51eab 09-26|-|-|-|
|77|46|Execu|Remove obsolete code, configuration, or temporary residue created by the change when|B|25d7d742 09-21 (ed 99e51eab(D))|-|-|-|
|78|48|Execu|Before writing a retained supporting artifact, resolve its destination through the applicable|B|25d7d742 09-21|plan:59; research:49|-|-|
|79|48|Execu|When a retained plan exists, keep its task list and current status|B|7cc491b4 09-21|PF:3|-|-|
|80|48|Execu|Update that status before reporting a delivery milestone, so the report and|U|683ba60d 09-24 (ed 7e8a627f(O))|plan:65; close:44|-|-|
|81|48|Execu|Preserve either established `<topic>-plan.md` or `<topic>.plan.md` names; do not rename a valid|B|13ab6b11 09-21|plan:61; research:58|-|-|
|82|48|Execu|When the build starts, record the change's base commit in `tasks.md`.|O|d393eef0 09-26|-|-|-|
|83|48|Execu|Mark a task complete only when its expected result is evidenced, as|B|7cc491b4 09-21 (ed d393eef0(O))|-|-|-|
|84|48|Execu|Record changed decisions, failed or omitted checks, and why a completed task|B|7cc491b4 09-21|-|-|-|
|85|48|Execu|Keep raw output separate and link the useful evidence.|B|7cc491b4 09-21|-|-|y|
|86|50|Execu|In a plan whose acceptance criteria are numbered `AC<n>`, tasks carry the|O|d393eef0 09-26|PF:22-38|-|-|
|87|50|Execu|Set `[/]` when a task starts.|O|d393eef0 09-26|-|-|-|
|88|50|Execu|A task whose test approach is `tdd` or `check` is verified by|O|d393eef0 09-26|PF:26,107|-|-|
|89|52|Execu|When the implementer reports the task, set `[?]`, then launch `hive-verify-task` as|O|d393eef0 09-26 (ed 36a1a1a4(OD))|-|-|-|
|90|52|Execu|The acknowledgement of a background launch is not the verdict: its completion|O|9abeca73 10-01|-|-|-|
|91|53|Execu|When every criterion is `met`, open the cited `path:line` or read the|O|d393eef0 09-26|-|-|-|
|92|54|Execu|When any criterion is `not met`, even if others are `cannot verify`,|O|d393eef0 09-26|-|-|-|
|93|54|Execu|Then set `[?]` again and launch one re-verification that checks only the|O|d393eef0 09-26 (ed 01074e8f(O))|-|-|-|
|94|54|Execu|If it again reports `not met`, set `[!]` again with the round-2|O|d393eef0 09-26|-|-|-|
|95|55|Execu|When the verdict has no `not met` but a criterion is `cannot|O|d393eef0 09-26|-|-|-|
|96|56|Execu|What makes the verdict independent is the verifier's context, not its model:|OD|36a1a1a4 10-15|-|-|-|
|97|56|Execu|Launch it on its configured `verifier` profile, which on the hosts that|OD|36a1a1a4 10-15|-|-|-|
|98|56|Execu|When the host cannot set a model, the configured model equals the|OD|36a1a1a4 10-15|-|-|-|
|99|56|Execu|Never launch the verifier on a smaller model or a lower effort|OD|36a1a1a4 10-15|-|-|-|
|100|56|Execu|In the plan's progress section, record the implementer's model and the verifier's,|O|e2070f30 09-29 (ed 36a1a1a4(OD))|DEL:73|-|-|
|101|56|Execu|When `hive-verify-task` cannot run, for example because its model is unavailable or|O|d393eef0 09-26 (ed 36a1a1a4(OD))|-|-|-|
|102|57|Execu|Apply the user's decision: accepted sets `[x]` with a line `accepted unverified|O|d393eef0 09-26|PF:35|-|-|
|103|59|Execu|Do not wait idle while a verification runs.|O|9abeca73 10-01|G:78; DEL:45|-|-|
|104|59|Execu|Start the next pending task that does not depend on the task|O|9abeca73 10-01|-|-|-|
|105|59|Execu|When several tasks wait for verification and their **Locations** share no file,|O|9abeca73 10-01|-|-|-|
|106|59|Execu|Only when no such task or verification remains, wait on the completion|O|9abeca73 10-01|DEL:55-57; VER:88|-|-|
|107|59|Execu|If a fix round must change a file that a task started|O|9abeca73 10-01|-|-|-|
|108|61|Execu|When the delivery mode makes commits, commit each task yourself when it|D|b1ed8315 10-08|GW (git-workflow); DD|-|-|
|109|61|Execu|Stage only that task's paths, keep the change folder out of these|D|b1ed8315 10-08|-|-|-|
|110|63|Execu|Other tasks, such as `characterization`, tasks without a test approach, and delivery|O|d393eef0 09-26|-|-|-|
|111|63|Execu|`hive-verify-task` is part of the authorized implementation and needs no separate approval.|O|d393eef0 09-26|VER:61; research:19|-|-|
|112|63|Execu|It checks the task against the plan; it is not a code|O|d393eef0 09-26|-|-|-|
|113|63|Execu|A plan without `AC<n>` criteria uses only `[ ]` and `[x]`, and|O|d393eef0 09-26|PF:38|-|-|
|114|65|Execu|For any multi-step work, with or without a retained plan, when the|U|847acc14 09-25 (ed 8c69c3ef(O))|-|-|-|
|115|65|Execu|A plan task completes in the native list when it reaches `[x]`,|O|d393eef0 09-26|plan:8|-|-|
|116|65|Execu|A retained plan's record, when it exists, still governs.|U|847acc14 09-25|-|-|-|
|117|67|Execu|Keep task-started processes identifiable.|B|a73c60ac 09-21|-|-|-|
|118|67|Execu|Before telling the user that a child is still in progress after|U|591b85de 09-25|591b85de role text|-|-|
|119|67|Execu|Silence alone does not mean a command is hung: inspect available liveness|B|a73c60ac 09-21|DEL:61|-|-|
|120|67|Execu|Use the host’s supported long-running execution controls when needed and continue independent|B|a73c60ac 09-21|DEL:45,61|-|-|
|121|67|Execu|Do not stop unrelated processes.|B|a73c60ac 09-21|close:16; G:37|-|-|
|122|71|Verif|Compare the final changes and observed behavior against the acceptance criteria and|B|7cc491b4 09-21|G:62|-|y|
|123|71|Verif|Local in-vivo verification and UI review are required when applicable, without a|B|6bb6e431 09-21|G:66; VER:49|-|-|
|124|71|Verif|Run affected tests and applicable build, typecheck, and lint commands that the|B|ceab31eb 09-21|-|-|y|
|125|71|Verif|Exercise changed runtime behavior through the real integration path; inspect changed user-facing|B|ceab31eb 09-21|-|-|-|
|126|71|Verif|In deployed environments, apply the shared gate authorization boundary before the functional|B|6bb6e431 09-21|G:66; VER:26|-|-|
|127|71|Verif|When a deployed walk runs, because the project declares it or the|O|60b9a547 09-30|VER:99|-|-|
|128|71|Verif|Any failed check or error observed after a deployment, such as an|O|60b9a547 09-30|VER:99|-|-|
|129|71|Verif|Tests passing or a successful build do not replace those observations.|B|ceab31eb 09-21|G:62|-|-|
|130|71|Verif|Before reporting a UI change complete, obtain the independent UI review and|U|f414c62e 09-22|VER:61; build:36|-|-|
|132|73|Verif|Use existing evidence only while its relevant code, configuration, and environment remain|B|ceab31eb 09-21|VER:100|-|-|
|133|73|Verif|While iterating, run the project's affected-test selection for the changed code, widened|O|01074e8f 09-29|VER:83|-|-|
|134|73|Verif|Close regression coverage through the project's actual CI gate.|B|ceab31eb 09-21 (ed cc016a93(O))|VER:87|-|-|
|135|73|Verif|When none will run it, run the full suite once on the|D|99e51eab 09-26 (ed cc016a93, 2765105a)|G:72|-|-|
|136|73|Verif|For a mechanical rename or move, find every reference to the old|U|3c6c243e 09-29|-|-|-|
|137|73|Verif|Do not rerun a passing suite for confidence.|D|99e51eab 09-26|VER:83 (identical)|-|-|
|138|73|Verif|When the verification the change needs is much larger than the change|D|99e51eab 09-26|-|-|-|
|139|73|Verif|After a fix, rerun the failing check and affected scope; broaden to|B|ceab31eb 09-21|VER:102 (identical)|-|-|
|140|73|Verif|Do not mark a required but unavailable check as passed or silently|B|ceab31eb 09-21|G:68; VER:110|-|-|
|141|73|Verif|Distinguish failed, not verified/blocked, and not applicable with a reason; resolve authorized|B|ceab31eb 09-21|VER:110|-|-|
|142|73|Verif|Check that the handoff's paths and next step still match the final|B|7cc491b4 09-21|-|-|y|
|143|75|Verif|For UI work awaiting human validation, leave the verified built local stack|B|f469db22 09-21 (ed e2fb6e4e(O))|VER:79; AL (artifact-layout.md)|-|-|
|144|75|Verif|Pending human review means these resources are still needed.|B|f469db22 09-21|-|-|-|
|145|75|Verif|Once no longer needed, stop only task-started temporary processes and apply the|B|f469db22 09-21|AL; close:78|-|-|
|146|75|Verif|Report remaining resources and how to stop them, and respect an explicit|B|f469db22 09-21|-|-|-|
|147|77|Verif|Report what changed, what was verified, unresolved failures or uncertainty, and the|B|7cc491b4 09-21|G:62; RPT:51|-|y|
|148|77|Verif|Use a separate report only when a retained deliverable needs it; a|B|7cc491b4 09-21|-|-|y|
|149|77|Verif|A successful tool exit, a checked box, or an agent's assertion alone|B|7cc491b4 09-21|G:62|-|y|
|150|79|Verif|Complete delivery actions already explicitly authorized, subject to their conditions; for a|B|7cc491b4 09-21 (ed 15e17ae3(O))|GW:36|-|-|
|151|79|Verif|Otherwise stop at the authorized result with remaining delivery steps identified.|B|7cc491b4 09-21|G:28; GW:36|-|-|
|152|79|Verif|Run [flow-close] when any one of these happens: the work item reaches|O|01074e8f 09-29|close:14; G:108|-|-|
|153|79|Verif|Inside an iterative route that [flow-research] defines, a verified round is not|OD|d5164f71 10-11|research:42-73|-|-|
|154|79|Verif|It closes the change record (a pause leaves it on disk), cleans|O|7e8a627f 09-29 (ed 01074e8f(O))|close:3-27; G:23|-|-|
|155|79|Verif|The record's versioned delivery follows the closing route recorded in the delivery|O|7e8a627f 09-29|close:67; DD:22|-|-|

Read-reference sentences (10): #4 L12 (d2673600 10-12); #5 L12 (d2673600 10-12); #6 L12 (d2673600 10-12); #17 L18 (ceab31eb 09-21); #22 L22 (23c1c790 09-21); #24 L24 (dd3e7db1 09-21); #30 L30 (d2673600 10-12); #31 L30 (d2673600 10-12); #55 L40 (25d7d742 09-21); #131 L71 (ceab31eb 09-21)

#### close

|#|L|sec|rule (first 12 words)|cls|commit|dup|C|D|
|-|-|-|-|-|-|-|-|-|
|1|8|Intro|This skill grants no effect of its own.|U|7e8a627f 09-29|-|-|-|
|2|8|Intro|The global guidance authorizes the deletions it names (task temporaries, task-started processes,|U|7e8a627f 09-29|G:108|-|-|
|6|14|When|**Work item close:** the item reached its completion point (merged into the|O|7e8a627f 09-29|VER:88; G:32|T|-|
|7|15|When|**Session end:** the user ends the session.|O|7e8a627f 09-29|-|-|-|
|8|16|When|**Explicit request:** the user asks for hygiene or cleanup.|O|7e8a627f 09-29|-|-|-|
|9|16|When|Run the parts of 1 and 2 that apply to the current|U|7e8a627f 09-29|-|-|y|
|10|20|Item|Run these steps in order; the last one ends the turn.|O|7e8a627f 09-29 (chg work-close-sequence 09-26)|-|-|-|
|11|22|Item|**Tracker.** Apply the global ticket-state rule to the ticket and to each|U|7e8a627f 09-29|G:32|T|-|
|12|23|Item|**Change record.** When the work has a change folder, close it as|O|7e8a627f 09-29|plan:94; DD:22; CR:16|-|-|
|13|24|Item|**Git cleanup.** Run [Clean up after a merge] when you merged.|R|8e5fba29 09-23 (mv 7e8a627f)|close:66; GW:44|-|-|
|14|24|Item|When the change record goes to the base branch through a closing|O|7e8a627f 09-29|-|-|-|
|15|24|Item|Each time you remove a worktree, merged or abandoned, update its record|O|593fd162 10-13|GW:18|-|-|
|16|25|Item|**Temporaries and processes.** Remove only reproducible temporary files this task created and|O|7e8a627f 09-29|G:108; AL:41|-|-|
|17|25|Item|Keep prior material, unique evidence, and anything whose ownership or disposability is|B|b63fdcd9 09-20 (mv 7e8a627f)|G:108; AL:41; G:37-38|-|-|
|18|25|Item|Restore the local state the user relies on that this task changed,|O|f8c03823 10-06|G:108|-|-|
|19|25|Item|First check that restoring loses none of the user's data; if it|O|f8c03823 10-06|-|-|-|
|20|26|Item|**Incidental findings.** List the findings the global guidance says to propose, each|O|7e8a627f 09-29 (ed 34b4edc9(OD))|G:31|-|-|
|21|26|Item|Each finding accepts three answers: open a ticket; decline (saved in persistent|O|01074e8f 09-29|G:31|-|-|
|22|26|Item|A guidance finding about the workflow itself, which has no project ticket,|O|01074e8f 09-29|VER:84 (workflow-cost findings)|-|-|
|23|26|Item|On a recurrence of a declined finding, update its count, raise the|U|cb25d22b 09-27 (mv 590061dc)|G:31|-|-|
|24|27|Item|**Completion report.** Follow [reporting], with a cleanup line stating what was removed|O|c6a626fc 09-24 -> 7e8a627f (chg work-close-sequence 09-26)|RPT:49-51|-|-|
|25|28|Item|**Close question.** Follow [Close question], after the report.|O|af339ac1 09-24 (chg work-close-sequence 09-26)|-|-|-|
|26|32|Close|**Condition:** the turn completes a work item, after its report and closing|O|af339ac1 09-24 -> 722c7fe6 10-08 (mv 590061dc, d2673600)|RPT:93; G:23|-|-|
|27|34|Close|**Actions:** ask the close question under [questions], ending with the pause line.|OD|75fe4ce5 10-08 (mv d2673600)|-|-|-|
|28|34|Close|This question replaces the Recommendations route question.|M|d2673600 10-12|RPT:77-79|-|-|
|29|36|Close|**Exceptions:** skip it when the user already answered it or explicitly asked|D|0fee2d02 09-26 (mv 590061dc, d2673600)|RPT:93; G:23|-|-|
|30|36|Close|That report still carries its cleanup label, delivery state, and reminders.|D|0fee2d02 09-26 (mv 590061dc)|RPT:55-65; G:72|-|-|
|31|36|Close|Investigation-only findings use [flow-research]'s route question instead.|M|d2673600 10-12 (text first in global.md)|G:23 (identical); RPT:79,93|-|-|
|32|38|Close|For every applicable close question, name the next ticket and the reason,|O|60b9a547 09-30 (mv 590061dc, d2673600)|RPT:77|-|-|
|33|38|Close|You choose it; an option that asks the user to name it|O|60b9a547 09-30 (mv 590061dc)|-|-|-|
|34|40|Close|Offer one option per route, each continue option naming the flow it|U|af339ac1 09-24 (mv 590061dc)|-|-|-|
|35|42|Close|Continue now with the recommended ticket when the loaded context helps it.|U|af339ac1 09-24 (mv 590061dc)|RPT:77|-|-|
|36|43|Close|Take that ticket to a new session when the context no longer|U|af339ac1 09-24 (mv 590061dc)|-|-|-|
|37|44|Close|End the session, stating what stays open if it ends: pending items|U|b97a6d25 09-25 (mv 590061dc)|RPT:51; close:80-81|-|-|
|38|46|Close|**Condition:** the next work continues a plan still in progress.|O|722c7fe6 10-08 (mv d2673600)|RPT:91; G:23|-|-|
|39|48|Close|**Actions:** the continue option covers all of its remaining tasks, and names|O|722c7fe6 10-08|RPT:91|-|-|
|40|48|Close|A production deployment keeps its own question unless the user already named|O|722c7fe6 10-08|-|-|-|
|41|50|Close|**Condition:** the next ticket reuses neither the closed item's files nor its|U|591b85de 09-25 (mv 590061dc, d2673600)|-|-|-|
|42|52|Close|**Actions:** recommend the new session, but never for a plan still in|U|591b85de 09-25 (mv 590061dc; edit 722c7fe6(O))|RPT:91|-|-|
|43|56|Recor|Run this after the code is integrated and any required post-merge CI|O|7e8a627f 09-29|VER:88|-|-|
|44|56|Recor|While the work runs, `tasks.md` and the status in `proposal.md` are updated|O|7e8a627f 09-29|plan:65; build:48; DD:22|-|-|
|45|58|Recor|For each delta file, apply it to the current-requirements home.|R|1fbec0a3 09-23 (e1d9e79b 09-23; moved by 7e8a627f)|CR (change-records.md)|-|-|
|47|58|Recor|In `<specs>/specs/`: apply ADDED requirements to `<specs>/specs/<capability>/spec.md`, creating it when absent; replace|R|1fbec0a3 09-23 (mv 7e8a627f)|CR:55|-|-|
|48|58|Recor|Edit only those blocks, leave the rest of the file untouched, and|R|1fbec0a3 09-23 (mv 7e8a627f)|-|-|-|
|49|59|Recor|Set the change's status to closed in `proposal.md`, citing the commit or|U|7e8a627f 09-29|-|-|-|
|50|60|Recor|Move the folder to `<specs>/changes/archive/YYYY-MM-DD-<change-id>`, using the closing date.|R|1fbec0a3 09-23 (mv 7e8a627f)|-|-|-|
|51|60|Recor|Use `mv` when the folder was never tracked by Git, and `git|U|7e8a627f 09-29|-|-|-|
|52|60|Recor|Never recreate the files by rewriting them.|R|1fbec0a3 09-23 (mv 7e8a627f)|-|-|-|
|53|61|Recor|Update `<specs>/project.md` only if the phase, an open decision, or a blocker|R|1fbec0a3 09-23 (mv 7e8a627f)|-|-|-|
|54|62|Recor|Version the record once, as the cases below say.|O|7e8a627f 09-29|-|-|-|
|55|62|Recor|Stage by named path, never a directory (see [git-workflow]).|U|847acc14 09-25 (mv 7e8a627f)|-|-|-|
|56|64|Recor|Versioning cases:|O|7e8a627f 09-29|-|-|-|
|57|66|Recor|**Specs repository with `Delivery: direct-base`:** one commit and one push with the|O|7e8a627f 09-29|DD:18|-|-|
|58|67|Recor|**Specs inside the code repository, or in a separate specs repository that|O|7e8a627f 09-29 (ed 01074e8f(O))|DD:22|-|-|
|59|67|Recor|Act only on that recorded answer; repository policy does not replace it.|O|7e8a627f 09-29|DD:22|-|-|
|60|67|Recor|Order: return to the base branch and fast-forward it, make the record|O|7e8a627f 09-29|DD:22|-|-|
|61|67|Recor|If no closing route was recorded, ask for it before publishing.|O|7e8a627f 09-29|DD:22|-|-|
|62|68|Recor|**`hold`:** the record stays unversioned and is reported as pending, like the|O|7e8a627f 09-29|DD:22 (identical)|-|-|
|63|69|Recor|**Pause:** do not archive.|O|7e8a627f 09-29|CR|-|-|
|64|69|Recor|Leave the folder on disk with its current status and report it|O|7e8a627f 09-29|-|-|-|
|65|70|Recor|**Abandon:** archive without applying deltas, with the reason in `proposal.md`, and version|O|7e8a627f 09-29|-|-|-|
|66|74|Branc|After verifying a merge you performed, clean up its branches.|R|8e5fba29 09-23 (mv 7e8a627f)|GW:44|-|-|
|67|74|Branc|The global guidance grants the deletion of merged work branches and their|U|7e8a627f 09-29|G:108; AL:41|-|-|
|68|74|Branc|First confirm that no open pull request uses the work branch as|R|8e5fba29 09-23 (mv 7e8a627f)|GW|-|-|
|69|74|Branc|`gh pr merge --delete-branch` skips that check, so delete separately.|R|8e5fba29 09-23 (mv 7e8a627f)|-|-|-|
|70|74|Branc|Delete the merged remote branch unless the host already removed it, then|R|8e5fba29 09-23 (mv 7e8a627f)|-|-|-|
|71|74|Branc|Return the checkout to the base branch, run `git fetch --prune`, and|R|8e5fba29 09-23 (mv 7e8a627f)|-|-|-|
|72|74|Branc|After a promotion, also fast-forward each local environment branch it moved, such|U|591b85de 09-25 (mv 7e8a627f)|-|-|-|
|73|74|Branc|Delete merged preview or temporary branches that the work created.|U|591b85de 09-25 (mv 7e8a627f)|-|-|-|
|74|74|Branc|Apply the same checks to this repository's other local and remote branches|U|2afca607 09-23 (mv 7e8a627f)|-|-|-|
|75|74|Branc|A branch with unmerged commits is a decision, not noise: report its|R|8e5fba29 09-23 (mv 7e8a627f)|-|-|-|
|76|74|Branc|Report the branches checked, the deletions, the fast-forwards, and the unmerged branches|U|591b85de 09-25 (mv 7e8a627f)|-|-|-|
|77|78|End|Delete the UI review images this task created that the user did|O|7e8a627f 09-29|AL; VER:77|-|-|
|78|78|End|When the work item is paused and continues in another session, keep|O|f8c03823 10-06|AL; VER:77|-|-|
|80|79|End|Stop any task-started process that is still running.|O|7e8a627f 09-29|close:25|-|-|
|81|80|End|List the reminders the user still controls, and what stays open with|U|7e8a627f 09-29 (b97a6d25 09-25)|RPT:83-85; close:44|-|-|

Read-reference sentences (5): #3 L10 (d2673600 10-12); #4 L10 (d2673600 10-12); #5 L10 (d2673600 10-12); #46 L58 (7e8a627f 09-29); #79 L78 (d2673600 10-12)