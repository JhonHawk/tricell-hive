# Design

## Approaches

- **E1, solo la guía:** sin código; cada trabajo con interfaz sigue en `open` + `sync` + `evidence` + `close` (4 o más llamadas) y la iteración sin número.
- **E2, comando `work capture` más recorte de la guía (elegido, D1 A):** una llamada por lote de capturas, iteración numerada; cuesta código Go y tests.
- **E3, visor como galería:** quitar etapas, tareas y documentos del binario y de la interfaz; el cambio más grande, rompe requisitos del spec y es el más difícil de revertir.

E2 es el único que deja una llamada por captura y numera la iteración, que es lo que el usuario usa del visor.

## Decisions

- **Cierre (D2 A):** un `work close` en flow-close. El cierre automático por fin de sesión cambiaría el contrato y partiría en dos un trabajo que sigue en otra sesión.
- **Número de iteración (D3 A):** en el título de la vista, sin cambiar la interfaz.
- **Supuestos (decisiones menores, sin pregunta):**
  - El comando vive bajo `hive view work capture`, junto a los demás verbos semánticos (`tooling/cli/view_work.go`).
  - El sufijo es «· #N», neutral al idioma, porque el binario se distribuye y la interfaz del visor tiene su propio idioma.
  - La pantalla es el título base (`--view-title`). No hay clave aparte, porque ningún registro del servidor la guardaría (revisión H3).
  - La etapa por defecto es `verify`; el flujo puede pasar `--stage build`.
  - **Estados de espera en el cierre (revisión H1):** `workflow-lifecycle` no cambia. flow-close conserva un solo `work sync` para dejar un trabajo con capturas en `in_review` (PR que espera merge o revisor humano), `paused` o `awaiting_validation`, y `work close` para el trabajo integrado o abandonado. Un trabajo sin capturas no existe en el visor y no necesita nada.
  - El requisito «Bounded publication interface» (`walkthrough-viewer/spec.md:512`) se cumple sin cambios: la validación humana sigue en la pregunta nativa del host, y las comprobaciones de evidencia, en el binario.

## Proposed shape

```
hive view work capture --view-title TEXT --image FILE[:caption] ...
    [--stage build|verify] [--key KEY] [--title TEXT] [--summary TEXT]
    [--candidate COMMIT] [--work-id ID]
```

1. **Trabajo:** si la sesión no tiene trabajo abierto, lo abre como `work open`. La clave sale de `--key` (la guía pide el ticket o la carpeta del cambio) y, si falta, de la rama actual; en `HEAD` separado sin `--key` responde `key_required`. Título por defecto: la clave. Resumen: `--summary` (la guía pide un párrafo en la primera captura) o, por defecto, «UI captures for <key>». Con un solo trabajo abierto lo reutiliza; con varios exige `--work-id`, como hoy.
2. **Etapa:** sin ocurrencia, arranca una en `in_progress` con la razón fija «captures published». En pausa (por ejemplo, porque otra etapa tomó el primer plano), la reanuda a `in_progress`. Terminada, abre una ocurrencia nueva enlazada con `previous-stage` igual a la terminada. Todo esto se resuelve al enviar, contra el estado remoto, no al encolar (revisión H10). El cambio de primer plano sigue las reglas actuales de `sync`.
3. **Candidato:** `--candidate` o, por defecto, `HEAD`, resuelto al ejecutar el comando (también cuando se encola), con la rama y la base que hoy calcula `evidenceCreateCandidate`. En `HEAD` separado o sin `refs/remotes/origin/HEAD`, responde `candidate_branch_required` con el paso siguiente, sin publicar a medias (revisión H9).
4. **Iteración (revisión H3):** se calcula al enviar, leyendo las vistas del trabajo en el servidor. Se queda con las de título base igual (quitando el sufijo ` · #N`), toma la de N más alto como `supersedes` y publica N+1; si no hay ninguna, N=1. No depende del journal local, que se compacta y no cruza sesiones.
5. **Repetición (revisión H11):** si la última iteración de esa pantalla tiene las mismas imágenes y leyendas y el mismo candidato, responde `unchanged` sin crear iteración.
6. **Cola (revisión H3):** `capture` se encola como acción propia (`validWorkAction`, `validateWorkQueue`, `flushEntry`), con sus imágenes congeladas como hoy en `evidence`. La apertura, la etapa, N y `supersedes` se resuelven al enviar, en orden. Así dos capturas offline de la misma pantalla salen como «#1» y «#2».
7. **Recibo:** el recibo compacto de `evidence`, más `iteration` y `screen` (el título base).
8. **Contrato:** `capture` pertenece a los flujos `build` y `verify` del contrato por partes (`walkthrough-viewer/spec.md:271`). `hive view work capture --help` es autosuficiente, para que la guía no mande leer el contrato.

`work evidence` y los demás verbos no cambian.

## Impact

- **Binario:** `tooling/cli/view_work.go` (despacho y opciones), `view_work_evidence.go` (carga y candidato reutilizados), `view_work_queue.go` y `view_work_journal.go` (acción encolable nueva), `view_work_lifecycle.go` (reanudar y volver a entrar en la etapa), `view_work_contract.go` (texto y flujos). Nada cambia en `tooling/viewer/` ni en la interfaz.
- **Guía:**
  - `content/skills/flow-walkthrough/`: SKILL, con su descripción, y la referencia.
  - `flow-plan/SKILL.md:8`, `flow-plan/references/plan-handoff.md:21`.
  - `flow-build/SKILL.md:8,15,20`.
  - `flow-verify/SKILL.md:8,15,17`, `flow-verify/references/ui-verification.md:25`.
  - `flow-close/SKILL.md:10,12,21`.
  - `workspace-conventions/references/artifact-layout.md:27`.
  - `content/agents/review/hive-review-ux.md:14`, solo si cambia su sentido.
- **Documentación:**
  - `README.md:480-486` y `llms.txt:57`.
  - `site/src/content/docs/viewer.md`, `flows/{overview,plan,build,verify,close,research}.md`, `skills.md:34`, `index.mdx:104`.
  - `_support/docs/architecture/deployment-manager.md:176,195` e `instruction-resources.md:17`.
- **Otros consumidores:** las sesiones abiertas con la guía anterior siguen funcionando, porque ningún verbo existente cambia. Ninguna superficie de producción.

## Recovery

Revertir el merge restaura la guía y quita `work capture`. Las capturas publicadas con `capture` quedan como vistas normales con «· #N» en el título; no hay datos que migrar.
