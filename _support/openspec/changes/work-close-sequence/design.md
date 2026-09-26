# Diseño

## Contexto verificado (2026-09-25)

**Sesiones con la falla** (release 6465f0e9d90c; la guía y `flow-build` estaban cargados en las dos):

| Hallazgo | Sesión | Qué pasó | Evidencia |
| --- | --- | --- | --- |
| X1 | Codex sample-project `01a0dab1` (gpt-6-sol medium) | El reporte de ARK-687 cerró el turno sin pregunta, y `task_complete` vino justo después. El modo por defecto de Codex dice "strongly prefer … executing the user's request rather than stopping to ask questions" | rollout líneas 8, 79, 82 |
| G3 | Grok globex `01a0daf9` (grok-4.7-build-fast high) | El reporte del despliegue a producción cerró el turno sin pregunta. `ask_user_question` se usó una sola vez, para el alcance | `chat_history.jsonl` líneas 138, 408 |
| G4 | Grok globex `01a0daf9` | `gh pr merge 165 --merge --delete-branch=false` y nunca borró la rama. El reporte no tiene línea de limpieza, ni dice nada del `git bundle` que dejó en `_support/workspace` y copió al servidor | líneas 382, 396–408 |

**Casos que sí cumplieron:** OpenCode sample-project `ses_f25205` hizo la pregunta de cierre en el mensaje 664 y respetó "Terminar por ahora". Codex ark `01a0dab3` usó `send_user_message_question` y `request_user_input_async` durante la tarea. Con eso, que la herramienta esté disponible no explica la falla.

**Texto actual:**
- `content/guidance/global.md:22`: la pregunta de cierre es la cuarta frase de un párrafo largo de Communication.
- `global.md:145`: la limpieza al cerrar.
- `content/skills/git-workflow/SKILL.md:36`: la limpieza de ramas después del merge. Ya advierte que `gh pr merge --delete-branch` se salta una comprobación y que hay que borrar por separado.
- `content/skills/flow-build/SKILL.md:60`: termina con "run the global task-close cleanup and report it, and end the report with the global close question".
- `tests/content/budget_test.go` limita `global.md`, que hoy pesa 36,359 bytes.

**Modelo de trazas** (`tests/pilot`):
- `trace.go:248-266` reconoce como pregunta solo `askuserquestion` y `ask_user_question`. Las herramientas de Codex (`request_user_input`, `request_user_input_async`, `send_user_message_question`) y la de OpenCode (`question`) caen como `tool` genérico.
- `question_after_detail` (S1) revisa el detalle antes de una pregunta, pero nunca detecta que falte una.
- Ningún caso de `tests/fixtures/flows/cases.json` llega a un merge con limpieza de rama. `git-delivery` trabaja sobre `main` y prohíbe limpiar.

**Mecánica de pilotos:**
- **Variante de guía:** el piloto #30 (`_support/workspace/2026-09-25-role-hints-pilot/run.sh`) la inyectó a mano con un home paralelo, `CODEX_HOME` en Codex y `HOME` con el `GROK_HOME` real en Grok, fuera de `tests/pilot/main.go`. El runner no admite brazos en `deployed-global` (`main.go:283-285`).
- **Engram:** ese piloto **no** aisló Engram: borró después los proyectos del almacén diario. El runner sí lo aísla (`memory.go:44-284`): comprueba un almacén vacío, levanta un servidor HTTP propio, pasa `-c mcp_servers.engram.env.*` a Codex (que filtra el entorno de sus servidores MCP) y limpia con marcador de dueño.
- **Evaluación:** `--assess` necesita el `run.json` que escribe el runner (`assess.go:37`).
- **Versiones instaladas:** codex-cli 0.157.0, que no admite `request_user_input` en `exec`, y Grok 1.0.38.

## Diseño elegido

### Guía (D11-A)

1. **`global.md`: condición única.** La pregunta de cierre conserva su condición actual ("when a work item completes"; con un pendiente, el trabajo no está completo) y solo vive en `global.md`. La frase se mueve a una viñeta propia, que va **después** del bloque de elementos de cierre (`global.md:25-30`). Así es lo último que el modelo lee de la plantilla del final del mensaje. La viñeta dice:
   - En un trabajo completado, después del reporte y de sus elementos de cierre, el último acto del turno es la pregunta de cierre. Un turno que termina con el reporte sin ella está incompleto.
   - Se hace con la herramienta nativa. Si no hay herramienta en la sesión, o el trabajo es desatendido, va en texto con las mismas opciones.
   - En un cierre, la pregunta de rutas de Recommendations se funde con la pregunta de cierre, para no hacer dos preguntas seguidas. Cada opción de seguir ahora nombra el flujo que arranca, como en Recommendations: una opción por ruta, más la de sesión nueva y la de terminar.
   - Solo se omite si el usuario ya respondió la pregunta de cierre de este mismo punto de cierre o pidió terminar explícitamente. Una pregunta anterior de alcance o de entrega no cuenta.
   - Las opciones y los criterios de sesión nueva son los mismos.

   Se mueve todo desde "After the report" hasta el final de `:22`: las opciones, los criterios de sesión nueva, lo que se dice al ofrecer terminar y "Name the ticket and the reason". `:22` termina con el contenido del reporte. El presupuesto está exacto (`globalGuidanceBudget = 36359`, `tests/content/budget_test.go:12`), así que se sube en esta misma entrega. El aumento se mide con `wc -c`, se justifica y se intenta compensar recortando redundancias del bloque que se mueve.
2. **`flow-build`: solo referencias, en orden.** La frase "Report what changed…" (`:58`) y el cierre de `:60` se funden en una secuencia numerada que no copia reglas:
   1. Aplicar la limpieza de cierre global y, si hubo merge, "Clean up after a merge" de `git-workflow`.
   2. Escribir el reporte de cierre con su línea de limpieza.
   3. Hacer la pregunta de cierre global.
3. **Sin cambios en `git-workflow`:** ya dice que hay que borrar la rama por separado. Es probable que G4 usara `--delete-branch=false` para seguirlo y luego no volviera a ese paso, aunque no está confirmado. El paso 1 de la secuencia lo trae al cierre.

### Criterios deterministas (D12-B)

- **Trazas:**
  - `toolKind` reconoce `question` (OpenCode). Para que S1 siga siendo válido, los eventos de texto y de herramienta de OpenCode llevan `Role="assistant"` y el `messageID` como `Message`, con un caso de prueba que lo cubra.
  - En Codex, los textos (`agent_message`) llevan `Role="assistant"`. Su pregunta **no** se mapea como herramienta: `codex exec` no admite `request_user_input` (0.157.0), así que en Codex la pregunta de cierre solo se observa como texto. A4 lo declara.
- **`close_question_after_report`:**
  - Considera solo eventos `Kind=="text"` con `Role=="assistant"` y excluye `final` (`trace.go:450`).
  - Se aplica solo si la corrida terminó en estado `completed`. Con un timeout queda como no evaluable.
  - **Pasa** si después del último texto del asistente hay un evento `question` (evidencia `tool`), o si ese texto termina en una línea que acaba en "?" (evidencia `text_fallback`, para revisar a mano los "?" retóricos).
  - **Falla** si el último texto no tiene pregunta después.
  - Solo se usa en los casos de regresión y en el caso de piloto `close-sequence`, que exige además que haya habido merge.
- **`merged_branch_deleted`:**
  - Solo cuenta comandos con `Success==true`, y reutiliza `shellSegments` y `gitInvocation` (que ya manejan `zsh -lc` y `-C`).
  - **Merge:** `gh pr merge …`, o `git merge` seguido de un `git push` a la rama base declarada.
  - **Borrado:**
    - `git push … --delete|-d <b>` y `git push origin :<b>` o `:refs/heads/<b>`;
    - `git branch -d|-D|--delete <b>`;
    - `gh pr merge … --delete-branch|-d` sin `=false`;
    - `gh api -X DELETE …/git/refs/heads/…`.
  - Como `gh pr merge <n>` no nombra la rama, cualquier borrado de rama después del último merge cuenta.
  - **Falla** si hubo merge y ningún borrado después. **No aplica** si no hubo merge.
  - Límite declarado: si GitHub borra la rama solo, el criterio falla.
- **Casos de regresión:**
  - Se reescriben al formato que lee `parseTrace`: `codex exec --json` para Codex y `streaming-messages-json` para Grok. Los originales vienen de rollout y `chat_history`, y `provenance.md` lo anota.
  - Codex: casos con texto, uno que falla (X1) y otro que pasa.
  - Grok: casos con `ask_user_question` o con texto, uno que falla (G3) y otro que pasa.
  - Borrado de rama: uno que falla (G4, `--delete-branch=false`) y otro que pasa.
  - `TestRegressionCriteriaReturnsAllThree` y el README se actualizan según dónde se registren los criterios.

### Piloto con el runner (D13-B, D15-A)

- **Runner (`tests/pilot`):** en el modo `deployed-global` se acepta `--guidance-source <checkout>` junto con una etiqueta de brazo `--arm A|B`. Por cada corrida, el runner:
  1. instala la guía de ese checkout con `tooling/management` (scope user), con `Options.Home` = home paralelo y **sin** `GROK_HOME` ni `CODEX_HOME` en el entorno del manager. Un test falla si algún destino queda fuera del home paralelo (R2);
  2. prepara la autenticación sin copiarla. En Codex se enlaza **solo** `auth.json`, y `config.toml` se genera en el home paralelo conservando la definición del MCP de Engram, que necesita `-c mcp_servers.engram.env.*`. Al terminar, se comprueba que `auth.json` sigue siendo un enlace; si Codex lo reemplazó por un archivo, se borra sin leerlo y se avisa de inmediato (R3);
  3. lanza el host con **`HOME` paralelo** en los dos. Codex lleva además `CODEX_HOME` paralelo, porque sus skills viven en `$HOME/.agents/skills` (R1). Grok lleva el `GROK_HOME` real para su autenticación, con el límite declarado de que carga los agentes desplegados y no los del brazo, lo que no afecta a la pregunta de cierre;
  4. registra en `run.json` los hashes de `AGENTS.md` y de `flow-build/SKILL.md` que el host va a leer. Un test con host falso comprueba que las rutas de lectura son las del home paralelo;
  5. extiende el control de rutas protegidas al `CODEX_HOME` paralelo y al real;
  6. aplica el home paralelo antes de `startHTTP` del Engram aislado, de modo que `ENGRAM_DATA_DIR` gana a cualquier ruta por defecto dentro del home paralelo, con un test que lo comprueba (R5).

  Se mantienen el aislamiento de Engram del runner (un almacén vacío verificado, servidor HTTP propio, `-c mcp_servers.engram.env.*` en Codex y limpieza con marcador de dueño), el control de rutas protegidas, la captura de trazas y la evaluación.
- **Caso `close-sequence`** en `tests/fixtures/flows/cases.json`: un repo con remoto bare local, base `main` y `user.name`/`user.email` fijados en el repo, como `setupGitDelivery` (`main.go:136`), porque con `HOME` paralelo no hay `~/.gitconfig` (R4). El pedido autoriza crear `fix/…`, hacer un cambio mínimo con su prueba, commit, push, integrar en `main` con push y cerrar. `assessFlows` le aplica los dos criterios nuevos y el estado final del repo (la rama ausente en local y en el remoto).
- **Corridas:** Codex `gpt-6-sol` (esfuerzo `medium`) y Grok `grok-4.7-build-fast` (esfuerzo `high`), brazos A (`HEAD`) y B (working tree), 2 corridas por celda. En total son 8, con límite de 600 s. Codex decide y registra `-a never -s workspace-write` o `--codex-bypass-sandbox` según `tests/fixtures/flows/README.md`.
- **Registro por corrida:** el comando, los hashes del brazo, el código de salida, el estado terminal, la duración, el estado de Engram, la ruta de la traza y los criterios.
- **Límites:**
  - el fixture sin `gh` no reproduce el camino exacto de G4 (`gh pr merge --delete-branch=false`), solo la limpieza con git local;
  - dos corridas por celda no establecen confiabilidad;
  - si el modelo se niega a hacer el merge y pregunta, la corrida no cuenta.

## Riesgos acotados

- No está verificado que `grok -p` 1.0.38 exponga `ask_user_question` sin interfaz. El criterio acepta la pregunta en texto.
- **Autenticación de Codex (sin verificar):** no se sabe si Codex 0.157 renueva `auth.json` reemplazando el archivo o escribiendo encima. Si lo reemplaza, la copia de credenciales queda en el home paralelo, donde se detecta y se borra. Además, el token real podría quedar invalidado por la rotación, y habría que volver a hacer `codex login`. Aceptado con esa protección (D16-A, 2026-09-26).
