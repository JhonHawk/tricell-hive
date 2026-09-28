# Tareas

Commit base: se registra al empezar `flow-build`.

## T1 — Reescribir la sección de comunicación

- [ ] `## Communication` reescrita como receta positiva con un ejemplo, en el working tree y sin commit, con el inventario de reglas aceptado por el usuario.

**Closes:** AC1, AC2, AC3.

**Depends on:** ninguna.

**Locations:** `content/guidance/global.md` líneas 14–32 (y las referencias de las líneas 85 y 154); `tests/content/budget_test.go`; el inventario en `_support/workspace/2026-09-27-communication-recipe/inventory.md`.

**Execution:** hilo principal, porque es texto acoplado a las decisiones D11 (validar con piloto) y D12 (toda la sección) y al research del 2026-09-27; un brief para un subagente sería tan largo como el propio cambio.

**Test approach:** check: el tamaño de la sección, el marcador del ejemplo, la regla de etiquetas y el inventario completo.

**Changes:**
- Hacer un inventario de cada frase de las líneas 16–32. Cada fila lleva la frase actual y su línea en la versión nueva, o un descarte con motivo.
- Escribir la sección siguiendo los cinco bloques de [design.md](design.md#enfoque).
- Mostrar al usuario solo los descartes y esperar su aceptación antes de T3.
- Bajar `globalGuidanceBudget` al nuevo tamaño.

**Verification:**
- `awk '/^## Communication/{f=1} /^## Scope and authorization/{f=0} f' content/guidance/global.md | wc -c` da ≤ 8400 (AC1).
- `rg -c '^Example completion report:' content/guidance/global.md` da 1, y el bloque cercado que le sigue es el único de la sección (AC2).
- La regla de etiquetas en negrita dice «three or more areas» (AC3).
- El inventario no tiene filas sin línea nueva ni descarte, y el usuario aceptó los descartes.
- `rg -n 'in-progress item|cleanup|close question|Pending|Next step|Recommendations|Reminders' content/guidance/global.md` sigue encontrando los nombres de los que dependen las líneas 85 y 154 y las skills.
- `go test ./tests/...` pasa.

## T2 — Preparar la comparación

- [ ] Existen los dos brazos verificados, el caso `backlog-status` y el extractor ciego.

**Depends on:** T1.

**Locations:** `tests/fixtures/flows/cases.json` y la carpeta de la fixture nueva; `_support/workspace/2026-09-27-communication-recipe/arms/{a,b}` (ignorado por Git); el extractor en el scratchpad de la sesión.

**Execution:** delegado a `test-engineer` para el caso y su prueba, porque la interfaz está fijada (suite `flows`, estructura de `cases.json`) y no se cruza con las escrituras de T1. Los brazos y el extractor quedan en el hilo principal: son pocos comandos y dependen de T1.

**Test approach:** check: `go test -race ./tests/pilot/...` con una prueba que cargue la fixture nueva y evalúe `assessFlows` sobre una traza sintética; el runner no tiene corrida en seco.

**Changes:**
- **Brazos:** A es `git archive <base> | tar -x -C arms/a`. B es el mismo archivo más `content/guidance/global.md` y `tests/content/budget_test.go` copiados del working tree.
- **Caso `backlog-status`:** un `AGENTS.md` con `## Hive` que declara un tracker local sin red; un `BACKLOG.md` con 6–8 tickets abiertos en 2–3 módulos con IDs tipo `ABC-123`; la pregunta «¿cuál es el estado del proyecto y qué hay en el backlog?»; `expected.skill_read: "flow-research"`. Sin cambios en `flows.go`, salvo que la prueba muestre que hacen falta.
- **Extractor:** script desechable que, por cada `events.json`, toma el último texto del asistente y la pregunta (nativa o en texto), reemplaza las rutas de la fixture y de la corrida por `<proyecto>` y escribe cada respuesta en un archivo sin marca de brazo.

**Verification:**
- `diff -r arms/a arms/b` muestra solo los dos archivos.
- `go test -race ./tests/pilot/...` pasa, incluida la prueba de la fixture nueva.
- El extractor, probado sobre un `events.json` de una corrida anterior que exista en `_support/workspace`, no deja ninguna ruta con `-A-` o `-B-`.

## T3 — Correr el piloto

- [ ] 12 corridas terminadas y evaluadas: `backlog-status`, `direct-build` y `adaptive-plan` × Grok y Codex × brazos A y B.

**Closes:** AC5.

**Depends on:** T2 y los descartes de T1 aceptados.

**Locations:** `_support/workspace/2026-09-27-communication-recipe/runs/r<NN>/`, con nombres numerados que no revelan el brazo.

**Execution:** hilo principal, porque orquesta procesos externos y registra la evidencia. Las corridas independientes se lanzan juntas, de a 4 como máximo.

**Test approach:** check: `--assess` de cada corrida y comparación por celda.

**Changes:**
- **Paso previo:** leer el modelo configurado en `~/.grok/config.toml` y `~/.codex/config.toml`, y usarlo en `--configured-model`.
- **Comando:** `go run ./tests/pilot --suite flows --case <case> --host <grok|codex> --arm <A|B> --guidance-source <arms/a|arms/b> --delivery deployed-global --model <m> --configured-model <m> --effort <high|medium> --timeout 600s --out <run-dir>`, y en Codex además `--allow-native-trust` (D13-A).
- **Registro:** `run.json` guarda host, versión, modelo, hashes de la guía y aislamiento de memoria. Anotar «sin intervención humana».
- **Hashes:** antes de seguir, comparar el `GuidanceBlockHash` entre brazos del mismo host: igual dentro de un brazo, distinto entre A y B.

**Verification:** las 12 corridas con `Terminal: completed`; una corrida que no termina se repite una vez. `MemoryIsolation.Cleanup` verificado en cada `run.json`. En cada celda (caso × host), ninguno de los criterios que nombra AC5 está en `pass` en A y en `fail` en B.

## T4 — Comparación a ciegas

- [ ] El usuario eligió en 6 pares y el resultado está registrado.

**Closes:** AC4.

**Depends on:** T3.

**Locations:** `_support/workspace/2026-09-27-communication-recipe/blind/`, con la clave en un archivo aparte.

**Execution:** hilo principal y usuario; la elección es del usuario.

**Test approach:** check: el recuento de elecciones.

**Changes:** armar 6 pares (caso × host) con el orden mezclado al azar, sin rastros del brazo, y presentarlos en un documento local. Opciones por par: primero, segundo o sin preferencia. La clave se revela después de la elección.

**Verification:** la versión nueva gana al menos 4 de 6 pares; «sin preferencia» cuenta en contra (AC4). Si no, una sola ronda de ajuste de T1 y repetir T3–T4; si vuelve a perder, se abandona el cambio.

## T5 — Integrar y desplegar

- [ ] Versión nueva en `rebuild/harness-engineering` y desplegada en las 6 CLIs.

**Depends on:** T4 con AC4 y AC5 cumplidos.

**Locations:** `content/guidance/global.md`, `tests/content/budget_test.go`, este registro.

**Execution:** hilo principal (entrega).

**Verification:** commit y push; `go run ./tooling/cli status --hosts codex,claude,grok,pi,opencode,cursor --scope user` muestra todos los recursos en la release nueva y verificados; `~/.claude/CLAUDE.md` contiene «Example completion report:».

## T6 — Cerrar

- [ ] Temporales borrados y cambio archivado.

**Depends on:** T5, o el abandono del cambio.

**Execution:** hilo principal.

**Verification:**
- `_support/workspace/2026-09-27-communication-recipe/` queda sin brazos ni corridas, salvo el inventario, los pares elegidos y el recuento.
- Ningún `auth.json` queda en un `shadow-home/`.
- El Engram de uso diario no tiene registros con los nombres de fixture tomados de los `run.json`.
- El historial nativo de Grok queda declarado como límite, sin borrarlo.
- `git mv` a `changes/archive/<fecha>-communication-recipe`.

## Verificación compartida

| Comprobación | Cuándo | Mecanismo y evidencia | Autorización |
| --- | --- | --- | --- |
| Tests de contenido | T1, T5 | `go test ./tests/...` en verde | parte de la implementación |
| Inventario de reglas | T1 | tabla sin filas abiertas; descartes aceptados | el usuario |
| Piloto A/B | T3 | 12 corridas aisladas; `run.json` y evaluaciones por celda | D11-A, D13-A, D14-A |
| Elección a ciegas | T4 | recuento de 6 pares | el usuario |
| Despliegue | T5 | `status` de las 6 CLIs | vigente en la sesión si se cumplen AC4 y AC5 |

## Revisión del plan y progreso

**Primera ronda** (candidato: proposal `1d178aed0413`, design `4ff16bc692d7`, tasks `34bd832d886d`), con dos `review-plan` de solo lectura:

- **Contenido de la guía:**
  - B1, la lista de restricciones cubría ~9 de ~30 comportamientos: aceptado; se agregó el inventario.
  - B2, el piloto no cubre la mayoría de las reglas: aceptado como límite declarado.
  - N1, ejemplo acotado a reportes de cierre, con placeholders y cumpliendo las reglas medidas: aceptado.
  - N2, AC1 relajado a ≤ 8 400 bytes: aceptado.
  - N3, marcador del ejemplo: aceptado.
  - N4, la línea de limpieza se conserva aunque haya menos de 3 áreas: aceptado.
  - N5, nombres anclados: aceptado.
- **Medición:**
  - B1, `--timeout 600s`: aceptado.
  - B2, el sandbox de Codex: resuelto cambiando `close-sequence` por `adaptive-plan` (sin push) y con `--allow-native-trust` autorizado (D13-A).
  - B3, cegado de rutas: aceptado.
  - B4, brazo B con `git archive`: aceptado.
  - N1, hashes: aceptado.
  - N2, paso previo y `--effort`: aceptado.
  - N3, criterio de la pregunta: aceptado.
  - N4, fixture de backlog: aceptado.
  - N5, AC4 como preferencia (D14-A): aceptado.
  - N6, AC5 por celda y con las 12 corridas completas: aceptado.
  - N7, limpieza: aceptado.

Casi todas las correcciones aplican lo que propusieron los revisores. El cambio de `close-sequence` por `adaptive-plan` fue del orquestador.

**Segunda ronda** (solo medición, mismo revisor): confirma que `adaptive-plan` corre solo con cada brazo en Grok y Codex, que pide decisiones y una pregunta, y que ningún otro caso encaja mejor. `--allow-native-trust` se mantiene en las tres corridas de Codex, porque la confianza se escribe al visitar el proyecto. Un hallazgo menor, AC5 y T3 con listas distintas, quedó alineado nombrando los mismos cinco criterios.

**Cobertura:** contenido de la guía y medición. No hay interfaz de usuario ni backend afectados. Límite declarado: el piloto no ejercita todas las reglas; el inventario de T1 las protege.

**Estado:** plan listo; la implementación espera tu visto bueno para arrancar `flow-build` en T1.
