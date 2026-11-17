import importlib.util
import pathlib
import textwrap
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/create-prd/scripts/check_prd.py"


def load_module():
    spec = importlib.util.spec_from_file_location("check_prd", SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


VALID_SPANISH_PRD = textwrap.dedent("""\
    # Reembolsos — v1 · iterating · 2026-11-16 · Owner: Ana

    ## Problema
    Hoy las solicitudes de reembolso se revisan por correo sin trazabilidad. Esto retrasa los pagos 5 días.

    ## Objetivos
    - Reducir el tiempo de resolución de reembolsos.

    ## Fuera de alcance
    - No incluye reembolsos en criptomonedas.

    ## Historias de usuario
    - Como supervisor, quiero aprobar reembolsos mayores a $500 para controlar el riesgo (R1, R2).

    ## Requisitos

    ### R1 · P0 · decidida

    Todo reembolso mayor a $500 requiere aprobación de un supervisor antes de pagarse.

    - Fuente: Ana, 2026-11-10
    - Veredicto: contradice; hoy se paga sin aprobación

    ### R2 · P1 (propuesta) · por decidir (Ana responde)

    El solicitante recibe notificación al aprobarse el reembolso.

    - Fuente: Documento inicial, sección 2
    - Veredicto: nueva

    ## Criterios de aceptación
    - Al solicitar un reembolso de $600 queda en espera de supervisor (R1).
    - Al aprobarse el reembolso se registra el envío de notificación (R2).

    ## Métricas de éxito
    - Tiempo medio de aprobación: hoy 5 días, meta < 24 horas.

    ## Preguntas abiertas
    - ¿Qué canal usa la notificación de R2? Responde: Ana (R2).

    ## Historial de versiones
    - v1 · 2026-11-16 · iterating
      - Agregadas: R1, R2
""")


class CheckPrdTests(unittest.TestCase):
    def setUp(self):
        self.mod = load_module()

    def test_valid_spanish_prd_passes(self):
        res = self.mod.validate_prd(VALID_SPANISH_PRD)
        self.assertTrue(res["ok"], res["findings"])
        self.assertEqual(res["findings"], [])

    def test_missing_section_or_table_rule_or_duplicate_code_or_vague_term_fails(self):
        # Duplicate R1 and vague term "etc." and table row rule
        bad = VALID_SPANISH_PRD.replace(
            "### R2 · P1 (propuesta) · por decidir (Ana responde)",
            "| R2 | P1 | regla en tabla |\n\n### R1 · P1 · decidida",
        ).replace("antes de pagarse.", "antes de pagarse, etc.")
        res = self.mod.validate_prd(bad)
        self.assertFalse(res["ok"])
        joined = "\n".join(res["findings"])
        self.assertIn("R1", joined)
        self.assertIn("table", joined.lower())
        self.assertIn("vague", joined.lower())

    def test_cancelled_p0_rule_exempt_from_ac_and_empty_sections_or_rule_body_fail(self):
        # Cancelled P0 rule does not require an AC citation.
        cancelled_only = (
            VALID_SPANISH_PRD.replace(
                "### R1 · P0 · decidida",
                "### R1 · P0 · cancelada",
            )
            .replace(
                "- Al solicitar un reembolso de $600 queda en espera de supervisor (R1).\n",
                "",
            )
        )
        res_cancelled = self.mod.validate_prd(cancelled_only)
        self.assertTrue(res_cancelled["ok"], res_cancelled["findings"])

        # Bold/italic markdown rule statement must pass
        bold_rule = VALID_SPANISH_PRD.replace(
            "Todo reembolso mayor a $500",
            "**Todo** reembolso mayor a $500",
        )
        res_bold = self.mod.validate_prd(bold_rule)
        self.assertTrue(res_bold["ok"], res_bold["findings"])

        # Empty required section (e.g., Objetivos) and rule without rule-statement paragraph must fail.
        missing_body = (
            VALID_SPANISH_PRD.replace(
                "- Reducir el tiempo de resolución de reembolsos.",
                "",
            )
            .replace(
                "Todo reembolso mayor a $500 requiere aprobación de un supervisor antes de pagarse.\n\n",
                "",
            )
        )
        res_bad = self.mod.validate_prd(missing_body)
        self.assertFalse(res_bad["ok"])
        joined = "\n".join(res_bad["findings"]).lower()
        self.assertIn("goals", joined)
        self.assertIn("r1", joined)


if __name__ == "__main__":
    unittest.main()
