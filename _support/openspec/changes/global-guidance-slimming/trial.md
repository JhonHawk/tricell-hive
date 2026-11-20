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

Pendiente de T7.

## Reglas movidas

Pendiente de T7: cada regla del «Mapa de reglas» con su detector del auditor o «revisión manual» y su resultado.
