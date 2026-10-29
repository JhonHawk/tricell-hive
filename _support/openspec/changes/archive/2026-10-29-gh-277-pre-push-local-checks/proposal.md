# Correr localmente, antes del push, lo que el CI va a rechazar

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por el PR #278 (merge abc5e2bf) |
| Tracker · GitHub Issues | • [#277 — Antes del primer push el agente no corre localmente lo que el CI va a rechazar](https://github.com/JhonHawk/tricell-hive/issues/277) |
| Git | `automatic` · PR a `development` y merge al pasar las pruebas · sin parada humana |
| Verificación | `go test ./...` · búsquedas · `astro build` · `/code-review` |
| Siguiente paso | Ninguno. Observar en sesiones de cam con la guía nueva si el agente corre los chequeos locales antes del push |

## Pedido

> hagamos research para ajustar esto, porque creo que es mucho más barato hacerlo local que esperar al CI

El usuario eligió D13-A el 2026-10-29: plan completo con issue. Criterios de [#277](https://github.com/JhonHawk/tricell-hive/issues/277), leídos el 2026-10-29, textuales:

> - Una regla distribuida exige, antes del primer push de un cambio no mecánico, las suites unitarias, el lint y el typecheck de los paquetes cambiados y de sus dependientes en el repo.
> - Una regla distribuida exige correr localmente los jobs del CI que son deterministas, sin servicios y aplicables a las rutas cambiadas, leídos de los workflows.
> - La integración, el e2e y los despliegues quedan en el CI y el reporte los nombra como no corridos localmente.
> - `flow-verify` deja de decir que la cobertura de regresión se cierra solo con el CI para lo que se puede correr localmente.
> - Cada archivo tocado sigue en 16,000 bytes o menos. `verification.md` se divide por responsabilidad.
> - `site/` describe el comportamiento nuevo, y `go test ./...` pasa.

El plan no toca `verification.md`, porque la regla vive en `review-and-gates.md` (ver el diseño). Por eso queda fuera la parte del criterio que pedía dividirlo; el límite de 16,000 bytes sigue aplicando a cada archivo tocado.

## Objetivo

Con la entrega `automatic` y un CI presente, la guía deja la cobertura de regresión al CI. Por eso el agente sube el cambio después de correr solo las pruebas de los archivos que tocó, y espera unos 3.5 minutos para ver un fallo que se podía detectar en uno o dos minutos con un comando local.

En `cam-monorepo`, del 15 al 29 de octubre, 11 de los 14 jobs fallidos antes del merge eran de este tipo: pruebas unitarias o de convención, cambios incompatibles de contrato e i18n.

El cambio agrega una regla para el momento antes del push. El agente corre localmente las suites de los paquetes cambiados y de sus dependientes, y los jobs del CI que son deterministas y no necesitan servicios. El CI queda para la integración con servicios, el e2e y los despliegues.

## Alcance y aceptación

**Incluido:**
- La regla en `flow-verify/references/review-and-gates.md`.
- Los punteros en `flow-verify/SKILL.md` y `git-workflow/SKILL.md`.
- La página `site/.../verify.md`.
- Una entrada en `CHANGELOG.md`.

**Excluido:**
- Correr la integración con servicios antes del push: necesita Postgres y Redis, tarda de 3 a 6 minutos o más y compite con otras sesiones por esos servicios.
- Las fallas intermitentes de integración después del merge.
- Los hooks de cam.
- Un cambio mecánico, que sigue verificándose con los chequeos que ya lo cubren.

**Restricciones:**
- Cada archivo queda en 16,000 bytes o menos.
- `global.md` no cambia.
- Se conservan la regla de hooks de `git-workflow/SKILL.md:28` y la regla de integración local estancada de `verification.md:53`.
- La revisión dedicada sigue corriendo después de los chequeos locales, como fija `review-and-gates.md:40`.
- `go test ./...` pasa.

**Búsqueda en los requisitos vigentes:** se buscaron "CI", "push", "local", "verify" y "suite" en `_support/openspec/specs/`, el 2026-10-29. `documentation-site` y `context-delivery` no fijan qué chequeos corren antes del push. No hay deltas.

**Criterios:**
- AC1. Antes de cada push de un cambio no mecánico a un repositorio cuyo CI correrá sobre él, la guía exige correr localmente el lint, el typecheck y las pruebas unitarias o de convención de cada paquete tocado y de cada paquete del repositorio que depende de él, con la selección de paquetes cambiados del proyecto cuando la tiene. *Falso en la base cuando* `rg -n "depends on" content/skills/flow-verify` no encuentra una regla así y `verification.md:50` limita "affected" a la iteración.
- AC2. La misma regla exige correr localmente los jobs del CI que ese push disparará para las rutas cambiadas, que son deterministas y no necesitan servicios, secretos ni despliegue, con los comandos que el job corre, leídos de los workflows. *Falso en la base cuando* ninguna regla de `content/` pide leer los workflows para reproducir sus jobs; `git-workflow/SKILL.md:28` solo los lee para comparar con los hooks.
- AC3. Los jobs que necesitan servicios, el e2e y los despliegues quedan en el CI, y el reporte los nombra como "solo CI". *Falso en la base cuando* ninguna regla nombra esa categoría.
- AC4. `flow-verify/SKILL.md` ya no deja la cobertura de regresión solo al CI para lo que esa regla corre localmente. *Falso en la base cuando* `flow-verify/SKILL.md:47` dice "Close regression coverage through the project's actual CI gate" sin esa excepción.
- AC5. `git-workflow/SKILL.md` apunta a la regla antes de un push, así que un push desde cualquier etapa la aplica. *Falso en la base cuando* `git-workflow/SKILL.md` no la nombra.
- AC6. `site/src/content/docs/flows/verify.md` describe los chequeos locales antes del push. *Falso en la base cuando* la página solo describe la suite completa antes de una parada de validación o sin CI (`verify.md:25`).

## Entrega

- **Repositorio:** `tricell-hive`. Rama propuesta `feat/gh-277-pre-push-local-checks`, desde `development` 0afaa7c6.
- **Respuestas reutilizadas de esta sesión** para el mismo repositorio: `automatic` (commits, `/code-review` de Claude Code sobre el candidato local, push, PR a `development` y merge al pasar las pruebas) y el cierre del change record con un PR pequeño sin revisión dedicada, con merge del agente. Después se corre `hive update`, por el marcador de maintainer.
- **Efectos adicionales previstos:** ninguno.
- **Requisito de `AGENTS.md`:** el PR lista los lugares revisados y dice que `global.md` no cambia.
