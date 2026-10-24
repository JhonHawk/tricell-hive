# Tareas · engineering-skills-adoption

Base del cambio: `3394a43b` (confirmada al arrancar el build el 2026-10-23, rama `feat/engineering-skills-adoption`).

## Diseño en vigor

- **K1–K4** se integran en los pasos que ya existen de `diagnosis.md`, sin pasos nuevos y con redacción propia:
  - Paso 1, tras la compuerta de reproducción: forzar el fallo cambiando código, un fixture o datos sirve para construir la reproducción y no es un arreglo; un fallo forzado (*forced failure*) se confirma con un diff contra la copia sin cambiar antes de confiar en él. Una vez reproducido, reducir la reproducción a la más pequeña (*smallest reproduction*) que aún muestra el síntoma y compararla con el caso más cercano que funciona (*closest case that works*). Ante un fallo intermitente (*intermittent*), subir la tasa de reproducción antes de formular hipótesis e informar la tasa observada; coherente con `global.md:31`, que pide guardar la salida del fallo antes de repetir.
  - Paso 3: cambiar una sola variable (*one variable*) por experimento; marcar la instrumentación temporal con una etiqueta única (*unique tag*).
  - Paso 5: la retirada de medidas temporales incluye buscar esa etiqueta y quitar todo lo que la lleva, en la misma frase que ya existe, para que no haya dos reglas de retirada.
- **K9 (D16-A, D17-A):** `ticket-body.md` es el único hogar de la regla del cuerpo del ticket. Se lee antes de escribir el cuerpo de cualquier ticket que Hive abre. Contenido: título por comportamiento; comportamiento actual (*current behavior*) y esperado (*expected behavior*); evidencia con `path:line` y el commit o la fecha en que se leyó, como evidencia y no como instrucción; criterios observables y falsos en el estado actual; fuera de alcance (*out of scope*). Para un hallazgo incidental, remite a la regla global de hallazgos incidentales (`global.md:31`), sin copiar sus campos. Si el repositorio tiene plantilla de issue, se sigue y esta regla completa lo que deja abierto. El idioma y los secretos ya los rigen las reglas globales; no se repiten. `flow-research/SKILL.md` y el paso 5 de `flow-close/SKILL.md` llevan solo un puntero.
- **F1 (D15-A):** en `agent-delivery.md:9` se quita el paréntesis entero, «(gentle-ai, superpowers, mattpocock-skills, OpenSpec)», y queda «every reference workflow inspected».
- **Créditos (D18-A):** entrada nueva en `inspiration.md`, con el formato de las demás, tras Superpowers; `CONTRIBUTING.md` la añade al final de su enumeración y conserva «in that order».
- **Changelog:** las entradas de #118 y #120 citan su PR, porque no tienen ticket; la de este cambio va sin número.

## Tareas

### T1 — Técnicas de diagnóstico en `diagnosis.md`

- [x] Los pasos 1, 3 y 5 de `diagnosis.md` contienen K1–K4 sin contradecir la compuerta de reproducción.

**Closes:** AC1, AC2, AC3, AC4, AC5, AC6.

**Depends on:** none.

**Locations:** `content/skills/flow-research/references/diagnosis.md`.

**Execution:** main thread; la redacción es corta y el brief sería más largo que el cambio.

**Test approach:** check — las seis búsquedas de AC1–AC6 y la validación de contenido.

**Changes:** ver «Diseño en vigor», K1–K4.

**Verification:** cada búsqueda de AC1–AC6 encuentra al menos una línea en `diagnosis.md`; la lectura del archivo muestra que el paso 1 conserva «form no hypothesis and edit no code» y que el texto nuevo presenta el fallo forzado como parte de la reproducción; `go test ./tests/content/... ./integrations/mdlinks/...` pasa.

### T2 — Referencia del cuerpo de tickets

- [x] Existe `ticket-body.md` con sus apartados y los dos sitios que abren tickets enlazan a ella.

**Closes:** AC7, AC8.

**Depends on:** none.

**Locations:** `content/skills/flow-research/references/ticket-body.md` (nuevo), `content/skills/flow-research/SKILL.md:38` (párrafo del ticket en «Deliver the finding»), `content/skills/flow-close/SKILL.md:26` (paso 5).

**Execution:** main thread; cambio corto acoplado al diseño D16–D17.

**Test approach:** check — las búsquedas de AC7 y AC8 y la validación de enlaces.

**Changes:** ver «Diseño en vigor», K9.

**Verification:** las cinco búsquedas de AC7 encuentran cada una al menos una línea en `ticket-body.md`; `rg -c 'ticket-body.md'` encuentra el puntero en `flow-research/SKILL.md` y en `flow-close/SKILL.md`; `rg -n -i 'severity' content/skills/flow-research/references/ticket-body.md` solo aparece dentro del puntero a la regla global; `go test ./tests/content/... ./integrations/mdlinks/... ./tooling/management/...` pasa.

### T3 — Nombres de referencias y créditos

- [x] `agent-delivery.md` ya no nombra referencias, y la página de créditos y `CONTRIBUTING.md` incluyen «Skills For Real Engineers».

**Closes:** AC9, AC10.

**Depends on:** none.

**Locations:** `_support/docs/architecture/agent-delivery.md:9`, `site/src/content/docs/inspiration.md`, `CONTRIBUTING.md:78-79`.

**Execution:** main thread; tres ediciones de una línea.

**Test approach:** check — las búsquedas de AC9 y AC10 y el build del sitio.

**Changes:** ver «Diseño en vigor», F1 y créditos.

**Verification:** `rg -c -i 'gentle-ai|superpowers|mattpocock|OpenSpec' _support/docs/architecture/agent-delivery.md` no encuentra nada; las tres búsquedas de AC10 encuentran cada una al menos una línea; `CONTRIBUTING.md` conserva «in that order»; `pnpm run build` en `site/` construye sin errores.

### T4 — Entradas del changelog

- [x] `CHANGELOG.md` tiene bajo `[Unreleased]` las entradas de #118, #120 y este cambio.

**Closes:** AC11, AC12, AC13.

**Depends on:** T1, T2, T3 (la entrada de este cambio describe lo que entregaron).

**Locations:** `CHANGELOG.md` (sección `[Unreleased]`, subsección «Changed»).

**Execution:** main thread.

**Test approach:** check — las búsquedas de AC11–AC13 sobre el bloque `[Unreleased]`.

**Changes:** tres entradas en «Changed», con el estilo de las existentes; ver «Diseño en vigor», changelog.

**Verification:** con `awk '/^## \[Unreleased\]/{f=1;next} /^## \[/{f=0} f' CHANGELOG.md`, `rg -c '#118'`, `rg -c '#120'`, `rg -c -i 'diagnos'` y `rg -c -i 'ticket'` encuentran cada una al menos una línea.

## Verificación compartida

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite local | Candidato final, checkout local | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'`, `pnpm run build` en `site/`; todo pasa | Implementación |
| Verificación por tarea | Tras cada tarea | `hive-verify-task` por tarea con los criterios de su `Closes:` | Implementación |

No hay efecto de UI renderizada que requiera la compuerta de UI: los cambios del sitio son de texto.

## Revisión del plan y progreso

- 2026-10-23: plan escrito (proposal `31f4f404`, tasks `88133eb4`).
- Revisión: `hive-review-plan`, dominio «guía distribuida y documentación», nativo; unos 115k tokens y 148 s según el host. Cuatro bloqueantes de verificación aceptados y aplicados como los propuso el revisor: AC1 dividido en tres criterios (ahora AC1–AC3), patrones con frase para la etiqueta y el fallo forzado, AC5 dividido en contenido (AC7) y punteros (AC8) con una búsqueda por apartado, el changelog verificado número a número dentro de `[Unreleased]` con la entrada de este cambio sin número (AC11–AC13), y el enlace de créditos dentro de AC10. No bloqueantes aplicados: el fallo forzado presentado como parte de la reproducción, la etiqueta enlazada con la retirada del paso 5, K2 coherente con `global.md:31`, un puntero a la regla global en vez de copiar la gravedad y la acción, las páginas `close.md` y `research.md` fijadas fuera de alcance, el requisito del PR en «Entrega» y el paréntesis completo de F1. Sin nueva ronda: las correcciones aplican las propuestas del propio revisor. Límite: el revisor no verificó en la web la URL ni el nombre del proyecto; los confirma el README del clon local (`# Skills For Real Engineers`) y el origen del clon (`github.com/mattpocock/skills`).
- 2026-10-23, build: T1–T4 verificadas por `hive-verify-task` (nativo, en el mismo checkout; implementador y verificador con el modelo configurado, no observado), todos los criterios cumplidos, sin rondas de corrección. Commits: T1 `f5c9cc63`, T3 `5d72cf0d`, T2 `29051003`, T4 `76f4fcd8`. Suite final: `go vet ./...`, `go test ./...`, pruebas de skills y build del sitio pasan. PR #121 fusionado como `5f249c60`. Observación del verificador de T1, sin cambio: el paso 1 no dice de forma explícita que forzar el fallo es una excepción a «edit no code»; el texto lo presenta como parte de construir la reproducción.
