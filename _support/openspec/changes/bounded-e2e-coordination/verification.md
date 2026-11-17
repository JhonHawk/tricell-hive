# Verificación de Fase 2

Estado al 2026-11-17: T1–T7 implementadas; T8 abierta. No hay push, PR ni integración de este cambio en development. Las respuestas D53–D55 siguen vigentes; el intento de D52 ya se consumió. D56 A autoriza un segundo intento acotado, cuya evaluación se registra separada.

## Candidato y comprobaciones

El piloto observó `b174a26a`, instalado temporalmente junto al trial existente. La revisión nativa de Codex encontró tres defectos: auto-verificación por un autor anterior, instrucciones contradictorias sobre el checkout del verificador y clasificación incorrecta de una modificación prohibida de la fixture. Se corrigieron con casos negativos; la revisión limitada de las correcciones no encontró regresiones concretas. Esa segunda revisión no cubre lo que la primera pudo omitir.

Pasaron `go vet ./...`, `go test ./...`, 172 pruebas Python y site check/lint/build. La suite race pasó todos los paquetes; tooling/cli necesitó ampliar el timeout predeterminado de 10 minutos y terminó en 638.094 s. Las correcciones posteriores se comprueban en sus paquetes afectados; se reutiliza la evidencia de paquetes sin cambios. Los logs y el checkpoint conservan los comandos y resultados.

La revisión nativa incluyó el inventario de supresiones/exclusiones del candidato. Aceptó las directivas de la fixture, sus rutas ignoradas, las excepciones acotadas del owner, los campos de progreso excluidos del fingerprint y las fronteras explícitas de medición. No identificó nuevos skips de tests ni exclusiones de CI/lint del repositorio. Esto no sustituye la revisión que debía realizar el coordinador dentro del piloto.

## Resultado de la primera observación nativa

Claude Code 2.1.326 usó Opus 5.5 como coordinador y verificador, y Sonnet 5.5 para los responsables. El ciclo completo de setup, ejecución y restauración duró aproximadamente 6 min 47 s, dentro del máximo de 20 min. No recibió instrucciones humanas adicionales de etapa.

| Criterio | Resultado | Alcance |
| --- | --- | --- |
| AC1 | Cumplido | Coordinó dos unidades, sustitución, verificación e integración; no acredita lectura de todas las skills. |
| AC2 | Cumplido | Responsables nativos con commits locales limitados a sus rutas. |
| AC3 | Cumplido | Rechazos determinísticos de referencias, conflictos y actualizaciones obsoletas. |
| AC4 | Cumplido | Rechazos Git/evidencia y verificación independiente de los heads reales. |
| AC5 | Cumplido | Sustitución después del cierre nativo, conservando head/checkpoint. El sustituto completó validaciones, sin necesitar otro commit. |
| AC6 | No cumplido | El generador y validador pasan sus tests; el piloto no produjo el contrato H1. |
| AC7 | No cumplido | No produjo la tabla H2 ni la disposición de prosa sustituida; la inyección no activó el ciclo verificador→corrección. |
| AC9 | Cumplido | Medición comparable y límites de 16,000 B respetados. |
| AC10 | Evidencia incompleta | Modelo, candidato, plazo y restauración corroborados; faltan lecturas completas de skills, ruta MCP de memoria y parte de la instrumentación. |
| AC11 | No cumplido | Los verificadores del piloto no recibieron ni produjeron el inventario H3 exigido. |
| AC12 | Cumplido | Ambos heads integrados en el worktree declarado y prueba conjunta anterior al estado integrated. |

El informe automático original fue inconclusive porque buscó fixture-integration en el checkout raíz, que seguía en main. La rama sí estaba en otro worktree de la misma fixture: Git, eventos nativos y una repetición independiente de acceptance confirman la integración. Se conserva el informe original; no se presenta una comprobación del revisor como si la hubiera ejecutado el driver.

La inyección A esperaba una respuesta síncrona completed del tool Agent. Los responsables se ejecutaron en segundo plano y sus cierres llegaron como notificaciones nativas. La mutación nunca ocurrió; esto no demuestra que el modelo se negara a corregir. Sí se observó que leyó las referencias de coordinación y registro, sin invocar ni leer las SKILL.md completas de los flujos. El encargo inicial tampoco enumeraba los entregables H1/H2/H3; su ausencia es real, pero esa omisión limita lo que el experimento permite concluir sobre su adopción.

La revisión independiente con anchors sobre evidencia cruda clasifica la aceptación como fallida/incompleta. Una sola observación no prueba fiabilidad, portabilidad o superioridad de un modelo.

## Restauración y conservación

Se restauró el release previo en los siete hosts registrados con un plan nuevo y fingerprints vigentes. Las 312 entradas comprobadas coinciden con sus hashes anteriores; el trial join quedó intacto. No hubo cambio de autenticación ni de preferencias detectado. La memoria aislada fue retirada; la búsqueda exacta del proyecto, sesión y cinco IDs de hijos en la memoria cotidiana devolvió cero sesiones, observaciones y prompts. No se ejerció una lectura MCP de memoria, por lo que esa ruta permanece sin verificar.

La fixture, sus ramas/worktrees y las trazas se conservan como evidencia única. Las identidades, rutas absolutas, hashes y resultados crudos están únicamente en `_support/workspace/2026-11-17-bounded-e2e-coordination/`; el informe independiente está en `native-run/independent-review.md` y el resultado validado en `native-run/reviewed-e2e-evidence.json`.

## Correcciones posteriores y siguiente verificación

Se corrigió el driver para resolver la rama declarada mediante Git worktree list y comprobar su repositorio, confinamiento, head y limpieza. La captura de cierres asíncronos relaciona la asignación con el evento nativo de cierre y entrega el aviso al padre, sin reactivar al hijo. El encargo ahora enumera los entregables H1/H2/H3 y la validez del plan como resultados esperados, sin dictar skills o pasos. No se añadieron reglas distribuidas a raíz de una sola observación.

Estas correcciones están en `3c1f8e87`; el paquete completo tests/pilot pasó con race en 125.819 s y vet pasó. Tienen regresiones locales con Git/procesos y entradas de hooks sintéticas; no equivalen a otra observación nativa y son posteriores a la revisión nativa limitada. Se identifican como Unreviewed en su commit. T8 solo podrá cerrarse con evidencia adicional o con una decisión explícita que cambie sus criterios; ninguna se presupone.

## Tamaño comparable

Contra development `15b2e402`: Markdown distribuido 558,726 → 592,573 B (**+33,847 B**); Python distribuido 194,684 → 232,423 B (**+37,739 B**). Global: **+454 B**. La unión de lectura E2E declara **224,188 B**, con 65,476 B de referencias, y mide por separado 105,283 B de helpers ejecutables. Excluir scripts no convierte esta fase en una reducción de prosa: incorpora capacidades nuevas. No son tokens observados.

El episodio de lectura/adopción queda registrado como watch en Engram `watch/e2e-skill-contract-adoption`; este documento no lleva otro conteo de recurrencias.

## Segunda observación autorizada (D56 A)

El segundo intento observó `d0e328d9` (código del driver `3c1f8e87`) junto al mismo trial. Setup y restauración completos: **371.585 segundos (6 min 12 s)**. Claude Code 2.1.326 sirvió `claude-opus-5-5` al coordinador; la traza conserva las identidades y modelos de los responsables y verificadores. No recibió comandos humanos adicionales. La inyección controlada sí ocurrió esta vez; el driver encontró el worktree real de integración y ejecutó la comprobación conjunta con exit 0. El resultado automático permanece `inconclusive` hasta la revisión semántica independiente.

La instalación previa se restauró mediante un plan nuevo: siete hosts, 312 entradas sin diferencias, trial join intacto. El proceso del driver terminó, el almacén aislado fue retirado y la búsqueda de identificadores exactos de fixture/sesión/agentes, también en campos de contenido de Engram cotidiano, encontró cero registros. La ruta de lectura MCP de memoria no fue ejercida. La telemetría informó USD 3.4975706 a precio de lista; no acredita un cargo real.

La evidencia se conserva separada en `attempt-2/native-run/` bajo la carpeta local de este trabajo. El primer intento y su guard permanecen intactos. D56 está consumida; no hay autorización para un tercer piloto. La revisión independiente con hashes de evidencia se conserva en `attempt-2/native-run/independent-review.md`; `pilot e2e review` validó los anchors y clasificó el run como `failed` (exit 1), sin modificar el informe original.

### Dictamen del segundo intento

| Criterio | Resultado | Alcance |
| --- | --- | --- |
| AC1, AC2, AC3, AC4, AC5, AC9, AC12 | Cumplidos | Coordinación, límites locales, rechazos determinísticos, verificación de heads, reemplazo cerrado e integración conjunta. |
| AC6 | No cumplido | Escribió un contrato Markdown; no generó el formato v1 ni ejecutó init/check. Launch además llamó build a go vet; no acredita arranque. |
| AC7 | Cumplido | Tabla regla/enforcer/caso negativo con falla real y disposición explícita de prosa previa inexistente. |
| AC10 | Cumplido con límites declarados | Registra candidato, modelos, roles, fuentes efectivamente leídas, terminal y restauración. No equivale a que se leyeran todas las skills; la ruta MCP no fue ejercida. |
| AC11 | No cumplido | Hay inventario y disposiciones, pero no se entregó al mecanismo independiente el brief H3 completo con su referencia. |

El recorrido específico de T8 verificador → solicitud → misma persona responsable → nuevo veredicto no se ejercitó: el coordinador detectó el fallo y pidió la corrección antes de asignar al verificador. La prueba negativa posterior del verificador sí corrobora el defecto y el nuevo head, pero no reemplaza esa secuencia. Algunos briefs de owners instruyeron commits Git directos, omitiendo el escáner del helper; B sustituto sí lo usó. No se identificaron secretos en los diffs ni efectos remotos no autorizados.

El encargo experimental tampoco explicitaba esa secuencia del verificador: pedía defecto demostrado y corrección. Asimismo pedía un contrato e inventario, sin precisar el formato validado ni la entrega del inventario al mecanismo independiente. Estas discrepancias entre encargo y evaluación limitan atribuir los resultados al modelo. Se alineó únicamente el encargo del driver con los criterios existentes: detección independiente antes de la corrección, contrato v1 validado y resolución independiente del inventario. No se dictan herramientas/skills ni comandos por etapa; no se cambió la guía distribuida ni se debilitó un criterio. Los pilotos originales no se reetiquetan y esta edición no tiene observación nativa.

Comprobaciones independientes repetidas: 17 tests del registro, 9 del contrato, contenido/footprint/tamaños/fronteras y E2E race pasaron. Para la alineación posterior del encargo se repiten los casos E2E y vet de tests/pilot; no se repiten las suites ajenas sin entradas modificadas. El resultado queda local, con T8 abierta y sin push/PR/merge. Una nueva observación requiere autorización propia; no forma parte de D56.
