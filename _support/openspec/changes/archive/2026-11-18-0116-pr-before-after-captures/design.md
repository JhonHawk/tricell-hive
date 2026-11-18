# Design

## Approaches for the «before» capture (D2)

- **A — Al empezar el build (elegida).** Antes de la primera edición de cualquier cosa que una pantalla existente muestra, el hilo principal sirve el checkout sin cambios y captura las regiones que el plan cambia, antes de delegar. Coste nuevo: hoy el build solo sirve la app para la primera vista, después de la tarea de UI y a veces con datos de prueba, así que A adelanta un arranque de la app con los datos que tenga el checkout sin cambios. Si el build ya editó la pantalla, el «antes» no se reconstruye y el PR lo dice; el checkpoint guarda sus rutas para retomarlo tras una compactación.
- **B — De la base en verify.** Un worktree de la rama base, compilado y servido con los mismos datos, captura los mismos estados que el candidato. Compara exactamente lo mismo, pero suma una segunda compilación y servidor por cada cambio de interfaz.
- **C — Solo «después».** Las capturas que verify ya toma van al PR sin columna «antes». No da la comparación pedida.

## Where the rule lives

- La captura «antes» vive en `flow-verify/references/ui-verification.md`, que ya gobierna qué se captura en un cambio de interfaz, incluida la primera vista del build; `flow-build/SKILL.md` solo adelanta el momento de leerla.
- La tabla del PR vive en una referencia nueva de `git-workflow`, porque `git-workflow/SKILL.md` (paso 6, hogar del cuerpo del PR) está a 70 bytes del límite de 16 000; el paso 6 lleva solo el puntero con su condición.
- El borrado al cerrar reutiliza la acción **Delete** de `support-material.md`, que ya admite material que el cuerpo de un PR fusionado reemplaza; solo se precisa la comprobación de la subida.

## Viewer (D6)

Las capturas «antes» no se publican en el visor: con el mismo título aparecerían como iteración #1 y la primera vista real como #2, y el visor guarda su propia copia de cada imagen. El visor y su referencia no cambian.

## Who sees an uploaded image (D3, D7)

Según la documentación de GitHub, en repos privados e internos de github.com solo quien tiene acceso al repo ve la imagen; en repos públicos, cualquiera; en GitHub Enterprise Server, cualquiera con la URL. Quitar el enlace no borra el archivo: solo el soporte de GitHub lo hace. Por eso un repo privado de github.com queda cubierto por la entrega, y un repo público o GitHub Enterprise Server exige datos de prueba (D3 A) y una autorización única por trabajo (D7 A). La visibilidad se detecta con `gh repo view --json visibility` y el host con la URL del remoto.
