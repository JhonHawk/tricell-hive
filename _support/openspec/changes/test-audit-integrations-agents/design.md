# Diseño

## Enfoque

Hay un solo enfoque razonable: el de la skill `test-audit`. Cada lane va en orden: primero los keepers con sus aserciones trasladadas y las reparaciones F, después los borrados. Al final, una revisión de preservación y una mutación por contrato. Las alternativas (borrar primero y reponer lo que falte, o reescribir los tests desde cero) dejan contratos sin prueba durante el cambio o pierden la evidencia del ledger.

## Dueños de cada contrato

`ledger.md` asigna a cada contrato su keeper, con la línea que ya lo comprueba o la aserción que hay que trasladar. Los más relevantes:

- Formato de salida de `Render` en los 7 hosts: `TestRenderGolden` (`render_golden_test.go:52`), que absorbe las pruebas de formato de `agents_test.go:115, :158, :236, :257, :381, :402`.
- Perfil → modelo y esfuerzo en `Resolve`: `resolve_test.go:19`, con las filas verifier de `:196`.
- Releases congeladas con menos hosts: `agents_test.go:423` como tabla.
- Reescritura de links `skill:`: `tooling/management/skill_links_test.go:61` en la frontera plan/apply, con `antigravity` agregado (D2 A).
- Rechazos de `Parse`: `agents_test.go:364` como tabla, que suma los caminos de error que hoy no tienen test (H2 B).

## CI

**Retired:** D4. Se agregó un workflow `tests.yml` que corría la suite en Linux, y el usuario decidió quitarlo: la suite sigue siendo un paso local.

## Fixes que mostró la suite en Linux

- **H7:** GNU tar llama a `gzip` para `-z`, así que `get-hive.sh` lo exige antes de cualquier request.
- **H8:** el `curl` de Linux quita el userinfo de `%{url_effective}`. El caso del test se salta solo en Linux, y el issue #543 sigue la regla.
- **H10:** `ValidateState` lee sin lock por diseño. `readStateFileAttempt` reintenta cuando un rename atómico deja sin enlaces el inodo abierto. Las otras dos instancias del patrón en `tooling/cli` (`readPrivateWorkFile` y `readSafeMapFile`) quedan como hallazgos.
