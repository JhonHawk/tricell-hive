# Design

## Approaches

Un solo camino es sensato: cambiar el criterio de qué es una entrada en su único hogar, «Local instructions» de la guía global, y hacer que la skill que ya cura los archivos con el usuario parta las entradas largas. Se descartó un validador: mediría el largo de una línea, pero no distingue un hecho de una bitácora, y sería maquinaria nueva para archivos que no están en el repositorio.

## Decisions

- **La línea dice qué aplica; el detalle va a la memoria persistente.** El tema es `local/<slug>`, nombrado por el asunto del hecho tras buscar uno existente, porque dos hechos con el mismo tema se sobrescriben. El detalle de un archivo de proyecto va al Engram de ese proyecto; el del archivo de máquina, al proyecto fijo `hive-machine` (D12 A). El puntero usa la forma que «Persistent memory» ya define, `engram:<project>/<topic_key>`, y la regla del nombre de proyecto admite el que un puntero nombra.
- **Cuándo hay detalle.** Cuando la línea no puede contener el comando con sus opciones en 200 caracteres, o reutilizar el hecho toma dos o más pasos.
- **Una limitación con su solución es un hecho anotable** (D13 A): la línea dice la limitación y los pasos van al detalle.
- **El puntero tiene ciclo de vida.** Un tema pertenece a la línea que lo señala: solo se actualiza desde el archivo de esa línea, porque varios archivos pueden compartir un proyecto de Engram. Un detalle que falla deja su línea obsoleta y se corrige cuando otra vía funciona; una línea quitada deja su tema, sin borrar memoria. El proyecto `hive-machine` se crea por la línea de comandos de Engram cuando la herramienta de memoria rechaza el nombre.
- **Los hijos reciben el detalle por el brief.** Siguen leyendo los archivos (D11 A), pero no se les pide leer la memoria persistente.
- **El detalle se lee solo cuando la tarea necesita el hecho.** No se carga al arrancar: ese es el ahorro.
- **«Persistent memory» gana un momento de escritura.** El detalle de un hecho local se guarda cuando se registra su línea; la excepción se nombra también en la frase que prohíbe guardar por descubrimiento, porque sin ella un modelo literal lo omitiría.
- **Sin Engram, la línea sola.** La memoria propia de un host no la leen los otros hosts de la máquina, así que un puntero hacia ella no serviría a la siguiente sesión.
- **Lo que no es un hecho de la máquina tiene otra casa.** Una instrucción del usuario es una decisión de la memoria persistente; el archivo y la librería de diseño van en `DESIGN.md`, y su línea se conserva hasta que ese registro los tenga, porque es lo único que deja dibujar en ese archivo; lo que vale para todo clon va a la guía versionada, como ya dice la regla.
- **Cuentas.** Las cuentas de prueba de un entorno siguen en `ACCOUNTS.md`; qué cuenta o identidad usa una herramienta en esta máquina es un hecho del archivo, sin su credencial.
- **La limpieza usa el modo de curación de la skill, con una excepción que el usuario pidió:** recorre los ocho archivos de la máquina en una sesión, no solo los del proyecto actual. Antes de reescribir nada se copian los ocho archivos, tal cual, a la carpeta ignorada de la limpieza, de modo que nada se pierde; el detalle de cada entrada partida se guarda en Engram y se lee de vuelta antes de tocar el archivo. Lo retirado no se vuelca a Engram: quedaría como memoria basura.
- **La tabla de limpieza no entra al repositorio.** Nombra hosts, cuentas y proyectos privados: vive en `_support/workspace/` (ignorado por Git) y se muestra en la conversación.

## Proposed shape

Forma de una entrada:

```
- <qué tiene, usa o aplica aquí> · <fecha> [· unverified] [→ engram:<project>/local/<slug>]
```

Ejemplos del tipo que el usuario pidió:

```
- acli signs in to Jira with the work account; attachments go through the REST API · 2026-11-10 → engram:<project>/local/jira-attachments
- This repository commits with the client's Git account, one of three on this machine · 2026-11-13
- Local stack: containers <names> on ports 5432 and 3000 · 2026-11-12 → engram:<project>/local/start-stack
```

Clases para la limpieza de cada entrada existente:

1. **Queda** como línea corta: ya dice qué aplica aquí.
2. **Se parte**: línea corta más un tema `local/…` en Engram con el procedimiento o la solución.
3. **Sale**: medición, evento, diagnóstico, estado de un trabajo, duplicado, obsoleta, o hecho que el repositorio ya dice; su texto queda en la copia literal de la carpeta de limpieza.
4. **Se muda**: instrucción del usuario a una decisión de Engram.
5. **Se reubica**: de archivo de máquina a archivo de proyecto o al revés, con su destino en la tabla.

La tabla lleva además una columna que marca la entrada cuyo detalle contiene un valor que debe sustituirse por el archivo que lo guarda.

## Impact

- Cada sesión y cada subagente cargan menos al arrancar: de unos 18 KB a unos 5 KB en el proyecto más cargado (estimado hasta hacer la limpieza).
- Reutilizar un procedimiento cuesta una lectura de Engram cuando la tarea lo necesita.
- En una máquina sin Engram el procedimiento no se guarda y se vuelve a descubrir.
- Otras sesiones abiertas pueden escribir en los mismos archivos durante la limpieza: se leen de nuevo justo antes de escribir y solo se tocan las entradas de la tabla aprobada.

## Recovery

`git revert` del merge y un refresco de la instalación devuelven la regla anterior. Para los archivos locales, la copia literal de la carpeta de limpieza restaura cualquiera de los ocho.
