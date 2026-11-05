# Design

## Approaches

Se presentaron tres enfoques el 2026-11-04 (D6); el usuario eligió A.

- **A, elegido: hora en el nombre.** Las carpetas nuevas son `YYYY-MM-DD-HHMM-<change-id>`. El orden queda visible y funciona sin historia de git ni dependencia de CI; el nombre es más largo.
- **B: dato en `proposal.md`.** Los nombres no cambian, pero el test leería un campo localizado («Estado» o «Status»), y eso es frágil.
- **C: dejar la tolerancia.** No cambia la convención y el chequeo sigue debilitado.

La fecha y la hora van en UTC (D8 A), para que el orden no dependa de la zona horaria del host. Para lo ya archivado (D7), el usuario eligió renombrar solo el par en conflicto del 2026-11-04, con la hora de su commit de archivo. No se renombra todo el archivo ni se deja todo como está.

## Proposed shape

**Nombre.** Una carpeta archivada tiene hora cuando su cuarto segmento son cuatro dígitos `HHMM` válidos (00–23 y 00–59). `2026-11-04-1403-viewer-…` tiene hora; `2026-11-04-gh-311-…` no. Un `<change-id>` que empiece con cuatro dígitos se leería como hora. No existe ninguno en el archivo, y los `change-id` empiezan con el ticket o con una palabra.

**Orden en el test.**
- `sort.Strings` ya ordena `2026-11-04-1403-…` antes que `2026-11-04-2003-…`.
- Dos carpetas no tienen orden conocido, y entonces la especificación puede coincidir con cualquiera de las dos, en tres casos:
  - el mismo día, cuando a alguna le falta la hora;
  - el mismo día, cuando las dos tienen la misma hora;
  - una sin hora (fecha local del formato anterior) y otra con hora a un día o menos, como en el paso a UTC.

  En los demás casos manda el orden del nombre. El test guarda los bloques de todas las carpetas anteriores de cada requisito y acepta los que no tienen orden conocido frente a la última, de modo que una carpeta intercalada no los oculta.
- Cualquier carpeta que ordene después supera a un duplicado, también una del mismo día sin hora.

**Duplicados (H29–H36).** Una pasada aparte, `duplicateRequirements`, devuelve por carpeta los requisitos que su delta lista más de una vez, con sus operaciones (H34).
- La recolección salta esos bloques (H29 y H32).
- `checkSpecDeltas` reporta el duplicado solo cuando la carpeta que lo tiene es la del delta más reciente de ese requisito (H31). En ese caso no compara la especificación, porque el resultado es ambiguo.
- El comentario de `checkSpecDeltas` describe los dos tipos de mensaje (H33).
- El bucle de coincidencias se escribe con una condición simple (H35) y la lista de fallos parte de la de duplicados sin copiarla (H36).
- Los subtests prueban que una carpeta posterior se compara después de un duplicado, con una especificación que la contradice: dos fallos esperados (H30).

## Impact

Lo usan `flow-close` en cada cierre de un trabajo con carpeta de cambio, en todos los hosts, y el test de contenido en CI y en local. `doctor` y los nombres antiguos no cambian. No toca producción.

## Recovery

Revertir el merge, que también deshace el renombre. Las rutas antiguas de las dos carpetas que pudieran citarse fuera del repositorio volverían a ser válidas con la reversión.
