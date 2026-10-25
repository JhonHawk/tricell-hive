# Diseño

## Contexto verificado (base `e438e14b`, 2026-10-25)

- `tdd.md` decide cuándo escribir la prueba primero y prohíbe dobles de los módulos propios, pero no habla de propiedades, de duplicados ni de puntos de entrada solo para pruebas (`rg -i "propert|seam|already owns"` sin resultados en `content/`).
- `hive-write-tests.md:15` pide nombrar el cambio realista que haría fallar cada prueba; es la base del filtro y se conserva.
- `verification.md:13-20` elige evidencia por superficie; no tiene fila para reemplazos.
- `flow-plan/SKILL.md:26` pide decidir qué caminos viejos se retiran al reemplazar un sistema, sin pedir comparación de comportamiento.
- Evidencia de necesidad: investigación del 2026-10-25 sobre dos proyectos de clientes, resumida en [#157](https://github.com/JhonHawk/tricell-hive/issues/157).

## D-1 — Hogar canónico

`tdd.md` es el hogar de qué prueba escribir: `flow-build` lo lee antes de tocar código o pruebas y `flow-plan` al elegir el enfoque de prueba. `hive-write-tests` solo remite. La comparación viejo contra nuevo vive en `verification.md`, porque es evidencia de verificación de un reemplazo, no una regla de prueba unitaria; `flow-plan` remite.

## D-2 — Contenido de la sección «Choose the kind of test»

- **Propiedades:** para lógica pura con un oráculo independiente del código bajo prueba (round-trip, idempotencia, invariantes, una versión optimizada contra la simple, un sanitizador contra un predicado independiente). No para un espacio finito que una tabla ya recorre entero. Sin librería (E1-A): el plan la propone como dependencia de desarrollo y remite a «Check a dependency before installing» de `security-boundaries.md`, sin repetir sus pasos. Un build sin plan o a mitad de tarea escribe pruebas por ejemplo, reporta el límite y propone la librería como seguimiento.
- **Antes de escribir cada prueba:** nombrar el cambio realista que la haría fallar (frase que hoy está en `hive-write-tests.md:13` y pasa aquí); buscar la prueba que ya es dueña de ese comportamiento en el límite más fuerte y extenderla en vez de duplicarla; rechazar los patrones de AC3, que absorben la lista de `hive-review-code.md:12`.
- **Puntos de entrada solo para pruebas:** no agregarlos; probar por el límite real. El punto de inyección en el límite del sistema (reloj, aleatoriedad, servicio externo) que producción usa con la implementación real está permitido, en línea con la regla de dobles de `tdd.md:9`.
- **Descripciones de `tdd.md`:** `tdd.md:5` y el puntero de `flow-build/SKILL.md:42` pasan a decir que la referencia también elige el tipo de prueba y filtra antes de escribirla (`flow-build/SKILL.md` tiene unos 790 bytes de margen).

## Clasificación según AGENTS.md, «Measurement»

Las reglas son funciones nuevas tomadas de referencias externas (el artículo de Addy Osmani y la skill `test-audit` de OpenClaw), con evidencia de necesidad en dos proyectos. Ninguna corrige una regla que haya fallado en una sesión real. Las adopta la decisión del usuario: ruta D2-A y E1-A del 2026-10-25. Endurecer la lista de `hive-review-code.md:12` es un movimiento a un solo hogar, no un cambio de criterio.

## D-3 — Fila «Replacement or migration»

Evidencia útil: las mismas entradas (aleatorias o una muestra de datos que la tarea está autorizada a usar) por el camino viejo y el nuevo, ruido normalizado, cada diferencia clasificada como defecto o cambio documentado en el plan. Disparador de escalamiento: datos persistidos o efectos externos, que siguen [persistent data changes](../../../../content/skills/flow-build/references/data-changes.md).
