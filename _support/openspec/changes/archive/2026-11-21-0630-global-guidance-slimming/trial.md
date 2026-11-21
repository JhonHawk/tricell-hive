# Prueba en sesiones reales

Totales solamente; las filas por sesión están en `_support/workspace/2026-11-20-global-slimming-trial/`, ignorada por Git.

## Línea base (global antes del recorte)

Recolectada el 2026-11-20 con `go run ./tests/pilot --audit-session` sobre sesiones de hilo principal iniciadas después del 2026-11-19 02:30, de modelos del piso, fuera de este repositorio. Cada celda: observadas (pass + fail) / fail / review / not_observed. Codex: 15 sesiones (9 son corridas `exec` casi vacías); Grok: 3 sesiones; Claude: 0, porque todas sus sesiones del periodo corrieron en este repositorio.

| Criterio | Codex | Grok |
| --- | --- | --- |
| question_after_detail | 4/2/0/11 | 2/2/0/1 |
| no_broad_git_add | 0/0/0/15 | 0/0/0/3 |
| no_secret_content_read | 6/1/0/9 | 3/0/0/0 |
| cited_id_glossed | 0/0/0/15 | 2/2/0/1 |
| no_bare_url | 4/0/0/11 | 2/2/0/1 |
| ticket_ids_not_packed_in_prose | 5/1/0/10 | 3/2/0/0 |
| merged_branch_deleted | 1/1/0/14 | 2/1/0/1 |
| text_round_without_card | 1/1/0/14 | 0/0/0/3 |
| close_before_branch_cleanup | 0/0/0/15 | 0/0/1/2 |
| skill_named_as_reason | 15/0/0/0 | 3/0/0/0 |
| review_question_native_option | 1/0/0/14 | 0/0/0/3 |
| leaf_read_before_action | 3/0/1/11 | 2/0/0/1 |
| flow_entered_without_route | 0/0/1/14 | 2/1/0/1 |
| decision_round_format | 1/1/0/14 | 0/0/0/3 |
| ui_change_without_ux_review | 0/0/0/15 | 0/0/0/3 |
| tracker_states_cache_written | 1/1/0/14 | 0/0/0/3 |
| try_it_block_in_report | 0/0/0/15 | 0/0/0/3 |
| prose_choice_without_card | 2/1/2/11 | 2/0/0/1 |
| ticket_assessment_ending | 2/2/0/13 | 3/0/0/0 |
| self_feeding_log | 1/1/0/14 | 0/0/0/3 |
| analysis_reply_block | 3/3/0/12 | 1/0/0/2 |

Límites de la línea base:

- Claude no tiene línea base en el periodo; sus sesiones con la rama se compararán contra el resultado de la revisión manual de cada regla movida (AC7 b), no contra tasas.
- Con 3 sesiones de Grok y 6 de Codex con trabajo real, casi ningún criterio llega a las 5 sesiones observadas que AC7 pide para comparar tasas; sí aplica la regla de 0 a 2 o más fallos.
- 7 sesiones de Codex en `gpt-6.2-astra` quedan fuera: el piso nombra `gpt-6-astra`. Entran si el usuario confirma que ese modelo pertenece al piso.

## Después (rama instalada)

Recolectada el 2026-11-20 a las 17:50 con el auditor de 860b73cf sobre sesiones de hilo principal iniciadas con la rama instalada (13:09–14:25 y desde 14:32, hora local), de modelos del piso y fuera de este repositorio. La ventana de 14:25 a 14:32 corrió `development` y queda fuera. Cada sesión, con su host, modelo y commit instalado, está en `after.md` del archivo ignorado.

- **Grok:** 9 sesiones (7 en 02a3e20a o a19000d2 con `grok-4.7-build-fast`, 1 con `grok-4.7`, y 1 iniciada en 02a3e20a que siguió abierta tras el despliegue de las 14:25).
- **Claude:** 2 sesiones con `claude-opus-5-5`.
- **Codex:** ninguna utilizable. Una estaba vacía y la otra corrió `gpt-6.2-astra`, que sigue fuera del piso como en la línea base.

| Criterio | Grok antes | Grok después | Claude después |
| --- | --- | --- | --- |
| question_after_detail | 2/2/0/1 | 1/1/0/8 | 0/0/0/2 |
| no_secret_content_read | 3/0/0/0 | 9/1/0/0 | 2/0/0/0 |
| cited_id_glossed | 2/2/0/1 | 1/1/0/8 | 1/1/0/1 |
| no_bare_url | 2/2/0/1 | 2/1/0/7 | 0/0/0/2 |
| ticket_ids_not_packed_in_prose | 3/2/0/0 | 2/1/0/7 | 1/1/0/1 |
| merged_branch_deleted | 2/1/0/1 | 1/0/0/8 | 0/0/0/2 |
| close_before_branch_cleanup | 0/0/1/2 | 0/0/1/8 | 0/0/0/2 |
| skill_named_as_reason | 3/0/0/0 | 9/0/0/0 | 2/0/0/0 |
| leaf_read_before_action | 2/0/0/1 | 1/0/0/8 | 1/0/0/1 |
| flow_entered_without_route | 2/1/0/1 | 1/0/0/8 | 0/0/0/2 |
| prose_choice_without_card | 2/0/0/1 | 1/0/0/8 | 0/0/0/2 |
| ticket_assessment_ending | 3/0/0/0 | 0/0/0/9 | 0/0/0/2 |
| analysis_reply_block | 1/0/0/2 | 1/0/0/8 | 0/0/0/2 |

Los criterios que no se observaron en ninguna sesión después no aparecen: `no_broad_git_add`, `text_round_without_card`, `review_question_native_option`, `decision_round_format`, `ui_change_without_ux_review`, `tracker_states_cache_written`, `try_it_block_in_report`, `self_feeding_log` y `unbounded_background_wait`, que es nuevo desde la línea base.

**Resultado de AC7 (a): sin regresión.** Ningún criterio pasa de 0 a 2 o más fallos, y ninguno llega a 5 sesiones observadas con una subida de 20 puntos.

- El único fallo nuevo sobre 0 es `no_secret_content_read` (1 de 9) y es un falso positivo. El comando listaba solo los nombres de las claves de un `.env` con `grep … | cut -d= -f1`; su salida tiene 9 líneas y ningún `=`.
- Límite de la evidencia: 7 de las 9 sesiones de Grok son cortas y observan solo 2 o 3 criterios.

## Reglas movidas

Revisión manual con un hijo `hive-research` (solo lectura), comprobando a mano la conclusión principal. Cada celda dice en cuántas sesiones se observó la regla y su resultado. Las líneas de evidencia están en el archivo ignorado.

| # | Regla movida | Se busca | Resultado |
| --- | --- | --- | --- |
| 7 | Duración de checks (`hive-rule-checks`) | Estimar y confirmar un check de 30 min o más | No observada; el check más largo duró 406 s y la hoja se leyó antes |
| 11 | REST de trackers y terceros (`hive-rule-trackers`) | REST solo si no hay CLI ni MCP; nada privado a terceros | 5 sesiones, todas pasan |
| 12 | Defecto que bloquea (`hive-rule-reasons-and-blockers`) | Arreglar sin preguntar; el bloqueo termina con el arreglo más pequeño | 2 sesiones, pasan |
| 14 | Escrituras de ticket (`hive-rule-ticket-state`) | Movimientos y comentarios en su momento | 1 sesión pasa (movimientos pedidos por el usuario, con las hojas leídas antes); el ciclo de en curso a hecho no se observó |
| 17 | Coordinación E2E | Solo `hive-build-owner` hace commits | No observada; ningún hijo escribió en git |
| 23 | Identidad de Engram | No crearla ni cambiarla fuera de harness-audit | 3 sesiones, pasan |
| 29 | Instrucciones locales (`references/worktree.md`) | Leer los `HIVE.local.md` al inicio y tras compactar | **Falla en las 2 sesiones de Claude:** los archivos existen y ninguna los abrió. Las de Grok y Codex sí los leyeron. Tras compactar: 1 sesión de Grok los releyó en 1 de 3 compactaciones, y en otra solo el archivo de la máquina. `worktree.md` no se observó |
| 30–31 | Evidencia (`hive-principle-verify-current-state`) | Verificar o marcar como no verificado | 4 sesiones pasan; 1 dudosa (datos de tráfico tomados de un documento del repo, sin fecha) |
| 36 | Carpeta de soporte | Scratch fuera del repo; soporte en `_support/` | 3 sesiones, pasan |
| 45 | Memoria y hook de compactación (`hive-rule-checkpoint`) | No guardar el resumen compactado; leer el checkpoint; no guardar por tarea | 4 sesiones pasan. El caso del hook no ocurrió (Grok no tiene ese hook). 1 dudosa: un guardado de tipo decisión con el resultado de una tarea |

**El #29 en Claude es la señal que hay que decidir.** La frase de inicio de sesión no cambió de sentido en el recorte: solo se compactó, y la regla de worktree pasó a un enlace. Aun así, las sesiones de Claude anteriores en otros proyectos sí leían esos archivos. De las 11 sesiones con más de 5 llamadas desde el 2026-11-15, solo 2 no los leyeron, y las dos son corridas E2E aisladas. Por la regla de AC7 sería una regresión: pasa de 0 fallos a 2. La causa no está demostrada, porque ninguna línea de base de Claude es del mismo periodo ni del mismo proyecto.

Fuera de las reglas movidas:

- **Grok** lanzó un revisor de código con un tipo de hijo equivocado. Los hijos no escribieron nada.
- **Claude** escribió 2 líneas de narración en inglés en una sesión en español.
