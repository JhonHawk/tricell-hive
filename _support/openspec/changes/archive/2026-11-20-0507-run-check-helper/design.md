# Diseño

## Forma del helper (D3)

| Enfoque | Qué hace | Costo | Elegido |
| --- | --- | --- | --- |
| Script Python `run_check.py` | Recibe el comando como argumentos tras `--`, escribe su salida a un archivo contando bytes, mata el grupo al pasar el tope, cierra el archivo y después imprime el final y sale con el código del comando | Un script y su prueba, como `work_time.py` | Sí (D3 A) |
| Script de shell con `ulimit -f` | Tope por el kernel | `SIGXFSZ` sin mensaje; el orden de redirección en Bash es lo que falló | No |
| Solo texto | La hoja describe la forma segura | Depende de que el modelo escriba bien el shell | No |

## Dónde vive cada cosa

- `hive-rule-checks/SKILL.md`: una frase que cambia «or save the whole output to a file in the host's temporary location and filter the file» por correrlo con el helper, enlazando su referencia. El tamaño neto queda en cero o menos, porque la huella de lectura de plan-build tiene 17 B de margen (revisión del plan, F2).
- `hive-rule-checks/references/run-check.md` (nuevo, condicional, fuera de la huella): cuándo se usa, la interfaz, la prohibición de envoltorios que lean su propio log y la forma de respaldo cuando `python3` no corre.
- `hive-rule-checks/scripts/run_check.py` (nuevo).
- La alternativa `${PIPESTATUS[0]}` para un pipe directo sigue permitida: un pipe no escribe en un archivo que se lea a sí mismo. El helper es obligatorio para una salida que se **guarda**.

## Interfaz y comportamiento

```
python3 -I <hive-rule-checks-dir>/scripts/run_check.py [--log PATH] [--max-bytes N] [--tail N | --grep REGEX] -- <comando> [args…]
```

- Compatible con Python 3.9 (`/usr/bin/python3` de macOS): sin `process_group=`; usa `start_new_session=True`.
- Escribe `log: <ruta>` en stdout y lo vacía antes de lanzar el comando. Sin `--log`, crea `hive-check-*.log` con `tempfile.mkstemp` en `tempfile.gettempdir()`. Con `--log`, abre con `O_CREAT|O_EXCL|O_NOFOLLOW` y rechaza un archivo o un enlace que ya existan (exit 2).
- Lanza el comando sin shell, con stdin de `/dev/null`, stdout y stderr al mismo pipe, en una sesión nueva. Un comando compuesto se pasa como `bash -c '…'` de forma explícita.
- Copia bytes crudos del pipe al log y los cuenta. Al pasar `--max-bytes` (64 MiB por defecto) deja de escribir, envía `SIGTERM` al grupo y, a los 5 s, `SIGKILL`; marca el resultado como tope.
- Cuando el hijo directo termina, espera como mucho 5 s a que el pipe llegue a EOF (un nieto puede tenerlo abierto); después mata el grupo y cierra el pipe. En toda salida (normal, excepción, `SIGINT`, `SIGTERM`, `SIGHUP`) mata el grupo en un `finally`; las señales recibidas se reenvían al grupo.
- Cierra el log y solo entonces lo lee para imprimir: las últimas `--tail` líneas (40 por defecto) o hasta 40 líneas que coinciden con `--grep` (decodificadas con `errors="replace"`), sin pasar de 16 KiB en total, escribiendo bytes a `sys.stdout.buffer`; una última línea sin salto se imprime igual. Imprime el final también cuando salta el tope.
- Código de salida: el del comando; 128+n si una señal lo terminó; 127 si el comando no existe o no se puede ejecutar; 125 cuando salta el tope, con una línea en stderr `run_check: stopped at <N> bytes; log: <ruta>` (125 también lo usan `docker run` y `git bisect run`: la línea de stderr es lo que lo distingue).
- Respaldo sin `python3` (en la referencia): correr el comando con su salida a un archivo en un paso y leer el archivo en un comando aparte, nunca en un `trap` ni dentro del bloque redirigido.

## Criterio del auditor

`self_feeding_log` detecta un log que se alimenta a sí mismo de verdad, no cualquier redirección:

- **Fuente del texto:** el comando de un evento de shell. Para un `exec` de Codex cuyo literal se interpola (`${…}`), el extractor guarda el texto con los `${…}` sin resolver en un campo nuevo, `CommandTemplate`; `Command` sigue vacío y `CommandWithheld` sigue en `true`, así ningún otro criterio cambia de resultado. Solo `self_feeding_log` lee `CommandTemplate`.
- **Falla** cuando, en un mismo comando, un archivo `F` (una ruta literal o el nombre de una variable, sin resolver su valor) es destino de la redirección de un bloque compuesto (`{ …; } >F`, `( … ) >F`, `>>F`, `&>F`) o de `exec >F`, y `F` lo lee un `trap … EXIT` (u otra señal) o un comando dentro de ese bloque (`tail`, `head`, `cat`, `grep`, `rg`, `sed`, `less`).
- **No falla** (`not_observed`): un comando simple con redirección más un `trap` (Bash deshace esa redirección antes del `trap`); `{ …; } >F 2>&1 || tail F`; la salida a `F` en un comando y la lectura en otro; `run_check.py`; el patrón dentro de una cadena o un heredoc de Python.
- Sin `review`. Grok y los hijos: el mismo análisis cuando la traza trae el comando.
