# Lecturas completas de skills y referencias, y referencias separadas por responsabilidad

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por [PR #213](https://github.com/JhonHawk/tricell-hive/pull/213) (merge `c4b7feaf`, 2026-10-26) |
| Tracker · GitHub Issues | • [#204 — Lecturas de skills y referencias recortadas por el propio modelo](https://github.com/JhonHawk/tricell-hive/issues/204) |
| Git | interactive · PR a `development` · parada antes del push para tu validación |
| Verificación | pruebas de `tests/content` · suite completa · `/code-review` |
| Siguiente paso | Ninguno: cerrado. La eficacia de la regla se observa en las próximas sesiones de Codex y Pi |

## Pedido

> D2-A Planear #204 (Rec.)

> Tambien podemos hacer que las referencias sean más pequeñas, no recortando, si no separando responsabilidades en archivos más pequeños

> D3-C Regla + separar ya — "La regla y, en este mismo plan, separar las 4 referencias de más de 16 KB (delivery-decisions, plan-format, diagram-grammar, verification) con prueba de tamaño."

El issue [#204](https://github.com/JhonHawk/tricell-hive/issues/204) no declara criterios de aceptación (leído el 2026-10-26).

## Objetivo

En dos sesiones reales del 2026-10-26 el modelo leyó skills y referencias de Hive recortando su propia salida: Codex fijó `max_output_tokens` por debajo del tamaño del archivo (13 lecturas truncadas) y Pi leyó con `cut -c`, que corta cada línea. Las reglas que quedan después del corte no se pueden seguir; ese recorte es una causa posible de los cierres incompletos (#208) y de las preguntas solo en texto (#207).

El cambio añade a la guía global una regla de lectura completa y separa las cuatro referencias que pasan de 16 000 bytes en archivos de una sola responsabilidad. También agrega una prueba que mantiene cada referencia dentro de ese límite, el mismo que ya rige para `SKILL.md`. El resultado esperado es que un agente lea entera cada instrucción que sigue y que ninguna referencia crezca hasta volver probable un corte.

## Alcance y aceptación

**Incluye:** la regla de lectura completa en `content/guidance/global.md` y sus punteros; la prueba de tamaño de referencias; la separación de `flow-plan/references/delivery-decisions.md`, `flow-plan/references/plan-format.md`, `flow-report/references/diagram-grammar.md` y `flow-build/references/verification.md`; los enlaces, el sitio y la documentación de arquitectura que nombran esos archivos.

**Excluye:**
- Recortar o reescribir el contenido de las referencias. D3-C pide separar responsabilidades, no recortar.
- Un mecanismo de host, como un hook o un adaptador, que fuerce la lectura completa. AGENTS.md pide primero una regla y una limitación documentada.
- #207 y #208: se reevalúan después de este cambio.

**Restricciones:** ninguna regla pierde su intención ni queda repetida en dos archivos (AGENTS.md, una casa canónica); cada `SKILL.md` sigue en 16 000 bytes o menos; la suite de `tests/` sigue pasando; el PR declara el cambio de tamaño de `global.md` y la lista de lugares revisados.

**Otras restricciones:**
- Cada enlace y ancla hacia una sección movida (en `content/`, `site/`, `_support/docs/` y `AGENTS.md`) resuelve a un encabezado que existe en el archivo que ahora la contiene.
- La release valida y despliega los archivos nuevos en cada host. Ningún archivo sale de la release, porque los cuatro originales conservan su nombre. Los enlaces relativos entre archivos se validan al planear (`tooling/management/references.go:23-46`); las anclas no.
- `README.md` y `llms.txt` no cambian: no nombran estos archivos.

**Sin deltas de requisitos:** el cambio altera la guía, las pruebas y la distribución de texto entre archivos. No cambia ningún requisito de `_support/openspec/specs/`. La huella de lectura declarada (`context-delivery`, "Declared source footprint") sigue igual en su contrato; solo cambian sus listas de fuentes.

- AC1. `content/guidance/global.md` pide leer completos el `SKILL.md` de una skill, sus referencias y un archivo de rol: sin topes de salida por debajo del tamaño del archivo, sin recortar líneas, sin filtrar, ni directo ni dentro de un script, y siguiendo desde donde se cortó cuando la herramienta trunca. *Falso en la base cuando* `global.md` en `2e45e292` no tiene esa regla: solo pide releer `SKILL.md` tras compactar (`:80`), y la regla de scripts no trata topes ni recortes (`:81`).
- AC2. Al armar el brief de un hijo, `flow-research/references/delegation.md` remite a esa regla para las fuentes que el hijo lee. *Falso en la base cuando* `delegation.md:39` dice solo "The child reads these sources before dependent work", sin pedir lectura completa.
- AC3. `go test ./tests/content/` falla con la ruta y los bytes cuando un archivo bajo `content/skills/*/references/` pasa de 16 000 bytes. *Falso en la base cuando* `skill_size_test.go` solo mide `SKILL.md`.
- AC4. Ninguna referencia pasa de 16 000 bytes. Las cuatro grandes quedan separadas por responsabilidad, cada archivo nuevo con su propia condición de lectura y sin recortar contenido. *Falso en la base cuando* `delivery-decisions.md` (21 084), `plan-format.md` (20 514), `diagram-grammar.md` (17 663) y `verification.md` (16 692) pasan el límite.
- AC5. La regla HA-SK-02 de `harness-audit` marca una referencia de más de 16 000 bytes, igual que hoy marca un `SKILL.md` largo. *Falso en la base cuando* HA-SK-02 (`skills-and-agents.md:8`) limita solo `SKILL.md` y no contiene "each reference stays at or under 16,000 bytes".

## Entrega

- **Modo:** `interactive` (D4-B, 2026-10-26) sobre `tricell-hive`, base `development`. Rama de trabajo y commits locales; verificación y `/code-review`; parada antes del push para tu validación; después, push, PR a `development` y merge cuando pasen las pruebas.
- **Revisión:** `/code-review` de Claude Code sobre el candidato local, antes de pedirte la validación (D5-A, 2026-10-26).
- **Registro de cambio:** D7-A (2026-10-26): se versiona una vez al cerrar, con un PR de cierre pequeño a `development` sin revisión dedicada que mergeo yo, como en #200.
- **Después del merge:** refresco de la instalación local con `hive update` (marcador de mantenedor en AGENTS.md); no es una release.
- **Revisión de seguridad:** no aplica; el cambio no cruza un límite de seguridad.
