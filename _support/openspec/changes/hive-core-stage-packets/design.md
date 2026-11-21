# Design

## Approaches

Comparadas el 2026-11-21 contra pstack, contra Hive en `2bb96478` y contra Hive actual.

| Opción | Qué es | Sobrevive a la compactación | Lecturas por etapa | Deriva | Veredicto |
|---|---|---|---|---|---|
| 1. `hive-core` como almacén; el índice global se queda | 62 skills ocultas pasan a ser 1 | sí: el puntero vive en `global.md`, que todos los hosts reconstruyen | igual que hoy | baja | base de la elegida |
| 2. Flujos de `2bb96478` reescritos al estilo pstack, con `sheets/` propias y repetición por fase | cada flujo con sus hojas | no: el puntero vive en el cuerpo de la skill, que se pierde en 6 de 7 hosts | menos lecturas, pero no cabe en 16 KB | alta (238–536 KB de copias) | descartada |
| 3. Híbrido: 1, más las 15 hojas de una sola fase dentro de su flujo | mixto | sí | casi igual que hoy | baja, con una prueba | no gana bytes |
| **4. (elegida) 1, más paquetes generados al desplegar, la línea `Rules:` y la línea `Reply:`** | el fuente conserva un solo hogar y el despliegue concatena | sí: el índice global más el paquete que nombra el checkpoint | una por etapa | sin copias a mano, porque el manager genera | elegida (respuesta A) |

Por qué gana la opción 4:
- Es la única que ataca la palanca medida: más de 19 lecturas fijas por trabajo.
- Conserva las dos ganancias del reforge: el índice que sobrevive a la compactación y un solo hogar por regla.
- Reutiliza el patrón existente de `agents/openai.yaml`: un archivo sintético, registrado, que se rechaza si se escribe a mano.
- Las hojas usan `references/`, que el manager ya despliega y valida, incluso anidado.

## Decisions

Ver `proposal.md` «Decisions»: D1–D9, la entrega, la revisión y los supuestos.

## Proposed shape

**Fuente (E1)**

```
content/skills/hive-core/
  SKILL.md                    # name, description, disable-model-invocation: true; ~400 B; no lista hojas
  references/<nombre>.md      # una por hoja (questions.md, git-commit.md…); sin frontmatter; # título, luego "Read …"
  references/<nombre>/…       # referencias propias de una hoja (p. ej. references/questions/unanswered.md)
  scripts/<nombre>/…          # scripts/checks/run_check.py, scripts/trackers/{jira_rest,work_time}.py
```

`global.md`, medido sobre `7f6277e0`: ~29,018 B tras la reescritura. Quedan ~980 B para la alineación del índice (T1), la frase del paquete (T7) y el orden tras compactar (T8). T1, T7 y T8 miden su efecto; si una tarea no cabe, recorta la línea de índice que alarga, nunca una regla.

**Rewrite map (T1)**

| Desde | Forma vieja | Forma nueva |
|---|---|---|
| `global.md`, roles | `skill:hive-rule-<n>/SKILL.md[#a]` | `skill:hive-core/references/<n>.md[#a]` |
| `global.md` | `skill:hive-rule-<n>/references/<f>` | `skill:hive-core/references/<n>/<f>` |
| flujos, referencias de flujo | `../hive-rule-<n>/SKILL.md[#a]` | `../hive-core/references/<n>.md[#a]` |
| flujos | `../hive-rule-<n>/references/<f>` | `../hive-core/references/<n>/<f>` |
| playbooks, referencias anidadas | `../../hive-rule-<n>/SKILL.md[#a]` | `../../hive-core/references/<n>.md[#a]` |
| entre hojas | `../hive-rule-<m>/SKILL.md[#a]` | `<m>.md[#a]` |
| hoja → sus referencias | `references/<f>` | `<n>/<f>` |
| hoja → sus scripts | `scripts/<f>` | `../scripts/<n>/<f>` |
| hoja → otra skill | `../<skill>/…` | `../../<skill>/…` |
| referencia de hoja → su hoja | `../SKILL.md[#a]` | `../<n>.md[#a]` |
| referencia de hoja → sus scripts | `../scripts/<f>` | `../../scripts/<n>/<f>` |
| referencia de hoja → otra hoja | `../../hive-rule-<m>/SKILL.md` | `../<m>.md` |
| referencia de hoja → otra skill | `../../<skill>/…` | `../../../<skill>/…` |
| bloques de código | `<hive-rule-checks-dir>/scripts/run_check.py` | `<hive-core-dir>/scripts/checks/run_check.py` |
| prosa, pruebas, docs | `hive-rule-<n>` como nombre suelto | `<n>` o «la hoja `<n>`», según la frase |
| repositorio (pruebas, docs, site) | `content/skills/hive-rule-<n>/…` | `content/skills/hive-core/references/<n>.md` o `…/<n>/…`, `…/scripts/<n>/…` |

Las anclas se conservan: cada hoja mantiene sus encabezados, y `references.go` valida los enlaces antes de instalar. Los nombres viejos siguen aceptados solo en la auditoría (`leafFormerNames`) y en el archivo histórico.

**Paquetes (E2)**

- Manifiesto: `<flujo>/assets/stage.json`, un objeto `{"stage": ["hive-core/references/questions.md", …]}`. En `flow-verify` es `{"stage-entry": […], "stage-exit": […]}`.
- Generación: el manager añade un origen generado `content/skills/<flujo>/<clave>.md`:
  - lo marca `generatedSource`;
  - sus destinos los calcula `stagePacketTargets` junto a `skillPolicyTargets` (`ownership.go:154`, `plan.go:1167`), para cada host que instala el flujo;
  - lo produce `renderStagePacket`, que quita el frontmatter, antepone `## <nombre>` a cada miembro y reescribe los enlaces, incluidas las definiciones de estilo referencia:
    - a otro miembro del mismo paquete sin ancla, al encabezado de ese miembro en el paquete;
    - a otro miembro con ancla, o un ancla interna de un miembro, a la hoja en `hive-core` con su ancla;
    - el resto, relativo a la carpeta del flujo.
- Validación: `validateRelease` genera cada paquete, lo añade a la lista de archivos y anclas de `validateInstructionReferences` y valida su tamaño, sus líneas y sus enlaces, como si fuera `content/skills/<flujo>/<clave>.md`.
- Registro: `validRecordedStagePacket` acepta el registro, y `rejectHandWrittenPolicies` rechaza un `stage*.md` en el fuente.
- Línea de entrada del flujo, en texto y no como enlace, porque el archivo no existe en el fuente: «At entry, read `stage.md` in this skill's folder whole: it holds the leaves <hojas>, so a step that links one of them needs no new read until a compaction, after which read the packet again.» Sigue el respaldo: si falta, las hojas enlazadas; si falta la línea de fin, el resto; si falta un encabezado, esa hoja; si la salida dice que se cortó, las hojas enlazadas.
- Fin del paquete: la última línea es `<!-- hive packet end: <clave> · <miembro>,<miembro>,… -->`.
- La línea de entrada del flujo añade: «If its last line is not the packet end line, read the rest».
- Nombre de miembro: su ruta del manifiesto sin `.md`, por ejemplo `## hive-core/references/questions`.
- Auditoría:
  - quita el prefijo de número de línea que antepone la herramienta de lectura, como el `N\t` de Claude Code;
  - toma los miembros de la línea de fin;
  - da el paquete por leído solo si el resultado trae esa línea y un `## <miembro>` exacto por cada miembro;
  - si no, reporta `not_observed`;
  - así detecta el corte por la mitad de Codex y no confunde un miembro con una sección propia de una hoja.

Membresía inicial estimada sobre `febf6665`. T7 la recalcula sobre la base de E2, cuando `next-step-block` y `close-footer` ya se habrán fundido en `report-endings` y `close-report` (`reporting-family-leaves`).

| Paquete | Miembros | Bytes |
|---|---|---|
| hive-mode | questions, report-endings (+ lo fundido) | ~13,000 |
| flow-plan | verification-floor, tests, questions, delivery-ask, plan-handoff (sin reuse-before-writing, que la ruta recibe en el paquete de build) | ~21,450 |
| flow-build | verification-floor, checkpoint, checks, tests, reuse-before-writing, git-commit, security-boundaries, stored-data | ~22,500 |
| flow-verify / stage-entry | verification-floor, no-comments, blast-radius, prose, suppression-review, project-verification, checks, checkpoint | 22,831 |
| flow-verify / stage-exit | reviewer-brief, questions, review-findings, merge-gate, pr-reviews | 20,739 |
| flow-close | close-modes, ticket-state, close-hygiene, close-report (+ lo fundido), support-material | ~20,000 |
| git-workflow | git-commit, foreign-commits, pr-reviews, merge-gate, prose | 14,414 |
| flow-deploy | promotion-path, external-effects, merge-gate, pr-reviews, deployment-checks | 18,801 |
| unattended-delegation | delivery-ask, external-effects, questions, delegation, unattended-effects | ~20,700 |

Lecturas estimadas en la ruta plan→build: ~26 frente a 39. La prueba de huella (AC13) da la cifra declarada.

**Checkpoint y cierre (E3)**

- `tasks.py checkpoint --rules "flow-build/stage.md · extra: defect-fix"` escribe `- Rules:` en `## Checkpoint`, y `--principles "boundary-discipline"` escribe `- Principles:` (D13 A), para que la línea que `flow-build` pasa a `flow-close` sobreviva una compactación. Su valor es `none` o nombres `[a-z0-9-]+` separados por `, `; el flag une nombres sin duplicados, un nombre reemplaza `none` y nada borra la línea; `flow-close` la lee en su paso 2, antes de borrar el checkpoint. Igual que `--worktree`, las llamadas siguientes la conservan, y `check_change.py` valida su forma. Un plan que se queda en la conversación escribe la misma línea a mano.
- Orden tras compactar, dicho en una línea de `global.md` y en la hoja `checkpoint`: checkpoint, `HIVE.local.md`, el paquete de `Rules:`, sus extras y la memoria persistente.
- `Reply:` en 7 flujos: la fila de `report-endings` o `close-report` que termina el flujo, más lo que el flujo añade y la fila no tiene.

## Impact

- Lo leen o lo enlazan ~1,500 sitios en ~200 archivos, contando nombres sueltos. Las rutas `/SKILL.md` son 851 menciones en 156 archivos fuera del archivo histórico. El archivo histórico no se toca.
- **Instalación:**
  - el refresco único tras E3 (D9 A) retira ~62 carpetas instaladas, sus enlaces de catálogo, sus `openai.yaml` y sus `__pycache__`; lo maneja `ownership.go:233-299`, más T4;
  - instala `hive-core` y los paquetes.
- **Hosts:**
  - Antigravity pasa de 62 entradas a 1;
  - Codex pasa de 72 políticas a 11;
  - Claude, Grok, Pi y OpenCode no cambian su catálogo.
- **Sesiones abiertas:** las que conservan el `global.md` viejo fallan al leer hojas hasta reiniciarse, incluida la paralela en este repositorio. Por eso el refresco de este trabajo espera tu aviso (D9 A). Si otra sesión refresca antes, cada unidad funciona sola (D10 A).
- No toca producción.

## Recovery

- **E1 y E3:** `git revert` del merge, y luego `hive update --source <checkout>` con el binario vigente. Reinstala lo anterior y retira lo nuevo.
- **E2:**
  1. Revertir antes el merge de E3, que nombra los paquetes en `Rules:` y en `tasks.py`.
  2. Revertir el merge de E2.
  3. Desplegar el contenido revertido con el binario de E2. Retira los paquetes y sus registros.
  4. Reconstruir el binario.

  Un binario sin E2 rechaza un estado que registra paquetes ("invalid recorded source").
- No hay datos almacenados ni contratos externos.
- Lo que no se deshace solo: las sesiones abiertas con la guía anterior, que hay que reiniciar.
