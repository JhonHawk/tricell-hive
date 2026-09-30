# Diseño

## Contexto verificado

Base: `6230c7a` (`development`), el 2026-09-30. El inventario lo hizo un subagente de investigación; sus números clave se reprodujeron y lo que es inferencia se indica.

- **Golden de render.** `integrations/agents/testdata/render.golden` ocupa 244.206 bytes: los 20 roles reales en 6 hosts (`render_golden_test.go:29,51`). Nació como ayuda de caracterización para extraer `Resolve` en #46 (archivo `gh-46-read-only-views/tasks.md:28-34`). Los 3 commits que tocaron roles o perfiles desde su creación tuvieron que regenerarlo, y uno (25b671e) llegó con el golden roto. Varias pruebas de render que parecían sintéticas leen el rol real `hive-research.md` y los perfiles reales: `TestPiObserve…`, `TestCodexTOMLIsFlat…` y `TestCursorObserve…` (`agents_test.go:46-50`, `:76-80`, `:134-139` y `:346-366`). `TestProfilesReject*` (`:125-189`) reemplaza fragmentos literales del JSON real. Todas pasan a entradas sintéticas en T1.
- **Conteos fijos de 20 roles:** `render_golden_test.go:29` y `agents_test.go:28`.
- **Tabla de esfuerzo** `TestRepositorySourcesMatchExpectedEffortLevels` (`agents_test.go:251`): 20 roles × 3 hosts, con valores que repiten el front matter de los roles y los perfiles. Ninguna especificación la exige.
- **Valores de perfil fijados.**
  - `TestResolveReturnsProfileModelAndProfileEffort` (`resolve_test.go:21`), `TestResolveRoleEffortWinsOnlyWhereHostAcceptsEffort` (`:62`) y `TestResolveVerifyTaskRoleAgainstRepositoryProfiles` (`:171`);
  - `tooling/management/models_test.go` (lee los perfiles reales en `:28`; valores fijos en `:150`, `:154` y `:179`);
  - la vista de modelos (`tooling/cli/tui_models_view_test.go:46`; valores fijos en `:121-127`, `:157`, `:162`, `:169` y `:296`).
- **Pruebas que solo necesitan un archivo de perfiles válido**, sin fijar valores:
  - `catalog_test.go:16,63` y `cursor_test.go:128` en `tooling/management`;
  - `writeUpdateCatalog` en `tooling/cli/update_test.go:70`;
  - los fixtures de bootstrap en `tooling/cli/bootstrap_test.go`, que desde #71 copian el archivo real porque el validador exige `version` 1 y los 5 hosts base.
- **Voces.** `tests/content/voices_test.go:34` fija 16 frases del preámbulo, y `:60` fija los IDs `jarvis`, `senior-direct` y `mentor`. El requisito AC10 del archivo `gh-46-voice-layer/proposal.md:56` pide cuatro puntos. `TestVoiceListShowsRepositoryVoices` (`tooling/cli/voice_test.go:66`) exige esos 3 IDs, y la lógica del listado ya la cubre `TestVoiceListStripsDuplicateNamePrefix` con una voz sintética.
- **Límite de `global.md`** (`tests/content/budget_test.go`): `globalGuidanceBudget = 43320`, `globalGuidanceSlack = 1024`, y el archivo mide 43.312 bytes. Se editó en 58 de 102 commits de `global.md`: 50 subidas y 7 bajadas.

## Enfoque

### Perfiles sintéticos

Cada paquete que necesita perfiles tiene su propio archivo fijo en `testdata/`:
- `integrations/agents/testdata/profiles.json`;
- `tooling/management/testdata/agent-profiles.json`;
- `tooling/cli/testdata/agent-profiles.json`.

Cada archivo es válido para `ReadProfiles` y el validador: `version` 1, los 6 hosts y los perfiles razonamiento, ejecución, hereda y verificador. Usa valores reconocibles que no existen en los datos reales (`model-reasoning`, `model-exec` y similares), para que una prueba no pueda pasar por coincidencia con los reales.

Las pruebas que fijaban valores pasan a leer este archivo y a exigir sus valores sintéticos. Así se mantiene la regla que cubren: modelo y esfuerzo del perfil, variante `#effort` en OpenCode, esfuerzo del rol que gana solo en los hosts que lo aceptan.

Se descarta un paquete compartido de datos de prueba: los paquetes no comparten `testdata/`, y la guía de Go ubica los datos de prueba junto a cada paquete. Tres archivos pequeños iguales cuestan menos que un paquete nuevo. `tooling/cli/testdata/minimal-source/integrations/agent-profiles.json`, de #71, puede pasar a este archivo completo si el bootstrap lo necesita; la decisión se toma al construir y se registra.

### Golden sintético

`TestRenderGolden` dibuja unos 6 roles sintéticos con los perfiles sintéticos en los 6 hosts, para cubrir todas las formas que hoy cubre el golden real:
- observación con perfil de ejecución;
- razonamiento con `effort` y `effort_claude` propios;
- hereda el modelo sin esfuerzo, y hereda el modelo con `effort`: en OpenCode la variante fija no cambia (`agents.go:243`);
- perfil de modelo `verifier`;
- accesos `implement` y `verify`;
- un cuerpo con comillas y unicode.

Los perfiles sintéticos conservan las formas de los reales:
- `inherit` de Codex sin modelo;
- modelos de Grok vacíos;
- OpenCode con modelo base más esfuerzo, e `inherit` con variante;
- listas de herramientas de Claude y Pi;
- nombres de modelo distintos por host, porque `tui_models_view_test.go:121-122` y `models_test.go:166-193` necesitan que Claude y Codex difieran.

El golden se regenera con el flag que ya existe (`-update-golden`) y ocupa como mucho 40 KB.

### Comprobaciones sobre los datos reales

En `integrations/agents`, sin conteos fijos:
- todo rol de `content/agents` se dibuja sin error en los 6 hosts;
- en Claude, Codex y Pi todo rol dibujado tiene esfuerzo, y en Grok, OpenCode y Cursor ninguno lo tiene (D2-A, ajustada). El conjunto de valores válidos no se repite en la prueba: lo imponen `Parse` (`agents.go:94`) y `ReadProfiles` (`agents.go:160`, que acepta `ultra`), y una lista propia rechazaría contenido válido;
- los perfiles reales pasan `ReadProfiles` y hay un perfil verificador en cada host. Esto ya existe en `TestRepositoryProfilesDeclareVerifierOnAllHosts`.

### Voces

- `voices_test.go` conserva una ancla por cada cosa que enumera AC10 (`gh-46-voice-layer/proposal.md:56`), incluidos la glosa de IDs y las secciones con etiqueta: unas 9 en vez de 16.
- Los IDs se obtienen del directorio `content/voices/*.md`, sin `preamble`.
- Se mantienen el tamaño máximo y la ausencia de marcadores reservados en cada voz real.
- `TestVoiceListShowsRepositoryVoices` exige al menos una voz y que `preamble` no aparezca.

### Límite de `global.md` (D1-A)

- Un solo techo: `globalGuidanceBudget` pasa a unos 44.336 bytes (tamaño actual más 1024).
- Se quitan `globalGuidanceSlack` y la comprobación de encogimiento.
- El comentario explica que el techo se sube a propósito cuando el crecimiento lo supera, y que encoger no pide nada.
- Consecuencia aceptada por el usuario: crecer por debajo del margen ya no se ve en el diff de la prueba.
