# session-identity-subagents — Identidad de sesión del visor con subagentes de Grok e hilos del daemon de Codex

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #398, merge 6b70dca2; H21 in PR #399, merge efa81bec |
| Tracker | ninguno (trabajo sin ticket); seguimiento de Cursor propuesto aparte |
| Git | `JhonHawk/tricell-hive`, base `development` en `da54bfec` (origin); rama propuesta `fix/session-identity-subagents` |

## Decisiones

2026-11-05
- **Dirección del usuario** — ¿Quién publica en el visor cuando hay subagentes? Solo la sesión principal; los hijos devuelven su resultado al padre y no reciben instrucciones ni identidad del visor.
- **D1 A** — ¿Cómo se tratan los eventos de un subagente de Grok? Como en Codex: el inicio del subagente se registra como `subagent_started` de la sesión padre y todo evento propio del hijo se descarta, sin tocar el mapa ni crear actividad de otra sesión. El visor no muestra el fin del subagente en Grok.
- **D2 A** — ¿Qué fuente de identidad usa Codex para los hilos de su daemon? La variable `CODEX_SESSION_ID` (sesión raíz), aceptada por el código fuente de Codex 0.167 y comprobada con una ejecución real de `codex exec` que la compare con el `session_id` del hook.

- **D3 A** — ¿Cómo se entrega? Modo `automatic`: revisión, push, PR y merge sin parada de validación.
- **D4 A** — ¿Qué revisión de código corre? La nativa de Claude Code, `/code-review`.

## Solicitud

> «Analiza la sesión sobre grok en carvento, no pudo generar la sesión porque no tiene id» · «y esa sesión en el visor, fuera de abrirse no recibió updates» · «A» (corregir Grok y Codex ahora, Cursor como seguimiento) · «Que el seguimiento del visor sea tarea solo del padre».

Fuera de alcance: Cursor (sus subagentes no están verificados), guardar el `work_id` en el frontmatter del plan, y cerrar la work huérfana «Retirar código que nadie usa» de carvento.

## Objetivo

Hoy, en Grok 1.0.77, un subagente corre en el proceso del padre con su propio `sessionId`; su `UserPromptSubmit` lo deja como única sesión viva del mapa y su `SessionEnd` deja cero. Las publicaciones del padre se atribuyen al subagente (`accepted_producer_required`) o se rechazan (`native identity ambiguous`). En Codex, el daemon `app-server` atiende varios hilos en un proceso y el último hilo que recibe un prompt desplaza a los demás en el mapa. Además, dos variables de identidad (host anidado) se rechazan siempre como ambiguas.

Al terminar, la sesión principal de Grok conserva su identidad con subagentes y el visor ve el inicio del subagente como actividad suya; cada hilo de Codex se identifica por su propia variable; y un host anidado resuelve la variable de su host más cercano.

## Alcance y aceptación

En alcance: el traductor de hooks de Grok, el resolver de identidad (variable de Codex y selección entre varias variables), sus tests, el delta del spec `walkthrough-viewer` y las dos referencias de identidad de host.

- AC1 — En Grok, durante y después de un subagente, `hive view work` sin flags sigue resolviendo la sesión principal · falso hoy porque la sonda `01a1af4a` (mapa `82328.json`) dejó al subagente como única sesión viva a las 05:18:10Z y cero sesiones vivas a las 05:18:18Z, y carvento recibió `accepted_producer_required` y `native identity ambiguous`.
- AC2 — El `SubagentStart` de Grok llega al visor como `subagent_started` de la sesión padre con el `subagentId`, también sin los hooks de Claude Code, y ningún evento que traiga `subagentType` escribe en el mapa ni se reporta con el id de sesión del hijo · falso hoy porque el hook de Grok no registra `SubagentStart` (`integrations/grok/hook_register.go`), el traductor lo trata como actividad genérica sin agente y el `UserPromptSubmit` del hijo quedó registrado en `82328.json`.
- AC3 — En una shell de un hilo de Codex, `hive view work` sin flags resuelve la identidad desde `CODEX_SESSION_ID` aunque otro hilo del mismo daemon haya recibido un prompt después · falso hoy porque `viewIdentityVariables` no incluye Codex y el mapa `35839.json` muestra que `01a1aebf` desplazó a `01a1adfa` mientras esta seguía trabajando.
- AC4 — Con dos variables de identidad presentes (un host lanzado desde la shell de otro), se usa la del host del proceso más cercano, reconocido también bajo un intérprete, y no se rechaza como ambiguo · falso hoy porque `resolveViewIdentity` devuelve `errViewIdentityAmbiguous` en cuanto encuentra más de una variable, por ejemplo Pi (proceso `node`) lanzado desde Claude Code con `PI_SESSION_ID` y `CLAUDE_CODE_SESSION_ID`.
- AC6 — Un host sin variable propia (Grok o Cursor) lanzado desde la shell de otro host resuelve por su propio mapa aunque herede la variable del otro · falso hoy porque `confirmVariableHost` rechaza la variable heredada como ambigua y no prueba el mapa: un Grok lanzado desde Claude Code no puede publicar.
- AC5 — `host-session-identity.md` da `CODEX_SESSION_ID` como la variable de Codex, y `host-session-events.md` describe los payloads de subagente de Grok y el daemon de Codex · falso hoy porque la fila de Codex nombra `CODEX_THREAD_ID` sin verificar y la documentación dice que los subagentes de Grok comparten la sesión del padre.

## Prerrequisitos

- [ ] Codex CLI 0.167 con sesión iniciada en esta máquina y una ejecución corta de cuota autorizada por D2 A · owner: el usuario (ya instalado y autenticado) · needed by: T4

## Contexto verificado

- Las fechas del documento son hora local (UTC−6); los timestamps citados están en UTC, que ya es 2026-11-06.
- Sesión Grok de carvento `01a1aed2-ced4-79d2-bc52-4b55d46873bf` (2026-11-05 local): `open` a las 03:52:28Z; `sync` rechazado a las 03:52:35Z, `accepted_producer_required` a las 03:52:42Z y 03:55:55Z y `native identity ambiguous` a las 03:59:10Z, con el subagente `01a1aefc` vivo de 03:52:35Z a 03:58:21Z.
- Sonda con Grok 1.0.77 en un repositorio desechable: payloads capturados con hook de proyecto. El `SubagentStart` lleva el `sessionId` del padre más `subagentId` y `subagentType`; el `UserPromptSubmit`, el `SubagentStop` y el `SessionEnd` del hijo llevan el id del hijo y `subagentType`; los eventos del padre no llevan `subagentType`.
- `integrations/codex/hook_event.go`: patrón de referencia (eventos con `agent_id` del subagente descartados salvo inicio y fin).
- `codex-rs/core/src/exec_env.rs:41` en `rust-v0.167.0`: `inject_session_env` pone `CODEX_SESSION_ID` en las shells. El binario instalado contiene la cadena. No hay documentación pública.
- Claude Code y los subagentes de Codex comparten el `session_id` raíz; OpenCode y Pi no usan el mapa (`hookUsesMap`); Cursor queda sin verificar (investigación delegada, 2026-11-05).
- Documentación de Grok 1.0.77 (`~/.grok/docs/user-guide/10-hooks.md:96`, `:106`, `:110`): `SessionStart` no se dispara para un subagente, `SubagentStart` existe, y `SessionEnd` trae `subagentType` para una sesión hija.
- El hook de Grok que registra Hive (`integrations/grok/hook_register.go`, `sessionEventNames`) no incluye `SubagentStart`. En esta máquina el evento llega por los hooks de Claude Code en `~/.claude/settings.json`, que Grok también ejecuta (compatibilidad activa por defecto).
- El spec «Host-observed events» fija que los eventos de host que no son de tareas solo mueven la última actividad; el visor valida `subagent_*` (`tooling/viewer/timeline_contract_v3.go:2082`) sin otro efecto.
- El resolver reconoce el proceso host solo por el nombre del comando (`hostOfProcessName`). Pi corre como `node` y solo lo reconoce el escritor del mapa (`processHost`, `hook_map.go`).
- `site/src/content/docs/viewer.md:49` describe la resolución sin nombrar hosts ni variables: no cambia.

## Entrega

- **D3 A** — `automatic`: tras la revisión de código, push de `fix/session-identity-subagents`, PR a `development` y merge, sin parada de validación.
- **D4 A** — Revisión nativa de Claude Code `/code-review` sobre el candidato final, antes del push.
- No autoriza: promoción a `master`, release, ni escribir en trackers (el ticket de Cursor es una acción aparte).

## Estado de revisión y progreso

2026-11-05 — Plan escrito.
2026-11-05 — Revisión `hive-review-plan`: 2 P1, 3 P2, 6 P3, todos aplicados. Cambios:
- registro de `SubagentStart` en el hook de Grok (T1, AC2);
- reconocimiento de Pi bajo `node` (T3, AC4);
- vuelta al mapa para un host sin variable, como decisión menor (AC6, nuevo);
- T4 compara también con el id de hilo y va antes de T3;
- la prueba en rojo de T2 restaura el archivo de la base en vez de usar `git stash`;
- el spec acota los subagentes a Claude Code, Codex y Grok;
- citas de la documentación de Grok, de las fechas y de la sesión en el commit.
