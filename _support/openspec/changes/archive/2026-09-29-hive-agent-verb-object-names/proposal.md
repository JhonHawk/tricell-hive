# Renombrar los agentes de Hive a `hive-<verbo>-<objeto>`

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado en `rebuild/harness-engineering` con el PR #51 (`fbab417`) y desplegado en los seis hosts el 2026-09-29 |
| Tracker · GitHub Issues | Sin issue: el cambio nace de la conversación del 2026-09-29 |
| Git | Automático · PR #51 integrado (`fbab417`) |
| Verificación | `go test` y `go vet` · unittest de skills · búsqueda de nombres viejos · vista previa del plan de instalación |
| Siguiente paso | Ninguno. Pendiente fuera de este cambio: la sesión de `gh-46-read-only-views` debe usar los ids nuevos |

## Objetivo

Los 20 agentes de Hive usan nombres de puesto genéricos (`backend-developer`, `review-code`) que chocan con agentes que la gente ya tiene instalados en los catálogos compartidos de los seis hosts. Además mezclan formas (puestos, verbos, sustantivos de agente), arrastran el prefijo `sdd-` de la versión anterior y no distinguen `review-task` de `sdd-verify`.

El cambio renombra cada agente al esquema E2, `hive-<verbo>-<objeto>`, que el usuario eligió el 2026-09-29 y que sustituye la decisión del 2026-09-26 (`hive-<rol>`). En el mismo cambio corrige seis descripciones (D1-A: H1–H6) cuyo texto oculta una responsabilidad o deja ambiguo el límite con otro agente. Al desplegar, el gestor retira los archivos con los nombres viejos en cada host y deja solo los nuevos.

## Alcance y aceptación

Incluye:

- Renombrar los 20 archivos de `content/agents/`, su campo `name` y su encabezado, según la tabla de [design.md](design.md#tabla-de-nombres).
- Aplicar los textos H1–H6 de [design.md](design.md#descripciones).
- Actualizar cada referencia activa: skills distribuidas, la documentación viva de arquitectura, las pruebas Go y Python, y el caso de delegación de `tests/fixtures/flows/cases.json`.
- Acotar la regla de `flow-build` que pide la suite completa cuando no hay CI: solo para cambios que no son mecánicos; un cambio mecánico cierra con las comprobaciones que ya lo cubren (D5-B, añadido el 2026-09-29 a petición del usuario, que la vio disparar la suite completa en este mismo cambio).
- Marcar como sustituida la decisión del registro [agent-names.research.md](../../../sessions/2026-09-26-hive-agent-names/agent-names.research.md).

Excluye, con su razón:

- Historia: `_support/openspec/changes/archive/`, `_support/sessions/` (salvo la nota de sustitución), `_support/evidence/` y la investigación fechada `_support/docs/harness-engineering/2026-09-20-workflow-map.md` describen lo que pasó con los nombres de entonces.
- `tooling/legacy/catalog.json`: inventario del Hive anterior que la migración usa para reconocer sus archivos; sus nombres son los del legado.
- `tests/fixtures/regression/`: AGENTS.md pide dejar esos casos sin cambios.
- Filas de evidencia observada en `_support/docs/architecture/agent-delivery.md` (tabla "Evidence through 2026-09-22"): registran los ids que se ejecutaron entonces; reciben una nota, no el id nuevo.
- El cambio activo `_support/openspec/changes/gh-46-read-only-views/`: otra sesión lo está escribiendo (sin versionar, modificado el 2026-09-29 a las 00:58) y cita `backend-developer`, `review-ux`, `sdd-verify` y `sdd-explore`. Editarlo desde aquí pisaría a otro escritor; queda pendiente actualizar sus ids en su propia sesión.
- Carpetas de `content/agents/`: se conservan; el nombre del archivo es lo que determina el id instalado.
- Arquetipos o nombres visibles para la TUI de #46: quedaron como idea aparte.

Restricciones que deben seguir siendo ciertas:

- `model_profile`, `access_profile`, `effort` y el cuerpo de cada rol no cambian, salvo el encabezado `# <name>`.
- `go vet ./...`, la suite completa `go test -count=1 -timeout 25m ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasan. La primera corrida con `-race` se detuvo (D4-A) con una premisa falsa: `tooling/distribution/bootstrap_shell_test.go:630` sí copia el `content/` real. Este cambio altera un contrato (los ids instalados), así que no es mecánico y su propia regla nueva exige la suite completa.

Criterios:

- AC1. Cada archivo de `content/agents/` declara en `name` el id nuevo de la tabla, igual a su nombre de archivo, y no queda ninguno con un id viejo. *Falso en la base cuando* `rg -N '^name:' content/agents` lista `"backend-developer"` y los otros 19 ids viejos.
- AC2. Ninguna referencia activa cita un id viejo fuera de las exclusiones. *Falso en la base cuando* la búsqueda de T2 devuelve 99 líneas: 40 en los propios roles y 59 en 14 archivos de skills, documentación y pruebas; cinco de ellas son filas de evidencia histórica que se conservan.
- AC3. Las descripciones de `hive-verify-change`, `hive-build-infra`, `hive-design-ui`, `hive-research`, `hive-design-architecture` y `hive-build-kmp` llevan los textos H1–H6. *Falso en la base cuando* ninguna de las seis frases distintivas de T1 aparece en `content/agents`.
- AC4. La vista previa de instalación desde el commit del cambio retira, en cada host registrado, los 20 archivos de agente con id viejo e instala los 20 nuevos, sin conflictos. *Falso en la base cuando* la misma vista previa sobre `82f2dcc` no propone ningún cambio de agentes.
- AC5. La regla de cierre de regresión de `content/skills/flow-build/SKILL.md` exige la suite completa sin CI para todo cambio que no sea mecánico según la guía global, y para un renombrado o traslado mecánico exige buscar cada referencia al nombre o ruta viejos, incluidas cadenas y rutas, y correr las pruebas de los paquetes que las contienen. *Falso en la base cuando* `rg -c 'any change that is not mechanical' content/skills/flow-build/SKILL.md` no encuentra nada.

## Entrega

Decidido el 2026-09-29 (D2-A y D3-A):

- **Modo:** automático. Rama `feat/hive-agent-verb-object-names` desde `rebuild/harness-engineering`, commits locales, push, PR a `rebuild/harness-engineering` y merge cuando la revisión quede resuelta. El repositorio no tiene CI; las comprobaciones son las locales de [tasks.md](tasks.md).
- **Revisión de código:** `/code-review` de Claude Code sobre el diff de la rama, antes del merge.
- **Registro del cambio:** esta carpeta viaja en la misma rama; se versiona con el primer commit, al abrir el PR y en el merge, y se archiva con el cierre.
- **Después del merge:** la regla temporal de AGENTS.md pide refrescar la instalación local: recompilar el binario (el cambio toca `integrations/`) y ejecutar `hive update`, que pide confirmación. Se informa el commit desplegado.
- **Fuera de alcance:** promoción a `master`.
