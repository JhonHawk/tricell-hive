# Tareas

Base: `development` en `27acf125`; rama propuesta `feat/off-repo-facts-and-effects`. T1 a T4 corren en el hilo principal: cada una edita tres archivos o menos y su redacción sigue las reglas de autoría de `AGENTS.md`. T5 va a un hijo `hive-write-docs` porque edita hasta siete archivos de documentación a partir de textos ya fijados. Las frases literales que buscan las comprobaciones son las de `design.md` «Proposed shape».

### T1 — Datos externos medidos al planear

- [ ] La referencia nueva define el sistema externo, los datos externos, su medición de solo lectura, el hijo por sistema, lo que pide el brief y el contraste
- [ ] `flow-plan` paso 2 manda leerla después de elegir el enfoque, cuando el enfoque lee o cambia un sistema externo
- [ ] `delegation.md` nombra esa recolección como tercera razón para delegar y entre los hijos que corren a la vez

**Closes:** AC12, AC13, AC14

**Depends on:** ninguno

**Locations:** `content/skills/flow-plan/references/outside-the-repository.md` (nuevo), `content/skills/flow-plan/SKILL.md`, `content/skills/flow-build/references/delegation.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** la cabecera y la sección «External facts» de la referencia, el puntero del paso 2 y la tercera razón de `delegation.md`, con «when neither applies» cambiado a «when none applies». No cambian «Prerequisites», `flow-build` paso 1 ni los roles.

**Verification:**
- AC12: `rg -n -i "external fact" content/skills/flow-plan/SKILL.md content/skills/flow-plan/references/outside-the-repository.md` encuentra al menos una línea en cada archivo; la del `SKILL.md` está en el paso 2 y enlaza la referencia; la cabecera de la referencia contiene «credential» y nombra lo que la entrega ya usa.
- AC13: `rg -n -i "one read-only child per external system" content/skills/flow-plan/references/outside-the-repository.md content/skills/flow-build/references/delegation.md` encuentra una línea en la referencia y una en `delegation.md` bajo «When a child does the work»; «When children run at the same time» de `delegation.md` nombra los `external facts`; y `rg -n "when neither applies" content/skills/flow-build/references/delegation.md` no encuentra nada.
- AC14: `rg -n -i "not measured" content/skills/flow-plan/references/outside-the-repository.md` encuentra una línea, y `rg -n -i "verbatim" content/skills/flow-plan/references/outside-the-repository.md` otra, en la sección «External facts».
- `go test ./tests/content/...` pasa (tamaños y enlaces).

### T2 — Ronda de efectos por clase y excepciones en la guía global

- [ ] La referencia define la clase de efecto, lo que deja fuera, los seis casos que conservan su pregunta, hasta dónde vale la respuesta y el registro de cada efecto cubierto
- [ ] `flow-plan` paso 4 pregunta una decisión por clase en cada plan, con ronda propia cuando entrega y revisión ya están respondidas
- [ ] `global.md` nombra las dos excepciones a la pregunta propia de un efecto crítico, y «Delivery» nombra la ronda de efectos

**Closes:** AC15, AC16, AC22

**Depends on:** T1 (escribe en el mismo archivo de referencia)

**Locations:** `content/skills/flow-plan/references/outside-the-repository.md`, `content/skills/flow-plan/SKILL.md`, `content/guidance/global.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** la sección «Effects round», la frase del paso 4 y las frases de las líneas 14 y 98 de `global.md`. No cambian la definición de efecto crítico, la línea 22 ni la 68.

**Verification:**
- AC15: `rg -n -i "class of effect" content/skills/flow-plan/SKILL.md content/skills/flow-plan/references/outside-the-repository.md` encuentra al menos una línea en cada archivo; la sección «Effects round» nombra la ronda propia, la exclusión de lo que la entrega ya autoriza y la tarea operativa.
- AC16: cada una de estas búsquedas encuentra una línea dentro de «Effects round»: `rg -n -i "production deployment"`, `rg -n -i "interrupts"`, `rg -n -i "hard-to-reverse|hard to reverse"`, `rg -n -i "replacing infrastructure"`, `rg -n -i "did not create"`, `rg -n -i "shared environment"`, todas sobre la referencia.
- AC22: `rg -n -i "effects round" content/guidance/global.md` encuentra dos líneas, una en «Scope and authorization» y otra en «Delivery»; `rg -n -i "unforeseen-spend amount" content/guidance/global.md` encuentra una; y la línea de «Scope and authorization» contiene «only spends or is reversible».
- Tamaño: `wc -c content/guidance/global.md` (hoy 46,635); el reporte y el pull request nombran la cifra nueva.
- `go test ./tests/content/...` pasa.

### T3 — Deploy, plantilla, resumen y listas de contenido del plan

- [ ] `flow-deploy` paso 4 y `flow-plan` paso 3 no repiten, en la tarea de promoción de un plan, lo que una clase respondida en ese plan cubre, y siguen preguntando producción aparte
- [ ] La plantilla `proposal.md` muestra la línea de un dato externo y la sección `## Effects`
- [ ] `plan-handoff.md` y las listas de `flow-plan` llevan los datos no medidos y las clases respondidas

**Closes:** AC23, AC19

**Depends on:** T2 (usa sus términos y edita `flow-plan/SKILL.md` después)

**Locations:** `content/skills/flow-deploy/SKILL.md`, `content/skills/flow-plan/SKILL.md`, `content/skills/flow-plan/assets/change-folder/proposal.md`, `content/skills/flow-plan/references/plan-handoff.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** un puntero en `flow-deploy` paso 4 y otro en `flow-plan` paso 3; el comentario de «Verified context» y la sección nueva en la plantilla; los puntos 6 y 9 del resumen; las listas de «Where the plan lives» y «Change folder». No cambia la pregunta de producción ni sus acciones de recuperación. Son cuatro archivos con una frase o una sección cada uno; se quedan en el hilo principal porque comparten términos con T2.

**Verification:**
- AC23: `rg -n -i "effects round" content/skills/flow-deploy/SKILL.md` encuentra una línea dentro del paso 4 que nombra `## Effects` y deja producción aparte; el comentario de `## Effects` en la plantilla dice que el encabezado queda en inglés; y `rg -n -i "effects round|class of effect" content/skills/flow-plan/SKILL.md` encuentra una línea en el paso 3.
- AC19: `rg -n -i "external fact" ` y `rg -n -i "class of effect"` encuentran cada una al menos una línea en `content/skills/flow-plan/assets/change-folder/proposal.md` y en `content/skills/flow-plan/references/plan-handoff.md`; `rg -n "^## Effects" content/skills/flow-plan/assets/change-folder/proposal.md` encuentra una; «Where the plan lives» y «Change folder» de `flow-plan/SKILL.md` nombran ambos términos.
- `python3 content/skills/flow-plan/scripts/check_change.py _support/openspec/changes/off-repo-facts-and-effects` termina en 0, `wc -c content/skills/flow-plan/SKILL.md` queda por debajo de 16,000 y `go test ./tests/content/...` pasa.

### T4 — Monto para gasto imprevisto en el encargo desatendido

- [ ] El paso 2 dice o pregunta el monto en toda activación, sin contarlo entre las respuestas abiertas ni detener el trabajo
- [ ] El paso 4 cuenta las clases del plan entre las respuestas en vigor y deja sin autorizar las no respondidas
- [ ] La viñeta de destino y techo exactos del paso 6 nombra su excepción con sus tres condiciones

**Closes:** AC24, AC25

**Depends on:** T2 (nombra las clases del plan y la excepción de `global.md`)

**Locations:** `content/skills/unattended-delegation/SKILL.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** los pasos 2, 4 y 6, como fija `design.md`. No cambia la lista de lo prohibido ni la regla de bucles.

**Verification:**
- AC24: `rg -n -i "unforeseen-spend amount" content/skills/unattended-delegation/SKILL.md` encuentra una línea en el paso 2, que dice que la pregunta no es una respuesta abierta del paso 4 y que sin respuesta el monto es cero; el paso 4 contiene `class of effect` o `effects round` y dice que una clase sin responder no añade autorización, no deja el trabajo en `hold` y que el paso 6 decide sus efectos.
- AC25: `rg -n -i "unforeseen spend" content/skills/unattended-delegation/SKILL.md` encuentra una línea en la misma viñeta que contiene «exact target and ceiling», y esa viñeta nombra el comando registrado, el vencimiento del encargo, el total dentro del monto y qué pasa cuando no se puede leer el precio.
- `wc -c content/skills/unattended-delegation/SKILL.md` queda por debajo de 16,000 y `go test ./tests/content/...` pasa.

### T5 — Sitio y registro de cambios

- [ ] Las páginas de plan y deploy describen la medición de datos externos y la autorización por clase
- [ ] Las páginas que describen el encargo desatendido nombran el monto para imprevistos
- [ ] `CHANGELOG.md` lo nombra en `[Unreleased]`

**Closes:** AC11

**Depends on:** T1, T2, T3, T4 (describe sus textos finales)

**Locations:** `site/src/content/docs/flows/plan.md`, `site/src/content/docs/flows/deploy.md`, `site/src/content/docs/flows/overview.md`, `site/src/content/docs/flows/build.md`, `site/src/content/docs/skills.md`, `site/src/content/docs/index.mdx`, `CHANGELOG.md`

**Execution:** `hive-write-docs`

**Changes:** las páginas describen la conducta y enlazan la skill, sin copiar sus reglas. En `flows/build.md`, `flows/overview.md` e `index.mdx` se edita solo la frase que quede inexacta, y el reporte de la tarea dice cuáles quedaron sin cambio y por qué. `README.md` y `llms.txt` no cambian.

**Verification:**
- AC11: `rg -n -i "external fact|class of effect|unforeseen" site/src/content/docs CHANGELOG.md` encuentra líneas en `flows/plan.md`, `flows/deploy.md`, al menos una página del encargo desatendido y `CHANGELOG.md`.
- El sitio construye en local con el comando de `site/package.json`, y las páginas cambiadas se abren desde el build.

## Verificación y revisión humana

| Check | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Comprobación de cada tarea | Al cerrar la tarea, en local | Los `rg` y comandos de la tarea | Cubierta por la implementación |
| Suite completa | Candidato final, en local | `go test ./...` y `python3 -m pytest tests/skills` | Cubierta por la implementación |
| Carpeta de cambio | Candidato final, en local | `check_change.py` termina en 0 | Cubierta por la implementación |
| Sitio | Candidato final, en local | El build del sitio termina sin errores | Cubierta por la implementación |
| Revisión de código | Candidato final, antes del push | `/code-review` de Claude Code sobre el diff | La respuesta de revisión (D5 A) |

Sin parada de validación: la entrega es automática (D4 B). El efecto que pidió el usuario, sesiones que no se detienen por datos tardíos ni por un gasto mínimo, solo se ve en sesiones reales que corran las reglas nuevas; las filas del 2026-11-08 de la rúbrica de revisión lo miden.

## Checkpoint

## Estado de la revisión y avance
