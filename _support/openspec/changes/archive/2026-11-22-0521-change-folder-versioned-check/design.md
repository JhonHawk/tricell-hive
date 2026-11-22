# Design

## Approaches

- **A — Subcomando `versioned` en `tasks.py` (elegido, D1 A).** Revisa exactamente el folder del trabajo, que `tasks.py` ya recibe, y reutiliza su lector de Git. `flow-close` paso 3 lo corre antes del reporte. Costo: un paso más en el cierre que el agente debe correr; D3 detecta cuando no lo hace.
- **B — Extender `cleanup_branches.py`.** Ya corre en cada cierre y cubre el repo de specs, pero trabaja por repo: listaría los folders sin commit de otros trabajos y atribuir cuál es de este volvería a ser juicio del agente.
- **C — Prosa más auditoría.** Detecta después, no impide; root `AGENTS.md` «Measurement» pide el mecanismo antes que la prosa. Queda solo como complemento (D3).

Peldaño de reutilización: 2 (el código ya tiene `git_reader`/`git_out` en `tasks.py` y `base_branch_of` en `workspace_repo.py`). `base_branch_of` se carga con `importlib`, como `git_reader` carga `git_state.py`, en lugar de copiarlo: una copia divergiría (condición d de «Reuse before writing»).

## Decisions

- **Repo desde el folder.** `git -C <folder> rev-parse --show-toplevel` da el mismo repo desde un repo, un subpaquete de un monorepo o la raíz de un workspace sin Git, porque no depende del directorio de trabajo.
- **Versionado** = el último commit del folder contenido en una rama remota (con `Delivery: direct-base` en el `## Hive` de ese repo, en `origin/<base>`) y, si el folder vive en un repo distinto del trabajo, ningún archivo sin rastrear ni modificado. En el repo del trabajo, un archivo sin rastrear o modificado después de ese commit es un aviso, porque `git-workflow` paso 4 deja que el archivo del lote de higiene commitee las ediciones posteriores (la línea del checkpoint). Bajo `ready` el folder está en la rama del PR, que es remota: pasa.
- **Sin comandos de entrega.** Con salida 1 el cierre vuelve a `git-workflow` paso 4, que fija el commit con `git_state.py commit` (escaneo de secretos), el PR o el push según el `Delivery` o el modo en vigor, y la revisión de commits ajenos. El subcomando solo nombra lo que falta.
- **Sin red.** Usa las ramas remotas locales, que el push del agente actualiza; no hace fetch. Un push hecho desde otra máquina puede verse como no versionado hasta un fetch: la salida lo dice.
- **`--work-repo` (D2 A).** Solo avisa: el folder queda versionado aunque el repo del código no declare `Specs`.

## Proposed shape

```
python3 -I tasks.py <change-folder> versioned [--work-repo DIR]

repo: <toplevel> · base: <base> (Hive | origin/HEAD) · delivery: direct-base | not declared · work repo: same | <ruta>
versioned: yes | no
untracked: <ruta>            (una línea por archivo)
modified: <ruta>             (fallo en otro repo; en el repo del trabajo: "warning: modified <ruta>: later edits wait for the archive")
not committed | not pushed: <sha> is on no remote branch | not on origin/<base>
next: commit and deliver the folder as git-workflow step 4 sets, then run versioned again
warning: <work-repo> declares no Specs | Specs <valor> does not hold the folder; declare Specs: <ruta relativa al repo del trabajo>
```

Salida: 0 versionado; 1 no versionado, fuera de Git o sin base resoluble bajo `direct-base`; 2 argumentos inválidos o carpeta ilegible, como el resto de `tasks.py`.

## Impact

`tasks.py` lo usan `flow-plan`, `flow-build`, `flow-verify` y `flow-close`; el cambio añade un subcomando y no toca los existentes. El criterio de auditoría solo lee transcripts. No toca producción.

## Recovery

Revertir el PR quita el subcomando, el paso y el criterio; nada persiste fuera del repo.
