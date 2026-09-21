# Mapa general de flujos de trabajo de Hive

Fecha de revisión: 2026-09-20. Estado: inventario documental del Hive anterior y propuestas para el rebuild; no es una implementación ni reactiva sus instrucciones.

## Propósito y alcance

Este documento responde qué trabajos cubría Hive, cómo se recorrían y qué podemos aprender de los repositorios de referencia. Un flujo se define por su objetivo, entradas, actividades y evidencia de cierre. Una skill, un agente o un comando puede participar en varios flujos; no equivale por sí solo a un flujo de trabajo.

La fuente principal es el clon local de `master` en `/path/to/reference-volume/dev-resources/tricell-hive-master`, commit `16e7d3357a3c41530d5e31460c3024872566f3c7`. El catálogo describe lo que sus fuentes prescriben, no garantiza que se ejecutara correctamente ni que siga instalado. Las referencias externas se revisaron en sus checkouts locales, sin actualizar remotos ni probar integraciones.

Las reglas del rebuild y la autorización del usuario siguen vigentes. Los requisitos de delegación, hooks, publicación y estructura del Hive anterior se estudian como material de referencia, no como instrucciones para esta sesión.

## 1. Las tres rutas de entrada

| Ruta documentada | Cuándo entra | Recorrido | Salida y límite |
|---|---|---|---|
| Investigación | Resolver una pregunta mediante código, documentos o fuentes externas | Delimitar pregunta → inspeccionar evidencia → contrastar → responder | Respuesta sustentada; archivo si se pide conservarlo. No implica implementación ni plan formal. |
| Trabajo directo | Obtener un resultado autorizado sin seleccionar el contrato formal de desarrollo | Entender entregable y efectos → explorar lo necesario → ejecutar → comprobar → cerrar | Resultado proporcional al trabajo. Puede tener un registro reanudable; no exige un plan formal por ser complejo. |
| Desarrollo formal | Petición explícita de `/flow-plan` y posteriormente `/flow-build`, o su equivalente | Explorar → definir → diseñar/planear → implementar → verificar/revisar → entregar/cerrar | Plan portable, permisos y evidencia de ejecución diferenciados. Planear, implementar y publicar son autorizaciones distintas. |

Fuentes: [flow-research][h-research], [task-routing][h-routing], [common intake y ruta directa][h-agents], [flow-plan][h-plan], [flow-build][h-build].

El trabajo directo también cubre análisis de negocio/BI, propuestas de arquitectura y operaciones de infraestructura. El tipo de entregable determina la evidencia necesaria; no convierte esos trabajos en desarrollo formal. Una consulta breve o un cambio mecánico no necesita toda esta estructura.

## 2. Catálogo de trabajos que cubría Hive

La agrupación siguiente es una síntesis del inventario, no una nueva lista de comandos obligatorios.

| Trabajo | Entrada habitual | Secuencia esencial | Entregable y criterio de cierre | Fuente de master |
|---|---|---|---|---|
| Investigar y comparar | Pregunta, afirmación o decisión pendiente | Precisar la pregunta, buscar evidencia, contrastar contradicciones y límites | Respuesta con fuentes; incertidumbre explícita cuando falta evidencia | `flow-research`; `adversarial-research` como variante explícita |
| Consultar estado y pendientes | Proyecto o repositorio identificado | Consultar Git, tracker declarado, PR y despliegues aplicables; reconciliar observaciones | Estado fechado y pendientes comprobados; consultar no decide qué ejecutar | `status-fetch` |
| Analizar datos o proponer arquitectura | Datos, necesidad de negocio o restricciones | Definir métricas/supuestos o alternativas, analizar y justificar | Cálculos reproducibles o propuesta con compromisos y decisiones abiertas | `agent-routing.md`, Common intake |
| Iniciar un proyecto | Idea, conversación, brief o código inicial | Refinar requisitos, resolver preguntas, establecer workspace y base técnica según lo que falte | Requisitos, convenciones, estructura y fundamento técnico; no equivale a desplegar | `bootstrap-playbook.md` |
| Especificar producto y revisar negocio | Requisitos o cambio de alcance | Crear/revisar épica y mapa de producto, revisar reglas y escenarios, derivar tareas después del gate de negocio | Especificación y criterios de aceptación; cambios respecto a lo ya construido identificados | `spec-writing-playbook.md` |
| Planear desarrollo | Cambio acotado que requiere contrato formal | Explorar archivos y contratos, resolver decisiones/prerrequisitos, escribir tareas y verificación, registrar aprobación | Plan con contrato, autorización y ejecución separados | `flow-plan` y `plan-format.md` |
| Implementar y reanudar | Petición directa o plan autorizado | Reconciliar estado real, ejecutar trabajo pendiente, conservar evidencia y resolver desvíos | Cambio acotado y comprobado; al reanudar no repetir trabajo ya aplicado | Ruta directa o `flow-build`, reconcile/execute |
| Diagnosticar y corregir | Fallo reproducible o comportamiento inesperado | Obtener evidencia, localizar causa, aplicar corrección autorizada, comprobar regresión | Causa sustentada y corrección verificada, o límite concreto del diagnóstico | `debugging.md`, `testing.md`, ruta directa/build |
| Revisar y verificar | Diff, cambio construido, spec o criterios | Seleccionar controles pertinentes, revisar, ejecutar comprobaciones y registrar resultados | Hallazgos o evidencia observable; una revisión aprobada no concede publicación | `flow-build/references/verify-gate.md`; matriz de `agent-routing.md` |
| Auditar deuda y riesgos | Código o arquitectura existentes | Revisar por lentes, validar hallazgos, clasificar severidad y priorizar | Hallazgos con evidencia y acciones propuestas; auditoría no autoriza arreglar todo | `audit-playbook.md` |
| Entregar y promover | Cambio verificado y destino autorizado | Reconciliar permisos/estado Git, cumplir controles, publicar o promover, comprobar entorno | Estado terminal observado, salud/smoke y validación funcional aplicable; notas de entrega | `flow-build` CLOSE, `git-mechanics.md`, `promotion-playbook.md` |
| Migrar y mantener el workspace | Proyecto anterior al modelo o documentación desordenada | Inventariar, clasificar, proponer movimientos/reparaciones, aplicar lo autorizado | Estructura reconciliada sin perder decisiones, referencias ni evidencias necesarias | `migration-playbook.md`, `workspace-hygiene-playbook.md`, `workspace-archive` |
| Consolidar repositorios | Migración explícita a monorepo | Inspeccionar dependencias y contratos, planear el corte, ejecutar y verificar consumidores | Corte acotado con evidencia de compatibilidad; no se confunde con mover documentos | `monorepo-cutover` |
| Comunicar y documentar | Resultados, conocimiento o entregable documental | Organizar contenido, elegir formato, generar y revisar legibilidad/enlaces | Documento, reporte o sitio verificable; crearlo no autoriza distribuirlo | `flow-report`, `starlight-docs-site` |

Los playbooks se encuentran en `global/skills/flow-core/references/`; las reglas citadas por nombre, en `global/rules-situational/`. El [árbol de skills][h-skills] permite localizar las demás entradas.

## 3. Cómo se conectan

No hay una única cadena obligatoria para toda petición. Estos recorridos ilustran las conexiones documentadas:

- **Proyecto nuevo:** necesidad → requisitos → fundamento del proyecto → especificación y revisión de negocio → desarrollo → verificación → entrega autorizada.
- **Cambio sobre producto existente:** petición → exploración → ruta directa o plan formal explícito → implementación → verificación → entrega dentro del alcance autorizado.
- **Pregunta o decisión:** pregunta → investigación → respuesta/propuesta. Se detiene ahí; implementar requiere otra instrucción que lo cubra.
- **Auditoría:** inventario → hallazgos validados → priorización → selección de correcciones. Solo las correcciones elegidas entran al flujo de desarrollo.
- **Incidente o bug:** evidencia → diagnóstico → corrección autorizada → comprobación; promoción al entorno si forma parte del alcance concedido.
- **Retomar trabajo:** leer registro/handoff → contrastar con estado actual → identificar pendiente real → continuar por la ruta correspondiente.

La revisión de negocio valida qué debe hacer el producto. La revisión técnica y las pruebas validan el cambio construido. La comprobación posterior al despliegue valida el entorno entregado. Son evidencias diferentes; ninguna sustituye automáticamente a las otras.

## 4. Capacidades transversales, no flujos nuevos

| Capacidad | Función en los recorridos |
|---|---|
| Intake y control de alcance | Identificar entregable, exclusiones, efectos autorizados, evidencia de cierre y punto de parada. |
| Memoria y continuidad | Recuperar decisiones y contexto; contrastarlos con fuentes actuales. `memory-policy`, `memory-sync` y `engram-init-workspace` cubrían partes de este ámbito. |
| Organización de artefactos | Separar código, documentos durables, sesiones, evidencia y temporales. Las rutas históricas de master no sustituyen las convenciones actuales del rebuild. |
| Delegación y revisión independiente | Distribuir trabajo cuando aporte valor y el host lo permita. No es una nueva fase de producto ni justifica instalar otro runtime. |
| Reglas técnicas | Orientar pruebas, debugging, seguridad y stack según el cambio. `language-rules` era una superficie de acceso, no un flujo completo. |
| Reportes y handoff | Comunicar resultados y dejar información suficiente para continuar. `flow-core` era una biblioteca compartida, no un workflow ejecutable. |
| Autonomía acotada | `unattended-delegation` trataba misiones delegadas con límites; no convertía toda tarea en trabajo desatendido. |

## 5. Qué aportan los repositorios de referencia

Estas comparaciones se refieren a las versiones locales inspeccionadas. Sus afirmaciones de ahorro, calidad o compatibilidad no fueron reproducidas en esta revisión.

| Referencia | Patrón documentado | Qué vale la pena estudiar para Hive | Diferencia o límite |
|---|---|---|---|
| **optional reference project** | ODD para el trabajo cotidiano; SDD por elección explícita, con proposal/spec/design/tasks y fases adicionales | Mantener pequeño el trabajo pequeño; registro único para trabajo sustancial; separar investigación de implementación y preservar continuidad | Su SDD local permite archivado sin verification como gate: `/sdd-verify` es diagnóstico opcional y puede archivarse trabajo parcial explicitando pendientes. No equivale al gate `built → verified` de Hive formal. |
| **optional reference project** | Brainstorming → entorno aislado → plan → ejecución → pruebas/revisión → cierre de rama | Especificaciones ejecutables, debugging sistemático, revisión contra intención y evidencia antes del cierre | Su README presenta el recorrido como obligatorio y prescribe prácticas fuertes de TDD. No trasladar esa obligatoriedad ni sus acciones Git al rebuild sin decisión propia. |
| **optional reference project** | Reconocer → auditar → validar hallazgos → priorizar → escribir planes; ejecución delegada y reconciliación como opciones | Un plan como entregable completo, con archivos, contexto, comandos y criterios; volver a comprobar backlog antes de ejecutarlo | Es una referencia de auditoría/asesoría, no cubre por sí sola todo el ciclo de Hive. La ventaja de usar ejecutores baratos es una propuesta del proyecto, no un resultado medido aquí. |
| **optional reference project** | Entender el problema y buscar la solución suficiente: reutilización, biblioteca estándar, plataforma y dependencias existentes | Reducir código y complejidad innecesarios sin quitar validación, seguridad o accesibilidad | Es principalmente criterio de implementación/revisión, no sustituto de planificación, autorizaciones o verificación. Sus benchmarks no prueban mejora para Hive. |

Fuentes: optional external research source, optional external research source, optional external research source, optional external research source, optional external research source.

La investigación del [corpus de harness engineering](2026-09-20-portable-harness-research.md), incluido el ZIP de Uber, aporta criterios de diseño y medición. No define por sí sola nuestros flujos ni demuestra que adoptar estas referencias mejoraría los resultados.

## 6. Propuesta para organizar el rebuild

**Propuesta, pendiente de decisión:** conservar este mapa de necesidades y elegir gradualmente qué procedimientos mínimos implementamos. No restaurar el catálogo antiguo entero ni combinar todos los frameworks.

Un primer corte útil sería:

1. **Investigar y decidir:** respuesta o propuesta con evidencia y límite claro de alcance.
2. **Cambiar y verificar:** una ruta directa proporcional; el plan formal queda disponible cuando se elige y aporta valor.
3. **Entregar y operar:** publicar/promover solo en el alcance autorizado, comprobar estado real y registrar pendientes.
4. **Conservar y retomar:** documentación y continuidad suficientes, sin acumular scratch como conocimiento durable.

Bootstrap, specs, auditoría, migración y documentación seguirían siendo procedimientos especializados que se seleccionan por necesidad. Este agrupamiento es una hipótesis de organización, no cuatro nuevas skills ni un compromiso de implementación.

El siguiente paso sería elegir un recorrido real y su criterio de éxito; después escribir la guía mínima y probar unos pocos casos comparables. La existencia de este mapa no valida aún la selección automática de procedimientos ni su funcionamiento entre hosts.

## 7. Dónde conservar las referencias

Ubicación actual: `/path/to/reference-volume/dev-resources/reference/`. El usuario autorizó mover las cuatro referencias y confirmó que el volumen es prácticamente fijo en esta Mac. Se trasladaron los checkouts completos, incluidos `.git`, archivos locales y el marcador `.jbcontextignore` del directorio padre. Los cuatro árboles de trabajo quedaron limpios y conservaron sus commits. No se actualizaron los remotos.

| Repo | Commit local inspeccionado | Tamaño aproximado del checkout |
|---|---|---|
| optional reference project | `95edf9ff9172ca82f18ef34ccca2776b15348bb1` | 106 MB |
| optional reference project | `5bf4e78011075bcfc0dc295f0724994cd123ee71` | 8.5 MB |
| optional reference project | `cac56e1ebd3c279aa9153616cfeac7b174ab90f9` | 284 KB |
| optional reference project | `e3ba2aa6f1e6f0bc4d69eb09c9f0d0a93af56156` | 3.5 MB |

**Decisión aplicada:** referencias externas juntas en el volumen de recursos; Hive en desarrollo permanece en su workspace habitual y el clon de `master` sigue en `dev-resources/tricell-hive-master`. Se retiró el directorio original después de verificar 3,505 entradas de archivos/enlaces por contenido, permisos de archivo y destino de enlaces; los archivos se compararon con SHA-256. No se dejó una segunda copia ni un enlace de compatibilidad.

Antes del traslado se comprobó que cada repositorio tenía únicamente su worktree principal y no contenía enlaces simbólicos absolutos. La búsqueda de la ruta anterior en el checkout activo de Hive y en las configuraciones principales de Codex/Claude solo encontró este documento; no fue una auditoría de todos los consumidores de la Mac. Los enlaces de procedencia fijados a commits no cambian. Si el volumen no está montado, la consulta local queda temporalmente indisponible.

## 8. Fuentes y límites de esta revisión

- Hive: lectura de las entradas de investigación, routing, planificación, construcción y biblioteca común; procedimientos de bootstrap, especificación, promoción e higiene; inventario de skills y estructura de auditoría/migración.
- Referencias: README y documentación de uso pertinente de los cuatro repositorios. Se contrastó en particular la separación ODD/SDD y la política local de verificación/archivado de optional reference project.
- No se ejecutaron estos workflows, no se probaron integraciones, no se validaron benchmarks y no se auditó todo el código de los proyectos externos.
- Los enlaces siguientes están fijados a commits inspeccionados, no a ramas móviles. Se incluyen para trazabilidad; la lectura realizada fue local.

[h-research]: https://github.com/JhonHawk/tricell-hive/blob/16e7d3357a3c41530d5e31460c3024872566f3c7/global/skills/flow-research/SKILL.md
[h-routing]: https://github.com/JhonHawk/tricell-hive/blob/16e7d3357a3c41530d5e31460c3024872566f3c7/global/skills/task-routing/SKILL.md
[h-agents]: https://github.com/JhonHawk/tricell-hive/blob/16e7d3357a3c41530d5e31460c3024872566f3c7/global/rules-situational/agent-routing.md
[h-plan]: https://github.com/JhonHawk/tricell-hive/blob/16e7d3357a3c41530d5e31460c3024872566f3c7/global/skills/flow-plan/SKILL.md
[h-build]: https://github.com/JhonHawk/tricell-hive/blob/16e7d3357a3c41530d5e31460c3024872566f3c7/global/skills/flow-build/SKILL.md
[h-skills]: https://github.com/JhonHawk/tricell-hive/tree/16e7d3357a3c41530d5e31460c3024872566f3c7/global/skills





