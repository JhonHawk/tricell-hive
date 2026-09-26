# Tareas

### T1: Guía de cierre

- [x] La pregunta de cierre tiene su propia viñeta después de los elementos de cierre en `global.md`, y `flow-build` termina con la secuencia numerada que solo hace referencias.

**Depende de:** nada.

**Ubicaciones:** `content/guidance/global.md` (`:22`, `:25-30` y la viñeta nueva después de `:30`); `content/skills/flow-build/SKILL.md` (`:58` y `:60`); `tests/content/budget_test.go:12`.

**Ejecución:** hilo principal. Es redacción corta de guía, acoplada a las decisiones de diseño.

**Cambios:** según "Guía (D11-A)" de `design.md`.

**Verificación:**
- `wc -c content/guidance/global.md` y el presupuesto nuevo coinciden, y `go test ./tests/content/...` pasa.
- `rg -n 'close question' content/guidance/global.md content/skills/flow-build/SKILL.md` muestra la condición definida una sola vez, en `global.md`.
- Una lectura final confirma que las opciones y los criterios de sesión nueva no cambiaron.

### T2: Trazas y criterios deterministas

- [x] OpenCode `question` se mapea como pregunta. Los textos de Codex y OpenCode llevan `Role`, y los de OpenCode también `Message`. Existen `close_question_after_report` y `merged_branch_deleted`, con casos que fallan y que pasan.

**Depende de:** nada. Corre en paralelo con T1.

**Ubicaciones:** `tests/pilot/trace.go` (`toolKind`, las ramas de codex y opencode); `tests/pilot/regression.go` y `regression_test.go`; `tests/fixtures/regression/close_question_after_report/` y `merged_branch_deleted/`, con su `provenance.md`; `tests/fixtures/regression/README.md`.

**Ejecución:** delegada a `test-engineer`, que puede modificar solo esas rutas y no hace commits. Recibe `design.md`, el README de regresión y los hallazgos X1 (rollout Codex `01a0dab1`, líneas 79–82), G3 (Grok globex `01a0daf9`, líneas 138 y 408) y G4 (líneas 382–408). Nunca imprime contenido de las sesiones y reescribe los casos al formato que lee `parseTrace`.

**Cambios:** según "Criterios deterministas" de `design.md`. Suma un caso de prueba que confirme que S1 sigue dando lo mismo con los fixtures actuales y que evalúa bien OpenCode con `Role` y `Message`.

**Verificación:**
- `go test -race ./tests/pilot/...` pasa.
- Al revertir por un momento cada rama del criterio que ejercita un caso que falla, la prueba falla; después se restaura.
- `rg -l -i -P '<patrón del README>' -g '!README.md' tests/fixtures/regression` no devuelve nada.

### T3: Variante de guía en el runner y caso `close-sequence`

- [x] `--guidance-source` con `--arm A|B` funciona en `deployed-global` para Codex y Grok, conserva el aislamiento de Engram y la evaluación, y el caso `close-sequence` aplica los criterios de T2 y el estado final del repo.

**Depende de:** T2, porque la evaluación llama a sus criterios.

**Ubicaciones:** `tests/pilot/main.go`, `launch.go`, `memory.go` (sin cambiar su contrato de aislamiento), `flows.go` (`assessFlows`) y sus pruebas; `tests/fixtures/flows/cases.json` y los archivos del fixture `close-sequence`; `tests/fixtures/flows/README.md`.

**Ejecución:** delegada a `backend-developer`, que puede modificar solo esas rutas y no hace commits ni lanza modelos. Recibe `design.md`, `AGENTS.md` (Measurement), `_support/docs/architecture/agent-delivery.md:84` y el precedente `_support/workspace/2026-09-25-role-hints-pilot/run.sh`.

**Cambios:** según "Piloto con el runner" de `design.md`, pasos 1 a 6 y el caso `close-sequence`.

**Verificación:**
- Pruebas unitarias con un host falso que comprueban:
  - que A y B instalan guías distintas en homes distintos;
  - que los hashes quedan en `run.json`;
  - que Codex recibe el entorno de Engram por `-c mcp_servers.engram.env.*`;
  - que Codex y Grok corren con `HOME` paralelo y que ningún destino del manager queda fuera de él;
  - que `auth.json` sigue siendo un enlace al terminar, y que si no lo es se borra y se avisa;
  - que `ENGRAM_DATA_DIR` gana dentro del home paralelo;
  - que los enlaces simbólicos se borran;
  - que `--arm` sin `--guidance-source` en `deployed-global` sigue fallando con un error claro.
- `go test -race ./tests/pilot/...` y `go vet ./...` pasan.
- Una corrida de humo sin modelo, como `--case deployed-smoke`, no aplica: los pilotos con modelo son T4.

### T4: Piloto (D13-B, D15-A)

- [x] 8 corridas (Codex y Grok × A y B × 2) del caso `close-sequence` con el runner, y una tabla que compara A y B.

**Depende de:** T1, porque B necesita la guía nueva, y T3.

**Ubicaciones:** la salida del runner en `_support/workspace/2026-09-26-work-close-sequence/runs/`, y el resumen en la sección de avance de este archivo.

**Ejecución:** delegada a `sdd-verify`, que corre solo el runner con los flags del diseño, escribe solo en esa carpeta y no toca la configuración global. Brazo A: una copia limpia de `HEAD` en `_support/workspace/…/arm-a`. Brazo B: el working tree.

**Verificación:**
- Por corrida: el comando, el brazo y sus hashes, host y versión, modelo, código de salida, estado terminal, duración, el resultado y el tipo de evidencia de `close_question_after_report`, el resultado de `merged_branch_deleted`, si la rama se borró en local y en el remoto, y si hay línea de limpieza.
- El runner informa que el almacén de Engram de cada corrida quedó limpio, y el almacén diario no tiene proyectos de fixture (se comprueba con una búsqueda de su nombre).
- No queda ningún proceso ni enlace simbólico de autenticación.

### T5: Cierre del cambio

- [ ] Suite completa, revisión de código y registro actualizados.

**Depende de:** T1 a T4.

**Ejecución:** hilo principal, con `/code-review` como proceso aparte.

**Verificación:**
- `go test ./...`, `go test -race ./...` y `go vet ./...` pasan.
- `/code-review` (D14-A) sin hallazgos abiertos, o refutados con evidencia.
- La limpieza de `_support/workspace/2026-09-26-work-close-sequence/` queda reportada, y la salida del piloto se conserva hasta que revises el resultado.

## Verificación y revisión humana

| Comprobación | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Pruebas | Después de T1, T2 y T3, en el checkout local | Los comandos de cada tarea pasan | Parte de la implementación |
| Piloto | T4, con el runner, homes paralelos y Engram aislado | La tabla por corrida y la limpieza | Autorizado (D13-B y D15-A, 2026-09-26) |
| Revisión del código | T5, sobre el diff del working tree | `/code-review` | Elegida (D14-A) |
| Tu revisión | Al final (`hold`) | El diff, la tabla del piloto y las diferencias A/B | Tuya: decides commit y despliegue junto con `subagent-effort-profiles` |

No hay interfaz visual, así que no aplica revisión de UI.

## Estado de la revisión y avance

- Borrador del 2026-09-26 con las decisiones D11-A, D12-B, D13-B, D14-A y D15-A.
- Ronda 1 de revisión del plan, en paralelo sobre la versión d116ca3820d1 / 5ff4b284f64b / de4a85e3cc48:
  - **Guía y portabilidad:** 3 bloqueantes (presupuesto exacto, condición duplicada, orden frente a los elementos de cierre) y 4 no bloqueantes (excepción amplia, duplicación en `flow-build`, la línea `:58`, trabajo desatendido). Todos corregidos en el diseño.
  - **Criterios y piloto:** 4 bloqueantes (evaluación sin runner, aislamiento de Engram, pregunta de Codex no observable en `exec`, S1 en OpenCode) y 4 no bloqueantes. Corregidos. El piloto se rediseñó con D15-A, que decidiste.
- Ronda 2, la única re-revisión, con los mismos revisores retomados sobre la versión 9edc2ed2f357 / fcbb4c66a4c3 / c200510ce023:
  - **Guía:** los hallazgos de la ronda 1 quedaron resueltos. Agregó R1 (mover completas las opciones de la pregunta) y R2 (formato de la pregunta fusionada), aplicados tal como los propuso.
  - **Criterios y piloto:** los hallazgos de la ronda 1 quedaron resueltos. El rediseño D15-A trajo R1 a R3 bloqueantes (`HOME` paralelo en Codex, el manager con `GROK_HOME` real, los enlaces de configuración de Codex) y R4 y R5. Todos aplicados tal como los propuso.
  - Queda un riesgo que no se puede verificar leyendo: cómo renueva Codex `auth.json`. Aceptado con la protección (D16-A, 2026-09-26).
- Plan listo. Implementación autorizada con `flow-build` (2026-09-26), en `hold`.
- T1 terminada (2026-09-26):
  - `global.md` tiene la viñeta nueva en la línea 31, después de los elementos de cierre. La línea 22 termina con el contenido del reporte.
  - El archivo pasó de 36,359 a 36,732 bytes (+373). Para compensar quité la frase "it never comes first…", que ya cubre la línea 22 ("before calling the question tool"). `globalGuidanceBudget` sube a 36732 en el mismo cambio.
  - `flow-build` cierra con los 3 pasos en orden, solo con referencias.
  - `rg 'close question' content` muestra que la condición vive solo en `global.md`.
  - `go test ./tests/...` pasa.
- T2 terminada (2026-09-26, `test-engineer`, revisada por el hilo principal):
  - `toolKind` reconoce `question`. Los textos de Codex y OpenCode llevan `Role`, y los eventos de OpenCode llevan `Message`.
  - `closeQuestionAfterReport` y `mergedBranchDeleted` quedaron como funciones, sin conectar a `regressionCriteria`, porque eso corresponde a T3.
  - Hay 7 casos nuevos registrados, con su `provenance.md`.
  - RED observado antes de GREEN. Se comprobó cada caso que falla revirtiendo la rama del criterio, y la restauración quedó idéntica byte a byte.
  - `go test -race ./tests/pilot/...` (89 pruebas) y `go vet` pasan, y el escaneo del README está limpio.
  - Desvío informado: un `jq` exploratorio imprimió en el transcript del hijo un fragmento de razonamiento de negocio de la sesión de Grok. No contenía secretos ni llegó a ningún archivo entregado.
- T3 terminada (2026-09-26, `backend-developer`, revisada por el hilo principal):
  - Hay `--guidance-source` y `--arm A|B` en `deployed-global`, solo para Codex y Grok. Los dos flags se exigen juntos.
  - `guidance_variant.go` resuelve las rutas reales con `codex.Resolve` y `grok.Resolve`. Grok lee su guía de `$HOME/.claude/CLAUDE.md`, y los dos hosts leen las skills de `$HOME/.agents/skills`, así que `HOME` paralelo cubre ambos.
  - El manager corre sin `CODEX_HOME` ni `GROK_HOME` y con control de destinos.
  - Codex recibe un `config.toml` generado, solo con `[mcp_servers.engram]`, y `auth.json` como enlace, con su protección.
  - Las rutas protegidas cubren el `CODEX_HOME` real y el paralelo.
  - El caso `close-sequence` quedó conectado a `assessFlows`, junto con la comprobación de la rama y de la línea de limpieza.
  - RED observado antes de GREEN, restaurado byte a byte.
  - `go test -race ./tests/pilot/...` (107 pruebas) y `go vet` pasan, y `go test ./...` también.
  - Desvíos aceptados: los campos se llaman `GuidanceBlockPath` y `GuidanceBlockHash`, y los dos flags se exigen en ambos sentidos.
- Desvío del piloto (2026-09-26, decidido por el hilo principal dentro del diseño autorizado):
  - Al instalar el brazo A (una copia limpia de `HEAD`), el parser del working tree fallaba con `unsupported agent field "claude_effort"`. Los agentes de `HEAD` todavía usan `claude_effort`, que el cambio en `hold` `subagent-effort-profiles` renombró.
  - Solución: el brazo A usa el `global.md` y las skills de `HEAD`, pero una copia de `content/agents/` del working tree. Así los brazos solo difieren en la guía que se mide (`global.md` y `flow-build`), y se comprueba con hashes y `diff -r` antes de correr.
  - Las dos corridas de sondeo fallidas se conservan como evidencia y no cuentan en la tabla.
- T4, primer resultado (2026-09-26, `sdd-verify`: 8 corridas, n=2 por celda). Versiones: codex-cli 0.157.0 con gpt-6-sol medium, Grok 1.0.38 con grok-4.7-build-fast high. Engram quedó aislado y limpio en todas las corridas, el almacén diario sin proyectos de fixture, `auth.json` siguió siendo un enlace en las 4 de Codex, y el worktree `arm-a` se borró.
  - **Lectura corregida por el hilo principal,** revisando las trazas:
    - Codex: A 0/2 y B 2/2 cierres con pregunta, en texto.
    - Grok: A 2/2 con `ask_user_question` nativo; B 1/2 en texto, a mitad de línea, y B-1 sin pregunta.
    - `merged_branch_deleted`, rama ausente y línea de limpieza pasan 8/8.
  - **Falsos negativos del criterio:**
    - Grok sin interfaz devuelve `ask_user_question` sin respuesta y sigue escribiendo, así que el criterio ve texto después de la pregunta.
    - Solo aceptaba "?" al final de la última línea.
  - **Hueco del runner:** Codex escribe la confianza del proyecto en el `config.toml` del home paralelo y el runner lo marca como cambio en un recurso protegido. No afecta a `~/.codex`.
  - **D17-A (2026-09-26):**
    - corregir los dos falsos negativos (la pregunta nativa cuenta si después solo vienen su resultado y texto; en texto cuenta un "?" en el último párrafo);
    - agregar casos desde estas trazas;
    - corregir el hueco del runner;
    - volver a evaluar las 8 corridas con `--assess`, sin correr modelos otra vez.
    
    La guía no cambia: en Codex mejora, y en Grok queda en observación sobre sesiones reales.
- T4 terminada con D17-A (2026-09-26, `test-engineer`, revisada por el hilo principal):
  - El criterio acepta una pregunta nativa seguida solo de su resultado y de texto, y un "?" en el último párrafo.
  - El caso G3 se ajustó para que conserve su forma real (trabajo después de la pregunta temprana) y sigue fallando.
  - Hay 3 casos nuevos sacados de las trazas del piloto.
  - La confianza escrita en el `config.toml` del home paralelo se permite cuando se usa `--allow-native-trust`.
  - **Reevaluación sin modelos (`criterion-assessment-v2.json`):**
    - `close_question_after_report`: Codex A 0/2 y B 2/2 (texto); Grok A 2/2 (herramienta) y B 1/2 (texto).
    - La limpieza de ramas y la línea de limpieza pasan 8/8.
    - En las corridas de Codex, el control de recursos protegidos sigue en fallo, porque `--assess` no vuelve a calcularlo; la corrección aplica a corridas nuevas.
  - **Conclusión:**
    - La guía nueva produce la pregunta de cierre en Codex.
    - En Grok no la mejora: con la guía nueva pasó de la herramienta nativa a texto, y una corrida no preguntó.
    - Con n=2 no hay confiabilidad. Grok queda en observación sobre sesiones reales.
- Hallazgo fuera del cambio: la sesión de Codex del instalador publicó en `239a9d7` el `deployment-manager.md` completo, con el párrafo de `claude_effort` del cambio en `hold` `subagent-effort-profiles`. La documentación publicada describe un comportamiento que `HEAD` todavía no tiene, y se corrige al entregar ese cambio.
- `/code-review` alto (2026-09-26, D14-A), 10 hallazgos, todos con fundamento:
  - **Corregido en el hilo principal (hallazgo 6):** la secuencia de `flow-build` aplicaba siempre. Ahora empieza con "When the work item completes", y la frase general "Report what changed…" vuelve a cubrir cualquier parada (espera de revisión, `hold`, bloqueo).
    - El piloto midió la versión anterior. Para el caso `close-sequence`, donde el trabajo sí se completa, la diferencia no cambia qué se pide.
  - **Delegados a `test-engineer`, en curso:** 1 (limpieza de `auth.json` en toda salida), 2 (exigir merge), 3 (notas preparadas y sin preparar), 4 (borrados entre merge y push), 5 (rutas del home paralelo), 7 (reporte antes de la pregunta), 8 (línea de limpieza en el reporte), 9 (registrar si se leyó `flow-build`) y 10 (casos de Codex para `merged_branch_deleted`).
  - **Comprobado por el hilo principal:** no quedó ningún `auth.json` en las carpetas del piloto.
- Entrega autorizada por el usuario (2026-09-26): "commit + push y deploy cuando acabe". Commit selectivo por rutas, push a `rebuild/harness-engineering` y despliegue a los seis hosts cuando termine `work-close-sequence`. Al integrar, la carpeta de cambio se cierra y se archiva.
- Correcciones de `/code-review` terminadas (2026-09-26, `test-engineer`, revisadas por el hilo principal):
  - Los 10 hallazgos quedaron corregidos con prueba antes y después.
  - Hallazgo 7: `grok-tool-then-text` pasa a ser un caso que falla, porque la pregunta venía antes del reporte.
  - **Reevaluación sin modelos (`criterion-assessment-v3.json`):**
    - `close_question_after_report`: Codex A 0/2 y B 2/2; Grok A 0/2 y B 1/2.
    - Borrado de rama, rama ausente, línea de limpieza, notas preservadas y lectura de `flow-build` pasan 8/8.
    - En las corridas de Codex, el control de recursos protegidos sigue en fallo por el límite de `--assess` ya documentado.
  - `go test ./...`, `go test -race` en `tests` e `integrations`, `go vet`, el escaneo de casos y `gofmt` pasan.
- T5: verificación completa hecha. Siguen el commit, el push y el despliegue autorizados.
- Cerrado (2026-09-26): integrado en `rebuild/harness-engineering` con `5f7ad9c`. La carpeta se archiva en `changes/archive/2026-09-26-work-close-sequence/`. No hay deltas de requisitos: el repositorio no tiene `specs/`.
