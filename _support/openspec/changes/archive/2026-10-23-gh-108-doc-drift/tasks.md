# Tareas · gh-108-doc-drift

Commit base: `2d89940e` (`development`); rama `docs/gh-108-doc-drift`.

Todas las tareas editan documentación sin cambiar comportamiento: no llevan enfoque de prueba y se marcan con la evidencia de su propia comprobación, sin `hive-verify-task`. Todas las ejecuta el hilo principal, porque cada brief sería más largo que el cambio.

### T1 — Spec de instalación sin pin de pi-subagents

- [x] El delta `specs/versioned-installation/spec.md` de esta carpeta (ya escrito con el plan, copiado del spec vigente sin el pin) se aplica al spec vigente al cierre, antes de marcar el cambio como Completado, como dice [change records](../../../../content/skills/flow-plan/references/change-records.md).

**Closes:** AC1.

**Locations:** `_support/openspec/changes/gh-108-doc-drift/specs/versioned-installation/spec.md`; al cierre, `_support/openspec/specs/versioned-installation/spec.md` (requisitos «Pi user-scope install declares pi-subagents» y «Verified release package installation»).

**Verification:** antes del cierre, `rg -c "0.74.0|pinned pi-subagents" <delta>` no devuelve nada y los dos nombres de requisito coinciden con el spec vigente; tras aplicarlo, `rg -c "0.74.0|pinned pi-subagents" _support/openspec/specs/versioned-installation/spec.md` no devuelve nada y `git diff` del spec solo toca esas tres líneas.

### T2 — Política de soporte de seguridad

- [x] `SECURITY.md` declara la última versión minor publicada (D11-A).

**Closes:** AC2.

**Locations:** `SECURITY.md:3-9`; `CHANGELOG.md` (`[Unreleased]`).

**Changes:** la tabla queda con dos filas: «Latest published minor release (fixes ship as a new patch release of it; upgrade to that patch)» → Yes; «Earlier minor releases» → No; y una frase que remite a GitHub Releases para saber cuál es la última. Entrada `### Changed` en `CHANGELOG.md` `[Unreleased]` con `#108`.

**Verification:** `rg -n "0\.[0-9]\.x" SECURITY.md` no devuelve nada; `rg -n "Latest published minor release" SECURITY.md` y `rg -n "github.com/JhonHawk/tricell-hive/releases" SECURITY.md` devuelven su línea; `rg -n "#108" CHANGELOG.md` devuelve la entrada.

### T3 — Lectura de reporting sin «obligatoria»

- [x] Las seis frases que llaman obligatoria la lectura de `reporting` dicen que se lee bajo sus condiciones.

**Closes:** AC3.

**Locations:** `llms.txt:22`; `_support/docs/architecture/repository-and-distribution.md:48,85`; `_support/docs/architecture/agent-delivery.md:50`; `_support/docs/architecture/instruction-resources.md:67,88`.

**Changes:** `llms.txt`, `repository-and-distribution.md:48` y `agent-delivery.md:50`: reporting se lee cuando aplican sus condiciones (no para saludos ni respuestas directas cortas), como `content/guidance/global.md:17`. Las líneas 85, 67 y 88, que describen el supuesto del escenario de medición, dicen «reporting» o «reporting, which the scenario counts» en vez de «mandatory reporting».

**Verification:** `rg -n "before every chat reply|mandatory reporting" llms.txt _support/docs/architecture/` no devuelve nada.

### T4 — Datos derivados y textos históricos

- [x] Recuento de skills, tabla de huella, notas de historia y duración de la suite al día.

**Closes:** AC4, AC5, AC6, AC7.

**Locations:** `_support/docs/architecture/repository-and-distribution.md:10`; `_support/docs/architecture/instruction-resources.md:79-86`; `_support/docs/harness-engineering/README.md:1-3`; `_support/docs/harness-engineering/2026-09-20-portable-harness-research.md:1-3`; `CONTRIBUTING.md:26`.

**Changes:**

- `repository-and-distribution.md:10`: «the authored skills under `content/skills/`» sin número; la fecha «Reviewed» no cambia, porque solo se corrige una frase.
- `instruction-resources.md`: correr `go test ./tests/content/ -run TestGuidanceReadingFootprint -v -count=1` sobre el commit base y reemplazar las cuatro filas y la línea «Snapshot measured on…» con la fecha y el commit; ajustar la frase que sigue a la tabla solo si la medición cambia su conclusión (hoy «Simple query» pasa de −21,65 % a +4,68 %).
- Notas de historia: una línea justo después del título H1 de cada índice: «Historical record (2026-09-20): it covers five hosts; Cursor CLI was added later, and the current host matrix is in the [README](<ruta relativa al README>).»
- `CONTRIBUTING.md:26`: quitar la duración; dejar que la suite se corre localmente antes de proponer un cambio.

**Verification:**

- AC4: `rg -n "Eleven|Twelve|Thirteen" _support/docs/architecture/repository-and-distribution.md` no devuelve nada.
- AC5: `rg -n "Snapshot measured on 2026-10-12" _support/docs/architecture/instruction-resources.md` no devuelve nada y la tabla coincide con la salida del test.
- AC6: `rg -n "^Historical record \(2026-09-20\)" _support/docs/harness-engineering/README.md _support/docs/harness-engineering/2026-09-20-portable-harness-research.md` devuelve una línea por archivo.
- AC7: `rg -n "under a minute" CONTRIBUTING.md` no devuelve nada.

## Verificación compartida

| Puerta | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite local | Candidato final | `go vet ./...`, `go test ./... -count=1`, `python3 -m unittest discover -s tests/skills -p '*_test.py'`: pasan | Parte de la implementación |
| Revisión | Tras T2–T4, en el PR | `/code-review` (D4-A) | Elegida en esta sesión |

Sin verificación de UI ni en vivo: no cambia `content/`, el sitio ni el instalador.

## Revisión y progreso

- 2026-10-23: plan escrito sobre `2d89940e`; delta de T1 escrito con el plan.
- 2026-10-23: revisión del plan (`hive-review-plan`, dominio documentación; 89 s, 103 011 tokens): 1 bloqueante (sexta frase en `agent-delivery.md:50`), 1 riesgo alto (AC1 sobre el spec vigente) y 4 sugerencias; todas aplicadas como las propuso el revisor, sin otra ronda. Delta correcto: mismos nombres de requisito y solo cambian tres líneas.
- Plan listo; build autorizado por D9.
- 2026-10-23: T2, T3 y T4 hechas y comprobadas con su `rg` (AC2–AC7); `TestGuidanceReadingFootprint` sobre `2d89940e`: global 33 710, simple query 49 107 (+4,68 %), question 59 365 (+26,55 %), plan/build 229 774 (+32,18 %); suite completa en verde. T3 y T4 comparten dos archivos y van en un solo commit. T1 queda para el cierre.
- 2026-10-23: validado (D12-A); PR #114. `/code-review` (113 s, 104 596 tokens), 10 hallazgos, ninguno P0/P1:
  - Corregidos sin re-revisión: «Closes» → «Refs #108» (la incidencia se cierra tras aplicar T1); frase ambigua de `CONTRIBUTING.md` quitada; nota de `harness-engineering/README.md` acotada a la lista de hosts; las tres descripciones de `reporting` pasan a un puntero a `global.md`; etiqueta «not a chat reply» quitada; nota del documento de investigación sin repetir los cinco hosts; frase que aclara que una respuesta corta lee solo `global`.
  - No aplicado: cambiar el predicado de `TestGuidanceReadingFootprint` (otra decisión; queda explicado en el documento).
  - Hallazgo incidental para el cierre: el spec dice que solo `pi-subagents` descarga, pero `SECURITY.md` documenta también `agent-browser`; viene de antes de este cambio.
- 2026-10-23: PR #114 fusionado (`9538de01`). Cierre: delta aplicado a `_support/openspec/specs/versioned-installation/spec.md`; `rg -c "0.74.0|pinned pi-subagents"` devuelve 0 y el diff toca solo las tres líneas del pin (AC1 cumplido). #108 se cierra tras el PR de cierre.
