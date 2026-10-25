# Diseño

## Contexto verificado

Leído en `development` `b7277d82` el 2026-10-25.

- **Verificación por tarea.** [plan-execution.md](../../../../content/skills/flow-build/references/plan-execution.md):17-24: cada tarea `tdd` o `check` pasa a `[?]` cuando el implementador la reporta y se lanza `hive-verify-task`; `[x]` solo con todos los criterios `met`; hasta tres rondas de arreglo. La independencia viene del contexto limpio del verificador, no de su modelo (paso 5).
- **Solape existente.** `plan-execution.md:26`: mientras corre una verificación se construye la siguiente tarea que no depende de la verificada ni comparte archivos con ella; la que depende espera `[x]`. Las verificaciones de tareas sin archivos en común se lanzan juntas.
- **Commits.** `plan-execution.md:30`: un commit por tarea al llegar a `[x]`, nada antes; las tareas con UI esperan además la puerta de UI.
- **Origen.** [archive/2026-09-26-per-task-verification](../archive/2026-09-26-per-task-verification/design.md): en el experimento R1 (ARK-730) un implementador de modelo pequeño omitió dos requisitos del plan, y el orquestador marcó tareas `[x]` antes del veredicto.
- **Costo observado** (memoria Engram, sin releer esas transcripciones): S3, unos 21 min en la cadena T2→T3→T4 de `14a79ad0` (verificación, arreglo, re-verificación); W16, unos 6 min de verificadores en serie en `01a1579f`, los cinco con `met` a la primera. En `b979e8b2` (leído en su transcripción) las 15 verificaciones duraron de 33 a 93 s y corrieron en paralelo; el costo se concentra en cadenas y rondas de arreglo.

## Enfoque

**Puntos de control elegidos al planificar.** Cada tarea `tdd` o `check` lleva `**Checkpoint:** C<n>`. Un punto de control agrupa tareas y se alcanza como dice «Alcanzar un punto». `tasks.md` lista los puntos con sus tareas y la razón de su lugar. La numeración sigue el orden del plan; la última tarea `tdd` o `check` cierra siempre el último punto.

**Criterio de riesgo.** Una tarea cierra un punto de control (es la última de su grupo) cuando:

1. dos o más tareas posteriores dependen de ella, directamente o a través de otras, de modo que un defecto suyo se propagaría; o
2. cambia un contrato público o persistido, datos guardados, permisos o un límite de seguridad, o concurrencia.

Las demás tareas se agrupan en el siguiente punto. Así, en la cadena T2→T3→T4, T2 cierra un punto (dos dependientes) y T3 y T4 van al siguiente. Una tarea solo depende de tareas anteriores del plan. Sin el campo `Checkpoint`, cada tarea es su propio punto y se verifica al reportarse, como hoy; esta compatibilidad vive solo en `plan-format.md`, y `plan-execution.md` aplica la regla general. El criterio de riesgo también vive solo en `plan-format.md`; `plan-review.md`, `plan-execution.md` y el sitio remiten a él.

**Alcanzar un punto.** Un punto se alcanza cuando cada una de sus tareas está en `[?]` o fuera de él: bloqueada antes de reportarse (`[/]` con `blocked: <prerequisito>`, forma que T1 agrega a la tabla de marcadores; no se le lanza verificador) o retirada por el usuario, con su línea. Una tarea que el usuario manda a rehacer vuelve a `[/]` y el punto espera por ella. Una dependencia retirada se trata como en `plan-execution.md` paso 6: la tarea que dependía de ella se replantea con el usuario.

**Ejecución (D5-A, con D12-A).** La regla vigente está en `content/skills/flow-build/references/plan-execution.md` («Do not wait idle» y «Commit tasks»); resumen:

- La dependiente del mismo punto arranca en cuanto su dependencia se reporta (y se commitea, si hay commits), sin esperar veredicto; la de un punto anterior, y la tarea sin punto que depende de una tarea pendiente de veredicto, esperan a esas tareas en `[x]`.
- Dos tareas que comparten archivo nunca corren a la vez: ni implementación, ni ronda de arreglo, ni verificación de una mientras la otra está en alguna de esas fases.
- Al alcanzar un punto se lanza un verificador por tarea, todos juntos salvo los de tareas que comparten archivo, que van en serie. La tarea que llega tarde recibe su propio verificador.
- `waiting: T<n>`, re-verificación completa de dependientes y de tareas que comparten archivo tras un arreglo, reapertura sin gastar rondas y tercera reapertura al usuario, como en la regla.

**Commits (D12-A, reemplaza D4-A y D8-A).** Cada tarea se commitea cuando el implementador la reporta, antes de verificarla, solo con sus rutas; como dos tareas que comparten archivo no corren a la vez, cada commit y cada diff tienen solo los cambios de su tarea, y el verificador revisa los commits de su tarea. Los arreglos, reaperturas y correcciones van en commits `(T<n>)` aparte, sin amend. El trabajo se publica para merge solo cuando todas las tareas están en `[x]`; la primera vista y la puerta de UI siguen antes de esa publicación. Una tarea retirada se deshace con un revert. Sin commits (`hold`), las tareas que comparten archivo van una tras otra y el verificador declara el límite. Se eliminan el commit conjunto, el orden de commits y sus desempates, que la re-revisión mostró que abrían bloqueos (con `hold`, en cadenas de UI) y mezclas de cambios rechazados.

## Riesgos y por qué se aceptan

- Un defecto de T3 puede afectar a T4 antes de detectarse. Queda acotado al punto de control y a una ronda de reapertura; el criterio pone un punto tras las tareas con más dependientes.
- El juicio de riesgo es del planificador, y no habría detectado la omisión de ARK-730. Esa defensa sigue siendo el veredicto independiente antes de `[x]` y del commit, que no cambia.
- Con varias tareas sin commitear en el árbol, el verificador ve trabajo de otras tareas en sus rutas; su contrato ya le pide declararlo (`hive-verify-task.md:8`).
