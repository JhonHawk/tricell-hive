# Diseño

## Varios trackers (D7 A)

Opciones consideradas:

1. **Clave `Other trackers` y `Completion` por prefijo (elegida).** Sigue el patrón del proyecto: una clave por concepto, validada por `hive doctor` y editable con `hive project set` y el TUI, que ya recorren `hiveSettingKeys`.
2. **Repetir `Tracker`.** Compacto, pero rompe «una línea por clave» del parser y del editor y saca `Completion` de su clave.
3. **Sección libre `## Trackers`.** Sin cambios en Go, pero sin validación; un proyecto real ya mostró que el texto libre no basta.

Forma, con marcadores (el repositorio es público):

```markdown
- Tracker: Linear · team Example · key ABC
- Other trackers: Jira · example.atlassian.net · keys XYZ
- Completion: XYZ: QA on qa · Done on production
```

Gramática:

- `Tracker` es el principal, sigue siendo obligatorio y no se analiza: su forma no cambia.
- `Other trackers`: uno o más trackers separados por `;` (con o sin espacio; se recortan y las entradas vacías se ignoran). Cada uno tiene al menos tres partes separadas por ` · ` y la última es `keys <PREFIJO>[, <PREFIJO>]`, con `<PREFIJO>` = `[A-Z][A-Z0-9_]*` (también se acepta `key` con un solo prefijo). Un prefijo pertenece a un solo tracker; repetirlo es un aviso.
- `Completion`: uno o más valores separados por `;`, con la misma separación. Un valor empieza con prefijo cuando coincide con `^[A-Z][A-Z0-9_]*(, [A-Z][A-Z0-9_]*)*: `; si no, no tiene prefijo. Cada valor, sin su prefijo, debe coincidir con la forma actual `<state> on <ref> · <state> on <ref>`, comprobada sobre ese valor y no sobre la línea entera.
- Un valor sin prefijo aplica al principal; como mucho uno. `Completion` puede tener solo valores con prefijo: entonces el principal no tiene punto de cierre posterior. Un prefijo de `Completion` debe estar en `Other trackers`; si no, aviso. Un prefijo repetido en `Completion` es un aviso.
- Un ticket cuya clave no declara ningún tracker se trata con el principal.
- Todos los hallazgos nuevos son avisos: nunca bloquean ni cambian el código de salida.
- `hive project set --set 'Other trackers: …'` muestra la forma en el texto del aviso cuando el valor no la cumple; el TUI la muestra como ayuda, como hoy `Completion`.

## Errores de `tasks register` (H13)

Los validadores devuelven textos fijos («duplicate task key», «invalid task count»…). El error conserva `invalid_input` y la sugerencia de leer el contrato, y añade la ruta y la regla, con la convención documentada `tasks[<i>].<campo>` para la lista de `tasks register` y `metadata.<campo>` para `task update --json`, por ejemplo `invalid semantic JSON: tasks[1].key: duplicate task key`.

La decodificación se hace elemento a elemento en `view_work.go`, sin tocar `tooling/viewer` ni su contrato HTTP: primero la lista como mensajes JSON crudos, luego cada elemento con campos desconocidos rechazados, y el error se reduce a ruta y clase (`unknown field`, `wrong type`), sin el valor recibido. Los nombres de campo que se muestran son los de la entrada, no los de la petición al visor.
