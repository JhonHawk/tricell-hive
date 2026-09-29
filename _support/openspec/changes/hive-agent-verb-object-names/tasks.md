# Tareas

Commit base: `82f2dcc016db02d2ec55692a169d680f5acd57cc` (rama `feat/hive-agent-verb-object-names`, worktree `.claude/worktrees/hive-agent-names`).

### T1 — Roles renombrados con sus descripciones corregidas

- [x] Los 20 archivos de `content/agents/` llevan el id nuevo en el nombre de archivo, en `name` y en el encabezado `# <id>`, y las seis descripciones H1–H6 llevan su texto nuevo.

**Closes:** AC1, AC3.

**Depends on:** ninguna.

**Locations:** `content/agents/*/*.md`; tabla y textos en [design.md](design.md#tabla-de-nombres).

**Execution:** hilo principal. Es un cambio mecánico de 20 archivos cuyo encargo sería más largo que el propio cambio.

**Test approach:** check con `go test -count=1 ./integrations/...`, que valida `name` igual al nombre de archivo en cada rol, y las búsquedas de abajo.

**Changes:** `git mv` de cada archivo dentro de su carpeta; sustituir `name` y el encabezado; reemplazar las descripciones H1–H6 sin tocar `model_profile`, `access_profile`, `effort` ni el resto del cuerpo.

**Verification:**

- `rg -N '^name:' content/agents | sort` lista exactamente los 20 ids nuevos de la tabla.
- `git diff --stat -M <base> -- content/agents` muestra 20 renombrados.
- `git diff <base> -- content/agents | rg '^[-+](model_profile|access_profile|effort):'` no devuelve nada.
- Cada descripción coincide completa con su texto de [design.md](design.md#descripciones): `rg -cFx 'description: "<texto>"' <archivo>` devuelve `1` para H1–H4 y H6 con el texto literal de design.md, y para H5 con el texto original tras la sustitución indicada.
- `rg -c "including specifications for proposed changes" content/agents` no devuelve coincidencias.

### T2 — Referencias activas actualizadas

- [x] Skills, documentación viva y pruebas citan los ids nuevos; el registro de investigación del 2026-09-26 marca su decisión como sustituida.

**Closes:** AC2.

**Depends on:** T1.

**Locations:** `content/skills/flow-build/SKILL.md`, `content/skills/flow-build/references/verification.md`, `content/skills/flow-build/references/browser-automation.md`, `content/skills/flow-plan/SKILL.md`, `content/skills/flow-plan/references/delivery-decisions.md`, `content/skills/flow-plan/references/plan-format.md`, `content/skills/flow-plan/references/plan-review.md`, `content/skills/flow-research/SKILL.md`, `_support/docs/architecture/agent-delivery.md`, `_support/docs/harness-engineering/harness-audit-rules.md`, `integrations/agents/agents_test.go`, `integrations/target/catalog_test.go`, `tests/skills/harness_audit_rules_test.py`, `tests/fixtures/flows/cases.json`, `_support/sessions/2026-09-26-hive-agent-names/agent-names.research.md`.

**Execution:** hilo principal, por la misma razón que T1 y porque depende de sus rutas.

**Test approach:** check con la búsqueda de nombres viejos y la suite completa.

**Changes:** sustituir cada id viejo por el nuevo según [design.md](design.md#referencias-en-el-texto); en las pruebas, actualizar rutas y las claves de la tabla de esfuerzo de `agents_test.go` sin cambiar sus valores. En `agent-delivery.md`, conservar las filas de evidencia y añadir la nota que indica [design.md](design.md#referencias-en-el-texto). En el registro de investigación, cambiar el estado y la sección de decisiones para indicar que E2 sustituyó a `hive-<rol>` el 2026-09-29, con enlace a este cambio.

**Verification:**

```sh
OLD='backend-developer|frontend-developer|kotlin-multiplatform-developer|database-specialist|devops-engineer|solution-architect|visual-designer|sdd-spec-writer|sdd-explore|test-engineer|sdd-verify|review-task|performance-engineer|state-fetcher|review-plan|review-code|review-security|review-ux|review-harness|review-refuter'
rg -nP "(?<![\w-])($OLD)(?![\w-])" content integrations tooling tests _support/docs/architecture _support/docs/harness-engineering/harness-audit-rules.md -g '!tooling/legacy/**' -g '!tests/fixtures/regression/**' \
  | rg -v '^_support/docs/architecture/agent-delivery.md:[0-9]+:\| (Claude Code|Codex|Grok Build|Pi|OpenCode V2) [0-9]'
```

No devuelve ninguna línea (en la base devuelve 94: las 99 de la búsqueda menos las cinco filas de evidencia, que se conservan). Además pasan `go vet ./...`, `go test -count=1 ./integrations/... ./tests/... ./tooling/management` y `python3 -m unittest discover -s tests/skills -p '*_test.py'`.

### T3 — Prueba de retiro al renombrar un agente

- [x] Una prueba del gestor instala un agente sintético en todos los hosts, lo renombra en el origen y comprueba que la instalación siguiente borra la ruta vieja y crea la nueva en cada host.

**Depends on:** ninguna.

**Locations:** `tooling/management/catalog_test.go`, junto a `TestAgentCatalogueFreezesRendererProfilesAndModes` y `TestCatalogueInstallRollbackAndSharedRetirement`, reutilizando `setup`, `put`, `plan`, `apply` y `absent`.

**Execution:** delegada a un rol de pruebas (`hive-write-tests` tras T1). Es una prueba nueva con interfaz fija y sin escrituras compartidas con T1–T2. El brief nombra el archivo que puede modificar y pide no tocar código de producción.

**Test approach:** characterization: fija el comportamiento actual de `ownership.go:179-207`, que hasta ahora solo tiene prueba para skills.

**Changes:** `TestAgentRenameRetiresOldTarget` o nombre equivalente: `o.Hosts` con `claude`, `codex`, `grok`, `pi` y `opencode` como en la prueba de retiro de skills, más `cursor` si `setup` lo admite sin la preparación de `cursor_test.go`; si no, una segunda variante para Cursor. Perfiles copiados como en la prueba vecina, instalar `content/agents/design/old-agent.md`, borrar ese origen, escribir `content/agents/design/new-agent.md`, planificar e instalar de nuevo.

**Verification:** `go test -count=1 ./tooling/management -run 'AgentRename'` pasa. Para comprobar que la prueba puede fallar, se ejecuta una vez sin borrar el origen viejo: debe fallar en la aserción de ruta ausente.

### T4 — Vista previa del despliegue real

- [x] La vista previa desde el commit del cambio contiene, en cada host registrado, las 20 rutas de agente viejas como retiro y las 20 nuevas como instalación, sin error.

**Closes:** AC4.

**Depends on:** T1, T2, T3 y el commit local del cambio (`update` solo despliega commits).

**Locations:** ninguna en el repositorio; lee la instalación real del usuario.

**Execution:** hilo principal: es una lectura de la instalación real.

**Test approach:** check con el plan guardado, que no escribe en destinos.

**Changes:** ninguno.

**Verification:** `go run ./tooling/cli update --out "$SCRATCH/plan.json"`, con `SCRATCH` en el directorio temporal de la sesión, sin reemplazar el binario global. El comando termina sin error. En el plan guardado, agrupando las rutas de agente por carpeta (el campo `Host` llega vacío en los registros del plan): `jq -r '[.changes[] | select(.target.Kind=="agent" and .after==null) | .target.Path] | group_by(split("/")[:-1]|join("/"))[] | "\(.[0] | split("/")[:-1] | join("/")) \(length)"'` da 20 por cada carpeta de agentes de los hosts registrados y sus nombres son los 20 ids viejos; el mismo filtro con `.after!=null and .before==null` da los 20 ids nuevos por carpeta. En `update`, `--out` guarda el plan en lugar de aplicarlo (`tooling/cli/update.go:38`). Una ruta vieja solo puede entrar al plan como retiro, porque su origen ya no existe.

### T5 — Suite completa sin CI solo para cambios no mecánicos

- [x] La regla de cierre de regresión de `flow-build` limita la suite completa sin CI a cambios no mecánicos y remite los mecánicos a las comprobaciones que ya los cubren.

**Closes:** AC5.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-build/SKILL.md:65`.

**Execution:** hilo principal: una frase.

**Test approach:** check con la búsqueda y las pruebas de contenido.

**Changes:** sustituir "Close regression coverage through the project's actual CI gate; when none will run it, run the full suite once on the final candidate, and rerun it only when a later edit could affect what it covers." por "Close regression coverage through the project's actual CI gate. When none will run it, run the full suite once on the final candidate of any change that is not mechanical, as the global guidance defines it, and rerun it only when a later edit could affect what it covers; a mechanical change closes with the checks that already cover it, even without CI."

**Verification:** `rg -c 'any change that is not mechanical, as the global guidance defines it' content/skills/flow-build/SKILL.md` devuelve `1`; `rg -c 'when none will run it, run the full suite once' content/skills/flow-build/SKILL.md` y `rg -c 'alters behavior, run the full suite' content/skills/flow-build/SKILL.md` no encuentran nada; `go test -count=1 ./tests/...` y el unittest de skills pasan.

## Verificación compartida

| Comprobación | Momento | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Pruebas y vet | Al cerrar T2 | Los comandos de T2 pasan; la suite completa con `-race` se detuvo por D4-A | Incluida en la implementación |
| Prueba de retiro | Al cerrar T3 | `go test ./tooling/management -run 'AgentRename'` | Incluida en la implementación |
| Vista previa de instalación | Tras el commit local | Plan guardado con `go run ./tooling/cli update --out` (T4) | Incluida: no escribe en destinos ni en el binario global |
| Refresco de la instalación local | Tras integrar en la base | Recompilar el binario y `hive update`; en ningún directorio de agentes de los hosts queda un archivo con uno de los 20 ids viejos, y están los 20 nuevos | Regla temporal de AGENTS.md; `hive update` pide confirmación |

No aplica verificación de interfaz ni recorridos en ejecución real: el cambio no tiene efecto visual y la vista previa del gestor cubre el despliegue. Las sesiones abiertas conservan los ids viejos hasta reiniciarse.

## Revisión y avance

Revisión del plan del 2026-09-29 sobre la revisión `1981ce3182b0`, con dos hijos `review-plan` de solo lectura (despacho nativo de Claude Code, en paralelo):

- **Contenido distribuido y pruebas:** sin bloqueantes. Aceptados: conservar las filas de evidencia de `agent-delivery.md` con una nota; reescribir H4 ("not when the task is only to collect…") para que no se lea como "not only for"; comparar cada descripción completa en T1 y comprobar que desaparece la frase vieja de H5; corregir el recuento a 14 archivos. El cambio activo `gh-46-read-only-views` queda excluido por tener otro escritor (ver proposal.md).
- **Despliegue:** un bloqueante aceptado (el resumen de `--dry-run` no distingue retiro de instalación: T4 ahora lee el plan guardado con `--out`), la vista previa pasa a `go run` para no reemplazar el binario global, la nueva T3 añade la prueba de retiro de agentes que faltaba y la comprobación posterior al merge busca que no quede ningún id viejo.

Todas las correcciones aplican lo que propusieron los revisores, así que no hubo segunda ronda. Límites: ninguna revisión ejecutó pruebas ni la vista previa; AC4 depende del estado instalado real, que no se inspeccionó.

Avance del 2026-09-29:

- T1: verificada por `review-task` (AC1 y AC3 cumplidos). Señaló que el índice solo tenía los renombrados; el commit nombra cada ruta. Formato de `agents_test.go` corregido con `gofmt`.
- T2: verificada por `review-task` (AC2 cumplido; búsqueda en 0, en la base 94). Pasan `go vet ./...`, `./integrations/...`, `./tests/...`, `./tooling/management` y 23 pruebas de skills. La suite completa con `-race` se detuvo a los 6 minutos, cuando solo quedaba `tooling/cli` (D4-A).
- T3: la prueba `TestAgentRenameRetiresOldTarget` pasa en los seis hosts; falla si no se borra el origen viejo.
- T4: plan guardado desde `500bd84`: 7 destinos, 306 archivos; en las seis carpetas de agentes (`~/.claude`, `~/.codex`, `~/.config/opencode`, `~/.cursor`, `~/.grok`, `~/.pi/agent`) retira los 20 ids viejos e instala los 20 nuevos, sin error. El filtro original por `.target.Host` no servía (campo vacío); se corrigió para agrupar por carpeta. `review-task` dio AC4 por cumplido: `source_commit` es `500bd84`, cada carpeta corresponde a un solo host, los ids coinciden con la tabla y los 66 cambios que no son de agentes solo modifican (8 con contenido nuevo: las skills que tocan T2 y T5).
- T5: añadida por D5-B después de la revisión del plan. `review-task` dio AC5 por cumplido y señaló que "alters behavior" y "mecánico" no eran complementarios (una actualización de dependencia quedaba sin suite); se aplicó su propuesta: la condición es "no mecánico según la guía global". Verificado con las búsquedas de T5; la cubre también `/code-review` antes del merge.
- Aparte: `gofmt` también señala `tests/pilot/regression.go`, que ya estaba así en la base; no se toca.

