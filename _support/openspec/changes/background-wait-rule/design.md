# Diseño

## Enfoques

Se presentaron tres opciones (D1):

- **A, elegida.** Hoja nueva más un criterio de auditoría. Sigue el patrón de las hojas de lectura rara (`hive-rule-loops`, `a9ccb0dd`), que se leen en el momento en que se lanza la espera.
- **B.** Meterla en `hive-rule-checks`. No suma bytes a la guía global, pero esa hoja se lee antes de correr una verificación, así que las esperas de despliegues no la activan.
- **C.** A más un script `wait_until.py`. Es lo más robusto, pero añade un script con sus pruebas sin un fallo medido del texto. Queda en el quinto escalón de reutilizar antes de escribir.

Para el detector se presentaron dos (D2):

- **A, elegida.** Fail solo si hay un cierre después del último merge. Reutiliza `isCloseReport`, igual que se hizo con `close_before_branch_cleanup` (D11 A, `736c17b9`).
- **B.** Review siempre. Pierde la señal de un cierre que no limpia.

## Forma propuesta

### Hoja `content/skills/hive-rule-background-wait/SKILL.md`

En inglés, de 4,000 bytes como máximo. La descripción empieza con «Hive rule leaf: read…» y no pasa de 200 caracteres. Lleva `disable-model-invocation: true`. Contenido:

1. Antes de escribir un bucle, preferir un comando que bloquee hasta el estado final y salga con su código, como un modo watch o un comando `wait` del proveedor.
2. Un bucle hecho a mano termina en todos los resultados posibles: éxito, fallo, objetivo que aún no existe, proceso colgado y salida que no reconoce. La salida no reconocida termina con error e imprime lo que vio. Se compara un campo estructurado o el código de salida, no texto libre. Si no queda más que comparar texto, primero se corre la consulta una vez y se prueba el patrón contra su salida real, incluidas líneas de más como un token de paginación. Un watch que termina diciendo que no hay objetivo, como `no checks reported`, no ha terminado: se relanza el mismo comando.
3. Toda espera tiene plazo: el bucle entero va envuelto en `timeout <segundos>` o lleva un contador que sale con el último estado visto. Un `timeout` dentro del cuerpo solo acota una vuelta.
4. Si el host solo despierta al agente cuando el comando termina, el plazo es lo que lo despierta. Cada tramo dura entre 5 y 10 minutos, nunca un tramo de 30 o 60 segundos. Al despertar se revisa el avance. Si avanza, se relanza el mismo comando. Si no avanza en dos tramos seguidos, se para y se reporta como bloqueo, con el último estado visto.
5. Las esperas de CI y de hijos siguen sus propias hojas, [merge gate](../hive-rule-merge-gate/SKILL.md) y [child brief](../hive-rule-child-brief/SKILL.md). Esta hoja cubre las demás.

Línea en «Rule leaves» de la guía global, junto a la de checks: `- **Before starting a polling loop or a background command that waits on a process or an external job:** [background wait](skill:hive-rule-background-wait/SKILL.md).`

### Criterio `unbounded_background_wait` (`tests/pilot/session_audit_waits.go`)

- Solo cuentan los comandos de shell lanzados en segundo plano: en Claude, `Input.run_in_background == true`; en Grok, `Input.block_until_ms == 0` en `run_terminal_command`. Codex da `not_observed`.
- Es un bucle de espera si el comando tiene `while`, `until` o `for` junto con `sleep`.
- Hay plazo en tres casos, detectados con `shellSegments` y no por subcadena:
  - `timeout` o `gtimeout` es el primer comando de un segmento que contiene el bucle, por ejemplo `timeout 600 bash -c '…'`; un `timeout` dentro del cuerpo no cuenta;
  - el bucle es un `for` sobre una lista finita (cualquier lista no vacía, incluidas `$(seq …)` y `{a..b}`) o un `for ((…))` cuya condición compara un contador o el tiempo con un límite;
  - `$SECONDS`, `date +%s` o un contador aritmético aparecen en una comparación con un límite: en la condición de un `while` con `-lt`, `-le`, `<` o `<=`; en la de un `until`, o en una prueba del cuerpo que lleva a `exit` o `break`, con `-ge`, `-gt`, `>=` o `>`. Leerlos solo para imprimir no cuenta. La condición del bucle se añadió durante el build: sin ella, un `while [ $i -lt 60 ]` bien acotado daba `fail`.
- El bucle se reconoce también después de `then`, `do`, `else`, `elif`, `!`, `(` o `{`.
- Límites aceptados, que dan `pass` sin ser un plazo real: una condición con `|| true` y un tiempo leído una sola vez antes del bucle. El criterio tampoco comprueba que el bucle, al agotar su plazo, salga mostrando el último estado.
- El estado es `fail` para un bucle sin plazo (evidencia: la línea), `pass` si todos los bucles tienen plazo, y `not_observed` si no hubo bucles en segundo plano o si la sesión es de Codex. No hay estado `review`: Claude y Grok no ocultan el comando.
- Los fixtures usan la estructura real de cada host con valores sintéticos: ni perfiles, ni IDs de cuenta o de app, ni organizaciones de otros proyectos.

### `merged_branch_deleted` en la auditoría

Dentro de `existingAuditCriteria`, si `mergedBranchDeletedWith` da `fail` y ningún mensaje del asistente que `isCloseReport` reconozca viene después del último merge, el estado pasa a `review`. La evidencia lo explica: «no close report after the last merge: the session may be open or the merge a promotion». El detector de los pilotos no cambia.

## Impacto

- **Guía global:** suma una línea, unos 140 bytes. La rama de adelgazamiento la absorbe al rebasarse (D3 A).
- **Auditoría:** cambian `existing_criteria` y sus pruebas (`session_audit_test.go:291-314`). `claude-project-merged-cleanup-fail.jsonl` y `claude-project-local-merge.jsonl` (`TestAuditLocalMergeNeedsBase`, que espera `fail`) no tienen cierre. Cada uno recibe un cierre después del merge, para que su test siga probando lo mismo: que un cleanup en dry-run no cuenta como borrado, y que el merge local necesita base. Un fixture nuevo sin cierre da `review`. `codex-rollout-running-merge.jsonl` ya da `not_observed` y no cambia.
- **Tests de contenido:** `rule_enforcers_test.go` valida nombres, tipos y estados de los criterios. `single_home_test.go` registra la frase canónica de la hoja.
- **Documentación:**
  - `session-review-rubric.md:7` añade cómo leer el `review` nuevo de `merged_branch_deleted`.
  - La fila de limpieza de ramas de `rule-enforcers.md` nombra también `tests/pilot/session_audit.go`.
  - `CHANGELOG.md` recibe su entrada en `[Unreleased]`, como pide `CONTRIBUTING.md:149-150`.
- **Producción:** no toca nada en producción.

## Recuperación

Para deshacerlo se revierte el commit integrado y se vuelve a desplegar con `hive update`. No hay datos que migrar.
