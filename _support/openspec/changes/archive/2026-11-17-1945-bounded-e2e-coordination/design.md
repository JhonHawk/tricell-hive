# Design

## Approaches

**D50 A, elegido:** integrar la coordinación E2E en los flujos existentes mediante un procedimiento canónico y condiciones de lectura explícitas. Reutiliza la continuidad build → verify → entrega; evita otra entrada y reglas duplicadas.

**D50 B, descartado:** añadir una skill de coordinación que invoque los flujos. Haría visible una entrada dedicada, pero añadiría una interfaz y fronteras con research, build y unattended-delegation que mantener.

## Decisions

- **D51 A, elegido:** responsable local con worktree/rama asignados y commits locales; principal dueño de publicación, verificación independiente e integración.
- **D51 B, descartado:** responsable dueño de un PR completo con operaciones remotas; amplía la superficie de concurrencia y autoridad más allá de esta fase.
- **D51 C, descartado:** conservar todos los commits en el principal; cambia menos contratos pero concentra más trabajo de cada unidad en él.

## Proposed shape

El siguiente contrato define el candidato que el plan propone construir; no está implementado. D50 y D51 están decididas, y la aprobación final del plan acepta los detalles de contrato siguientes.

### Procedimiento y roles

La modalidad E2E necesita objetivo, criterios, destino de entrega y límites explícitos. No se activa por duración ni por interpretar una petición de análisis como implementación. Separa el contrato de coordinación de las excepciones que unattended-delegation aplica cuando el usuario está ausente.

Un nuevo rol específico de responsable local permite la excepción D51 A sin ampliar indiscriminadamente todos los implementadores. Usa capacidades nativas disponibles; la ausencia de escritura, aislamiento o lanzamiento se informa como limitación, nunca como capacidad comprobada por instalar un archivo de rol. El principal crea y asigna los worktrees dentro de la autorización del trabajo. Los responsables no crean otros, no publican y no lanzan hijos.

El principal asigna archivos y recursos disjuntos, interfaces fijas y evidencia requerida. Cada responsable construye, comprueba, corrige y entrega su head y las limitaciones. El principal lanza verificadores independientes, reconcilia las evidencias con la revisión exacta e integra. Los responsables detienen sus propios procesos antes de devolver el trabajo; la verificación arranca los recursos que necesite con asignaciones propias.

### Registro de ejecución

Un contrato versionado representa unidades, tareas/criterios asociados, dependencias, propietario, rama/worktree, base, rutas permitidas, head devuelto, evidencias, estado y siguiente acción. El principal es su único escritor. `tasks.md` sigue siendo la entrada humana y enlaza el registro.

Reutilizar `tasks.py` para validación/actualización atómica, con rechazo de actualizaciones obsoletas. No duplicar Git: `git_state.py` observa repositorio/head y hace commits locales acotados. La validación estructural nunca afirma que un criterio pasó por el simple hecho de contener un campo de evidencia.

Los IDs nativos de agentes son pistas de recuperación, no prueba de que sigan vivos. Ante una interrupción, el principal inspecciona el estado real y confirma que no queda otro escritor antes de reasignar. Una nueva revisión invalida las evidencias cuyo alcance no siga siendo válido. La recuperación no arranca un agente solo porque un timeout o un ID ausente parezca indicar que murió.

La ubicación concreta debe distinguir datos portables del plan de rutas/IDs efímeros. Los datos locales de ejecución van en el workspace ignorado de este trabajo, conforme al layout vigente; no se versionan rutas personales ni trazas crudas. El checkpoint conserva el puntero.

Formato v1, en `_support/workspace/<fecha>-<work>/execution.json`: `schema_version`, `generation`, `plan_ref`, `plan_revision`, `grant_ref`, `integration_target`, `units`. Cada unidad tiene `id`, `tasks`, `criteria`, `depends_on`, `assignment_id`, `owner_id`, `role`, `repo`, `worktree`, `branch`, `base_sha`, `allowed_paths`, `resources`, `state`, `returned_head`, `evidence`, `verifier`, `limitations`, `next_action` y `cleanup`. Los campos que todavía no proceden son nulos o listas vacías, nunca una aprobación implícita. `grant_ref` apunta a las decisiones vigentes, no concede autoridad por existir en JSON.

Estados: `assigned → running → returned → verified → integrated`. Desde running/returned/verified puede pasarse a `blocked`; una corrección vuelve a running e invalida el veredicto. La sustitución registra el cierre comprobado del assignment anterior y crea otro antes de volver a assigned. El mismo ID de unidad conserva su historia. `integrated` solo lo registra el principal después del readback de integración y la comprobación conjunta en el destino declarado. El registro distingue `integration_scope: local|remote` y `integration_head`; la publicación tiene su propio resultado, no se infiere de integrated. El piloto integra ambas unidades en una rama local de la fixture y termina en integrated local, sin publicación remota.

El principal usa validación de estructura y transiciones con `expected_generation`; el helper conserva el archivo anterior ante error y usa exclusión de escritura breve más reemplazo atómico para que dos llamadas del principal no pierdan actualizaciones. No hay locks con recuperación destructiva por edad. Las ramas/worktrees o rutas de escritura coincidentes de unidades activas se rechazan; los recursos declarados exclusivos tampoco pueden compartirse. Una generación o revisión del plan diferente obliga a reconciliar la asignación.

`validate-handoff` consulta el Git local sin red: identidad del repo/worktree/rama, head completo, descendencia desde la base, estado limpio y diff completo de rutas permitido. Los checks semánticos quedan con el principal y verificadores: un JSON que dice PASS no demuestra que se ejecutó una prueba. Un veredicto identifica head, criterio, comando, resultado y evidencia disponible, con verificador distinto al autor. Una transición a verified rechaza datos incompletos, evidencia ausente o head distinto; aceptar esos campos no reemplaza inspeccionar la evidencia.

### Contrato de verificación del proyecto

H1 será un índice versionado de comandos y recorridos reales del proyecto, con Launch, Doctor, Drive, Evidence y Cleanup, más IDs de features y sus rutas/recorridos. Los scripts de la aplicación permanecen en sus ubicaciones de código; el contrato reutiliza sus comandos. La ubicación canónica propuesta es `_support/docs/verification/project.json`, como conocimiento mantenido del proyecto.

El generador crea una estructura inicial sin sobrescribir archivos existentes y sin inventar comandos. El agente la completa desde el proyecto y mantiene las entradas afectadas por el cambio. Una capacidad sin comprobar queda marcada como tal; una no aplicable lleva razón. El contrato no ejecuta shell ni descubre ejecutables arbitrariamente.

Un validador Python con biblioteca estándar comprueba versión, campos, referencias, rutas confinadas y features. El flujo inspecciona requisitos y efectos y ejecuta los comandos normales del proyecto. Salidas, revisiones y códigos de salida pertenecen a la evidencia de cada ejecución; nunca se confunden con validación estructural.

Formato v1: `schema_version`, `targets[]` e IDs de features por target. Cada target declara `id`, `cwd`, `capabilities` y `features`. Las cinco capacidades tienen `status` (`unverified`, `verified`, `unsupported` o `not_applicable`), una razón para las dos últimas, y referencias a comandos/documentos existentes, requisitos/efectos y condición observable de éxito. Una entrada verified necesita evidencia con revisión y resultado; esto es metadato declarado que el flujo vuelve a contrastar antes de reutilizarlo. Cada feature identifica fuentes y referencias a drive/evidence. No se usa un número artificial mínimo de features: se cubren los recorridos afectados por el objetivo.

La CLI nueva `flow-verify/scripts/check_contract.py` ofrece `init` y `check`, con repo/path explícitos; no usa subprocess ni ejecuta los comandos que el contrato menciona. Init se niega a sobrescribir; check rechaza JSON duplicado, versión desconocida, referencias inexistentes y rutas que escapan por `..` o symlink. El flujo conserva contratos del proyecto ya equivalentes: los referencia o convierte de forma explícita, nunca crea dos fuentes activas.

### H2 y H3

H2 se integra en el procedimiento existente de corrección y auditoría: preferir un mecanismo determinístico y un caso que demuestre su rechazo; retirar la prosa que ese mecanismo sustituye. La tabla regla ↔ mecanismo ↔ comprobación es acotada al trabajo y tiene un solo hogar, no un catálogo universal duplicado.

H3 tiene un único hogar en `flow-verify/references/suppression-review.md`, requerido por flow-verify para cualquier mecanismo de revisión y enlazado por hive-review-code. Examina el diff de directivas de supresión, tests omitidos y configuración que debilita controles; pide inventario de candidatos y resolución fundamentada por cada uno, o que no hubo cambios candidatos. El brief de revisión nativa incluye ese requisito y su referencia, usando el mecanismo del host para instrucciones de revisión. Si no puede suministrarse o verificarse la cobertura, H3 queda no verificado; no se sustituye silenciosamente la revisión seleccionada. Un hallazgo necesita escenario y evidencia de su efecto; una coincidencia textual o un comentario explicativo no es un defecto por sí mismo. Mantener las excepciones legítimas de defect-fix.

## Impact

Guidance global; flujo de plan/build/verify/close y delegación; unattended-delegation; git-workflow; rol nuevo y pruebas de fronteras; scripts de estado/contrato y sus pruebas; specs y páginas que describen esas conductas. No afecta datos productivos ni añade infraestructura.

Las mediciones de lectura separan prosa de código y cuentan las referencias requeridas por escenarios explícitos. Comparar también los escenarios existentes con la misma lista de fuentes antes/después para no atribuir un cambio de metodología a un ahorro. Ningún resultado en bytes equivale a tokens o a comportamiento demostrado.

## Recovery

Los contratos tienen versión y diagnósticos explícitos; las operaciones existentes conservan sus interfaces. El cambio se entrega por commits reversibles. Una retirada de contenido no borra estados ni trabajo de responsables: los preserva para integración o recuperación manual. No se permite limpiar una rama/worktree activo o de propiedad incierta.

## Validation

D52 autoriza una única observación de comportamiento, después de construir el candidato, con 20 minutos como techo del intento completo. La sugerencia posterior del usuario prioriza instalar el contenido mediante Hive y ejecutar un host ya instalado. D55 A aprueba la instalación temporal compartida y su restauración. Se usa Claude Code nativo: la autenticación habitual responde y la config limpia no, por lo que no se añade un adaptador de autenticación ni se copian credenciales.

Antes de aplicar, preparar los planes de instalación y retorno con el gestor existente, leer el estado actual, preservar el trial vigente y rechazar drift. El candidato de prueba une el cambio con el trial sin publicar ni mover la rama trial. Nombrar todos los commits incorporados y confirmar que el conjunto coincide con la autorización. No escribir `refs/hive/trial-join` para un ensayo temporal. Al terminar, restaurar el release anterior con un plan fresco que respete fingerprints; si otra sesión modificó el contenido entretanto, detener el reemplazo y reconciliar, nunca restaurar ciegamente.

Usar un único caso nuevo en el runner existente, no una campaña, otro framework ni un motor de agentes. Mantener el host instalado y la autenticación habitual. T7 debe configurar explícitamente las herramientas nativas del caso: el allowlist actual de Claude en launch.go no enumera delegación/skills, y Codex usa --ephemeral salvo native-agent-dispatch. Confirmar nombres y controles en la versión instalada, añadir solo los permisos locales necesarios a este caso y probar los argumentos sin modelos. Si se optara por Codex antes del primer lanzamiento, aplicar la excepción de persistencia necesaria al nuevo caso y conservar su historia; no suponer que la excepción por otro nombre lo cubre. Si un control indispensable no puede verificarse por preflight sin modelos, el intento lo observa y puede quedar inconcluso; nunca se lanza otra sesión de sondeo que consuma un segundo intento.

La memoria del caso se aísla con las primitivas ya existentes (directorio nuevo vacío, sincronización desactivada, routing comprobado); las historias nativas permanecen preservadas. El reloj monotónico empieza antes del setup efectivo del ensayo, limita cada proceso por el tiempo restante y reserva los últimos dos minutos para parada y retorno. No iniciar llamadas nuevas cuando se consume esa reserva. Un retorno que detecta cambios ajenos preserva esos cambios y se informa como no completado, nunca se fuerza para cumplir el reloj.

Fixture local con dos unidades y archivos/recursos disjuntos. El prompt inicial nombra objetivo, criterios y límites, pero no dicta las skills ni cada etapa. Una unidad devuelve un candidato con un defecto controlado y documentado por la fixture; su verificador debe demostrarlo y pedir una corrección sin aceptar una supresión del test como arreglo. La fixture incluye también una supresión legítima justificada para comprobar que no existe prohibición indiscriminada. La inyección, si es necesaria, ocurre con el responsable detenido, queda registrada como mutación de la fixture y cambia el head antes de verificar; no se atribuye al modelo ni se usa evidencia anterior. La otra unidad devuelve trabajo parcial con cierre nativo confirmado, y el coordinador debe asignar un sustituto desde ese estado. Esto comprueba sustitución tras devolución parcial, no tolerancia a cualquier caída de VM.

El piloto mide selección/lectura de skills, dos responsables, commits locales, revisión independiente, corrección, sustitución e integración local conjunta sin instrucciones humanas adicionales. El principal integra los dos heads comprobados en una rama local declarada, verifica el head resultante y ejecuta la prueba que cruza ambas unidades. No hace push ni prueba merge remoto: esos comportamientos existentes se conservan y se verifican en la entrega real del cambio, sin afirmar que esa entrega ejercita responsables E2E. La lectura de una skill se prueba por invocación/Read y contenido, no por catálogo. Se registran modelo observado, hashes, comandos, heads, autor/verificador y resultado. Una única observación no califica modelos ni demuestra fiabilidad o portabilidad.

## Review status and progress

R1 (2026-11-17): concretar permisos/persistencia del launcher del caso; H3 canónico para cualquier revisión; integración local del piloto; AC11 sustituye AC8 y AC12 cubre el resultado combinado. Las decisiones D50–D55 se conservan; ninguna tarea estaba implementada.

No hay un segundo intento automático ni cambio de host después de lanzar el modelo. Un fallo se clasifica y puede corregirse dentro del trabajo autorizado, pero la repetición del piloto requiere una nueva autorización. AC sin evidencia conserva estado no verificado.
