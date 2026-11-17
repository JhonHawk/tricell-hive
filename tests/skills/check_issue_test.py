import importlib.util
import pathlib
import textwrap
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/create-issue/scripts/check_issue.py"


def load_module():
    spec = importlib.util.spec_from_file_location("check_issue", SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


VALID_STORY = textwrap.dedent("""\
    # Exportar reportes de ventas en CSV

    ## Resultado
    Los administradores pueden descargar el reporte mensual en CSV para conciliar pagos.

    ## Para quién y por qué
    Para el equipo de finanzas, que hoy copia las tablas manualmente cada cierre de mes.

    ## Hecho cuando
    - El botón de exportar descarga un archivo CSV con las filas filtradas.
    - Un usuario sin rol administrador no ve la acción de exportar.

    ## Fuera de alcance
    - Exportación en formato Excel (.xlsx).

    ## Preguntas abiertas y supuestos
    - Supuesto: el delimitador es coma y codificación UTF-8.

    ## Dependencias
    - Ninguna.

    ## Pista, no vinculante
    Revisar `src/reports/export.ts` y `src/ui/ReportsTable.tsx` para reutilizar el filtro actual.
""")


class CheckIssueTests(unittest.TestCase):
    def setUp(self):
        self.mod = load_module()

    def test_valid_story_ticket_passes(self):
        res = self.mod.validate_issue(VALID_STORY, ticket_type="story")
        self.assertTrue(res["ok"], res["findings"])

    def test_more_than_7_criteria_fails_unless_finding_group(self):
        eight_criteria = "\n".join(f"- Criterio observable número {i}." for i in range(1, 9))
        over = VALID_STORY.replace(
            "- El botón de exportar descarga un archivo CSV con las filas filtradas.\n"
            "- Un usuario sin rol administrador no ve la acción de exportar.",
            eight_criteria,
        )
        res_normal = self.mod.validate_issue(over, ticket_type="story", finding_group=False)
        self.assertFalse(res_normal["ok"])
        self.assertTrue(any("7" in f for f in res_normal["findings"]))

        res_group = self.mod.validate_issue(over, ticket_type="story", finding_group=True)
        self.assertTrue(res_group["ok"], res_group["findings"])

    def test_bug_missing_repro_steps_or_hint_with_too_many_locations_fails(self):
        bad_bug = VALID_STORY.replace(
            "Revisar `src/reports/export.ts` y `src/ui/ReportsTable.tsx` para reutilizar el filtro actual.",
            "Revisar `a.ts`, `b.ts`, `c.ts`, `d.ts`, `e.ts` y `f.ts`. Segunda oración prohibida.",
        )
        res = self.mod.validate_issue(bad_bug, ticket_type="bug")
        self.assertFalse(res["ok"])
        joined = "\n".join(res["findings"]).lower()
        self.assertIn("bug", joined)
        self.assertIn("hint", joined)

    def test_empty_required_sections_fail(self):
        empty_outcome = VALID_STORY.replace(
            "Los administradores pueden descargar el reporte mensual en CSV para conciliar pagos.",
            "",
        ).replace("- Ninguna.", "")
        res = self.mod.validate_issue(empty_outcome, ticket_type="story")
        self.assertFalse(res["ok"])
        joined = "\n".join(res["findings"]).lower()
        self.assertIn("outcome", joined)
        self.assertIn("dependencies", joined)


if __name__ == "__main__":
    unittest.main()
