# Documentación: alinear versiones, Pi y datos derivados

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en [PR #114](https://github.com/JhonHawk/tricell-hive/pull/114), merge `9538de01`, el 2026-10-23; delta aplicado al spec vigente en el cierre |
| Tracker · GitHub Issues | • [#108 — Documentación: alinear versiones, Pi y datos derivados](https://github.com/JhonHawk/tricell-hive/issues/108) |
| Git | Interactivo (D3-A de esta sesión) · `docs/gh-108-doc-drift` → PR → `/code-review` (D4-A) → merge a `development` · parada antes del push |
| Verificación | `rg` por criterio · `go test ./tests/content/ -run TestGuidanceReadingFootprint` · `go vet` · `go test` · unittest de skills |
| Siguiente paso | Ninguno; sin promoción a `master` |

## Solicitud

> #108 ahora

(Respuesta del usuario a D9 en el cierre de #110, 2026-10-23.) La incidencia #108, leída el 2026-10-23, no declara criterios; su «Acción sugerida» dice: «Actualizar los textos y el spec actual, distinguir explícitamente los documentos históricos y añadir comprobaciones de consistencia para recuentos o datos derivados cuando resulte práctico. Sustituir la duración prometida de la suite por una formulación que no dependa de una máquina concreta.»

Decisiones del usuario (2026-10-23):

- D11-A: `SECURITY.md` declara soporte solo para la última versión minor publicada, sin números.

## Objetivo

La revisión del 2026-10-23 encontró siete derivas documentales, y todas siguen en `development` (`2d89940e`): una política de soporte de seguridad que se quedó en 0.2.x, un requisito del spec que fija `pi-subagents@0.74.0` aunque el instalador ya no fija versión, un recuento de skills, cinco frases que llaman obligatoria la lectura de `reporting`, una tabla de huella de lectura caducada, índices de cinco hosts sin marca de historia y una duración de la suite que depende de la máquina. Este cambio corrige el texto sin tocar comportamiento.

## Alcance y aceptación

Incluye los siete puntos de la incidencia. Excluye, con su motivo:

- Comprobaciones automáticas de consistencia: la incidencia las pide «cuando resulte práctico»; quitar los números que envejecen (skills, duración, versiones) evita la deriva sin código nuevo, y la tabla de huella ya tiene su prueba de medición.
- Promoción a `master`: va con la próxima release.

Criterios:

- AC1. Tras el cierre, `_support/openspec/specs/versioned-installation/spec.md` no contiene `0.74.0` ni «pinned pi-subagents» en los requisitos «Pi user-scope install declares pi-subagents» y «Verified release package installation», y describe la fuente `npm:pi-subagents` sin versión, como el instalador (`integrations/pi/package_settings.go:11`). *Falso en la base cuando* el spec vigente tiene 3 apariciones de esos textos. El estado no pasa a Completado sin esta evidencia: el cambio archivado de 2026-10-15 tenía el mismo delta y nunca se aplicó.
- AC2. `SECURITY.md` declara soporte para la última versión minor publicada (las correcciones salen como un parche nuevo de ella) y ninguno para las minor anteriores, sin números de versión, y remite a GitHub Releases para saber cuál es. *Falso en la base cuando* la tabla lista `0.2.x` y `0.1.x`.
- AC3. Ni `llms.txt` ni `_support/docs/architecture/` describen `reporting` como obligatorio antes de toda respuesta. *Falso en la base cuando* `rg -n "before every chat reply|mandatory reporting" llms.txt _support/docs/architecture/` devuelve 6 líneas.
- AC4. `repository-and-distribution.md` no da un número de skills. *Falso en la base cuando* contiene «Eleven authored skills».
- AC5. La tabla de `instruction-resources.md` muestra la medición de `TestGuidanceReadingFootprint` sobre el commit base de este cambio, con su fecha y commit. *Falso en la base cuando* dice «Snapshot measured on 2026-10-12».
- AC6. `_support/docs/harness-engineering/README.md` y `2026-09-20-portable-harness-research.md` abren con una nota que los marca como históricos de cinco hosts y remite a la matriz actual. *Falso en la base cuando* ninguno contiene esa nota.
- AC7. `CONTRIBUTING.md` no promete una duración de la suite. *Falso en la base cuando* contiene «under a minute».

Restricciones: la suite completa sigue pasando; `content/` no cambia, así que la instalación desplegada no cambia; `README.md` ya es correcto y no se toca.

## Entrega

Reutiliza las decisiones de esta sesión para `tricell-hive`: D3-A interactivo (parada antes del push para la validación del usuario; después push, PR, `/code-review`, merge) y D4-A `/code-review`. La carpeta de cambio se versiona al cierre con un PR pequeño sin revisión que fusiona el agente, tras aplicar el delta al spec vigente. Sin `hive update`: no cambia `content/`.
