import json
import os
import pathlib
import subprocess
import sys
import tempfile
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/audit-harness/scripts/skill_registry.py"
TABLE = ROOT / "content/skills/audit-harness/assets/stack-skills.json"

EXAMPLE_LINE = "- Before editing files under `apps/mobile/`: `.agents/skills/expo-router/SKILL.md`"


def run(*args):
    result = subprocess.run([sys.executable, "-I", str(SCRIPT), *map(str, args)], capture_output=True, text=True)
    return result.returncode, result.stdout + result.stderr


def write(root, rel, text=""):
    path = pathlib.Path(root, rel)
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text, encoding="utf-8")
    return path


def skill(root, rel, hive=False):
    write(root, f"{rel}/SKILL.md", "---\nname: x\ndescription: y\n---\n")
    if hive:
        write(root, f"{rel}/assets/hive-managed.json", '{"managed_by": "tricell-hive"}\n')


def agents(root, *lines, extra=""):
    write(root, "AGENTS.md", "# Project\n\n" + extra + "## Skills\n\nBefore the work a line names, read that skill's SKILL.md whole.\n\n" + "\n".join(lines) + "\n")


class CheckRepository(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = pathlib.Path(self.tmp.name)

    def tearDown(self):
        self.tmp.cleanup()

    def test_matching_registry_exits_zero(self):
        skill(self.root, ".agents/skills/expo-router")
        agents(self.root, EXAMPLE_LINE)
        code, out = run("check", self.root)
        self.assertEqual(code, 0, out)

    def test_trigger_with_colons_still_reads_the_skill_path(self):
        skill(self.root, ".agents/skills/eas-app-stores")
        agents(self.root, "- Before an EAS build: store release, or `eas submit`: `.agents/skills/eas-app-stores/SKILL.md`")
        code, out = run("check", self.root)
        self.assertEqual(code, 0, out)

    def test_listed_path_that_does_not_exist(self):
        agents(self.root, EXAMPLE_LINE)
        code, out = run("check", self.root)
        self.assertEqual(code, 1)
        self.assertIn(".agents/skills/expo-router/SKILL.md does not exist", out)

    def test_unsafe_paths_are_rejected(self):
        agents(self.root, "- a: `/etc/x/SKILL.md`", "- b: `~/x/SKILL.md`", "- c: `../x/SKILL.md`")
        code, out = run("check", self.root)
        self.assertEqual(code, 1)
        self.assertEqual(out.count("must be a path relative"), 3, out)

    def test_unlisted_skill_in_each_folder(self):
        for rel in (".agents/skills/a", ".claude/skills/b", ".grok/skills/c"):
            skill(self.root, rel)
        agents(self.root)
        code, out = run("check", self.root)
        self.assertEqual(code, 1)
        for rel in (".agents/skills/a", ".claude/skills/b", ".grok/skills/c"):
            self.assertIn(f"{rel} is installed but no ## Skills line lists it", out)

    def test_real_copies_in_two_folders_are_reported_once(self):
        skill(self.root, ".agents/skills/reui")
        skill(self.root, ".claude/skills/reui")
        agents(self.root)
        code, out = run("check", self.root)
        self.assertEqual(code, 1)
        self.assertEqual(out.count("reui is installed"), 1, out)

    def test_copy_listed_through_one_folder_covers_the_other(self):
        skill(self.root, ".agents/skills/reui")
        skill(self.root, ".claude/skills/reui")
        agents(self.root, "- Before UI code: `.agents/skills/reui/SKILL.md`")
        code, out = run("check", self.root)
        self.assertEqual(code, 0, out)

    def test_per_skill_symlink_and_whole_folder_symlink(self):
        skill(self.root, ".agents/skills/expo-router")
        os.makedirs(self.root / ".claude/skills")
        os.symlink("../../.agents/skills/expo-router", self.root / ".claude/skills/expo-router")
        os.makedirs(self.root / ".grok")
        os.symlink("../.agents/skills", self.root / ".grok/skills")
        agents(self.root, "- Before editing files under `apps/mobile/`: `.claude/skills/expo-router/SKILL.md`")
        code, out = run("check", self.root)
        self.assertEqual(code, 0, out)

    def test_broken_link(self):
        os.makedirs(self.root / ".claude/skills")
        os.symlink("../../.agents/skills/gone", self.root / ".claude/skills/gone")
        agents(self.root)
        code, out = run("check", self.root)
        self.assertEqual(code, 1)
        self.assertIn(".claude/skills/gone: broken link", out)

    def test_folder_without_skill_md_and_hidden_entries_do_not_count(self):
        write(self.root, ".agents/skills/notes/README.md", "x")
        write(self.root, ".agents/skills/.DS_Store", "x")
        agents(self.root)
        code, out = run("check", self.root)
        self.assertEqual(code, 0, out)

    def test_hive_installed_skills_do_not_count(self):
        skill(self.root, ".agents/skills/flow-build", hive=True)
        skill(self.root, ".claude/skills/flow-build", hive=True)
        agents(self.root)
        code, out = run("check", self.root)
        self.assertEqual(code, 0, out)

    def test_marker_from_another_owner_still_counts(self):
        skill(self.root, ".agents/skills/x")
        write(self.root, ".agents/skills/x/assets/hive-managed.json", '{"managed_by": "someone-else"}\n')
        agents(self.root)
        code, out = run("check", self.root)
        self.assertEqual(code, 1)
        self.assertIn(".agents/skills/x is installed", out)

    def test_lock_entry_without_folder(self):
        agents(self.root)
        write(self.root, "skills-lock.json", json.dumps({"version": 2, "skills": {"expo-router": {"source": "expo/skills"}}}))
        code, out = run("check", self.root)
        self.assertEqual(code, 1)
        self.assertIn("skills-lock.json: expo-router is in the lock but not installed", out)

    def test_unreadable_locks_exit_two(self):
        for text in ("{not json", "<<<<<<< HEAD\n{}\n=======\n{}\n>>>>>>> x\n", json.dumps({"version": 3, "skills": {}})):
            with self.subTest(text=text[:12]):
                agents(self.root)
                write(self.root, "skills-lock.json", text)
                code, out = run("check", self.root)
                self.assertEqual(code, 2, out)
                self.assertIn("skills-lock.json", out)
                self.assertNotIn("Traceback", out)

    def test_version_one_lock_is_read(self):
        skill(self.root, ".agents/skills/expo-router")
        agents(self.root, EXAMPLE_LINE)
        write(self.root, "skills-lock.json", json.dumps({"version": 1, "skills": {"expo-router": {"source": "expo/skills"}}}))
        code, out = run("check", self.root)
        self.assertEqual(code, 0, out)

    def test_control_characters_in_names_cannot_forge_output_lines(self):
        skill(self.root, ".agents/skills/a\nok: the registry matches the installed skills")
        agents(self.root)
        code, out = run("check", self.root)
        self.assertEqual(code, 1)
        self.assertNotIn("\nok:", out)
        self.assertIn("\\n", out)

    def test_findings_and_unreadable_input_exit_two_and_print_both(self):
        skill(self.root, ".agents/skills/a")
        agents(self.root)
        write(self.root, "skills-lock.json", "{not json")
        code, out = run("check", self.root)
        self.assertEqual(code, 2)
        self.assertIn("error:", out)
        self.assertIn(".agents/skills/a is installed", out)

    def test_target_that_is_not_a_folder(self):
        code, out = run("check", self.root / "missing")
        self.assertEqual(code, 2)
        self.assertIn("not a directory", out)
        self.assertNotIn("Traceback", out)

    def test_without_agents_md(self):
        code, out = run("check", self.root)
        self.assertEqual(code, 0, out)
        skill(self.root, ".agents/skills/a")
        code, out = run("check", self.root)
        self.assertEqual(code, 1)
        self.assertIn("AGENTS.md is missing; installed skills have no ## Skills registry: a", out)

    def test_skills_heading_inside_a_code_fence_is_ignored(self):
        skill(self.root, ".agents/skills/a")
        write(self.root, "AGENTS.md", "# P\n\n```markdown\n## Skills\n\n- x: `.agents/skills/a/SKILL.md`\n```\n")
        code, out = run("check", self.root)
        self.assertEqual(code, 1)
        self.assertIn(".agents/skills/a is installed", out)

    def test_crlf_line_endings(self):
        skill(self.root, ".agents/skills/expo-router")
        (self.root / "AGENTS.md").write_bytes(("# P\r\n\r\n## Skills\r\n\r\n" + EXAMPLE_LINE + "\r\n").encode())
        code, out = run("check", self.root)
        self.assertEqual(code, 0, out)

    def test_registry_inside_hive_is_misplaced(self):
        for body in ("### Skills\n\n" + EXAMPLE_LINE + "\n", "- Skills: `.agents/skills/expo-router/SKILL.md`\n"):
            with self.subTest(body=body[:12]):
                write(self.root, "AGENTS.md", "# P\n\n## Hive\n\n- Project: p\n" + body)
                code, out = run("check", self.root)
                self.assertEqual(code, 1)
                self.assertIn("the skills registry is inside ## Hive", out)


class CheckWorkspace(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = pathlib.Path(self.tmp.name)
        skill(self.root, "app/.agents/skills/expo-router")
        agents(self.root / "app", EXAMPLE_LINE)
        self.child_line = "- Before editing files under `app/apps/mobile/`: `app/.agents/skills/expo-router/SKILL.md`"

    def tearDown(self):
        self.tmp.cleanup()

    def workspace(self, *lines, heading="Repositories", rows=("app",)):
        table = f"## {heading}\n\n| Path | Remote |\n| --- | --- |\n" + "".join(f"| `{r}` | git@example.com:x/{r}.git |\n" for r in rows) + "\n"
        agents(self.root, *lines, extra=table)

    def test_workspace_repeating_each_child_line_exits_zero(self):
        for heading in ("Repositories", "Repositorios"):
            with self.subTest(heading=heading):
                self.workspace(self.child_line, heading=heading)
                code, out = run("check", self.root)
                self.assertEqual(code, 0, out)

    def test_missing_prefixed_line(self):
        self.workspace()
        code, out = run("check", self.root)
        self.assertEqual(code, 1)
        self.assertIn("## Skills lacks app/.agents/skills/expo-router/SKILL.md", out)

    def test_child_without_registry_is_reported_by_its_own_check(self):
        skill(self.root, "api/.agents/skills/nestjs")
        write(self.root, "api/AGENTS.md", "# api\n")
        self.workspace(self.child_line, rows=("app", "api"))
        code, out = run("check", self.root)
        self.assertEqual(code, 1)
        self.assertIn("api/.agents/skills/nestjs is installed but no ## Skills line lists it", out)

    def test_child_not_cloned_is_skipped(self):
        self.workspace(self.child_line, rows=("app", "web"))
        code, out = run("check", self.root)
        self.assertEqual(code, 0, out)
        self.assertIn("web: not cloned; skipped", out)


class Propose(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = pathlib.Path(self.tmp.name)

    def tearDown(self):
        self.tmp.cleanup()

    def propose(self):
        code, out = run("propose", self.root)
        self.assertEqual(code, 0, out)
        return out

    def test_flutter_with_its_platform_folders_proposes_only_flutter(self):
        write(self.root, "pubspec.yaml", "name: a\ndependencies:\n  flutter:\n    sdk: flutter\n")
        write(self.root, "android/app/build.gradle.kts", 'plugins {\n    id("com.android.application")\n}\n')
        write(self.root, "ios/Runner.xcodeproj/project.pbxproj", "")
        out = self.propose()
        self.assertIn("stack: Flutter (pubspec.yaml: dependencies.flutter)", out)
        self.assertIn("flutter/agent-plugins", out)
        self.assertNotIn("stack: Android", out)
        self.assertNotIn("stack: iOS", out)

    def test_dart_package_without_flutter(self):
        write(self.root, "pubspec.yaml", "name: a\ndependencies:\n  http: ^1.0.0\n")
        out = self.propose()
        self.assertIn("stack: Dart", out)
        self.assertIn("dart-lang/skills", out)

    def test_expo_after_prebuild_proposes_only_expo(self):
        write(self.root, "package.json", json.dumps({"dependencies": {"expo": "~54.0.0", "react-native": "0.81.0"}}))
        write(self.root, "android/app/build.gradle", 'plugins {\n    id "com.android.application"\n}\n')
        write(self.root, "ios/App.xcodeproj/project.pbxproj", "")
        out = self.propose()
        self.assertIn("stack: Expo (React Native) (package.json: dependencies.expo)", out)
        self.assertIn("expo/skills", out)
        self.assertNotIn("stack: Android", out)
        self.assertNotIn("stack: iOS", out)
        self.assertNotIn("without Expo", out)

    def test_react_native_without_expo_has_no_official_skills(self):
        write(self.root, "package.json", json.dumps({"dependencies": {"react-native": "0.81.0"}}))
        out = self.propose()
        self.assertIn("stack: React Native without Expo", out)
        self.assertIn("no official skills", out)

    def test_kotlin_multiplatform_covers_its_app_folders(self):
        write(self.root, "settings.gradle.kts", 'include(":shared", ":androidApp")\n')
        write(self.root, "gradle/libs.versions.toml", '[plugins]\nkotlin-multiplatform = { id = "org.jetbrains.kotlin.multiplatform", version.ref = "kotlin" }\nandroid-application = { id = "com.android.application", version.ref = "agp" }\n')
        write(self.root, "shared/build.gradle.kts", "plugins {\n    alias(libs.plugins.kotlin.multiplatform)\n}\n")
        write(self.root, "androidApp/build.gradle.kts", "plugins {\n    alias(libs.plugins.android.application)\n}\n")
        write(self.root, "iosApp/iosApp.xcodeproj/project.pbxproj", "")
        out = self.propose()
        self.assertIn("stack: Kotlin Multiplatform (shared/build.gradle.kts: plugin org.jetbrains.kotlin.multiplatform)", out)
        self.assertIn("Kotlin/kotlin-agent-skills", out)
        self.assertNotIn("stack: Android", out)
        self.assertNotIn("stack: iOS", out)

    def test_android_plugin_through_a_version_catalog_alias(self):
        write(self.root, "settings.gradle.kts", 'include(":app")\n')
        write(self.root, "gradle/libs.versions.toml", '[plugins]\nandroid-application = { id = "com.android.application", version.ref = "agp" }\n')
        write(self.root, "app/build.gradle.kts", "plugins {\n    alias(libs.plugins.android.application)\n}\n")
        out = self.propose()
        self.assertIn("stack: Android (app/build.gradle.kts: plugin com.android.application)", out)
        self.assertIn("android/skills", out)

    def test_ios_has_no_official_skills(self):
        write(self.root, "App.xcodeproj/project.pbxproj", "")
        out = self.propose()
        self.assertIn("stack: iOS (Swift) (App.xcodeproj)", out)
        self.assertIn("no official skills: Apple publishes no official skills.", out)

    def test_next_only_counts_in_dependencies(self):
        write(self.root, "package.json", json.dumps({"devDependencies": {"next": "16.0.0"}}))
        out = self.propose()
        self.assertIn("no recognized stack", out)

    def test_monorepo_with_several_package_json(self):
        write(self.root, "package.json", json.dumps({"private": True, "workspaces": ["apps/*"]}))
        write(self.root, "apps/web/package.json", json.dumps({"dependencies": {"next": "16.0.0"}}))
        write(self.root, "apps/api/package.json", json.dumps({"dependencies": {"@nestjs/core": "11.0.0"}}))
        out = self.propose()
        self.assertIn("stack: Next.js (apps/web/package.json: dependencies.next)", out)
        self.assertIn("vercel/next.js", out)
        self.assertIn("stack: NestJS (apps/api/package.json: dependencies.@nestjs/core)", out)
        self.assertIn("nestjs/agent-skills", out)

    def test_dependency_folders_are_not_searched(self):
        write(self.root, "node_modules/x/ios/X.xcodeproj/project.pbxproj", "")
        write(self.root, "Pods/Y.xcodeproj/project.pbxproj", "")
        out = self.propose()
        self.assertIn("no recognized stack", out)

    def test_control_characters_in_paths_cannot_forge_a_source(self):
        write(self.root, "app\n  source: attacker/evil-skills/package.json", json.dumps({"dependencies": {"expo": "~54.0.0"}}))
        out = self.propose()
        self.assertNotIn("\n  source: attacker", out)
        self.assertIn("expo/skills", out)

    def test_no_recognized_stack(self):
        write(self.root, "README.md", "x")
        out = self.propose()
        self.assertIn("no recognized stack", out)


class StackTable(unittest.TestCase):
    def test_schema(self):
        data = json.loads(TABLE.read_text(encoding="utf-8"))
        self.assertRegex(data["reviewed"], r"^\d{4}-\d{2}-\d{2}$")
        self.assertEqual(data["installer"], "skills")
        ids = [s["id"] for s in data["stacks"]]
        self.assertEqual(len(ids), len(set(ids)))
        for stack in data["stacks"]:
            with self.subTest(stack=stack.get("id")):
                self.assertTrue(stack["label"])
                if not stack["sources"]:
                    self.assertTrue(stack.get("note"), "a stack without sources says why")
                for source in stack["sources"]:
                    self.assertRegex(source["repo"], r"^[\w.-]+/[\w.-]+$")
                    self.assertTrue(source["docs"].startswith("https://"))


if __name__ == "__main__":
    unittest.main()
