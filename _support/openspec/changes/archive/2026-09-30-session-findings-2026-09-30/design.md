# Diseño

## Contexto verificado

Revisión hecha el 2026-09-30 sobre las sesiones de ese día. La evidencia completa está en Engram, tema `monitoring/sessions-2026-09-30-ark-sample-project-globex`.

- **Versión cargada.** Todas las sesiones revisadas corrían con el contenido de `fb8f61b` (versión a5f3d1f) más `eddcf41`. Se observó en `prompt_context.json` de Grok y en los primeros registros de Claude Code. El único cambio de contenido posterior, `9cad70f`, llegó a las 14:04, cuando esas sesiones ya habían terminado.

- **H1, causa probable.** Se observó en los transcripts, sin prueba en vivo.
  - El modelo redacta el informe en su razonamiento y luego emite solo la llamada a la pregunta. Lo muestran el razonamiento en globex, índice 128 ("emitiré el texto junto con la herramienta de pregunta"), y en OpenCode, secuencia 281.
  - `updates.jsonl` de globex no tiene ningún fragmento de texto antes de las preguntas (filas 292–293).
  - De 12 preguntas precedidas por anotaciones (Engram, archivo de memoria, Linear, `todo_write`, archivos del plan) sin una corrección en contexto, las 12 salieron sin texto. De 3 preguntas sin esas anotaciones, las 3 tuvieron texto.
  - Tras una compactación cuyo resumen nombraba el fallo, 5 de 5 preguntas tuvieron texto. Lo mismo ocurrió en 01a0e9b9 T150, tras una queja del usuario.
  - La regla actual (`content/guidance/global.md:22`) ordena las anotaciones antes del mensaje, pero no dice que la respuesta que llama la pregunta deba contener texto.

- **H2.**
  - La regla vigente, `content/skills/flow-build/references/verification.md` (bullet de despliegue), solo hace el recorrido si el proyecto lo declara o el usuario lo pide, y no lo ofrece. Por eso no ofrecerlo en globex y en sample-project fue correcto.
  - Los fallos reales fueron dos, ambos en sample-project 01a0f05b:
    - Cuando el usuario pidió "haz los walks en paralelo", el hilo principal armó el navegador él mismo, con 47 llamadas, hasta que el usuario dijo "obvio puedes lanzar subagentes". La prohibición de programar el recorrido solo está en esa referencia.
    - El login de producción devolvió 401 en el corte, y el comentario de Linear siguió diciendo que la API estaba sana.
  - `flow-build/SKILL.md:63` menciona el recorrido funcional, pero no dice quién lo ejecuta.

- **H3.**
  - La sesión de Grok en ark 01a0f11f leyó `flow-research/SKILL.md`, cuya línea 18 ya pide delegar las búsquedas amplias o de varias familias de fuentes. Aun así hizo unas 35 búsquedas en el hilo principal y compactó 4 veces.
  - OpenCode con qwen pasó de 24k a 232k de contexto.
  - sample-project leyó páginas de Atlas de 31 KB en el hilo principal.
  - El criterio es de juicio, sin umbral.

- **H4.** Tres casos hoy:
  - ark supuso los dominios de Tricell y los hosts `*-wh` como "ya elegidos";
  - sample-project propuso recrear el proyecto de ARK en Atlas ("ark usa postgres");
  - OpenCode escribió "0 commits en ambos sentidos, verificado" cuando `qa` tenía 1849 commits más.

  Se suma un caso anterior ("Tú ya validaste…"). `global.md:24` trata las premisas del usuario, no las del agente.

- **H5.** Grok en ark escribió `linear.app/sample-workspace/issue/ARK-659`; el real es `linear.app/tricell/...`. El ajuste `Tracker` de ark nombra el equipo, no el espacio de trabajo. `global.md:16` pide la URL estable más corta, pero no dice de dónde sale.

- **H6.**
  - Hallazgos perdidos (sample-project):
    - Un hijo de verificación calificó como "Major" las tablas cortadas.
    - Otro hijo vio el año escrito "2,018".
    - El relevo los redujo a una línea o los omitió, y no se propuso ticket.
  - Enlaces sin resolver:
    - En 9 sesiones de Grok se buscaron referencias con `find /path/to/home…`.
    - Los roles enlazan sus referencias como `skill:owner/path`; por ejemplo, `content/agents/quality/hive-verify-change.md:12` enlaza `browser-automation.md`. Hay 9 roles con estos enlaces, y 2 enlaces en `global.md` (líneas 20 y 96).
    - En ámbito de usuario, las skills se instalan en `<home>/.agents/skills/` en los seis hosts, y `~/.claude/skills/<skill>` enlaza ahí. En ámbito de proyecto, cada host usa su propio directorio dentro de la raíz.
    - El renderizador de roles es `integrations/agents/agents.go:290` (`Render`), invocado desde `tooling/management/plan.go:480`. Hoy no transforma el cuerpo.
    - La validación de publicación (`tooling/management/references.go`) exige `skill:` en los roles fuente.

- **H7.**
  - globex no tiene sección `## Hive`.
  - Grok buscó `Base branch`, no lo encontró, usó `develop` porque así lo decía la documentación del repo, y no preguntó.
  - `global.md:108` ya manda preguntar y registrar.

- **H8.**
  - qwen escribió "Sí,现在开始" en una etiqueta.
  - Grok en sample-project escribió "(Recommended)" en inglés en D1–D4.
  - La sección de idioma de `global.md` (líneas 5–12) cubre informes y registros, pero no menciona las etiquetas de las opciones.
  - Las otras faltas de qwen ya tienen regla explícita, así que no se duplican:
    - D7–D10 sin glosa (`global.md:18`);
    - pendientes sin dueño (bullet de cierre);
    - mensaje de plan listo de una frase (`flow-plan/references/plan-review.md`).

- **H9.**
  - sample-project ofreció "Seguir aquí: Nombra el ticket" con cuatro tickets abiertos.
  - En globex, la opción única "Terminar" era correcta: no había tracker ni siguiente ticket.
  - `global.md:32` ya pide nombrar el ticket.

## Diseño elegido


### Texto de la guía (H1–H9, salvo la reescritura de enlaces)

Cada regla tiene un único lugar, en el archivo que está cargado cuando ocurre la conducta. Una regla que ya existe no se duplica: se aplica al caso que no cubría. La redacción final es del implementador, sujeta a estos contenidos y a las frases que buscan los criterios.

- **H1** (`global.md`, junto a la regla de esperas de `:22`). Es una oración general, válida para toda pregunta nativa, también las de `:21`, `:23` y `:32`:
  - *any response that calls the native question tool* empieza con el texto completo del mensaje en esa misma respuesta;
  - el razonamiento no se muestra al usuario ("reasoning is not shown");
  - una respuesta con solo la llamada es el fallo.

  Va en su propia oración, al inicio del bullet, para que no se pierda en el párrafo (D2-A). Si recurre con este texto, el siguiente paso es D2-B.
- **H4** (`global.md:22`): el texto y las opciones de una pregunta afirman como hecho solo lo "verified in this session"; lo demás va marcado como supuesto a confirmar.
  - Cubre los casos de ark, sample-project y "Tú ya validaste…".
  - El "0 commits, verificado" de OpenCode no era una premisa sin marcar, sino una verificación falsa. Lo cubren `global.md:89` y la sección Evidence, que no cambian.
- **H5** (`global.md:16`, oración de enlaces):
  - la URL de un ticket es la que devolvió el tracker o un enlace existente verificado, recortada solo a su forma estable más corta;
  - sin una URL así, se "cite the ID without a link";
  - nunca se compone el espacio de trabajo ni el dominio.

  Así no choca con "without a title slug" ni con "using known context".
- **H6, relevo** (`global.md:90`, como aplicación de `:40` al resumen, no como regla nueva): el hilo principal releva "every defect a child reports", en cualquiera de sus escalas (Blocker/High/Medium de `hive-review-ux`, major/minor de `hive-verify-change`), o lo propone bajo la regla de hallazgos incidentales. Ninguno se omite del resumen.
- **H7** (`global.md:108`): un valor hallado solo en otra documentación es una "recommendation for that question", no el ajuste. Se pregunta y se registra.
- **H8** (`global.md:8`):
  - Se agregan las "option labels" de las preguntas al texto en el idioma de la sesión.
  - La marca de recomendado ya está cubierta en `:21` ("mark one option Recommended in the session language") y no se repite.
- **H9** (`global.md:32`): el agente elige el ticket. Una opción que "asks the user to name" el ticket no cuenta.
- **H2** (`flow-build/SKILL.md:63`, en "Verify and close"), dos frases:
  - Condicional, sin repetir la referencia: cuando corre un recorrido desplegado (porque el proyecto lo declara o el usuario lo pide, también a mitad de sesión), va a los hijos que nombra el bullet de despliegue de la referencia de verificación, y el hilo principal "does not script" el recorrido. La regla de cuándo corre sigue solo en `verification.md:97`.
  - Cualquier comprobación fallida o error observado después de desplegar, como un 401 en el login, deja el despliegue sin verificar y se reporta así, "never as healthy".
- **H3** (`flow-research/SKILL.md:18`): se reescribe la oración "Keep a short, bounded lookup…" con un disparador que se cumple sobre la marcha:
  - en la "sixth search" o lectura sin respuesta, parar y delegar el resto;
  - no cuentan la lectura de la guía del proyecto ni la de skills y referencias.

  `flow-plan/SKILL.md` ("Ground the decisions") remite a ese criterio para la exploración. El umbral previsto antes de empezar se descartó: depende de la misma predicción que falló, y "dos familias de fuentes" delegaría toda pregunta de estado.
- **Presupuesto.**
  - Se condensan los mismos bullets editados.
  - `tests/content/budget_test.go:12` sube al tamaño final más unos 1 KiB, según la convención de su propio comentario (`:7-10`), con la razón en el PR.
  - Se enlaza #47 (reevaluar el peso de la guía global).

### Enlaces `skill:` resueltos al instalar (H6, D3-A, D6-A)

- **Directorio de skills por ámbito.** Hay que leerlo de la configuración, no escribirlo a mano.
  - **Usuario:** `<home>/.agents/skills`, el almacén compartido real. Hoy los seis hosts instalan ahí, y `~/.claude/skills/<skill>` son enlaces a él (`integrations/claude/claude.go:22-25`). Con un solo directorio para todos, el bloque compartido y los roles compartidos salen iguales en cada host.
  - **Proyecto:** la ruta relativa a la raíz donde ese host instala las skills, por ejemplo `.claude/skills` para Claude (`claude.go:15-18`) o `.agents/skills` para Codex (`codex.go:13-15`), según D6-A. Así no se escribe una ruta personal en archivos versionados.
    - Limitación aceptada: la ruta se resuelve desde la raíz del proyecto. Una sesión abierta en un subdirectorio, o un worktree sin las skills del proyecto versionadas, no la encuentra directamente.
    - `instruction-resources.md` lo documenta, y la frase de `global.md:88` dice "relative to the project root".
- **Interfaz.**
  - Cada paquete de host en `integrations/<host>` expone su directorio de skills para una `target.Config`, con su ámbito.
  - La reescritura es una función pura sobre el Markdown, `RewriteSkillLinks(body, dir)`. Recorre el texto por líneas, respeta los bloques de código y comparte esa lógica con el validador.
  - Esa función vive en un paquete pequeño que `integrations/agents` puede importar sin ciclo (`management` importa `agents`, `plan.go:169`). Por ejemplo, un `integrations/mdlinks` nuevo. La detección de bloques que hoy está en `outsideFences` (`references.go:23-52`) pasa a ese paquete y el validador la reutiliza.
  - `agents.Render` recibe el directorio de skills y reescribe `r.Body` antes de codificar, igual en todos los hosts. Así, en Codex la reescritura ocurre antes de `tomlQuote` (`agents.go:376-377`); después ya no se puede, porque el cuerpo es una sola línea escapada.
  - Se actualizan las llamadas a `Render` en `plan.go:480` y en `render_golden_test.go`.
  - El bloque global (`plan.go:468`) llama a la misma función desde `management`.
  - Nota: `tomlQuote` usa `json.Marshal`, que escapa `<` y `>` como `\u003c` y `\u003e`. Es TOML válido y se decodifica bien, pero un `grep` sobre el `.toml` no ve los `<...>` literales.
- **Qué se reescribe.**
  - Sí: los enlaces en línea `[label](skill:owner/path)` fuera de bloques de código, con el mismo criterio que `outsideFences` (`references.go`). El destino va entre `<...>`, por si la ruta tiene espacios.
  - No: los enlaces de referencia (`[x]: skill:...`). El validador los admite, pero hoy no hay ninguno en `content/`; se documenta.
- **Consistencia.**
  - El bloque se calcula por cada host que lo usa, con la misma comprobación de igualdad que ya tienen los roles ("shared … rendering differs by host", `plan.go:483`). Así, un archivo compartido entre Claude y Grok nunca recibe dos versiones distintas.
  - `hive status`, la recuperación y el estado por instalación comparan con los bytes gestionados del registro (`plan.go:805`, `versioning.go:146`), así que las rutas no producen falsos cambios. Un segundo `hive update` sin cambios no propone escrituras.
- **Sin subir `agents.Version`.**
  - Subirla impediría reinstalar versiones anteriores con `--release`, porque `validateRelease` rechaza otro `Renderer` (`plan.go:312`).
  - No hace falta: `LoadPlan` recalcula y compara (`plan.go:639-655`), y un plan guardado con el binario anterior ya falla de forma segura.
- **Sin caso de error nuevo.** Con la configuración actual no existe un host sin directorio de skills, así que no se agrega manejo para ese caso.
- **Qué no cambia.**
  - Las fuentes en `content/` conservan `skill:` (`references.go:113`).
  - La validación de publicación no cambia.
  - Los enlaces relativos dentro de una skill no se reescriben.
- **Documentación:**
  - `_support/docs/architecture/instruction-resources.md`, sección "Runtime resolution";
  - `_support/docs/architecture/deployment-manager.md:133`, la frase "Agent files are rendered in memory…";
  - `global.md:88`, la frase sobre resolver `skill:`, que queda para los casos sin reescritura;
  - el requisito nuevo de la spec `versioned-installation`.
- **Alternativa descartada:** D3-B, reforzar el texto para que el hilo principal pase las rutas. Esa regla ya existía (`global.md:88`) y falló en 9 sesiones.
