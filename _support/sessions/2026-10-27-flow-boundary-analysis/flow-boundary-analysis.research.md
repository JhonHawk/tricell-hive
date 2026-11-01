# Análisis de flujos de Hive por skill (fronteras)

- Fecha: 2026-10-27 · Base: `development` @ 4ef04e01
- Pregunta: ¿qué falla en cada flujo y en sus fronteras con los demás y con `global.md`, y cómo se resolvería?
- Método (decisión D3-A): un subagente `hive-review-harness` dueño por skill (research, plan, build, verify, security, deploy, close) más un abogado de `content/guidance/global.md`. Fase 1: autoevaluación con la rúbrica v2 y libro de fronteras. Fase 2: una ronda de réplica (K1–K23) en las fronteras disputadas. Fase 3: el hilo principal juzga y verifica las citas.
- Engram: `decisions/flow-analysis-rubric` (rúbrica y diseño).

## Rúbrica v2 (D1-A)

1. Entrada: disparadores, precondiciones y qué pasa sin ellas.
2. Salida y traspaso: cuándo termina, qué entrega, salidas alternas y bucles.
3. Dependencias globales: qué asume de `global.md`, qué repite y si funciona sin leerlo.
4. Contratos con otras skills: referencias compartidas, traspasos y contradicciones.
5. Decisiones del usuario y autorización: preguntas, efectos externos, arrastre de respuestas, silencio y delegación desatendida.
6. Cambios de alcance: a qué etapa regresa y qué pasa con lo producido.
7. Modos de falla e impacto: cómo falla, radio de impacto y quién lo detecta.
8. Robustez de ejecución: compactación (16 KB), modelo menos capaz, hosts y delegación.
9. Proporcionalidad y costo: camino ligero y contexto consumido.

Formato de cada hallazgo: veredicto (sólido, frágil o roto), evidencia `archivo:línea`, y si es OBSERVED o INFERRED.

## Veredictos por skill

| Skill | Roto | Punto más débil |
|---|---|---|
| flow-research | — | Salida: sondas de diagnóstico sin retirar; tracker inferido |
| flow-plan | — | Aloja 8 referencias transversales; `delivery-decisions.md:5` ilegible |
| flow-build | — | Margen de 409 B; condición de lectura de `verification.md` |
| flow-verify | — | La revisión UI del implementador no se lee en build; orden de seguridad |
| flow-security | Salida (2) | Hallazgos previos al merge sin dueño |
| flow-deploy | — | Recuperación sin autorizar; commits ajenos en un fix a mitad de corrida |
| flow-close | — | Margen de 150 B; cierres diferidos sin disparador |
| global.md | Contratos (4) | Exención de lo mecánico frente a flow-build; 42 KB sin techo |

Las fallas OBSERVED que encontraron los dueños en el historial (commits con ID de sesión) ya están corregidas. Ningún hallazgo nuevo es OBSERVED; todos son INFERRED de la lectura del texto.

## Fallos de la ronda de réplica

| K | Frontera | Resultado |
|---|---|---|
| K1 | "una parada de validación no es pausa" | Acuerdo: el hogar es `flow-close` "When to run it"; build y verify solo apuntan |
| K2 | Línea de contexto | El juez decide: `reporting.md` es canónico y `global.md:16` queda como invariante mínimo (close lo defiende; global aceptaba cualquiera de las dos) |
| K3 | Procedimiento de incidentales | Acuerdo: baja a `incidental-findings.md`; `global.md:34` conserva el invariante |
| K4 | Criterio plan/build en `reporting.md:96` | Acuerdo: queda un puntero a flow-research |
| K5 | Sondas de diagnóstico | Acuerdo: el hogar es `diagnosis.md`, no flow-close |
| K6 | Momento de la revisión dedicada | Acuerdo: pasa a `review-and-gates.md`; verify retira su cesión |
| K7 | Commits de correcciones fuera de build | Acuerdo: regla general en `flow-build/SKILL.md:50` y puntero en el Prepare de verify |
| K8 | `delivery-decisions.md:24` | El juez decide: se queda en plan (es contenido del contrato de entrega); la parte de ejecución ya está en git-workflow |
| K9 | Hallazgos de seguridad antes del merge | Acuerdo: `review-and-gates.md` los cubre (P0–P3, refutación, re-revisión) |
| K10 | Revisión UI del implementador | Acuerdo: puntero desde `flow-build/SKILL.md:40` (+241 B); el contenido se queda en ui-gate |
| K11 | Regla de la nueva superficie de interacción | El juez decide: va a `scope-changes.md` clase 2 (build lo defiende; plan proponía build o ui-gate) |
| K12 | Estado de un build devuelto a plan | El juez decide: va a `scope-changes.md` "Build" y plan solo lleva un puntero |
| K13 | "including deployment of existing code" | Acuerdo: se elimina |
| K14 | Regla "pedir una promoción es la elección de entrega" | Acuerdo: pasa a flow-deploy paso 4 |
| K15 | Registro de la promoción | Acuerdo: baja de `global.md:35` a flow-deploy; una release que publica es efecto crítico |
| K16 | Puntero de deploy a security-boundaries | No hace falta |
| K17 | Recorrido funcional pendiente frente al estado final del ticket | Acuerdo: se corrige `project-settings.md:29` |
| K18 | Inferencia del tracker | Acuerdo: el hogar es `project-settings.md:41`; también se corrige `plan-format.md:30` |
| K19 | Cambio nuevo durante la investigación | Acuerdo: entra en la pregunta de ruta (`scope-changes.md:13,35`) |
| K20 | `global.md:78` "and its diagnosis reference" | Global cede: se elimina |
| K21 | Entrada de un defecto urgente | `global.md:78` se queda como está; deploy agrega un puntero en su entrada |
| K22 | Redacción para cambios mecánicos | `global.md:79`: "A mechanical change skips verify and the dedicated review, not the delivery question…"; `:76`: "needs no plan or delegation" |
| K23 | Número de paso de memoria en close | Acuerdo: se cita por nombre, no por número |

## Hallazgos consolidados y propuesta

Ver la tabla del chat del 2026-10-27, con IDs H1–H19, que es la fuente para la discusión. Resumen de prioridad:

- **Alta:** H1 (recuperación en producción sin autorizar) y H2 (commits ajenos en un fix a mitad de promoción).
- **Prerrequisito:** H7 (tamaño de `flow-close`, `reporting.md` y `flow-build`), sin el cual no caben las demás correcciones.
- **Decisiones del usuario:** H1, H12 (dónde viven las referencias transversales) y H17 (techo de tamaño para `global.md`).

## Decisiones del usuario (2026-10-27)

- D4-A: platicar los hallazgos antes de elegir la ruta (seguimos en research).
- D5-A (H1): la pregunta de autorización del deploy a producción incluye la acción de recuperación de cada disparador, con su objetivo y su reversa.
- D6-B (H12): migrar `questions`, `scope-changes`, `effect-authorization` y `project-settings` a una skill neutral compartida y reescribir sus enlaces. El nombre y el alcance exacto se fijan al planear. Los hogares acordados en la réplica (K11, K12, K17, K18, K19) se mantienen; solo cambia la ruta del archivo.
- D7-B (H17): `global.md` sin techo de tamaño. Sigue en pie la otra mitad de H17: los resúmenes de clases de efecto y de "saltar verify" que se apartan del original.
- D8-B (skill core): la skill neutral de D6-B se convierte en una skill core de tres capas y entra al plan de corrección, con hook. La capa 3 reinyecta la capa 1 tras compactar donde el host lo soporte. Esto no contradice `AGENTS.md:61`: esa regla admite un mecanismo cuando hay una limitación verificada y documentada, y Hive ya tiene dos con ese criterio (corregido el 2026-10-27). Son el hook `PreToolUse` read-whole de Grok, instalado, y el plugin de contexto de OpenCode, instalado pero sin activar.
- D9-A (alcance del hook): el hook se diseña para los seis hosts. La primera tarea del plan arma una tabla de qué sobrevive a la compactación en cada host, y el hook se activa solo donde haya una pérdida verificada. Esto corrige una premisa anterior: Claude Code vuelve a leer el `CLAUDE.md` del proyecto después de compactar (el global de usuario está sin verificar) y recorta las skills a 5.000 tokens, así que ahí el hook serviría para la skill core más que para la capa 1.

## Investigación: skill core con carga bajo demanda y valor de un host vanilla

Pregunta: ¿se puede mover parte de `global.md` a una skill core (ampliando D6-B) y cargarla bajo demanda, y qué aporta Hive frente a un host sin Hive? Hay cuatro frentes.

### Hive legacy

- La retirada en este repositorio es `abad754d`; el issue #24 vive en `JhonHawk/tricell-hive-private`.
- El legacy se quedó con un núcleo de ~87 KB siempre cargado. El resto lo entregaba con enrutadores, un hook read-gate y packs.
- Mediciones:
  - el router `task-routing` se usó 0 de 9 y 0 de 10 veces (OBSERVED; las corridas ya no están en disco);
  - el read-gate se leyó en las cuatro CLIs;
  - forzar la lectura de una referencia no hizo que la regla gobernara (n=3).
- Se retiró por lo inevaluable de toda la capa acumulada, no por la carga bajo demanda.

### Proyectos de referencia

- Ninguno pone sus reglas críticas en una skill que el modelo decida cargar.
- Patrón que se repite: un arranque pequeño siempre presente (una fuente de referencia, 3,2 KB por hook SessionStart, que también corre tras compactar; otra, 3,5 KB por hook más AGENTS.md) y el procedimiento bajo demanda.
- Otra fuente de referencia mantiene de 18 a 26 KB siempre cargados por host y controla la deriva con tests.
- openspec no escribe instrucciones globales.

### Documentación de los hosts

- En ninguno de los seis hay una carga garantizada de una skill.
- Solo Claude Code reinyecta contexto tras compactar, con un hook y un tope de 10.000 caracteres. Los cuerpos de skills reinyectados se truncan a 5.000 tokens.
- Codex: el archivo global está exento de los 32 KiB (lo confirmó el usuario; ya se había evaluado antes).
- Cursor CLI no carga ningún archivo global.

### Valor de un host vanilla

- Nunca se midió una línea base vanilla: todas las corridas comparan variantes de Hive.
- Los hosts cubren lo genérico: concisión, commit solo si se pide (Claude y Codex), no revertir trabajo ajeno y cambios mínimos.
- No cubren autorización por efectos, secretos, idioma, proyecto actual ni flujos.
- Pi trabaja en modo YOLO, sin confirmaciones. Codex y OpenCode-astra empujan a actuar.
- Quitar lo cubierto ahorra de 1 a 3 KB.
- `global.md` creció 54 % desde la poda de 2026-10-15 (de 27.318 a 42.001 B).
- Externo: Gloaguen 2026 encuentra que los archivos de contexto apenas mejoran el éxito y suben el costo, pero recomienda usarlos para prácticas no estándar. La guía de Claude Code dice que un archivo largo reduce la adherencia.

### Propuesta en tres capas

1. Invariantes siempre cargados con disparadores literales.
2. Skill core con el procedimiento, leída por ruta desde esos disparadores.
3. Reinyección por hook solo donde el host la soporte; choca con la política de no tener sistema de hooks y es decisión del usuario.

Falta una medición: un brazo sin Hive, el actual y el de tres capas.

**Discrepancias:**
- La memoria `project_deprecate_rules_dir.md` cita `27e804b`, pero en este repositorio la retirada es `abad754d`. Puede ser un commit del repositorio privado; no se corrigió.
- Las cifras de Vercel difieren entre las fuentes secundarias.

## Verificación del juez

Se abrieron las citas que deciden H1, H2, H3, H4, H8, H9, H11, H13, H14 y P3 (claves retiradas) y se confirmaron. Los tamaños se midieron con `wc -c`. El resto se apoya en la cita del dueño y, en las fronteras, en el acuerdo de dos partes; el acuerdo no prueba nada por sí solo.

**No revisado:** las páginas `site/`, los detectores de `tests/pilot/` (no cubren comportamiento de estos flujos) y la carga real de `global.md` por host.

**Hipótesis abierta:** `flow-deploy/references/urgent-fixes.md` podría editar código sin pasar por build (planteada por el dueño de build). El dueño de deploy afirma que `urgent-fixes.md:5` presupone build y verify. No se verificó.
