# Diseño

## Contexto verificado

Leído en `development` 0afaa7c6 el 2026-10-29.

**Las reglas de hoy:**

- `content/skills/flow-build/references/verification.md:50` define "affected" para la iteración: "the project's affected-test selection while iterating … or the changed package's tests when the project has none". El archivo tiene 15,989 bytes, 11 de margen.
- `content/skills/flow-verify/SKILL.md:25` exige la suite completa "before a validation stop, or once on the final candidate when no CI gate runs it".
- `content/skills/flow-verify/SKILL.md:47` dice "Close regression coverage through the project's actual CI gate".
- `content/skills/flow-verify/SKILL.md:14` pide leer `review-and-gates.md` "Before the dedicated code review, the local runtime checks or the human UI handoff".
- `content/skills/flow-verify/references/review-and-gates.md:10`, en la fila "Publication / integration", dice "Inspect actual CI coverage". El archivo tiene 11,500 bytes.
- `content/skills/flow-verify/references/review-and-gates.md:40` corre la revisión dedicada "after the affected tests and the full suite when it runs locally".
- `content/skills/git-workflow/SKILL.md:28` hace obligatorios los hooks y, si no pueden correr, compara el CI con el hook. El archivo tiene 13,269 bytes.
- `site/src/content/docs/flows/verify.md:24-25` describe las pruebas afectadas y la suite completa.

**Evidencia de cam** (hijo `hive-research`, corridas revisadas por el hilo principal):

- Fallaron antes del merge 14 jobs, en 8 de 77 PRs. De esos, 11 se podían detectar localmente:
  - 6 pruebas unitarias o de convención: #226 (corrida 39232905374) y cuatro en #296.
  - 4 cambios de contrato detectados por oasdiff: #237 (39379773229), #240 (39846070019) y #296.
  - 1 de i18n: #214 (39112524802).
- Los otros 3 eran 2 de integración con servicios (#227, #246) y 1 incierto (#272).
- La mediana del CI de un PR es de 234 s, y ver el fallo tarda unos 3.5 min. El tiempo local se estima en 0.5 a 2 min por comando, a partir de los tiempos de los jobs; no se midió en esta máquina.
- El `lefthook.yml` de cam corre en pre-commit biome y oxlint, y en pre-push `turbo run typecheck`. Dice que las pruebas unitarias "stay in CI".

## Enfoque

**Dónde vive la regla.** Una sección nueva, "Local checks before publication", en `review-and-gates.md`, junto a "Local candidate and publication". Ahí tiene espacio (11,500 bytes), y es la referencia que flow-verify lee antes de publicar.

`verification.md` no se toca. Su línea 50 define la selección para la iteración, y la regla nueva define el alcance antes del push. Son dos momentos distintos con un solo hogar cada uno.

**Texto de la regla** (borrador, se ajusta al editar):

> **Condition:** a change that is not mechanical is about to be pushed with commits not yet integrated, in a repository whose CI runs on that push or its pull request, and the user did not skip verify or the check. A push that carries only commits already integrated, such as a merge from the base or a promotion, is not covered.
>
> **Actions:** before the push, run locally:
> 1. the lint, typecheck and unit or convention tests of each package the change touches and of each package in the repository that depends on one of them, with the project's changed-package selection when it has one, such as `turbo run lint typecheck test --filter=...[origin/<base>]`, otherwise each such package's own commands; a full local suite already run on the same candidate counts for its tests;
> 2. each job that the push or its pull request will trigger for the changed paths, read from the workflow files, that is deterministic and needs no external service, secret or deployment, such as a contract-compatibility or generated-artifact check, by the commands it runs.
>
> The service-dependent jobs that this push or its pull request triggers, such as an integration suite against a database, end-to-end jobs and deployments, stay in CI, beyond the affected tests and runtime checks verify already selects; name them CI-only in the report. A suite that CI runs only after merge follows [verification](../../flow-build/references/verification.md#frequency-and-coverage). On a later push, rerun the checks whose packages or paths changed since the last run. A failure here is a defect that blocks the work, under the global rule. A check that cannot run here, such as one whose tool is missing, does not hold the push when a job this push triggers runs it: report it as not run, as [verification](../../flow-build/references/verification.md#report-check-status) sets. These are routine checks; one expected to run longer than 30 minutes is costly, as [verification](../../flow-build/references/verification.md#gate-selection-and-timing) sets.

**Otros cambios:**

- La fila "Publication / integration" de `review-and-gates.md:10` apunta a la sección nueva.
- `review-and-gates.md:40` ("after the affected tests and the full suite when it runs locally") agrega "and the local checks before publication", así la revisión ve el candidato ya verificado.
- `flow-verify/SKILL.md`:
  - el puntero de la línea 14 agrega "before a push";
  - la tabla de la línea 25 suma la fila "Local checks before publication | before each push of a change that is not mechanical | review and gates";
  - el paso 2 de la secuencia antes de validar (`flow-verify/SKILL.md:40`) corre también estos chequeos locales, así tienen un solo lugar en el orden y no repiten la suite completa;
  - la línea 47 cambia "Close regression coverage through the project's actual CI gate" por "Leave to the CI gate only what [local checks before publication] leaves to it", y conserva la excepción de la suite completa antes de validar.
- `git-workflow/SKILL.md`, en "Prepare a coherent delivery", agrega una oración: "Before a push of a change that is not mechanical, run the checks that [review and gates](../flow-verify/references/review-and-gates.md#local-checks-before-publication) sets." Así lo aplica también un push desde `flow-build` cuando verify se salta.
- `site/src/content/docs/flows/verify.md:24-25` suma un punto que describe el comportamiento, y `verify.md:28` y `build.md:52` dicen que la revisión local corre después de la suite completa y de estos chequeos.
- `git-workflow/SKILL.md:28` sigue siendo solo para los hooks.
- `CHANGELOG.md` suma una entrada en `[Unreleased]`, sección Changed.

**Por qué los dependientes:** la prueba de convención del portal en #296 vive en `apps/customer-portal`, que depende del paquete de contratos, y revisa todo el código fuente del portal. Solo la suite del paquete dependiente la detecta.

**Alternativa descartada:** pedir la suite completa antes de todo push (opción C de la investigación). Cuesta lo mismo que la suite del monorepo, y la selección por paquete ya cubre los 11 casos.

## Riesgos

- Cuando cambia un paquete que muchos otros importan, la lista de dependientes crece. El límite de costo está en la regla de los 30 minutos.
- Clasificar un job como determinista y sin servicios es un juicio del agente. El reporte nombra lo que dejó en el CI, así que un error en esa clasificación queda a la vista.
