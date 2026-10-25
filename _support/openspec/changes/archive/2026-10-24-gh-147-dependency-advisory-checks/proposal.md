# Seguridad de dependencias: comprobar avisos antes de instalar y revisar la exposición a un aviso (2026-10-24)

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` el 2026-10-24 por el [PR #148](https://github.com/JhonHawk/tricell-hive/pull/148), merge `b93af714`; #147 cerrado; registro archivado el mismo día |
| Tracker · GitHub Issues | • [#147 — Seguridad de dependencias: comprobar avisos antes de instalar y revisar la exposición a un aviso en el workspace](https://github.com/JhonHawk/tricell-hive/issues/147) |
| Git | `interactive` (D5-A) · rama desde `development` · parada antes del push para tu validación |
| Verificación | `hive-verify-task` por tarea · `go vet ./...`, `go test ./...`, `unittest` y `pnpm build` en `site/` sobre el candidato final · `/code-review` de Claude Code (D6-A) |
| Siguiente paso | Ninguno; registro cerrado |

## Solicitud

Mensaje del usuario, 2026-10-24:

> En hive legacy, teniamos algunas reglas para validar CVE antes de cada instalación de libreria, tenemos eso en este nuevo hive? creo que esto nos obliga a tener una herramienta de apoyo.
>
> - Considera tambien este articulo https://www.perplexity.ai/hub/blog/llm-security
> - Considera que en esta semana pasada, hubo muchos CVE criticos (Que de hecho tenemos que validar si nos agente), lo que me da la idea de que necesitamos una skill y agente de seguridad

Respuestas siguientes, 2026-10-24, en tarjetas nativas:

- D1-A: «Regla + skill»: antes de instalar, comprobar que el paquete existe y consultar OSV con niveles por severidad, más el periodo de espera; skill nueva de revisión de exposición que reutiliza `hive-review-security`; `osv-scanner` opcional, con `curl` si no está; sin agente nuevo. Endurecer la regla actual es una corrección: el usuario la adopta sin esperar una segunda sesión (AGENTS.md, Measurement).
- D2-A: revisión de solo lectura de los CVE de la semana en todos los repos de `~/Development/projects` (hecha el mismo día; sus hallazgos se reportaron en la conversación y pertenecen a la sesión de cada proyecto).
- D3-A: abrir el issue [#147](https://github.com/JhonHawk/tricell-hive/issues/147).
- D4-A: un puntero de una línea en `content/guidance/global.md`, para que la comprobación aplique también fuera de un build.
- D5-A: entrega `interactive`. D6-A: `/code-review` de Claude Code.
- D7-A: construir el plan (aprueba AC1–AC7).
- D8-A (durante el build): restaurar el lockfile del sitio tal cual y compilar, aunque `pnpm audit` muestre 11 avisos altos en dependencias transitivas; se propone un issue aparte.
- D10-A (validación): aprobar y publicar.
- D9-A (durante el build): corregir el paso 5 para que, en una restauración, los avisos ya presentes sean un hallazgo incidental sin cambiar versiones.

Criterios de aceptación de #147, citados textualmente (leído el 2026-10-24):

> - La guía de instalación de dependencias indica cómo consultar avisos por versión exacta y la regla de resolución por severidad, con la pregunta obligatoria en CRITICAL/HIGH sin camino seguro.
> - La guía pide comprobar existencia y origen del paquete antes de añadirlo.
> - Existe una skill distribuida para revisar la exposición a un aviso en un workspace, de solo lectura, que delega en `hive-review-security`, y respeta la regla de datos enviados a terceros (#140).
> - Hive no empaqueta ni exige ninguna herramienta: usa la instalada y cae a la API de OSV.

## Objetivo

El Hive anterior consultaba OSV.dev por versión exacta antes de cada instalación y decidía por severidad, preguntando siempre ante un aviso crítico o alto sin camino seguro (`03bf6aff:global/rules-situational/security.md`, sección *Supply Chain Security*). La reconstrucción conservó solo una frase genérica (`content/skills/flow-build/references/security-boundaries.md:8`): no dice cómo consultar, qué hacer según la severidad, ni cuándo preguntar. Tampoco hay control de que el paquete exista y sea el esperado, que es el control de cadena de suministro que destaca el [artículo de Perplexity](https://www.perplexity.ai/hub/blog/llm-security) para nombres que un modelo inventa.

La semana del 14 al 21 de octubre de 2026 mostró el segundo hueco: responder «¿nos afecta este aviso?» en todos los repos (Next.js GHSA-2464-fhr4-w9xw y GHSA-673q-mjm6-r5xr, Node.js CVE-2026-111497, compromiso de LiteLLM 1.107.0) exigió armar a mano un reparto entre agentes. `hive-review-security` revisa un cambio concreto, no una exposición transversal.

Tras el cambio, antes de instalar o actualizar un paquete, dentro o fuera de un build, el agente confirma el paquete, respeta el periodo de espera, consulta avisos de la versión exacta sin enviar nombres privados y resuelve por severidad. Una skill nueva, `security-advisory`, guía la revisión de exposición de solo lectura y deja la corrección a la sesión de cada proyecto.

## Alcance y aceptación

Incluido:

- Sección nueva en `security-boundaries.md`, hogar canónico de la comprobación previa a instalar.
- Punteros en `tools-and-dependencies.md` y en `global.md` (D4-A).
- Skill nueva `content/skills/security-advisory/SKILL.md`.
- Lista de skills de apoyo en `site/src/content/docs/skills.md` e `index.mdx`, y entrada en `CHANGELOG.md`.

Excluido:

- Un rol de agente nuevo: `hive-review-security` ya cubre dependencias (D1-A).
- Empaquetar, instalar o exigir herramientas (`osv-scanner`, `govulncheck`, Socket), hooks o configuración global.
- Remediar los repos afectados por los avisos de la semana: pertenece a la sesión de cada proyecto.
- Deltas de requisitos: el cambio es guía distribuida, como en los cambios anteriores de guía (p. ej. `2026-10-24-gh-143-144-…`), y ninguna capacidad en `_support/openspec/specs/` cubre dependencias.

Restricciones que deben seguir valiendo: `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` en verde; cada `SKILL.md` en 16 000 bytes o menos; el sitio compila.

Criterios de aceptación (base `5894fbf4`):

- AC1. `security-boundaries.md` tiene una sección «Check a dependency before installing» que pide confirmar que el nombre exacto existe y es el paquete buscado antes de instalarlo, incluida una ejecución puntual como `npx`. *Falso en la base:* `rg -c "Check a dependency before installing" content/` no da coincidencias.
- AC2. Esa sección indica cómo consultar avisos: antes de instalar un paquete público, la API de OSV para su versión exacta; para el lockfile, el escáner del proyecto, `osv-scanner` en modo offline cuando está instalado, o la auditoría del gestor solo cuando todas las dependencias vienen de su registro por defecto; nunca un nombre de paquete privado hacia OSV. *Falso en la base:* `rg -ci "osv" content/` no da coincidencias.
- AC3. La sección resuelve por severidad: versión corregida compatible que el periodo de espera acepta → instalarla y reportar; MEDIUM/LOW sin corrección → seguir y reportar; CRITICAL/HIGH sin camino seguro o registro malicioso → preguntar siempre; un aviso sin calificación ni alias calificado cuenta como CRITICAL/HIGH; en cualquier instalación, los avisos en versiones que la instalación no cambia se reportan como hallazgo incidental sin cambiar versiones, y ante CRITICAL/HIGH o `MAL-` se pregunta antes de instalar (D9-A). *Falso en la base:* `rg -c "CRITICAL or HIGH" content/` no da coincidencias.
- AC4. La sección pide respetar el periodo de espera del proyecto también en ejecuciones puntuales, nombrar la versión exacta comprobada y no rebajar ni saltar el periodo sin preguntar. *Falso en la base:* `rg -ci "minimumReleaseAge|release cooldown" content/` no da coincidencias.
- AC5. `tools-and-dependencies.md` y `global.md` remiten a esa sección sin repetir sus pasos; el puntero global aplica también fuera de un build. *Falso en la base:* `rg -c "Check a dependency before installing" content/guidance/global.md content/skills/flow-build/references/tools-and-dependencies.md` no da coincidencias.
- AC6. Existe `content/skills/security-advisory/SKILL.md`: revisión de exposición de solo lectura que lee el aviso de una fuente primaria, delega el inventario en hijos `hive-research` cuando excede una lectura directa corta, usa `hive-review-security` para el alcance del código vulnerable, reporta por repo con veredicto y localizador, remite a la regla de datos enviados a terceros y deja la corrección a la sesión de cada proyecto. *Falso en la base:* el archivo no existe.
- AC7. Las dos listas de skills de apoyo del sitio nombran `security-advisory`, y `CHANGELOG.md` (Unreleased · Added) nombra la skill y la comprobación previa a instalar. *Falso en la base:* `rg -c "security-advisory" site/src CHANGELOG.md` no da coincidencias.

## Entrega

- Repositorio `tricell-hive`, base `development`. Modo `interactive` (D5-A, 2026-10-24): rama de trabajo, commits locales y verificación; parada antes del push para tu validación; tras ella, push, PR, `/code-review` de Claude Code (D6-A) sobre el diff de la rama y merge a `development` cuando pasen las comprobaciones locales.
- Tras el merge, como el marcador de mantenedor existe, `hive update` despliega el contenido (refresco local; no es un release).
- El registro del cambio se versiona al cerrar con un PR pequeño sin revisión dedicada que el agente fusiona.
- Mover #147 a cerrado al integrar en `development`.
