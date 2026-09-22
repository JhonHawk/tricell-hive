from pathlib import Path
import subprocess
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
HELPER = ROOT / "content/skills/workspace-archive/scripts/archive_sessions.py"


class WorkspaceArchiveHelperTest(unittest.TestCase):
    def run_helper(self, *args):
        return subprocess.run(["python3", str(HELPER), *args], text=True, capture_output=True)

    def test_dry_run_and_apply_preserve_selected_closed_research_and_links(self):
        with tempfile.TemporaryDirectory() as tmp:
            support = Path(tmp) / "_support"
            closed = support / "sessions/2026-09-20-closed"
            research = support / "sessions/2026-09-19-research"
            open_session = support / "sessions/2026-09-18-open"
            closed.mkdir(parents=True); research.mkdir(parents=True); open_session.mkdir(parents=True)
            payload = "# Plan\n\nStatus: completed\n\nexact bytes remain\n"
            closed.joinpath("closed-plan.md").write_text(payload)
            research.joinpath("research.research.md").write_text("# Research\n\nStatus: closed\n")
            open_session.joinpath("open.plan.md").write_text("# Plan\n\nStatus: planned\n")
            index = support / "sessions/README.md"
            index.write_text("[closed](2026-09-20-closed/closed-plan.md)\n")
            dry = self.run_helper("--support-root", str(support), "--session", "2026-09-20-closed")
            self.assertEqual(dry.returncode, 0)
            self.assertTrue(closed.exists())
            applied = self.run_helper("--support-root", str(support), "--session", "2026-09-20-closed", "--session", "2026-09-19-research", "--apply")
            self.assertEqual(applied.returncode, 0, applied.stderr)
            moved = support / "sessions/archived/2026-09-20-closed/closed-plan.md"
            self.assertEqual(moved.read_text(), payload)
            self.assertTrue((support / "sessions/archived/2026-09-19-research").exists())
            self.assertFalse(closed.exists())
            self.assertTrue(open_session.exists())
            self.assertIn("archived/2026-09-20-closed/", index.read_text())

    def test_open_or_colliding_session_is_left_in_place(self):
        with tempfile.TemporaryDirectory() as tmp:
            support = Path(tmp) / "_support"
            session = support / "sessions/2026-09-20-open"
            session.mkdir(parents=True)
            session.joinpath("open-plan.md").write_text("Status: planned\n")
            result = self.run_helper("--support-root", str(support), "--session", session.name, "--apply")
            self.assertEqual(result.returncode, 2)
            self.assertTrue(session.exists())
            closed = support / "sessions/2026-09-20-closed"
            closed.mkdir(); closed.joinpath("closed.plan.md").write_text("Status: completed\n")
            collision = support / "sessions/archived/2026-09-20-closed"
            collision.mkdir(parents=True)
            result = self.run_helper("--support-root", str(support), "--session", closed.name, "--apply")
            self.assertEqual(result.returncode, 2)
            self.assertTrue(closed.exists())

    def test_duplicate_or_symlinked_selection_cannot_partially_move(self):
        with tempfile.TemporaryDirectory() as tmp:
            support = Path(tmp) / "_support"; source = support / "sessions/2026-09-20-closed"
            source.mkdir(parents=True); source.joinpath("closed.plan.md").write_text("Status: completed\n")
            duplicate = self.run_helper("--support-root", str(support), "--session", source.name, "--session", source.name, "--apply")
            self.assertEqual(duplicate.returncode, 2)
            self.assertTrue(source.exists())
            linked = support / "sessions/2026-09-21-linked"
            linked.symlink_to(source, target_is_directory=True)
            result = self.run_helper("--support-root", str(support), "--session", linked.name, "--apply")
            self.assertEqual(result.returncode, 2)
            self.assertTrue(source.exists())

    def test_conflicting_plans_dot_name_and_outbound_link_are_safe(self):
        with tempfile.TemporaryDirectory() as tmp:
            support = Path(tmp) / "_support"; source = support / "sessions/2026-09-20-closed"
            docs = support / "docs"; source.mkdir(parents=True); docs.mkdir()
            source.joinpath("closed.plan.md").write_text("Status: completed\n[docs](../../docs/guide.md)\n")
            source.joinpath("open-plan.md").write_text("Status: planned\n")
            docs.joinpath("guide.md").write_text("guide\n")
            dot = self.run_helper("--support-root", str(support), "--session", "..", "--apply")
            self.assertEqual(dot.returncode, 2)
            conflicting = self.run_helper("--support-root", str(support), "--session", source.name, "--apply")
            self.assertEqual(conflicting.returncode, 2)
            source.joinpath("open-plan.md").write_text("Status: completed\n")
            applied = self.run_helper("--support-root", str(support), "--session", source.name, "--apply")
            self.assertEqual(applied.returncode, 0, applied.stderr)
            moved = support / "sessions/archived/2026-09-20-closed/closed.plan.md"
            self.assertIn("../../../docs/guide.md", moved.read_text())

    def test_symlinked_session_tree_or_broken_destination_blocks_preflight(self):
        with tempfile.TemporaryDirectory() as tmp:
            support = Path(tmp) / "_support"; source = support / "sessions/2026-09-20-closed"
            source.mkdir(parents=True); source.joinpath("closed.plan.md").write_text("Status: completed\n")
            destination = support / "sessions/archived/2026-09-20-closed"
            destination.parent.mkdir(); destination.symlink_to(Path(tmp) / "missing")
            blocked = self.run_helper("--support-root", str(support), "--session", source.name, "--apply")
            self.assertEqual(blocked.returncode, 2)
            self.assertTrue(source.exists())

    def test_explicit_spanish_closed_and_open_plan_states(self):
        with tempfile.TemporaryDirectory() as tmp:
            support = Path(tmp) / "_support"
            closed = support / "sessions/2026-09-20-cerrado"; open_session = support / "sessions/2026-09-20-abierto"
            closed.mkdir(parents=True); open_session.mkdir()
            closed.joinpath("cerrado.plan.md").write_text("# Plan\n\nEstado: cerrado\n")
            open_session.joinpath("abierto.plan.md").write_text("# Plan\n\nEstado: abierto\n")
            applied = self.run_helper("--support-root", str(support), "--session", closed.name, "--apply")
            self.assertEqual(applied.returncode, 0, applied.stderr)
            blocked = self.run_helper("--support-root", str(support), "--session", open_session.name, "--apply")
            self.assertEqual(blocked.returncode, 2)
            self.assertTrue(open_session.exists())

    def test_flow_plan_control_sheet_rows(self):
        with tempfile.TemporaryDirectory() as tmp:
            support = Path(tmp) / "_support"
            closed = support / "sessions/2026-09-20-sheet"; open_session = support / "sessions/2026-09-20-sheet-open"
            closed.mkdir(parents=True); open_session.mkdir()
            closed.joinpath("sheet.plan.md").write_text("| Campo | Valor |\n| --- | --- |\n| Estado | Completado · merge verificado |\n")
            open_session.joinpath("sheet.plan.md").write_text("| Field | Value |\n| --- | --- |\n| Status | In progress · implementation authorized |\n")
            applied = self.run_helper("--support-root", str(support), "--session", closed.name, "--apply")
            self.assertEqual(applied.returncode, 0, applied.stderr)
            blocked = self.run_helper("--support-root", str(support), "--session", open_session.name, "--apply")
            self.assertEqual(blocked.returncode, 2)
            self.assertTrue(open_session.exists())


if __name__ == "__main__":
    unittest.main()
