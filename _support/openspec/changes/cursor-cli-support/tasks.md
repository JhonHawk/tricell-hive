# Tareas — soporte de Cursor CLI

## T1 — Adaptador `cursor` y roles en el gestor

- [x] `plan install --hosts cursor --scope user` propone el bloque en `~/.cursor/AGENTS.md`, los skills compartidos y los roles en `~/.cursor/agents/`; el scope proyecto se rechaza.

**Depende de:** nada.

**Ubicaciones:** `integrations/cursor/cursor.go` y `cursor_test.go` (nuevos); `integrations/agents/agents.go` (lista de hosts `:126-129`, validación de claves de acceso `:142-182`, renderizado `:222-280`) y `agents_test.go`; `integrations/agent-profiles.json` (host `cursor`); `integrations/target/target.go` (`Config.CursorHome`, `ExpandHostHomes`) e `integrations/target/target_test.go` (valores por defecto, aislamiento sintético y `TestNewHomesOmitFromLegacyConfigJSON`); `tooling/management/plan.go:23` (`resolve`) y `:517` (directorios); `tooling/management/types.go:163` (`validateHosts`); `tooling/cli/main.go:28` y `:55` (ayuda y descripción de `--hosts`).

**Ejecución:** delegada a `backend-developer`: la interfaz está fijada en [design.md](design.md) y no se solapa con los archivos de T2–T3. El hilo principal revisa el diff.

**Cambios:** seguir el patrón de `integrations/pi/pi.go` con las decisiones de [design.md](design.md): bloque, skills compartidos y `target.ExpandAgents` hacia `<cursor-home>/agents` con extensión `.md`. Perfil de Cursor con `model: inherit` y `readonly: true` para `observe`; `ReadProfiles` con `cursor` opcional, `cursor` sin `effort` y `readonly` solo como booleano `true` ([design.md](design.md)). Antes de fijar la ruta, comprobar en la documentación de Cursor (Context7 `/websites/cursor`) si existe una variable de entorno para el directorio de configuración. `CursorHome` se declara `json:"cursor_home,omitempty"` para que el `planID` de journals pendientes anteriores no cambie (`apply.go:315`). Pruebas nuevas o extendidas para Cursor:

- instalación, actualización idempotente, retirada con otro consumidor de skills todavía registrado, conflicto de marcadores y recuperación;
- actualización de un skill compartido con `--hosts cursor` solo: falla con `shared resource update requires all consumers in --hosts` (patrón de `shared_test.go:82-113`);
- `~/.cursor/` previo con archivos ajenos: se conservan el directorio y sus hermanos y solo se retira `AGENTS.md`; `~/.cursor/` ausente: se crea y se limpia; `AGENTS.md` previo con texto del usuario: se preserva;
- catálogo con `content/agents/` (como `catalog_test.go:14`): Cursor produce un `[agent]` por rol en `~/.cursor/agents/` y los bytes renderizados coinciden con el formato de [design.md](design.md);
- `~/.cursor/agents/` previo con un archivo ajeno del mismo nombre que un rol: conflicto, sin sobrescribir;
- perfil de Cursor con una clave de acceso distinta de `readonly`, con `readonly: "true"` como cadena o con `effort`: rechazado; perfil con un host desconocido: rechazado;
- `TestCatalogueRendersAllRolesForEveryHost` (`agents_test.go:37`) ampliada con `cursor`;
- compatibilidad: una versión congelada con perfiles de cinco hosts y agentes sigue validándose e instalándose para esos hosts, y Cursor sobre esa versión falla con `unsupported agent host "cursor"`;
- `status --hosts cursor`: `installed`, `retained_shared` y `not_installed`.

**Verificación** (desde la raíz del checkout; `$TMP` es un directorio de `_support/workspace/2026-09-25-cursor-cli-support/`):

- `go test ./integrations/... ./tooling/...` pasa, con las pruebas de arriba (A2).
- `go run ./tooling/cli plan install --hosts cursor --scope user --home "$TMP/home" --state-dir "$TMP/state" | rg '^install .* \[(block|skill|agent|symlink)\]$'` muestra un solo `[block]` en `$TMP/home/.cursor/AGENTS.md`, entradas `[skill]` bajo `$TMP/home/.agents/skills/` y un `[agent]` por rol bajo `$TMP/home/.cursor/agents/` (A1). Sobre la salida sin filtrar, `rg -n 'readonly: true'` aparece en los roles `observe`.
- El mismo comando con `--scope project --root "$TMP/proj"` termina con el error de scope no soportado del adaptador, no con `project scope requires --root`.

## T2 — Contenido distribuido

- [x] `global.md` tiene la fila de Cursor y `harness-audit` tiene el puntero canónico con su comprobación.

**Depende de:** el hallazgo de T1 sobre una variable de entorno para `~/.cursor` (rutas del puntero y de la fila de roles).

**Ubicaciones:** `content/guidance/global.md:85` (tabla "Native role selection hints"); `tests/content/budget_test.go:12`; `content/skills/harness-audit/references/instruction-files.md` (HA-IF-18 junto a HA-IF-17), `references/hierarchy.md` (fila de carga de Cursor y observación de puntos de entrada), `references/native-tools.md:9` y `SKILL.md:16` (detección de hosts); `_support/docs/harness-engineering/harness-audit-rules.md` (alta de HA-IF-18 como `active`).

**Ejecución:** hilo principal: el texto depende de las decisiones de esta conversación y el brief sería más largo que el cambio.

**Cambios:**

- Confirmar de forma estática el nombre de la herramienta de delegación en el bundle de Cursor ([design.md](design.md)); si no se confirma, la fila nombra la capacidad.
- Agregar la fila de Cursor de [design.md](design.md).
- Agregar HA-IF-18 con el texto canónico del puntero, su ubicación, severidades, `keep` para un puntero existente y la regla de `Hive guidance: required`; la carga de Cursor en `hierarchy.md` como párrafo al estilo del de Grok (`:37`), separando lo documentado (proyecto: `AGENTS.md`, `CLAUDE.md`, `.cursor/rules`) de lo observado el 2026-09-25 (una corrida por caso), sin afirmar nada sobre `AGENTS.md` de niveles superiores o anidados; la detección de Cursor en `SKILL.md`; en `native-tools.md` (HA-ME-02), `cursor-agent --version` y que Cursor no tiene una vista conocida de las instrucciones cargadas, así que sus afirmaciones de carga son observaciones o hipótesis.
- `global.md` pesa hoy exactamente el presupuesto (34958 bytes): subir `globalGuidanceBudget` al tamaño nuevo e informar el antes y el después.

**Verificación:**

- `go test ./tests/... ./tooling/management/...` pasa, incluidos el presupuesto y la resolución de referencias (A4).
- `python3 tests/skills/harness_audit_rules_test.py` pasa con HA-IF-18 activo en el catálogo.
- Lectura del diff (A5): el texto del puntero coincide con [design.md](design.md), y HA-IF-18 tiene ID, disparador, severidades y la línea propuesta.

## T3 — Documentación

- [x] La documentación nombra a Cursor con sus destinos, límites y versión observada.

**Depende de:** T1 (rutas finales y variable de entorno).

**Ubicaciones:** `_support/docs/architecture/deployment-manager.md` (tabla de destinos, versiones observadas, párrafo de hosts de scope usuario); `_support/docs/architecture/repository-and-distribution.md` (fila de `integrations/`); `_support/docs/architecture/agent-delivery.md` (`:11` columna de perfil, `:41-45` tabla de destinos, `:51` versiones inspeccionadas, `:67` "five-host selection hints", `:69-75` tabla de evidencia; formato y perfil de Cursor, precedencia de `.cursor/` sobre `.claude/` y `.codex/`, bug del CLI 2026.09.18, que al retirar solo Cursor pasa a los roles de Claude y que el contrato leído por un hijo genérico no aplica `readonly`); `AGENTS.md` (lista de hosts objetivo del primer párrafo); `README.md:3` y `:21`.

**Ejecución:** hilo principal, junto con T2.

**Cambios:** registrar los destinos de Cursor, la versión 2026.09.18-9a7762b observada, que el archivo no se carga de forma nativa y depende del puntero (enlazando a HA-IF-18 en vez de copiar el texto), que el CLI 2026.09.18 no lista los roles de usuario aunque estén desplegados, y que los planes guardados antes del cambio deben regenerarse. Separar lo documentado por Cursor de lo observado.

**Verificación:** `rg -n -i cursor _support/docs/architecture AGENTS.md README.md` muestra cada mención esperada; ninguna afirma que Cursor carga `~/.cursor/AGENTS.md` por sí mismo.

## T4 — Verificación in vivo (requiere autorización aparte)

- [ ] Con la guía desplegada, el puntero funciona en Cursor con dos modelos y no duplica la guía en Claude Code ni Codex.

**Depende de:** T1–T3 y tres condiciones:

1. El diff aceptado y entregado. El estado del gestor valida el host de cada consumidor (`ownership.go:72`): si se despliega Cursor desde un diff que luego se descarta, el gestor del commit falla con `unsupported host "cursor"` en todo `plan`, `status` y `apply`. Si aun así se despliega antes, la reversión es `plan remove --hosts cursor --scope user` con el código nuevo antes de descartar el diff.
2. Autorización para desplegar en el home real por una de dos rutas, que se te preguntará con la vista previa delante:
   - **Solo Cursor:** materializar la versión instalada (`80ff36cc114d`, con sus modos) en un árbol desechable de `_support/workspace/2026-09-25-cursor-cli-support/`, sustituir ahí solo `integrations/agent-profiles.json` por el de seis hosts y correr `plan install --hosts cursor --scope user --source <ese árbol>` con el gestor nuevo. Escribe `~/.cursor/AGENTS.md` (con el `global.md` instalado, sin la fila de Cursor; A6 y A8 no dependen de ella) y `~/.cursor/agents/*.md`, y registra a Cursor como consumidor de los skills sin cambiar sus bytes (`plan.go:467-471`). No toca los archivos de los otros cinco hosts. Es un procedimiento manual, y la instalación completa sigue siendo necesaria al entregar el diff.
   - **Seis hosts:** `plan install --hosts claude,codex,cursor,grok,opencode,pi --scope user` con la versión nueva. Reescribe los bloques de `~/.claude/CLAUDE.md`, `~/.codex/AGENTS.md`, Pi y OpenCode, los skills compartidos y los roles de todos los hosts con el contenido de T2. `--release <hash anterior>` no sirve para Cursor: su perfil congelado no tiene el host y `Render` falla.
3. Autorización para reanudar pilotos de CLI en estas corridas.

**Ubicaciones:** fixtures desechables en el scratchpad de la sesión, fuera de cualquier repositorio para que ningún `AGENTS.md` ancestro contamine la prueba.

**Ejecución:** hilo principal para desplegar y preparar los fixtures; `sdd-verify` para correr y juzgar los casos de forma independiente del implementador.

**Cambios:** ninguno en código. Casos:

1. Cursor, puntero y control, con Grok 4.6 High Fast y un segundo modelo de otra familia (`cursor-agent --model`, elegido de `--list-models`): pregunta "¿dónde pondrías un script temporal?". Una corrida por caso y modelo.
2. Claude Code y Codex en un fixture con el mismo puntero, más un canario en el mismo archivo que pruebe que el puntero llegó al modelo. En Claude Code el fixture necesita un `CLAUDE.md` con `@AGENTS.md`, o el puntero no carga y el caso pasa en vacío. Comprobar en el stream de herramientas que no leen `~/.cursor/AGENTS.md`.

Protocolo (sección Measurement de `AGENTS.md`):

- `--mode ask` en Cursor (su `mcp.json` configura Neon con herramientas destructivas) y modos de solo lectura equivalentes en Claude Code y Codex.
- Memoria cotidiana declarada. Identificar por sesión los registros que las corridas escriban en Engram y borrarlos al final; la corrida del 2026-09-25 solo llamó a `mem_current_project`.
- Registrar host y versión, modelo, instrucciones leídas, escrituras, intervención humana y estado final.
- A6 y A7 son una comprobación de humo, no de fiabilidad.

3. Cursor con los roles desplegados: pedir la lista de tipos de subagente y comprobar que ningún rol de Hive aparece dos veces (A8). Si no lista roles de usuario, pedir que delegue en `review-refuter` y comprobar que lee `~/.cursor/agents/review-refuter.md` y se lo pasa a un hijo genérico.

Registrar además si Cursor lista cada skill dos veces, porque descubre tanto `~/.agents/skills` como los alias de `~/.claude/skills`. El runner de `tests/pilot/` no conoce Cursor (`launch.go:24`, `protected.go:18-42`, `main.go:244`): las corridas se hacen a mano con `cursor-agent`, `claude -p` y `codex exec`.

**Verificación:** A6: en Cursor, con los dos modelos, el caso con puntero lee `~/.cursor/AGENTS.md` y responde `_support/workspace/<fecha>-<work>/`, y el control no. A7: el canario aparece y no hay ninguna lectura de `~/.cursor/AGENTS.md` en Claude Code ni Codex. A8: ningún rol duplicado y, si aplica, el contrato leído de `~/.cursor/agents/`.

## Verificación compartida

| Compuerta | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Pruebas | Al cerrar T1–T3, working tree | `go test ./...`, `go test -race ./...`, `go vet ./...` y `python3 tests/skills/harness_audit_rules_test.py` sin fallos (A2, A3, A5) | Incluida en la implementación |
| Revisión de código | Después de las pruebas, antes de dar el working tree por verificado | `/code-review` de Claude Code sobre el diff | Elegida en D6-A; incluida en la implementación |
| In vivo en Cursor, Claude Code y Codex | Después de la revisión | T4 | Pendiente: despliegue global y pilotos |

## Estado de la revisión y avance

- Revisión del plan, ronda 1 sobre `6a4742902a55` (dos `review-plan` de solo lectura, despacho nativo en Claude Code):
  - Gestor Go: dos bloqueantes aceptados y corregidos (actualizar skills compartidos exige todos los consumidores, `plan.go:467`; el estado valida hosts de consumidores, `ownership.go:72`) y cinco sugerencias incorporadas.
  - Contenido: tres bloqueantes aceptados y corregidos (prefijo del puntero en conflicto con `Hive guidance: required`; fila de roles sin fuente de contrato, resuelto con D7-C; regla de `harness-audit` sin forma de hallazgo, ahora HA-IF-18) y ocho sugerencias incorporadas. Descartado: nada.
- D7-C (2026-09-25) amplió T1 con los roles de Cursor.
- Ronda 2 sobre `55a075b54f85`:
  - Gestor Go: un bloqueante aceptado y corregido (exigir seis hosts en `ReadProfiles` impedía reinstalar cualquier versión anterior; ahora `cursor` es opcional). La ruta de despliegue solo-Cursor sin `--release` se incorporó a T4 como opción. Dos sugerencias incorporadas (formato de `json.Marshal`, pruebas de perfil).
  - Contenido: un bloqueante aceptado y corregido (la condición "the only mention" fallaba en repos que mencionan la guía; ahora depende del encabezado como línea propia). Cinco sugerencias incorporadas: fila más corta alineada con `global.md:80`, "sin duplicados" como inferencia, justificación de `high`, cambios concretos en `native-tools.md` y `hierarchy.md`, líneas de `agent-delivery.md`, y la evidencia in vivo acotada a lo que se probó.
- Ronda 3 sobre `9aeb7f6d8238`, acotada a las correcciones de la ronda 2: sin bloqueantes en ninguno de los dos dominios. Incorporados sin nueva ronda: rechazo de hosts desconocidos en `ReadProfiles` y una errata de `design.md`.
- Límites de la revisión: los revisores no ejecutaron código, pruebas ni CLIs de modelos; el renderizado de Cursor, la precedencia de roles y el puntero actual quedan para T1 y T4.
- Avance (2026-09-25): T1–T3 implementados y verificados en el working tree.
  - T1 delegado a `backend-developer` (despacho nativo en Claude Code). Hallazgo: Cursor no documenta una variable para mover `~/.cursor` (`CURSOR_CONFIG_DIR` solo reubica `cli-config.json`, cursor.com/docs/cli/reference/configuration), así que la ruta es fija. TDD: el hijo no escribió todas las pruebas antes del código; comprobó el RED quitando temporalmente el host `cursor` de los perfiles (fallos por `unsupported agent host "cursor"`).
  - Revisión del hilo principal: agregada `TestCursorRecoveryAtWriteAndStateBoundaries` (A2 pedía recuperación de Cursor y faltaba) y corregido `gofmt`.
  - T2: presupuesto de `global.md` de 34958 a 35183 bytes (RED observado); HA-IF-18 en el catálogo (RED observado en la prueba Python).
  - Evidencia: `go vet ./...` limpio; `go test -race ./...` con 11 paquetes `ok`; `python3 tests/skills/harness_audit_rules_test.py` OK. Vista previa en home sintético: 1 `[block]` en `.cursor/AGENTS.md`, 49 `[skill]`, 20 `[agent]`, 8 `readonly: true`; scope proyecto: `hive: Cursor project scope is unsupported`.
  - `/code-review` (D6-A, nivel medium): un hallazgo `low` aceptado y corregido. Un plan guardado antes del cambio fallaba con `invalid root`; ahora pide regenerarlo (`plan.go:510`, prueba `TestPlanSavedBeforeCursorSupportAsksToRegenerate`, RED observado). También se corrigió el texto "all five hosts" de `ReadProfiles`.
- Diff presentado el 2026-09-25; el usuario eligió seguir en `hold` (D8-C). Nada versionado.
- D10-A (2026-09-25), después del piloto del #30: R2 aplicado dentro de este cambio. `global.md` queda idéntico a la variante B probada (sin tabla, con la frase nueva); presupuesto de 35183 a 34271 bytes; la tabla con la fila de Cursor pasa a `agent-delivery.md`; `AGENTS.md` ya no usa la tabla como ejemplo de nombre por host. T2 queda así reemplazado en lo que toca a la fila de Cursor.
- Entrega (D11-A, 2026-09-25): `e678cd9` (Cursor) y `2fd052c` (R2), con push a `rebuild/harness-engineering` junto con el `756bbf1` pendiente, por decisión del usuario. El árbol de `e678cd9` pasó `go vet`, `go test ./...` y la prueba Python en un worktree aislado. #30 cerrado.
- Siguiente paso: autorizar T4. La condición 1 (diff entregado) ya se cumple.
