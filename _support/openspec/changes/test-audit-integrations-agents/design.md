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

Un solo workflow, `tests.yml`, con un job que corre los mismos tres comandos de `CONTRIBUTING.md`. Se usa `pull_request` y no `pull_request_target`, para que el código del PR corra sin acceso a secretos ni token con escritura. El nombre del job no puede ser `source-branch`, porque el ruleset de `master` exige ese check solo de `master-source-branch.yml`. Volverlo un check obligatorio en el ruleset queda fuera de este trabajo.
