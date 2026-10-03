# Historia y procedencia técnica de Hive

Preparado el 2026-10-02, America/Mexico_City. Este documento describe cinco hitos conservados en el historial de Hive y separa el contenido de las reglas, las fechas registradas por Git y la corroboración temporal disponible en GitHub. Documenta la evolución del proyecto; no demuestra quién inventó estas prácticas, su prioridad frente a otros proyectos ni independencia respecto de otros proyectos. No se realizó un estudio comparativo.

## Historial que puede contrastarse

La base congelada es `39dfdb62fc73a91f373f3e46769891f329cea405`; el candidato curado local es `0b30b459eb3f7ab8f7e91ba5c03e79b289e6b72c`. Ambos contienen 934 commits y 70 merges. La curación conserva nombres, fechas con zonas horarias, padres y su orden, incluidos los commits que quedaron vacíos. Sustituye correos por direcciones GitHub noreply verificadas, excluye material privado y añade atribuciones clasificadas. Estos cambios producen nuevos identificadores de commit; se retiraron 51 cabeceras de firma, sin comprobar su validez criptográfica.

Los cinco pares de commits siguientes pertenecen al mapa original/curado verificado durante la preparación. Los enlaces curados apuntan al destino previsto en este repositorio; su disponibilidad remota requiere la entrega posterior. Los identificadores originales son correspondencias históricas y no requieren que los objetos originales sigan accesibles. La política noreply se comprobó en el candidato; las referencias de PR y caches pueden conservar historia y correos anteriores.

## Cinco hitos conservados

Las fechas siguientes son fechas de autor y committer de Git, iguales entre sí para cada hito y preservadas en el candidato. Se expresan en UTC para evitar diferencias de día entre zonas.

1. **Delegación y verificación separada, 2026-06-26T10:32:02Z.** En la raíz conservada, `global/rules/workflow/agent-routing.md`, sección `Delegation Thresholds`, asigna exploraciones acotadas a un subagente cuando el flujo abarca varios archivos. La regla `Verification runs in fresh context` separa verificación e implementación. Esto acredita instrucciones de delegación general y verificación independiente, sin equipararlas todavía al contrato posterior de investigación en tres frentes. Original `d066234b87d101354091e54e09fd48a626753911` → [commit curado d91404d](https://github.com/JhonHawk/tricell-hive/commit/d91404deed85475e3d990a739b5e5297fac912df) · [archivo del hito](https://github.com/JhonHawk/tricell-hive/blob/d91404deed85475e3d990a739b5e5297fac912df/global/rules/workflow/agent-routing.md).

2. **Consultas externas delegadas, 2026-08-16T05:15:25Z.** En `global/rules/workflow/agent-routing.md`, sección `Delegation Gates`, aparece un umbral explícito para delegar varias lecturas externas o varias consultas necesarias para fijar un hecho. Es una ampliación documentada del reparto de investigación y recopilación de evidencia. Original `24663660af78ff22b87778468b029390c6a5133d` → [commit curado 874acff](https://github.com/JhonHawk/tricell-hive/commit/874acff0491f61cd81dd0e22c117fb23047b7b86) · [archivo del hito](https://github.com/JhonHawk/tricell-hive/blob/874acff0491f61cd81dd0e22c117fb23047b7b86/global/rules/workflow/agent-routing.md). Esa fecha corresponde al 15 de agosto en America/Mexico_City.

3. **Actividad específica de investigación, 2026-09-20T20:16:21Z.** Se incorpora `global/skills/flow-research/SKILL.md`, con investigación de código y documentos, contraste de afirmaciones, contradicciones y límites. Su selección no autoriza implementación o publicación. Original `0ebb5b242182b5a7b09e720c3f897eac0c301c0f` → [commit curado a083c88](https://github.com/JhonHawk/tricell-hive/commit/a083c88b0f5cb2a71fc5414b2cf246d23850f674) · [archivo del hito](https://github.com/JhonHawk/tricell-hive/blob/a083c88b0f5cb2a71fc5414b2cf246d23850f674/global/skills/flow-research/SKILL.md).

4. **Tres frentes y revisión crítica, 2026-09-21T19:29:34Z.** `AGENTS.md`, sección `Complete research`, exige revisar la implementación previa de Hive, evidencia externa y fuentes de referencia. Delega los tres frentes a subagentes independientes en paralelo cuando hay capacidad. El hilo principal contrasta fuentes, cuestiona supuestos, resuelve contradicciones y sintetiza, sin aceptar una conclusión sólo porque la afirmó otro agente. La curación conserva ese contrato y generaliza la ubicación de fuentes privadas. Original `5a51c8f8788560f977757887d8f19aa1d6412539` → [commit curado c43ca2b](https://github.com/JhonHawk/tricell-hive/commit/c43ca2bbfb86ec2add88c3c42d6381862de74799) · [archivo del hito](https://github.com/JhonHawk/tricell-hive/blob/c43ca2bbfb86ec2add88c3c42d6381862de74799/AGENTS.md).

   ```text
   Implementación Hive ----\
   Evidencia externa -------+--> Hilo principal: contraste y síntesis
   Fuentes de referencia --/
   (subagentes independientes, en paralelo cuando hay capacidad)
   ```

5. **Rol explícito de investigación, 2026-09-29T07:12:07Z.** `content/agents/review/hive-research.md` define un rol para investigación delegada de sólo lectura sobre código o fuentes contradictorias, con síntesis apoyada en evidencia. El rol conservado documenta la intención y el contrato; su existencia no prueba que todos los hosts lo carguen o lo ejecuten correctamente. Original `9b110408cce6555a8b68e37e601d8d54f2e4507f` → [commit curado 1262b1a](https://github.com/JhonHawk/tricell-hive/commit/1262b1a1a8a143f3cd4b3cabae03cbdf793d7e10) · [archivo del hito](https://github.com/JhonHawk/tricell-hive/blob/1262b1a1a8a143f3cd4b3cabae03cbdf793d7e10/content/agents/review/hive-research.md).

## Qué corroboran las fechas externas

La API de GitHub de las PR enlazadas a continuación fue consultada el 2026-10-03T02:12:22Z. Las fechas Git son metadatos modificables; también lo son cuando GitHub los devuelve. La fecha de creación de una PR fecha su contenedor, sin demostrar qué commits contenía entonces. Un commit que es antecesor del merge registrado por GitHub tiene una corroboración posterior de presencia, sin demostrar su fecha de concepción.

- [PR2](https://github.com/JhonHawk/tricell-hive/pull/2), creada 2026-08-13T07:27:49Z y fusionada 2026-08-13T07:38:46Z: la raíz d066234 es antecesora del merge. No es miembro exacto de la lista de cambios de esta PR; no se afirma que PR2 la introdujera.
- [PR3](https://github.com/JhonHawk/tricell-hive/pull/3), creada 2026-08-17T09:31:24Z y fusionada 2026-08-17T11:36:33Z: el hito de consultas externas es antecesor del merge, sin membresía exacta en sus cambios.
- [PR48](https://github.com/JhonHawk/tricell-hive/pull/48), creada 2026-09-29T06:11:39Z y fusionada 2026-09-29T06:11:58Z: flow-research y los tres frentes son antecesores del merge, sin membresía exacta en sus cambios.
- [PR51](https://github.com/JhonHawk/tricell-hive/pull/51), creada 2026-09-29T07:14:42Z y fusionada 2026-09-29T07:22:50Z: el commit de hive-research es miembro exacto de la lista consultada y antecesor del merge.

La relación entre [PR1](https://github.com/JhonHawk/tricell-hive/pull/1), fusionada el 26 de junio, y la raíz actual no quedó demostrada en la exploración acotada. No se usa como corroboración de esa raíz ni se afirma que la relación sea inexistente. La primera corroboración observada para la raíz entre los merges comprobables es la de PR2, en agosto. Las PR y sus relaciones se verificaron mediante API con autenticación antes de la transición. Su acceso público depende de la visibilidad del repositorio. Los enlaces curados se comprobaron contra objetos locales; no se afirma que ya estén disponibles remotamente.

## Atribuciones y reproducción

Conservar la evolución de Hive incluye reconocer adaptaciones reales. Los recursos adaptados de diagram-design tienen un [aviso MIT completo](../../../content/skills/flow-report/references/license-diagram-design.md). La historia curada conserva además el aviso Superpowers asociado al adaptador histórico en [global/hooks/flow-session-context/LICENSE.superpowers](https://github.com/JhonHawk/tricell-hive/blob/8a2abb1a42246620dd38069379887d918812ddc5/global/hooks/flow-session-context/LICENSE.superpowers), en sus instantáneas aplicables. Estos créditos no se eliminan para sostener una afirmación de originalidad.

Para contrastar un hito, desde una copia del repositorio con el historial curado, sustituye `<curated-commit>` y `<repository-path>` por el hito de esta cronología:

```sh
git show <curated-commit>:<repository-path>
git show -s --format='%H%n%aI%n%cI' <curated-commit>
```

Este documento fue preparado el 2026-10-02 para una incorporación documental posterior. No existía en los 934 commits anteriores ni se insertó retrospectivamente en ellos. La preparación actual deja sus archivos sin commit; el commit documental futuro requiere autorización expresa y una fecha actual. Añadirlo después no demuestra que el documento existiera en las fechas de los hitos. La evidencia acredita reglas conservadas y relaciones históricas concretas; no sustituye pruebas de comportamiento, evidencia criptográfica de fechas o un estudio comparativo de prioridad.
