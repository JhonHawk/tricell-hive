# Design

## Approaches

- **A — Línea acotada en la regla de preguntas (elegida, D4 A).** Agrega un caso a la tabla de canales que ya existe. No prohíbe la async cuando es la única herramienta, y resuelve de forma explícita el choque con la plantilla del modo Default de Codex. Es la opción que más reutiliza: un archivo de regla que ya se lee en cada etapa.
- **B — «Nunca async» en `global.md` para Codex.** Actúa aunque no se lea el archivo de la regla. En contra: la sesión `01a20c02` sí lo leyó, consume el tope de 30.000 bytes de la guía global y deja sin canal el caso con solo async (sesión `01a1f2f5`).
- **C — Solo auditor.** Es el cambio más barato, pero detecta después de la sesión y no previene.

Las tres versiones comparadas (2026-11-24):
- pstack solo «prefiere la herramienta de preguntas sobre el texto libre» y mapea la herramienta por host; no trae plantilla ni caso async.
- Hive antes (`6b9e0d49`, 2026-10-02) usaba la async solo mientras seguía trabajo independiente y mandaba al texto la pregunta que bloquea.
- Hive actual trata la async como «sin tarjeta» y no resuelve el choque con la plantilla.

## Decisions

- Un recomendado entre backticks con un modo de entrega da `fail`; un pendiente sin recomendación da `review`. La razón: el primero es una pregunta de sesión sin opciones, y el segundo puede ser un hecho que se informa.
- El criterio async es nuevo y no se mete en `card_option_format`, que juzga el formato de las tarjetas que esperan y salta la async a propósito.
- El `▶` es la señal del bloque de siguiente paso porque es el único elemento de la plantilla que no depende del idioma.

## Impact

- `codexCompletedItem` alimenta todos los criterios que leen cargas de skill: `flow_entered_without_route`, `plan_without_build_route`, `analysis_reply_block` (`loadsHiveMode`), `try_it_block_in_report` y la activación desatendida de `prose_choice_without_card` y `text_round_without_card`. Sus estados pueden cambiar en sesiones Codex reales. Ninguna fixture trae ítems `UserMessage`, así que T3 mide el cambio sobre diez rollouts reales. `leaf_read_before_action` no lee `skill_invocation`.
- La regla de preguntas entra en los paquetes de `hive-mode`, `flow-plan`, la salida de `flow-verify` y `unattended-delegation`. La suma de sus archivos deja unos 2,8 KB bajo 23.500 en el más grande (`flow-plan`, 20.644 bytes).

## Recovery

Revertir el commit del PR. No toca datos guardados.
