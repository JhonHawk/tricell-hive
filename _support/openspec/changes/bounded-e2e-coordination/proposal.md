# bounded-e2e-coordination — Coordinación E2E con responsables locales

| Campo | Valor |
| --- | --- |
| Estado | Build · T1 en curso |
| Tracker | Ninguno |
| Git | tricell-hive · development · 15b2e402c72543b74d63e50df1bc2b188638212a · feat/bounded-e2e-coordination |

## Decisions

2026-11-17
- **D50 A** — ¿Cómo activar la coordinación E2E? Integrarla como modalidad explícita de los flujos existentes, con un procedimiento canónico; no añadir una nueva skill de coordinación. Su activación no exige que el usuario se ausente.
- **D51 A** — ¿Hasta dónde llega cada responsable? Construye, prueba, corrige y hace commits locales en la rama y worktree asignados; el principal asigna verificadores independientes, publica e integra.
- **D52 aprobada** — ¿Se observará el comportamiento real? Sí: un único piloto después de construir el candidato, con techo total de 20 minutos, dos unidades independientes, una corrección solicitada por un verificador y una sustitución controlada de responsable. El usuario permite Codex o Claude y después indica considerar instalar las skills y ejecutar los hosts ya instalados. Se prioriza Claude Code instalado con entrega nativa del candidato; la instalación temporal compartida quedó aprobada después en D55 A, tras exponer su alcance. No se autorizan varios intentos ni se levanta la pausa general de pilotos.
- **D53 A** — ¿Cómo entregar? Automatic: revisión, push, PR y merge a development, manteniendo los bloqueos por criterios, CI o aprobación humana pendiente.
- **D54 A** — ¿Qué revisión de código? Nativa de Codex `/review`, antes de publicar y de instalar el candidato de prueba.
- **D55 A** — ¿Instalar temporalmente el candidato? Sí: vía Hive en la instalación compartida, ejecutar Claude Code ya instalado y restaurar la revisión anterior preservando el trial y cambios ajenos. El usuario fue informado de que otros hosts/sesiones pueden acceder temporalmente al contenido compartido.

Las decisiones describen el comportamiento futuro del producto. El usuario inició la implementación con `$flow-build` el 2026-11-17; los hijos de esta construcción conservan sus permisos actuales.

## Business rules

rule: La coordinación E2E se activa explícitamente dentro de los flujos existentes y continúa el objetivo dentro de los límites acordados sin exigir la ausencia del usuario. · screen none · approved 2026-11-17: D50 A · found by plan
rule: "A child does not commit" en content/skills/flow-build/references/delegation.md se aplica sin excepción a los hijos de implementación. · screen none · cancelled 2026-11-17: D51 A introduce una excepción solo para el responsable local E2E · replaced by Los responsables locales E2E pueden hacer commits únicamente en su rama y worktree asignados; los demás hijos conservan sus límites actuales. · found by plan
rule: Los responsables locales E2E pueden hacer commits únicamente en su rama y worktree asignados; los demás hijos conservan sus límites actuales. · screen none · approved 2026-11-17: D51 A · found by plan
rule: El coordinador conserva la publicación, los PRs, la integración y la asignación de verificadores independientes; el responsable local no adquiere autoridad remota ni lanza hijos. · screen none · approved 2026-11-17: D51 A · found by plan

## Request

> El chiste es no hacer microgestión sobre un solo agente

El usuario eligió reformular Fase 2 alrededor de coordinación E2E y después invocó `$flow-plan`. El alcance propuesto incluye H1 (arnés y mapa de features), H2 (corrección determinística) y H3 (revisión de supresiones). H4 (verificación formal) y H5 (fuzzing de UI) permanecen fuera de esta fase. No hay ticket ni criterios de un tracker.

## Objective

Después de acordar objetivo, criterios, límites y entrega, el coordinador selecciona las capacidades necesarias, asigna unidades entregables independientes, comprueba la evidencia y conduce la entrega autorizada. El usuario decide cambios de alcance o efectos no cubiertos; no administra la secuencia de herramientas ni reconstruye el trabajo de cada hijo.

La implementación actual ya encadena build, verify y entrega. Esta fase añade un contrato coherente de responsabilidades, evidencia y recuperación; reutiliza ese recorrido.

## Scope and acceptance

Dentro del diseño:
- Modalidad E2E explícita y procedimiento canónico en los flujos actuales, compatible con el uso ordinario y distinguido de unattended-delegation.
- Responsables locales aislados con commits acotados y coordinación de verificadores por el principal.
- Registro de ejecución recuperable de varias unidades, escrito solo por el principal y enlazado desde el checkpoint existente.
- H1: contrato de verificación por proyecto, mantenido desde comandos existentes, con Launch, Doctor, Drive, Evidence, Cleanup y mapa de features.
- H2: regla, mecanismo que la hace cumplir y caso negativo; arquitectura, tipos, lint y tests antes de ampliar prosa.
- H3: revisión del diff que debilita comprobaciones; las supresiones son señales que examinar, no defectos automáticos.
- Pruebas determinísticas, medición separada de prosa/scripts/referencias y documentación coherente.

Fuera: nuevo runtime de agentes, scheduler, motor universal de shell, publicación por los responsables, delegación anidada, entrada automática desde servicios externos, operación persistente garantizada entre sesiones, H4, H5 y equivalencia de comportamiento entre todos los hosts.

Los criterios siguientes distinguen contratos ejecutables de comportamiento observado. Todos se contrastaron con la base el 2026-11-17; la falsedad indicada describe la ausencia actual, no una ejecución de modelos sobre la base.

- AC1 — La modalidad E2E activa las etapas necesarias para un objetivo autorizado y conduce dos unidades independientes sin pedir al humano qué herramienta o etapa ejecutar. Falso hoy: no existe el contrato E2E con responsables locales; el contrato general prohíbe sus commits. Observed in: traza y resultado del único piloto nativo autorizado sobre el candidato.
- AC2 — Un responsable local entrega commits solamente desde su rama/worktree y rutas asignados; el principal conserva las operaciones remotas. Falso hoy: no existe el rol con esta excepción y las pruebas de fronteras prohíben todo commit de un implementador. Observed in: rol distribuido, comprobación de handoff sobre repositorios Git locales reales y commits/traza del piloto.
- AC3 — El registro multiunidad rechaza referencias inválidas, ciclos, asignaciones activas incompatibles y actualizaciones obsoletas sin perder el estado anterior. Falso hoy: tasks.py conserva un checkpoint de un solo worktree sin este registro ni sus operaciones. Observed in: CLI de estado ejecutada contra archivos y repositorios locales de prueba.
- AC4 — Un handoff de otra rama/head, con cambios fuera de alcance o sin evidencia independiente de la revisión exacta no puede marcarse como verificado mediante el helper. Falso hoy: el comando done valida sintaxis del commit y texto, sin contrato de handoff. Observed in: rechazos de la CLI sobre Git local y transición aceptada con evidencias inspeccionadas en el piloto.
- AC5 — El coordinador reasigna una unidad parcialmente completada solo después de comprobar el cierre del anterior responsable, preserva su trabajo y continúa desde el checkpoint sin reconstrucción humana. Falso hoy: falta el protocolo y el registro de sustitución multiunidad. Observed in: IDs, cierre/ack del primer responsable, checkpoint y commits del sustituto en el piloto nativo.
- AC6 — H1 genera sin sobrescribir un contrato de proyecto con Launch/Doctor/Drive/Evidence/Cleanup y mapa de features; su validador rechaza referencias inválidas y nunca presenta capacidades no comprobadas como funcionales. Falso hoy: no existe ese generador/contrato/validador. Observed in: CLI de contrato sobre el proyecto local desechable y sus comandos reales durante el piloto.
- AC7 — La corrección de un fallo repetible deja una tabla acotada regla → mecanismo → caso negativo, y sustituye la instrucción procedural que ese mecanismo cubre. Falso hoy: la auditoría recomienda automatizar pero el flujo de corrección no tiene este contrato completo de selección y evidencia. Observed in: referencia distribuida de corrección y comprobación negativa del defecto controlado en el piloto.
- AC8 — La revisión de un cambio que introduce una supresión o debilita un test exige demostrar su efecto antes de aceptarlo o rechazarlo; una excepción legítima documentada no se clasifica automáticamente como defecto. **Retired:** la base ya exige revisar debilitamientos y demostrar escenarios; esta formulación no distingue suficientemente el comportamiento nuevo. Sustituida por AC11. Observed in: contrato distribuido de revisión, diff y veredicto fundamentado del caso controlado del piloto.
- AC9 — La medición informa por separado prosa, scripts y referencias, añade el escenario E2E con sus lecturas obligatorias y conserva una comparación antes/después con igual conjunto de fuentes; cada skill/referencia sigue bajo 16,000 B. Falso hoy: reading_footprint_test.go solo mide escenarios anteriores y no el recorrido ni el código añadido. Observed in: salida de las pruebas de contenido y tabla de medición del candidato.
- AC10 — El piloto registra exactamente qué candidato y roles se leyeron, modelos observados o configurados según el host, intervenciones humanas y estado terminal; detiene sus procesos dentro del techo y preserva configuración, autenticación e historias ajenas. Falso hoy: no existe ese escenario acotado y el timeout del runner comienza después del setup. Observed in: manifiesto, trazas, comprobación de procesos y readback de instalación antes/después del piloto.
- AC11 — Con cualquier mecanismo de revisión elegido, flow-verify exige enumerar supresiones/skips/exclusiones añadidos o ampliados en el diff y obtener un veredicto fundamentado por cada candidato; una excepción legítima no se rechaza por su forma textual. **Replaces:** AC8. Falso hoy: ni flow-verify ni su contrato de entrega a revisiones nativas exige ese inventario explícito y su resolución. Observed in: contrato y brief entregado a revisión nativa o por rol, inventario del diff y veredictos del piloto.
- AC12 — El coordinador integra localmente los heads verificados de ambas unidades en un destino local declarado, lee de vuelta ese resultado y ejecuta la comprobación conjunta antes de registrar integrated. Falso hoy: no existe el protocolo de integración de los nuevos responsables ni su registro. Observed in: Git local de la fixture, head de integración y salida de la comprobación conjunta del piloto; no implica publicación remota.

## Prerequisites

No hacen falta cuentas externas para los contratos ni sus pruebas determinísticas.
- [x] Instalación temporal del candidato con restauración de la revisión previa y preservación del trial: aprobada D55 A · needed by: T8.
- [ ] Disponibilidad de los controles nativos para delegar y confirmar el cierre de un responsable con el candidato instalado: el principal la comprueba antes de lanzar modelos; no se presupone por el catálogo · needed by: T8.

## Verified context

- Base comprobada el 2026-11-17: `git -C /Users/jmartinez/Development/projects/tricell/tricell-hive rev-parse HEAD` devuelve `15b2e402c72543b74d63e50df1bc2b188638212a`.
- `content/skills/flow-build/SKILL.md`, paso 5: continuación en el mismo turno a flow-verify.
- `content/skills/flow-build/references/delegation.md`: los hijos actuales no hacen commits y usan normalmente el checkout principal; los cambios en paralelo exigen interfaces fijas y archivos disjuntos.
- `content/agents/development/hive-build-backend.md` y `tests/content/agent_boundary_test.go`: la prohibición de commits forma parte del contrato de los implementadores y de su comprobación de contenido.
- `content/skills/flow-plan/scripts/tasks.py`: checkpoint Markdown con un worktree, escritura atómica y evidencia textual; no es un registro multiunidad ni prueba de ejecución.
- `content/skills/git-workflow/scripts/git_state.py`: el commit acotado acepta rutas explícitas, escanea secretos y preserva staging ajeno; reutilizarlo.
- `content/skills/flow-close/scripts/cleanup_branches.py`: admite ramas activas protegidas; el coordinador alimentará esa protección desde las unidades abiertas.
- `tests/content/reading_footprint_test.go`: unión de fuentes declaradas, no lecturas transitivas ni tokens observados. Medición actual: 130,562 B; global 52,219 B. Headroom: flow-build 735 B, git-workflow 131 B, con límite individual de 16,000 B.
- Las specs relevantes son workflow-lifecycle y context-delivery. Las descripciones históricas de lectura en instruction-resources y la spec contienen referencias antiguas; el plan reconciliará las afirmaciones que toque con el test actual.
- `claude --version`: 2.1.326. `codex --version`: 0.169.0. Ambas son lecturas locales sin modelos.
- `claude --setting-sources project auth status --json`, filtrado a estado sin identidad ni valores sensibles: configuración habitual autenticada; `CLAUDE_CONFIG_DIR` recién creado no autenticado. No se ejecutó un modelo ni se modificó la autenticación. No elegir config limpia como si preservara el acceso actual.
- Docs oficiales mediante Context7: https://code.claude.com/docs/en/debug-your-config y https://code.claude.com/docs/en/sub-agents. Establecen mecanismos documentados, no pruebas del comportamiento local del candidato.
- `hive update --help` y deployment-manager.md: update desde commit actualiza todos los hosts registrados; las skills de usuario son compartidas. `plan install --release` permite planear una restauración de un snapshot retenido. Aplicar exige fingerprints vigentes.
- El trial cambió durante esta planificación por trabajo ajeno: `refs/hive/trial-join` y `feat/question-card-alone` se releerán inmediatamente antes de preparar la instalación. No restaurar el hash antiguo de esta sesión sobre cambios posteriores.

## Effects

- Piloto nativo de modelos · target: un único run local de Claude Code instalado (Codex sigue siendo alternativa autorizada, no segundo intento) · ceiling: una ejecución, máximo total 20 minutos, cuota existente sin nuevas compras · tasks: T8 · answer: approved D52.
- Instalación temporal de contenido en la instalación compartida de Hive y restauración verificada · target: hosts registrados del usuario en esta máquina, con trial preservado · ceiling: cero gasto adicional de infraestructura · tasks: T8 · answer: approved D55 A, una instalación candidata y su retorno; no incluye reemplazo de autenticación ni preferencias ajenas.

El piloto no publica, envía mensajes a terceros, cambia datos compartidos ni despliega aplicaciones. La entrega Git del cambio es distinta de esa fixture.

## Delivery

Automatic (D53 A), con revisión nativa Codex `/review` (D54 A), sin parada rutinaria previa al push y con integración en development solo cuando los criterios y revisiones lo permitan. Se respeta cualquier aprobación humana requerida. La instalación temporal D55 A se realiza después de construir y revisar el candidato. Pedir el plan y responder esta ronda no inicia la construcción. D52 no se ejecuta durante planificación. No hay promoción a master ni release de producto en este alcance.

## Review status and progress

2026-11-17 — D50 A y D51 A registradas. Investigación delegada de scripts y pruebas recibida; contrastadas en el principal las restricciones de roles, el checkpoint y el helper de commits. No se ha solicitado la revisión del plan: faltan criterios y tareas tras decidir la evidencia conductual.
2026-11-17 — D52 autorizada; el usuario añadió la ruta de instalación nativa. Se incorporan AC1–AC10, tareas y deltas. Se descarta la configuración limpia de Claude tras comprobar que pierde autenticación. La ruta nativa evita ampliar el runner con aislamiento de configuración para Claude.
2026-11-17 — D53 A, D54 A y D55 A aprobadas mediante «ok». Revisión independiente: ajustar launcher del caso, dar hogar H3 independiente del mecanismo de revisión, integrar localmente ambas unidades y precisar el criterio H3. Revisión R1 del plan: AC8 retirado, AC11 lo sustituye y AC12 añade la observación de integración local; tareas/diseño/deltas actualizados. No se invalidó implementación: ninguna había empezado.

2026-11-17 — Revisión R1 independiente completada sin hallazgos materiales residuales. check_change.py pasa. Plan listo para flow-build; no hay implementación ni piloto ejecutado.
