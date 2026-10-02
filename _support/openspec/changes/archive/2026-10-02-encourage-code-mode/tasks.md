# Tareas

Commit base: `b3af4df` (`development`, 2026-10-02). Rama de trabajo: `feat/encourage-code-mode`.

Modelos: T1 la implementa `hive-build-backend`, con `sonnet` configurado y no observado, y la verifica `hive-verify-task`, con `opus` configurado. T2 la implementa el hilo principal, con Opus 5.5 observado, así que su verificación se lanza con `sonnet` para no usar el mismo modelo.

## T1 — Fila `Pi codemode` en `hive doctor` y en la TUI

- [x] Con Pi registrado, `hive doctor` y la vista Integrations muestran la fila `Pi codemode` descrita en [design.md](design.md#fila-pi-codemode), sin leer la configuración de Pi; sin Pi registrado no la muestran.

**Closes:** AC1, AC2, AC3.

**Depends on:** ninguna.

**Locations:**

- `tooling/cli/doctor_integrations.go`:
  - `collectIntegrationRows`, con la fila condicionada a `management.RegisteredHosts(o)`;
  - `nextStep`: un caso explícito para `pi-codemode` y el comentario de las líneas 173-174;
  - la constante de la fuente, junto a `agentBrowserSource`.
- `tooling/cli/tui_integrations_view.go`: `integrationsFixedRows` y el comentario de cabecera.
- `tooling/cli/tui_views.go:275`: la descripción del menú Integrations.
- Pruebas:
  - `tooling/cli/doctor_integrations_test.go`: las listas de nombres en la línea 122 y siguientes, y el fixture `integrationsFixture` con Pi registrado (`doctorHome(t, "claude,pi")`). El caso sin Pi necesita un fixture propio, como un parámetro de hosts en `integrationsFixture` o `doctorHome(t, "claude")` con `nonSyntheticOptions`.
  - En `collectIntegrationRows`, la llamada a `RegisteredHosts` va después de fijar `Scope = "user"` (líneas 112-114), porque filtra por ese valor.
  - `tooling/cli/tui_integrations_view_test.go`: `len(rows) != 4`, las listas de nombres en las líneas 46 y 80, y el cursor en las líneas 99, 102, 104, 236 y 237, que ahora caen en `Pi codemode`.
- Documentación: `_support/docs/architecture/deployment-manager.md`:
  - línea 73: las integraciones que lista, la condición de Pi registrado y la frase del único comando sugerido;
  - línea 143: "lists the four integrations".

**Execution:** delegada a `hive-build-backend`. La interfaz está fijada en design.md, no se pisa con T2 y la puede revisar el hilo principal sobre el diff.

**Test approach:** `tdd`.

1. Primero las pruebas que fallan:
   - con Pi registrado, el nombre y el orden de las cinco filas; sin Pi, las cuatro de hoy;
   - el texto de `Found`, `Record`, `Source` y `Next` de la fila nueva;
   - que con `integrationsFixture`, que no es sintético, y un `<home>/.pi/agent/settings.json` que declara `"defaultTools": ["+codemode"]`, la fila siga en `not checked`. No vale un `--home` sintético: ahí `forOptions` anula `lookPath` y la prueba no distinguiría nada.
2. Después la implementación.

**Changes:** según [design.md](design.md#fila-pi-codemode). Las pruebas existentes que fijan cuatro filas se actualizan a cinco; ninguna otra prueba cambia de expectativa.

**Verification:**

- `go test ./tooling/cli/ -run 'Integrations|Doctor|PiCodemode'`: pasa (el patrón inicial, sin `PiCodemode`, dejaba fuera las pruebas de AC1 y AC3; lo corrigió la verificación), y las pruebas nuevas fallan si se quita la fila (comprobado antes de implementar).
- `go run ./tooling/cli doctor` desde el checkout: en esta máquina, que tiene Pi registrado, imprime la fila `Pi codemode` con el siguiente paso que contiene `"defaultTools": ["+codemode"]`. Las pruebas unitarias cubren AC1 en cualquier máquina.

## T2 — Regla de code mode en la guía global

- [x] `content/guidance/global.md` tiene en Proportionality la regla de [design.md](design.md#regla-de-guía), justo antes de la regla de delegación.

**Closes:** AC4.

**Depends on:** ninguna.

**Locations:** `content/guidance/global.md`, sección `## Proportionality`, entre las líneas 85 y 86.

**Execution:** hilo principal. Es un párrafo cuya redacción ya está fijada.

**Test approach:** `check`.

**Changes:** insertar el punto tal como está en design.md.

**Verification:**

- `rg -n "runs a script able to call other tools|The main thread coordinates work" content/guidance/global.md`: la regla sale en la línea inmediatamente anterior a la regla de delegación.
- `rg -n "skip the host's approval" content/guidance/global.md`: encuentra la frase sobre escrituras en esa misma línea.
- `rg -n -i "codemode|\bexecute\b" content/guidance/global.md`: no encuentra nada, porque la regla no nombra herramientas.
- `go test ./tests/content/`: pasa. La prueba de tamaño (`budget_test.go`) lo admite: unos 43304 bytes frente al tope de 43622.

## Verificación compartida y revisión humana

| Prueba | Cuándo y dónde | Mecanismo y evidencia esperada | Autorización |
| --- | --- | --- | --- |
| Suite completa | Sobre el candidato final, en el checkout local | `go vet ./...` y `go test ./...` sin fallos | Parte de la implementación |
| Despliegue local | Después del merge | `go build -o "$(command -v hive)" ./tooling/cli`, luego `hive update` desde el checkout. Después, `rg -n "runs a script able to call other tools"` encuentra la regla en el bloque global de cada host registrado, como `~/.claude/CLAUDE.md` y `~/.pi/agent/AGENTS.md` | Excepción temporal de `AGENTS.md` (cada merge a `development` la pide) |
| Observación (D2-A) | Próximas sesiones reales de OpenCode y Pi | Si el modelo usa `execute` o `codemode` en cadenas de llamadas y si los scripts fallan | Fuera de este cambio. Se apunta en la memoria de seguimiento de sesiones |

No hay efecto visual que requiera `hive-review-ux` ni `hive-verify-change`: la fila nueva es texto en una lista que ya existe, y las pruebas de la vista comprueban su posición.

## Revisión del plan y avance

- **Revisión del plan:** ronda 1 sobre el candidato `466b65819af8`, con `hive-review-plan` nativo y de solo lectura, un revisor por dominio.
  - **Guía:**
    - B1 (AC4 se podía cumplir con líneas que ya existían): aceptado. AC4 y su comprobación usan ahora la frase exacta de la regla.
    - N1 (`go test ./tooling/...` no comprueba la guía): aceptado. Se usa `go test ./tests/content/` y la regla cabe en el tope de tamaño.
    - N2 a N4 (cuándo usar script o subagente, escrituras sin aprobación, secretos en el script): aceptados. La regla se condensó a partir del texto que propuso el revisor y se movió antes del bloque de delegación.
  - **Gestor en Go:**
    - B1 (el delta borraba cuatro escenarios): aceptado.
    - B2 (AC3 no se podía probar con `--home` sintético): aceptado. Se usa `integrationsFixture`, y la búsqueda en `PATH` pasó a restricción.
    - N1 a N4 y N6: aceptados.
    - N5 (la razón para mostrar siempre la fila era inexacta): aceptado. Se vuelve a lo que eligió el usuario en D4-A, la fila solo con Pi registrado.
  - **Ronda 2, guía:** sin bloqueantes. R1 a R4 (contradicción con una sola llamada, alcance de "as below", confirmación de escrituras y comprobación de posición) se aplicaron tal como los propuso el revisor, así que no necesitan otra ronda.
  - **Ronda 2, gestor en Go:** sin bloqueantes. Comprobó que se puede llamar a `RegisteredHosts` en esa función y que el fixture puede registrar Pi. N1 a N3 (fixture sin Pi, la condición de Pi en AC1 y la comprobación manual que depende de la máquina) se aplicaron tal como los propuso. Límite: no observó `writeOnboardingRecord` con Pi ya registrado.
- **Construcción (2026-10-02):**
  - T2 `[x]`: `hive-verify-task` con `sonnet` dio AC4 cumplido.
  - T1 `[x]`: `hive-verify-task` con `opus` (configurado) dio AC1, AC2 y AC3 cumplidos. Evidencia de fallo antes del cambio, según el implementador: `TestIntegrationsSectionListsTheFiveRowsInOrderWithPi`, `TestPiCodemodeRowTexts` y `TestPiCodemodeRowIgnoresPiSettings` fallaban antes de implementar.
  - Desvío aceptado: la descripción del menú usa comas y no "and", para no pasar de 80 columnas.
  - Sin prueba: la rama en la que `RegisteredHosts` falla y la fila se muestra igual. Ningún criterio la exige.
  - Suite completa (`go vet ./...`, `go test ./...`): pasa sobre el candidato final.
- **Plan:** listo. Las dos rondas cubrieron los dos dominios; las correcciones finales son propuestas literales de los revisores.
- **Entrega:** PR #101 fusionado en `development` (`edcaa72`), binario recompilado y `hive update` aplicado; la regla aparece en el bloque global de Claude Code, Codex, OpenCode, Pi y Cursor (Grok lee `~/.claude/CLAUDE.md`) y `hive doctor` muestra la fila.
- **Siguiente paso:** ninguno en este cambio.
