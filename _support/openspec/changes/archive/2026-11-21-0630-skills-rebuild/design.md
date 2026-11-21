# Diseño — flujo objetivo y capa global (F1)

Estado: decidido el 2026-10-30. El usuario eligió la opción recomendada en D1–D5, añadió D4b y aprobó el reparto de `hive-core`; D6 se mantiene como está. El diseño híbrido estricto ya estaba elegido (ver `tasks.md`, F1).

## Contexto verificado

### Costo de COM-829 por fase (Grok `01a18c56`, cam, 2026-10-30)

Hilo principal, según los turnos de `usage.json`, sumado a los subagentes que registraron uso. Las horas son CST.

| Fase | Turnos | Hilo principal | Subagentes | Lo que aportó |
|---|---|---|---|---|
| Research del ticket, plan, T1 y T2 | 2 (04:33–05:37) | 36.7M · $7.93 | `hive-research` $0.23 · `hive-review-plan` ×3 $0.77 | El plan, sin la región de acciones |
| Build T3–T6 con verificador por tarea | 3–8 (05:41–07:42) | 68.4M · $14.39 | `hive-verify-task` ×5 ≥ $0.99 (dos sin registro) | 0 defectos nuevos; confirmó un fallo previo |
| UI gate | 9–13 (07:51–08:06) | 20.5M · $4.10 | `hive-review-ux` ×2 $3.94 · `hive-verify-change` $1.16 | H1/H2 Medium, que no bloquearon |
| Suite, code review, refutación, re-revisión | 14–16 (08:06–08:54) | 27.6M · $5.56 | revisor, 2 refutadores, verify-task T6, re-revisión (sin registro) | 2 P1 reales (uno confirmado, otro plausible); la refutación no cambió nada; la re-revisión encontró 0 |
| Walkthrough y V1 | 17 (08:54–09:36) | 4.6M · $1.10 | — | Recomendó aprobar con H1/H2 dentro |
| Ronda de corrección de los botones | 18–21 (12:17–14:52) | 48.5M · $10.71 | — | El arreglo de lo que el plan no decidió |

El defecto visual, que no se decidió antes de construir, costó $10.71 en su corrección, más que toda la UI gate.

### Valor medido por rol (reporte `_support/sessions/2026-10-29-token-consumption/`, todos los hosts)

| Rol | Corridas | Hallazgos altos (C1 funcional + C2 seguridad) | Altos por corrida | Nota |
|---|---|---|---|---|
| `hive-review-security` | 26 | 37 | 1.42 | Solo corre a pedido |
| `hive-review-code` | 54 | 19 | 0.35 | |
| Revisión nativa del host | 171 | 55 | 0.32 | |
| `hive-review-plan` | 206 | 44 | 0.21 | Riesgos evitados, no fallas observadas; 26 % brechas de verificación |
| `hive-verify-change` | 91 | 11 | 0.12 | |
| `hive-review-ux` | 51 | 5 | 0.10 | 15 Medium |
| `hive-verify-task` | 649 | 20 | 0.03 | 60 % de lo que reporta es guía o estilo |
| `hive-refute-claim` | 49 | 1 | 0.02 | 34 % falsos positivos |

Una sola persona clasificó cada hallazgo por su descripción; el reporte estima que ~10 % de los casos límite podrían cambiar de clase.

## Flujo objetivo

```
 [Q] = pregunta al usuario    [C] = subagente    ★ = paso barato nuevo o reforzado

 ENTRADA ─ pregunta o investigación → RESEARCH
         ─ cambio chico entendido o "sin plan" → BUILD
         ─ resto → PLAN (o RESEARCH si no tiene alcance)

 RESEARCH   leer directo · [C] hive-research si crece · reporte
            [Q] ruta (opciones por ticket cuando preguntan qué tomar)

 PLAN       contratos, tareas y criterios de aceptación
            ★ decisiones de UI: patrón, región de acciones y jerarquía de botones
            [Q] una ronda: entrega · revisión · UI           (D2: [C] hive-review-plan)
            change record

 BUILD      por tarea: implementar · tests afectados · commit
            ★ primera vista renderizada en cuanto cambia un layout ([Q] si es interactivo)
            (D1: [C] hive-verify-task solo en tareas de riesgo)
            avance: una línea por tarea                      (D5)

 VERIFY     suite completa local · in-vivo
            UI: [C] hive-review-ux (+ [C] hive-verify-change si cambia el recorrido)   (D4)
            [C] code review, una vez                         (D3: sin refutación de rutina)
            [Q] validación con walkthrough, si la entrega para

 INTEGRAR   push · PR · CI · merge según la entrega
 CLOSE      limpieza · reporte · memoria
```

Comparado con hoy:
- Desaparecen el verificador por tarea de rutina, la refutación de rutina, el registro de IDs, modelos y duraciones, y el montaje de la base aparte por defecto.
- Entran o se refuerzan dos pasos antes de construir: las decisiones de UI en el plan y la primera vista renderizada.

## Ceremonias: decisiones

| # | Ceremonia | Hoy | Propuesta (recomendada) | Por qué |
|---|---|---|---|---|
| D1 | Verificador por tarea | Cada tarea de un plan, más los recheques | Solo tareas de riesgo (datos guardados, migración, seguridad, dinero, concurrencia, contrato público); el resto lo cubren sus tests y el verify final | 0.03 altos por corrida en 649 corridas; en COM-829, 5 corridas sin defectos nuevos |
| D2 | Revisión del plan | Un revisor por dominio afectado (3 en COM-829) | Un solo revisor, y solo si el plan cruza contrato, datos, seguridad o dependencias; ninguno en los demás | 0.21 por corrida, en riesgos evitados; vale en los planes riesgosos, sobra triplicado |
| D3 | Code review | Revisión, refutación de cada P0/P1, arreglo, recheque y re-revisión | Una revisión; el hilo principal comprueba las líneas citadas; re-revisión solo si un arreglo cambia un contrato; sin `hive-refute-claim` de rutina | La revisión es lo que más rinde (0.35); la refutación rinde 0.02 con 34 % de falsos positivos |
| D4 | UI gate | `hive-review-ux` + `hive-verify-change`, base servida en otro puerto, 5 viewports × 2 temas | La primera vista en el build; un `hive-review-ux` contra la guía de UI, sin servir la base salvo cambios de shell o layout compartido; `hive-verify-change` solo si cambia un recorrido; un Medium introducido en las vistas tocadas se corrige antes de validar o se ofrece como opción, nunca se recomienda aprobarlo | COM-829: la gate costó ~$9 y el defecto se pagó otra vez en la corrección ($10.71) |
| D4b | Gate tras aceptar una corrección | `hive-review-ux`, `hive-verify-change` y `hive-verify-task` vuelven a correr sobre lo que el usuario acaba de aprobar | Solo el code review del arreglo y los tests afectados. `hive-review-ux` solo si la corrección cambia vistas o tamaños de pantalla que el usuario no vio. El walkthrough de cada corrección incluye capturas en móvil (390 y 320) para que su aprobación cubra esos tamaños | COM-829, tramo final: 56 min entre la aceptación (13:10) y el merge; `hive-review-ux` fue el camino crítico (25 min) y solo encontró un Medium a 390/320 que unas capturas móviles habrían mostrado |
| D5 | Registro de avance | IDs de verificadores, modelo «configurado, no observado», segundos, citas releídas | Una línea por tarea: marcador, commit y lo que quedó sin verificar | Nadie lo lee y alarga cada turno |
| D6 | Preguntas de entrega, revisión y UI | Una ronda al cerrar el plan o al iniciar el build, reutilizada en la sesión | Igual (se mantiene) | Ya se pregunta una vez; el problema no estaba ahí |

## Capa global

`global.md` lleva solo los invariantes que toda sesión necesita, cada uno completo, sin punteros «antes de X, lee Y». Hoy pesa 19 169 bytes; la meta es no crecer pese a absorber invariantes, compensando con los ~20 punteros que se quitan (< 200 líneas; Codex corta `AGENTS.md` en 32 KiB).

Sus secciones: idioma; comunicación (línea de contexto, empezar por la respuesta, glosar IDs); el contrato mínimo de las preguntas; alcance y autorización (incluido el invariante de efecto crítico); preservación; secretos; evidencia; proporcionalidad y ruteo; los ajustes del proyecto (claves de `## Hive` y cómo se resuelven); memoria (cadencia y nombre de proyecto); carpeta de soporte.

### Reparto de las referencias de `hive-core`

| Referencia | Bytes | Lecturas en sesiones principales | Destino propuesto |
|---|---|---|---|
| `questions` | 15 981 | 120 | Contrato mínimo → global; preguntas propias → cada flujo (ruta en research, entrega en plan, validación en verify, cierre en close) |
| `project-settings` | 5 605 | 84 | Claves y resolución → global; casos multi-repo → referencia rara de plan |
| `effect-authorization` | 6 482 | 19 | Invariante → global; autorizaciones previstas → plan y build |
| `secrets` | 2 674 | — | Invariantes → global (ya están); obtener o documentar un secreto → referencia rara con disparador |
| `language` | 2 760 | 7 (96 con subagentes) | Invariante → global; nombres e i18n → roles de construcción |
| `memory` | 2 148 | 14 | → global (una línea más sobre el nombre de proyecto) |
| `evidence-tools` | 1 706 | — | Invariantes → global; detalle de ctx7 y CodeGraph → research |
| `third-party-requests` | 1 559 | 3 | Invariante → global; detalle → research |
| `option-scope` | 2 261 | 24 | → research (opciones de ruta) y plan (opciones de alcance) |
| `scope-changes` | 7 075 | 6 | Clasificación en 5 líneas → global; el resto se va |
| `loops` | 5 501 | 9 | → referencia rara de research (ruta iterativa) y `unattended-delegation` |
| `context` | 2 370 | 2 | Una línea en el cierre de fase de plan y verify; el resto se va |
| `tool-scripts` | 1 144 | 2 | Se va (sin evidencia de uso) |

## Presupuesto por skill

- `SKILL.md`: ≤ 16 000 bytes (≈ 4 000–4 500 tokens), por debajo de los 5 000 tokens que Claude Code vuelve a adjuntar tras compactar. Lo crítico va arriba.
- Referencia rara: ≤ 8 000 bytes, un solo nivel, con disparador explícito en el `SKILL.md`.
- La prueba sigue midiendo bytes (`tests/content/skill_size_test.go`). Añadir un tope total por skill (SKILL más referencias de su camino principal) queda a decisión en F7.
