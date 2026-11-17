# Tasks

Base: development `15b2e402c72543b74d63e50df1bc2b188638212a`. Sin rama de implementación. Las decisiones D50/D51 describen el producto futuro: esta construcción sigue las reglas actuales para hijos; sus commits los hace el principal. Solo el piloto autorizado ejerce el nuevo contrato en su fixture.

### T1 — Fijar el procedimiento canónico de coordinación

- [x] Crear el procedimiento E2E y su frontera con los flujos ordinarios y unattended-delegation.

**Closes:** AC1, AC5; observación conductual en T8.

**Depends on:** none.

**Locations:** `content/skills/flow-build/references/e2e-coordination.md`; `content/guidance/global.md`; deltas de este cambio.

**Execution:** principal; decide el orden y las excepciones del contrato aprobado.

**Test approach:** check.

**Changes:** activación explícita con objetivo/criterios/destino/límites; procedimiento único para asignar, esperar señales nativas, comprobar evidencia, corregir, sustituir e integrar. Los cambios de alcance o efectos sin autorización siguen preguntándose. La ausencia del usuario no activa por sí sola E2E. Mantener fuera cualquier scheduler o persistencia prometida por instalar instrucciones.

**Verification:** revisar cada requisito del delta contra su hogar y condiciones de lectura. Pruebas existentes de enlaces y límites; T8 observa AC1/AC5. Nada de pruebas textuales que se presenten como comportamiento del modelo.

### T2 — Incorporar el responsable local con una frontera explícita

- [ ] Añadir el rol local y su asignación con ramas/worktrees y recursos exclusivos.

**Closes:** AC2; commits y comportamiento reales en T8.

**Depends on:** T1.

**Locations:** `content/agents/development/hive-build-owner.md`; `content/skills/flow-build/references/delegation.md`; `tests/content/agent_boundary_test.go`; `_support/docs/architecture/agent-delivery.md`.

**Execution:** `hive-build-backend`, con ownership solo de estos archivos; principal comprueba y hace commits.

**Test approach:** tdd.

**Changes:** excepción nominada del rol owner con perfil implement existente; no cambiar los demás roles ni ampliar perfiles globalmente. Reutilizar `git_state.py commit` y su escaneo, con rutas explícitas. Prohibir remoto, trackers, hijos, cambio de worktree y acceso a recursos no asignados. Principal prepara aislamiento, cuentas e instalaciones y asigna verificadores. Confirmar que el renderer acepta el nuevo rol sin tocar goldens sintéticos de contenido fijo.

**Verification:** prueba de frontera del nuevo rol y de los implementadores existentes; paquetes `tests/content` e `integrations/agents`; inspección de salida renderizada en los hosts soportados, sin afirmar ejecución en ellos. Verificador de tarea independiente por cambiar autoridad.

### T3 — Registrar unidades y rechazar handoffs inválidos

- [ ] Extender los helpers existentes con registro multiunidad, actualización segura y validación Git local.

**Closes:** AC3, AC4, AC5, AC12; AC5 y la parte conductual de AC4 se observan en T8.

**Depends on:** T1.

**Locations:** `content/skills/flow-plan/scripts/tasks.py`; `content/skills/flow-plan/scripts/check_change.py`; `content/skills/flow-plan/references/execution-record.md`; `tests/skills/tasks_test.py`; `tests/skills/check_change_test.py` si existe o su prueba propietaria actual.

**Execution:** `hive-build-backend`, sin solapar T2/T4. Principal integra.

**Test approach:** tdd.

**Changes:** operaciones opcionales `execution init/check/update/validate-handoff` y puntero desde checkpoint; mantener CLI anterior compatible. Registro local ignorado fuera del change folder versionado. Único escritor principal, generación esperada, escritura atómica/excluida y estados definidos en design. Reutilizar observación Git existente donde alcance; cualquier consulta local adicional es factual y no ejecuta comandos del contrato H1. Registrar integration_scope/local|remote, integration_head y resultado de comprobación conjunta antes de integrated; la publicación se registra aparte. No convertir metadatos PASS en prueba de aceptación.

**Verification:** unittest con archivos y repositorios Git locales: duplicados, versión, IDs/ciclos, conflictos de asignación, generación obsoleta, escritura interrumpida, rama/head/ancestry incorrectos, rutas fuera de alcance, dirty handoff, evidencia ausente o antigua y autor=verificador. Probar también CLI anterior. Verificador de tarea independiente por estado/concurrencia.

### T4 — Generar y mantener el contrato de verificación del proyecto

- [ ] Añadir H1 con plantilla, init no destructivo, validación estructural y mantenimiento por cambios afectados.

**Closes:** AC6; comandos reales de la fixture se observan en T8.

**Depends on:** T1.

**Locations:** `content/skills/flow-verify/references/project-verification.md`; `content/skills/flow-verify/assets/project-verification.json`; `content/skills/flow-verify/scripts/check_contract.py`; `tests/skills/verification_contract_test.py`.

**Execution:** `hive-build-backend`, con archivos disjuntos de T2/T3/T5.

**Test approach:** tdd.

**Changes:** formato v1 de design, destinos dentro del support home existente y scripts en ubicaciones de código. Init produce campos no comprobados y nunca sobreescribe. Check valida estructura/referencias y no ejecuta shell. Los flujos generan/completan el contrato desde el proyecto cuando hace falta y mantienen solo entradas afectadas; no duplican contratos equivalentes ni instalan infraestructura de tests por su cuenta.

**Verification:** unittest para no sobrescritura, múltiples targets, referencias de features, capacidades no aplicables/no soportadas, JSON duplicado, paths/symlinks que escapan y evidencia incompleta. Verificador independiente del contrato y su frontera datos/ejecución.

### T5 — Encauzar correcciones y revisar debilitamientos de controles

- [ ] Integrar H2 y H3 en sus procedimientos existentes sin prohibición universal de comentarios/supresiones.

**Closes:** AC7, AC11; resultado aplicado se observa en T8.

**Depends on:** T1.

**Locations:** `content/skills/flow-build/references/defect-fix.md`; `content/skills/harness-audit/references/instruction-files.md`; `content/agents/review/hive-review-code.md`; `content/skills/flow-verify/references/suppression-review.md`.

**Execution:** `hive-write-docs`, con los contratos de design como fuente y sin editar el change folder.

**Test approach:** check.

**Changes:** escalera arquitectura → tipos → lint accionable → tests → prosa; tabla acotada regla/mecanismo/caso negativo y eliminación de pasos sustituidos. Dar a H3 un hogar canónico requerido por flow-verify con cualquier revisión, incluido el brief de la revisión nativa, y enlazarlo desde el rol. Inventariar supresiones y skips añadidos o ampliados en el diff, incluyendo configuración; exigir fallo/contrato concreto. Conservar el comentario de causa de workarounds autorizados. No añadir escáner universal de regex ni fixtures bajo tests/fixtures/regression.

**Verification:** auditoría de consistencia con HA-IF-05, defect-fix y review evidence; tamaño/enlaces. Casos positivo y negativo de supresiones y corrección controlada en T8. La prosa por sí sola no cierra AC7/AC11.

### T6 — Conectar los flujos y medir el costo completo de lectura

- [ ] Integrar las llamadas al procedimiento y helpers, actualizar documentación y medir el antes/después comparable.

**Closes:** AC1, AC2, AC4, AC6, AC9; AC conductuales esperan T8.

**Depends on:** T2, T3, T4, T5.

**Locations:** `content/skills/flow-research/SKILL.md`; `flow-plan/SKILL.md` y `references/plan-handoff.md`; `flow-build/SKILL.md`; `flow-verify/SKILL.md`; `flow-close/SKILL.md` y `references/branch-cleanup.md`; `unattended-delegation/SKILL.md`; `git-workflow/SKILL.md`; `tests/content/reading_footprint_test.go`; `_support/docs/architecture/instruction-resources.md`; `site/src/content/docs/flows/{overview,research,plan,build,verify,close}.md`; `CHANGELOG.md`; README/llms solo si sus afirmaciones cambian.

**Execution:** principal para contratos y enlaces; `hive-write-docs` puede recibir las páginas y fuentes finales con ownership disjunto.

**Test approach:** check.

**Changes:** continuation E2E autorizada sin otro comando humano por etapa; conservar los stops legítimos. Reutilizar flags de ramas activas de cleanup, sin motor de limpieza nuevo. Una fuente para cada procedimiento, sin duplicar pasos de scripts. Medir prosa/scripts/referencias por separado; escenario E2E completo y escenario anterior con mismos paths a ambos lados. Reconciliar las descripciones obsoletas que se tocan en arquitectura/specs. No ampliar global para alojar el procedimiento.

**Verification:** `go -C <repo> test ./tests/content -run 'Footprint|Skill.*Size|AgentBoundary' -count=1 -v`; validar enlaces; comprobar que cada SKILL/ref no supera 16,000 B. Reportar delta global y lecturas de ambos escenarios sin afirmar tokens. `pnpm --dir <repo>/site check`, `lint` y `build`. Completar matriz de hogares/consumidores para el PR.

### T7 — Preparar una única observación nativa y su límite

- [ ] Añadir al soporte de pilotos existente un caso E2E acotado y sus comprobaciones sin invocar modelos durante esta tarea.

**Closes:** AC10; la evidencia real se obtiene en T8.

**Depends on:** T1.

**Locations:** `tests/pilot/e2e_case.go` y prueba enfocada; `tests/pilot/launch.go`, `launch_test.go`, wrappers de deadline en `memory.go` y registro del nuevo caso; fixtures generadas por el caso, nunca fixtures/regression.

**Execution:** `hive-build-backend`; no lanza modelos ni instala contenido global.

**Test approach:** tdd.

**Changes:** reutilizar memoria aislada, captura de trazas, preservación y grupos de procesos existentes. Caso desplegado nativo, sin nuevo adaptador Claude ni manipular autenticación. El launcher actual de Claude solo preautoriza Read/Write/Edit/Glob/Grep/Bash: definir para este caso las herramientas nativas comprobadas de skills, delegación y comunicación necesarias. Probar argumentos y permisos efectivos antes del único lanzamiento. Si se elige Codex previamente, ampliar a este caso la excepción actual de --ephemeral de native-agent-dispatch y probarla. No ampliar permisos de todos los casos ni usar bypass de sandbox implícito. Dos unidades, defecto controlado para corrección/verificación y devolución parcial para sustitución; marcar inyección y condiciones de fixture como tales. Prompt inicial sin instrucciones por etapa. Registrar modelo, hashes, source reads y eventos; una lista de catálogo no prueba lectura. Techo monotónico 20 minutos desde setup efectivo, cada subprocess usa tiempo restante y se reserva parada/restauración. Sin retries ni cambio de host tras lanzar modelos.

**Verification:** pruebas ordinarias del caso, contador de intentos, timeout global y cleanup de grupos usando procesos locales inocuos. El caso nunca puede iniciarse por ejecutar `go test ./...`. Verificar serialización de evidence y clasificación pass/failed/blocked/inconclusive. No simular respuestas de modelos para afirmar autonomía. El caso termina integrando los heads de ambas unidades en un destino local y corriendo una comprobación conjunta; las pruebas del driver cubren que verified no equivale a integrated. Como cambia control concurrente de procesos, ejecutar tests de race del paquete.

### T8 — Verificar el candidato y ejecutar el piloto autorizado

- [ ] Completar suites/revisión y la única ejecución nativa; restaurar instalación y reportar cada AC con su evidencia o falta de ella.

**Closes:** AC1, AC2, AC3, AC4, AC5, AC6, AC7, AC9, AC10, AC11, AC12.

**Depends on:** T6, T7.

**Locations:** candidato completo; evidencia ignorada en `_support/workspace/<fecha>-bounded-e2e-coordination/`; checkpoint de este plan.

**Execution:** principal; revisores independientes según respuestas de entrega, y verificador de tarea para los contratos de riesgo. El modelo del piloto usa únicamente el candidato instalado y sus roles nativos.

**Test approach:** check.

**Changes:** ninguna función adicional. Leer rubric de sesiones y pilot-procedure; aplicar la autorización D55 A de instalación temporal. Preparar candidato que preserve trial y planes de retorno, inspeccionar diff exacto y todos los commits que se instalarían. Capturar release previo inmediatamente antes de aplicar. Usar Claude Code instalado y autenticación habitual; no instalar otro host. Conservar una sola ejecución; Codex es alternativa previa al lanzamiento si el preflight obliga, no un segundo intento.

**Verification:** suites de CONTRIBUTING (`go vet ./...`, `go test ./...`, unittest discover, site check/lint/build), cada una en ruta absoluta; `go -C <repo> test -race ./...` cuando el runner cambie concurrencia, conforme a CONTRIBUTING; usar umask 022 en la suite. Revisión antes del piloto para no desplegar defectos conocidos. Piloto máximo 20 minutos; readback de habilidades/roles y lectura efectiva; corroborar Git y outputs, no resúmenes; integrar ambas unidades solo localmente y comprobar su resultado conjunto antes de marcar integrated. Al terminar parar solo los grupos/PIDs del caso y restaurar sin pisar cambios ajenos. Un resultado parcial no cierra AC. Mostrar evidencia y completar entrega Git según respuesta vigente, seguida de flow-close. No ejecutar una segunda observación sin otra autorización.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Contratos y Git local | T2–T4, fixtures locales sin modelos | Casos negativos y positivos, stdout/exit | Implementación futura |
| Fronteras y lecturas | T2/T6 | Renderer, enlaces, cuotas y medición comparable | Implementación futura |
| Suite completa | T8, candidato exacto | Go, Python y site | Implementación futura |
| Revisión de código | T8 antes de instalar/publicar | Nativa Codex /review, findings resueltos y brief H3 | D54 A |
| Instalación temporal | T8, instalación real compartida | Release, plan/apply/readback/retorno | D55 A |
| Piloto de modelos | T8, un host instalado | Una traza, modelo, fuentes y resultados | D52, un intento/20 min |

No hay UI de producto ni cuentas de aplicación. Las páginas documentales se compilan; no se añade revisión UX por cambios de texto. Autenticación del host se usa sin imprimirla ni copiarla. Memoria ficticia se aísla y su limpieza identifica únicamente los registros del caso.

## Checkpoint

- Updated: 2026-11-17
- Stage: build · T2 implemented; T3/T4/T7 delegated; T6 wiring prepared
- Last commit: 596aeb688f802414c0b2343456805e450b98e3c4
- Open findings: none
- Next: Independent T1/T2 verification; T5 documentation when slot opens; finish T3/T4/T7 and T6

### Evidence

- Investigación local de estado y pilotos completada; ningún modelo ni instalación ejecutado.
- T1 content suite exit 0; T2 AgentBoundary/Render/SkillSize and integrations/agents exit 0; native behavior pending T8

## Review status and progress

2026-11-17 — T1–T8 propuestas. No hay tareas implementadas. D50 A, D51 A y D52 aprobadas; D53 A automatic, D54 A revisión nativa Codex y D55 A instalación temporal aprobadas posteriormente.

2026-11-17 — R1: launcher específico del caso, H3 independiente del mecanismo, integración local conjunta; AC8 retirado y sustituido por AC11, AC12 añadido. Todas las tareas siguen pendientes.

2026-11-17 — R1 revisada sin hallazgos materiales residuales; validador de plan pasa. D50–D55 registradas. Ninguna tarea implementada; esperar aceptación de la ruta a build.
2026-11-17 — T1 · 596aeb688f802414c0b2343456805e450b98e3c4 · unverified: AC1/AC5 behavioral evidence pending T8; independent contract verification pending

2026-11-17 — Build iniciado por `$flow-build`. T7 incluye wrappers compatibles de deadline en memory.go: los helpers actuales usan timeouts fijos; el límite total debe cubrir también su setup. No cambia el criterio ni habilita otro piloto.
