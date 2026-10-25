# Diseño

## Contexto verificado

Base `5894fbf4` (`development`, 2026-10-24).

- **Regla actual.** La única regla sobre avisos está en `content/skills/flow-build/references/security-boundaries.md:8`. `tools-and-dependencies.md:13` cubre necesidad y encaje de una dependencia, no su origen ni sus avisos. Leen `security-boundaries.md`: `flow-build/SKILL.md:22`, `flow-plan/SKILL.md:24`, `hive-review-security.md:11` y `hive-review-code.md:20`. `flow-build/SKILL.md:26` manda leer `tools-and-dependencies.md` antes de instalar. Los roles de construcción (`content/agents/development/*`) instalan dependencias como hijos y nunca hablan con el usuario.
- **Regla anterior.** `03bf6aff:global/rules-situational/security.md`, sección *Supply Chain Security*: consulta a OSV por versión exacta antes de cada instalación, el mapa de ecosistemas, los tres niveles de severidad y la auditoría tras instalar desde el lockfile. `03bf6aff:global/core-sections/authorization-safety.md:12` la resumía en la guía siempre cargada.
- **Datos enviados a terceros.** `content/guidance/global.md:31` (#140) incluye explícitamente las bases de vulnerabilidades y los registros de paquetes como servicios de terceros, y el nombre de un paquete privado como identificador privado.
- **`npm audit`** ([documentación de npm](https://cdn.jsdelivr.net/npm/npm@10.8.2/docs/output/commands/npm-audit.html)) envía el nombre y las versiones de cada paquete del árbol al registro configurado por defecto; desde npm 7 ya no omite los nombres con scope. Con paquetes privados de un registro por scope, esos nombres terminan en el registro público. Para `pnpm audit` se infiere lo mismo (no verificado).
- **Severidad en OSV.** El campo `severity` es opcional en el esquema; la base de Go no publica severidad ([API de OSV](https://google.github.io/osv.dev/post-v1-query/), [base de Go](https://go.dev/security/vuln/database)).
- **Enlaces `skill:` en `global.md`.** `tests/content/skill_links_test.go:13-25` resuelve cada destino con `os.Stat`, así que un ancla (`…md#section`) rompería la prueba: el puntero global enlaza el archivo y nombra la sección en el texto.
- **Despliegue de una skill nueva.** `tooling/management/plan.go:230` toma todos los `content/skills/*/SKILL.md`; `plan.go:368` exige nombre en kebab-case; `tooling/management/references.go:48-95` resuelve los enlaces relativos. No hay registro por host que editar. `tests/content/skill_size_test.go` aplica el límite de 16 000 bytes a toda skill.
- **Sitio.** No hay página por skill. Las skills de apoyo se listan en `site/src/content/docs/skills.md:17-26` e `index.mdx:79-90`.
- **osv-scanner** (Context7 `/google/osv-scanner`, v2.6.0, README y `docs/offline-mode.md`; [modo offline](https://google.github.io/osv-scanner/usage/offline-mode/), consultado el 2026-10-24; la última versión publicada es v2.7.0): `--offline` no envía información del proyecto ni de las dependencias; `--download-offline-databases` descarga solo el zip de cada ecosistema y funciona únicamente con `--offline`. Comando: `osv-scanner --offline --download-offline-databases <project-dir>`.
- **API de OSV** (`https://google.github.io/osv.dev/api/`): `POST https://api.osv.dev/v1/query` con `{"version", "package": {"name", "ecosystem"}}`. Que devuelva registros `MAL-` de paquetes maliciosos en npm y PyPI viene del blog de OpenSSF; no se confirmó con fuente primaria.
- **pnpm** (Context7 `/websites/pnpm_io`, `pnpm install`, consultado el 2026-10-24): `pnpm install --lockfile-only` solo actualiza `pnpm-lock.yaml` y `package.json`, sin escribir en `node_modules`. `minimumReleaseAge` va en minutos (desde v10.16.0; 1440 por defecto desde v11); en esta máquina vale 10080 a nivel global.
- **govulncheck** `./...` consulta la base de Go, cuyos endpoints no reciben rutas de módulos.

## Enfoque

- **Un hogar canónico.** El procedimiento completo vive en una sección nueva de `security-boundaries.md`, la referencia de seguridad que ya leen el plan, el build y los dos roles de revisión. La viñeta de la línea 8 se queda con lo que no es la comprobación (scripts de instalación, exposición, controles de lockfile, procedencia y avisos del proyecto, remediación automática) y remite a la sección. `tools-and-dependencies.md` y `global.md` solo remiten, sin repetir pasos (AGENTS.md, «one canonical home»).
- **Puntero global (D4-A).** Hace falta porque una instalación puntual (`npx`, `uvx`, `pnpm dlx`) o un paquete agregado por una petición conversacional ocurre sin cargar `flow-build`. El Hive anterior lo tenía en el núcleo siempre cargado.
- **Paquetes nombrados frente a restaurar el lockfile.** Los pasos de confirmación, periodo de espera y consulta por versión aplican a cada paquete nombrado en el comando o agregado a un manifiesto. Una restauración sin paquete nombrado solo escanea el lockfile y resuelve por severidad, para que un modelo literal no intente confirmar cada entrada.
- **Transitivas antes de instalar.** Cuando el gestor puede actualizar el lockfile sin instalar (`pnpm install --lockfile-only`, `npm install <name>@<version> --package-lock-only`), se escanea antes de que corran los scripts de las dependencias transitivas; si no puede, se escanea después y se reporta ese límite. El Hive anterior tenía el mismo hueco.
- **Sin herramienta empaquetada y sin nombres privados hacia terceros.** La consulta previa usa la API de OSV solo para paquetes públicos. El escaneo del lockfile usa `osv-scanner` en modo offline si está instalado, o el escáner que el proyecto ya configure. La auditoría del gestor (`pnpm audit`, `npm audit`) solo se usa cuando todas las dependencias vienen de su registro por defecto, porque envía cada nombre del árbol a ese registro; con paquetes de otro registro, el escaneo se reporta como no ejecutado.
- **Sin calificación de severidad.** Toma la de un alias CVE o GHSA; si no hay, se trata como CRITICAL/HIGH, para que la pregunta obligatoria no se pierda.
- **Ejecuciones puntuales y periodo de espera.** Se nombra la versión exacta comprobada (`npx <name>@<version>`) y debe cumplir el periodo de espera aunque el ejecutor no lo aplique. Una versión corregida que el periodo de espera todavía rechaza cuenta como «sin camino seguro». La validez en la sesión se limita al mismo paquete y versión exacta.
- **Hijos delegados.** Donde la sección pide preguntar al usuario, un hijo se detiene y reporta a su padre.
- **Skill de exposición.** `security-advisory` es de solo lectura y sigue la forma de `harness-audit` (alcance, procedimiento numerado, reporte, límites). Delega el inventario solo cuando la revisión excede una lectura directa corta, bajo `delegation.md`, sin repetir sus reglas de reparto; usa `hive-review-security` solo para el alcance del código vulnerable donde decide la prioridad. A ambos hijos les pasa la línea de red: búsquedas solo con IDs de aviso y nombres públicos, sin auditorías ni escaneos con red. La corrección se reporta para la sesión de cada proyecto (global.md, «Change only the current project»).
- **Sin deltas de requisitos.** Ninguna capacidad en `_support/openspec/specs/` cubre la guía de dependencias, y los cambios de guía anteriores no agregaron deltas.

Alternativa descartada: un rol nuevo de agente de seguridad. `hive-review-security` ya tiene dependencias en su descripción, acceso de solo observación y esfuerzo máximo; lo que faltaba era el procedimiento, no el rol (D1-A).

## Lugares revisados (para el PR)

Comportamiento afectado: qué hace el agente antes de instalar o actualizar un paquete, y cómo revisa la exposición a un aviso.

- `content/guidance/global.md:31` (datos a terceros) y `:81` (patrón de puntero «Before …, read [...]»).
- `content/skills/flow-build/SKILL.md:22`, `:26`, `:58`; `content/skills/flow-plan/SKILL.md:24`.
- `content/skills/flow-build/references/security-boundaries.md`, `tools-and-dependencies.md`.
- `content/agents/review/hive-review-security.md`, `hive-review-code.md:20`; `content/agents/development/*` (instalan como hijos).
- `content/skills/flow-research/references/delegation.md:11` (reparto de hijos, al que remite la skill).
- `site/src/content/docs/` (sin texto sobre dependencias; listas de skills en `skills.md` e `index.mdx`), `README.md` y `llms.txt` (solo listan los cuatro flujos), `_support/docs/architecture/` (sin texto sobre dependencias), `tests/pilot/` (sin detector de dependencias).

## Redacción propuesta

### `security-boundaries.md`

Viñeta de la línea 8, reemplazada:

```markdown
- Before introducing, updating or installing a dependency, run [Check a dependency before installing](#check-a-dependency-before-installing). Also check its installation and build scripts and its runtime or build exposure, and use the project's existing lockfile, provenance and advisory controls. Review automated remediation before running it: it may execute project-controlled scripts or use configured registries.
```

Sección nueva, al final del archivo:

```markdown
## Check a dependency before installing

Run this before you add, update or install a package or tool from a language package registry, such as npm, PyPI, Maven Central, the Go module proxy, crates.io, NuGet, RubyGems or pub.dev, inside or outside a build, including a one-off run such as `npx`, `uvx` or `pnpm dlx`. An install can run the scripts of the package and of its dependencies on this machine, so the check comes first, wherever the project will run. A system package manager such as Homebrew, apt or winget is outside this section: for it, confirm the package name and its source as step 1 sets.

Which steps apply:

- A package you name on the command or add to a manifest: all five steps.
- An update with no package argument, such as `pnpm update`: steps 2, 4 and 5, for the versions it changes.
- An install that only restores the lockfile, such as `pnpm install` with no package argument: steps 4 and 5.
- A one-off run or a global install, which has no project lockfile: steps 1, 2, 3 and 5.

Where this section says to ask the user, a delegated child stops and reports to its parent instead, and the parent passes the user's answer to the children it launches later. A check holds for the same package and exact version for the rest of the session, and the user's answer about advisories already in a lockfile holds for that unchanged lockfile for the rest of the session.

1. **Confirm the package.** Check that the exact name exists in its registry and is the package you mean: its repository, publisher and description match the purpose, and the name is not a near-miss of a better-known package. A name a model suggested, including your own, is unverified until this check. For a private or internal package, confirm instead that the project's configuration maps its scope or index to the private registry, so the public registry cannot serve it, such as a scope entry in `.npmrc` or a single private index for pip. Do not look the private name up in a public registry: that request sends it to a third party, as the global rule on requests to third-party services sets. When no such mapping exists, a public package under the same name could be installed instead: say so and ask the user.
2. **Pick a version the release cooldown accepts.** When the project or the user's package-manager configuration sets a minimum release age, such as pnpm's `minimumReleaseAge`, pick a version at least that old, also for a one-off run whose runner does not apply it. Never lower or bypass the cooldown without asking the user. For a one-off run, name that exact version, such as `npx <name>@<version>`, never `@latest` or no version, so what runs is what you checked.
3. **Look up advisories for that exact version.** A version can be safe while the next patch is compromised. Query `POST https://api.osv.dev/v1/query` with `{"version": "<version>", "package": {"name": "<name>", "ecosystem": "<ecosystem>"}}`. Ecosystems: npm, pnpm, yarn and bun → `npm`; pip and uv → `PyPI`; Maven and Gradle → `Maven`; Go modules → `Go`; Cargo → `crates.io`; NuGet → `NuGet`; RubyGems → `RubyGems`; Pub → `Pub`. Use the name the ecosystem publishes: the scoped name for npm, such as `@scope/name`, `groupId:artifactId` for Maven, and the module path for Go; a short name returns an empty result that looks clean. Send only a public package's name: never query OSV with a private or internal package, as the global rule on requests to third-party services sets.
4. **Scan the lockfile**, since the package's dependencies install with it and run their own scripts.
   - When the install only restores an existing lockfile, scan it before installing. When the package manager can update the lockfile without installing, such as `npm install <name>@<version> --package-lock-only`, or `pnpm install --lockfile-only` after adding the exact version to the manifest, do that and scan before installing. Otherwise scan right after the install and report that the scan came after it. When that later scan finds what step 5 would ask about, stop: do not run, import or build with the package, and ask the user; for a `MAL-` record, also say that its install scripts may already have run on this machine.
   - Use only a scanner that keeps private package names off third-party services, in this order: the project's configured advisory scanner when it meets that condition; otherwise `osv-scanner --offline --download-offline-databases -r <dir>` when `osv-scanner` is installed, where `<dir>` holds the lockfile, such as a workspace root, since offline mode sends no project or dependency names; otherwise the package manager's own audit, such as `pnpm audit` or `npm audit`, only when every package name in the tree, including the project's own name and its workspace packages, is public and comes from that manager's default registry, because the audit sends the tree's package names there. For Go modules, also run `govulncheck ./...` when it is installed.
   - An empty result means no known advisory, not a safe package. A record whose ID starts with `MAL-` marks a known malicious package.
5. **Resolve by severity**, using the advisory's rating. An advisory without a rating takes the rating of a CVE or GHSA alias; with none, treat it as CRITICAL or HIGH. In any install, advisories on versions this install does not change are an incidental finding, not this task's change, as the global rule on incidental findings sets: do not change versions to resolve them; report them, and ask the user before installing when one is CRITICAL or HIGH or a `MAL-` record. When the scan could not run in an install that only restores the lockfile, say so and proceed. For the packages you name and the versions the install changes:
   - A compatible fixed version that the cooldown accepts exists, meaning a patch or minor, or a major you verified as non-breaking: install that version and report the change. A fix that exists only in a new major is a version-bump decision: for MEDIUM or LOW, proceed and report the bump; for CRITICAL or HIGH, treat it as having no safe path, and so is a fixed version the cooldown still rejects.
   - MEDIUM or LOW with no fixed version: proceed and report the severity and the affected range.
   - CRITICAL or HIGH with no safe path, or a `MAL-` record: stop and ask the user before installing, every time. The install itself is the exposure, and a compromised install script cannot be rolled back.
   - A private package confirmed in step 1: proceed and report it as not checked against advisories. Only its own lookup is skipped; its public dependencies still go through step 4 and these rules.
   - Any other check that could not run, for lack of a tool, network, an ecosystem in the list above or a scanner that keeps names private: say so and ask the user before installing.

Report what this check changed or left open in the work's report: the versions you swapped, the advisories you accepted, and the packages or scans you could not check.
```

### `tools-and-dependencies.md`

Al final de «Install a missing tool» y de «Add a dependency», la misma frase:

```markdown
Before installing it, run [Check a dependency before installing](security-boundaries.md#check-a-dependency-before-installing).
```

### `global.md`

Una viñeta nueva en «Proportionality», justo antes de la de `browser automation` (`:81`):

```markdown
- Before you add, update or install a package or tool from a registry, including a one-off run such as `npx` and outside a build, read and run the section "Check a dependency before installing" in [security boundaries](skill:flow-build/references/security-boundaries.md).
```

### `content/skills/security-advisory/SKILL.md`

```markdown
---
name: security-advisory
description: Check whether a published security advisory, CVE or package compromise affects a set of repositories, and report each one's exposure with evidence, without changing them. Use when the user asks whether an advisory or a supply-chain incident affects their projects, or asks for an exposure sweep across repositories; not for reviewing a change's security boundaries or for checking a package before installing it.
---

# Security advisory exposure

Answer "does this advisory affect us, and where?" with evidence for each repository. The sweep is read-only: it changes no repository, dependency, tracker or deployment. Fixes run afterward, in each project's own session.

## Frame the sweep

Before delegating, read [delegation and waits](../flow-research/references/delegation.md). Before the report, read [reporting](../flow-close/references/reporting.md); before a user decision, read [questions](../flow-plan/references/questions.md). Requests to advisory databases, registries and search engines follow the global rule on requests to third-party services: send advisory IDs and public package names, never the name of a repository, client or private package.

1. **Read each advisory from a primary source**: the vendor's or project's advisory, GHSA, OSV, NVD, or the CISA KEV catalog for exploitation in the wild. Record its ID, product and ecosystem, affected and fixed versions, any mitigation, the conditions it needs (platform, feature, configuration), its severity, and whether it is exploited. Mark a fact you found only in secondary reporting, and say so in any verdict that rests on it.
2. **Resolve the scope**: the repositories or folders the user names. When the user names none, ask. Reading other projects for this sweep is allowed, as the global rule on changing only the current project sets.

## Collect the evidence

3. **Inventory each repository.** When the sweep covers more than a short direct read, delegate it to `hive-research` children, split as [delegation and waits](../flow-research/references/delegation.md) sets. Give the inventory, yours or each child's, these instructions together with the advisory facts from step 1:
   - For each repository and application, record the current branch, because only the working tree is read.
   - Record the version the lockfile resolves, not the manifest range, with its `path:line`.
   - When the advisory affects a runtime, record its pins: `.nvmrc`, `.node-version`, `engines`, Dockerfile base images, CI and hosting build settings.
   - Record whether the advisory's conditions appear, such as a configuration flag, a feature in use or the platform.
   - Run no install, audit, networked scan, Git fetch or checkout; web lookups carry only advisory IDs and public package names; never print a file that may hold secrets.
   - Report what cannot be seen from the repository, such as deployed versions, images behind a digest or hosting settings.
4. **Check the deciding claims.** Open the locators behind each affected or mitigated verdict, as [flow-research](../flow-research/SKILL.md) sets for checking a child's claims.
5. **Assess reachability where it sets priority.** For an affected application whose exposure depends on how the code reaches the vulnerable path, launch a `hive-review-security` child with the application, the advisory and its conditions, asking it to trace untrusted input to that path without exploiting anything, under the same limits on commands and web lookups as step 3.

## Report

Lead with the affected repositories. Report critical or actively exploited exposure as soon as it is confirmed, not only at the end. For each repository and application, give:

- the verdict: affected, mitigated (and by what), not affected, or unknown (and what would settle it);
- the evidence locator;
- the fixed version or mitigation it needs, and anything in the repository that blocks it, such as a version cap or override;
- reachability, when step 5 assessed it.

State the coverage: repositories and branches read, sources consulted, and what was not checked.

## After the report

The fixes belong to each project's own session, under the global rule on changing only the current project: report, for each affected project, the fixed version or mitigation it needs as a finding for that session. When the current project is affected, end with the route question that [flow-research](../flow-research/SKILL.md) sets, with the fix as one of its alternatives.
```

### Sitio y `CHANGELOG.md`

- `site/src/content/docs/skills.md`, en «Supporting skills», tras «Harness audit»: `- [Security advisories](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/security-advisory/SKILL.md): check whether a published advisory affects your repositories, without changing them.`
- `site/src/content/docs/index.mdx`, tras `harness-audit`: `- [security-advisory](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/security-advisory/SKILL.md): find which of your repositories a published advisory or package compromise affects, with the evidence for each one; the fixes stay in each project's own session.`
- `CHANGELOG.md`, `## [Unreleased]` · `### Added`, dos entradas:

```markdown
- Check a dependency before installing: before adding, updating or installing
  a package, including a one-off `npx` run, the agent confirms the package,
  picks a version the release cooldown accepts, looks up advisories for the
  exact version without sending private package names to third parties, scans
  the lockfile, and asks before installing when a critical or high advisory
  has no safe path.
- The `security-advisory` skill checks, read-only, which repositories a
  published advisory or package compromise affects, with evidence for each
  one, and leaves the fixes to each project's own session.
```
