# Tareas

Base del cambio: `32e2a709` (`feat/delegation-threshold-global`), confirmada al empezar el build el 2026-10-17.

## T1: regla de memoria en `global.md`

- [x] "Continuity" define la memoria persistente, qué va a Engram y qué a la memoria del host, y qué hacer cuando se contradicen.

**Closes:** AC1, AC2, AC3.

**Depends on:** ninguna.

**Locations:** `content/guidance/global.md`, sección `## Continuity` (líneas 86-89), al final de la lista.

**Execution:** hilo principal. Es un párrafo, y un brief para delegarlo sería más largo que el cambio.

**Test approach:** check, `rg`.

**Changes:** insertar el párrafo de [design.md § Diseño elegido](design.md#diseño-elegido), tal cual.

**Verification:** una comprobación por cláusula.
- `for p in 'Persistent memory in this guidance means Engram when its tools work in this session' "otherwise the host's own memory" 'not its instruction files' 'fails to connect, authenticate, or write' 'instead of falling back silently' 'When you write memory yourself' "decisions, restrictions, and work or delivery state to Engram, each under its own \`topic_key\`" 'only what concerns that host' 'engram:<project>/<topic_key>' "holds over the host's own memory instructions" "asks to save something in the host's own memory" 'automatic capture may still copy project facts' 'or a memory and a current source, disagree' 'act on either, and correct an entry, only when a current source settles which one is stale' 'report both without editing' "following Engram's conflict prompts" 'replacing only that fact with the pointer' 'this rule authorizes that edit' 'Change only the entries the task touches'; do echo "$p: $(rg -n -F "$p" content/guidance/global.md | cut -d: -f1)"; done` → un número de línea por cláusula.
- Las cláusulas hasta "asks to save something in the host's own memory" caen en una línea, y las siguientes en la línea inmediata posterior. Ambas líneas están entre `## Continuity` y `## Specs and change records` (`rg -n '^## ' content/guidance/global.md`).

## T2: hosts de Engram en `tool-dependencies.md`

- [x] La fila de Engram nombra los hosts que hoy reciben Engram.

**Closes:** AC4.

**Depends on:** ninguna.

**Locations:** `_support/docs/architecture/tool-dependencies.md:22`, columna "Relevant hosts".

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check, `rg`.

**Changes:** reemplazar `Codex, Claude, OpenCode` por el texto de [design.md](design.md#diseño-elegido). Las demás columnas no cambian.

**Verification:**
- `rg -n -F "| Engram | Persistent memory and optional host integration | Claude, Codex, Cursor, OpenCode, Pi; Grok reads Claude's MCP configuration |" _support/docs/architecture/tool-dependencies.md` → una línea.
- `git diff --stat -- _support/docs/architecture/tool-dependencies.md` → 1 inserción y 1 borrado.

## Verificación y revisión

| Compuerta | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite local | Sobre el candidato final, en este checkout | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` sin fallos; `tests/content/budget_test.go` registra el tamaño nuevo de `global.md` | Incluida en la implementación |
| Tamaño de `global.md` | Candidato final | `wc -c content/guidance/global.md` frente a 27 488 bytes en la base; la diferencia va en el commit y en el PR | Incluida |
| Revisión del alcance | Candidato final | Releer `global.md:31,37,39,59,61`, `git-workflow/SKILL.md:18` y `flow-close/SKILL.md:24,26` con el párrafo nuevo y confirmar que no hay contradicción | Incluida |
| Revisión de código dedicada | No aplica | Cambios solo de texto; preferencia guardada para Hive | — |
| Despliegue local | Tras los commits de T1 y T2 | `hive update --out <plan>` + `hive apply --plan <plan>`; el párrafo nuevo aparece en el global instalado de los seis hosts | C1-A, elegido en esta sesión |
| Prueba en sesiones reales | Tras el despliegue | Tú la ejecutas, en sesiones nuevas | Tuya |

## Revisión y progreso

- 2026-10-17: borrador escrito tras R2-A. Las decisiones de diseño se resolvieron con R2-A, `global.md:39` y C1-A, sin preguntas nuevas.
- 2026-10-17: revisión del plan sobre `877cae78e88f` con un `hive-review-plan` nativo (dominio: guía distribuida y comportamiento entre hosts). Duró 92 s y usó ~98k tokens.
  - **Bloqueantes:**
    - B1 (la captura automática del host no la controla el modelo): aceptado. La regla gobierna lo que escribe el agente, y la cláusula de contradicción es la defensa ante esa captura.
    - B2 (corregir sin una fuente que resuelva cuál está vieja; el conflicto propio de Engram): aceptado. Se corrige solo con una fuente actual; Engram sigue su procedimiento; se reemplaza solo el hecho.
  - **No bloqueantes:**
    - N1 (la descripción también cubría `AGENTS.md`/`CLAUDE.md`): aplicado, "not its instruction files".
    - N2 (Engram configurado pero caído): aplicado, se remite a la regla de integraciones caídas.
    - N3 (formato del puntero y `topic_key` propio): aplicado, `engram:<project>/<topic_key>`.
    - N4 (precedencia sobre las instrucciones de memoria del host y `/remember`): aplicado.
  - **Laguna de verificación:** "or a memory and a current source" no se comprobaba en T1; añadido.
  - **Límite:** si un hijo delegado puede escribir memoria queda fuera de alcance.
  - B1, B2 y N2–N4 cambian AC1–AC3, así que van a una re-revisión, la única ronda.
- 2026-10-17: re-revisión sobre `4a8a99556a59`, reanudando al mismo revisor. Duró 42 s y usó ~105k tokens.
  - Sin bloqueantes. Las tres comprobaciones obligatorias se cumplen.
  - **No bloqueantes, aplicados como los propuso el revisor, sin otra ronda:**
    - N1: dos puntos y sin "not a memory migration".
    - N2: "fails to connect, authenticate, or write".
    - N3: "this rule authorizes that edit".
    - N4: "following Engram's conflict prompts when it raises them".
    - N5: la excepción se limita a guardar en la memoria propia del host.
  - Los criterios AC1–AC3 se reescribieron para seguir ese texto. La verificación de T1 ahora comprueba dos líneas consecutivas.
- **Cobertura:** el dominio de la guía distribuida quedó revisado en `877cae78e88f` y re-revisado en `4a8a99556a59`. Los cambios posteriores aplican las propuestas del revisor.
- 2026-10-17, build: T1 y T2 se implementaron en el hilo principal (claude-opus-5-5, observado). Cada una pasó a `[x]` con veredicto `met` de `hive-verify-task`, cuyo modelo es `opus` con esfuerzo `high` (configurado, no observado). Las dos verificaciones corrieron en paralelo:
  - T1, AC1–AC3: 43 s.
  - T2, AC4: 24 s.
- **Compuertas del candidato final:**
  - `go vet ./...` sin hallazgos.
  - `go test ./...` con salida 0.
  - Tests de Python: 23 OK.
  - `global.md` pasó de 27 488 a 28 865 bytes (+1 377).
  - Revisión del alcance: con la definición nueva, "persistent memory" en `global.md:31`, `git-workflow/SKILL.md:18` y `flow-close/SKILL.md:26` apunta a Engram. `:37` y `:39` quedan cubiertos por "this rule authorizes that edit" y "Change only the entries the task touches".
- **Commits en `feat/delegation-threshold-global`:** `2825a561` (T1) y `c6c1ee16` (T2), sin push.
- **Despliegue:** `hive update --out` + `hive apply --plan` desde `c6c1ee16`, transacción `73b00555…`. El texto nuevo está en `~/.claude/CLAUDE.md` (que también usa Grok), `~/.codex/AGENTS.md`, `~/.config/opencode/AGENTS.md`, `~/.pi/agent/AGENTS.md` y `~/.cursor/AGENTS.md`.
- **Estado:** desplegado en local para tus pruebas en sesiones reales. Siguiente paso: tú lo pruebas y decides el merge a `development`, junto con los cambios anteriores de la rama.
- **Límites:** no se comprobó que Cursor y Pi tengan Engram instalado más allá de su configuración. No se observó en la práctica `mem_update` ni `include_history`. Queda fuera de alcance si un hijo delegado puede escribir memoria.
