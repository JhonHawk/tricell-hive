#!/usr/bin/env python3
"""Regression tests for read-only generated-tree parity checks."""

import contextlib
import hashlib
import io
import importlib.util
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
BUILD_PATH = ROOT / "harness" / "build.py"
SPEC = importlib.util.spec_from_file_location("harness_build", BUILD_PATH)
BUILD_MODULE = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(BUILD_MODULE)


def snapshot(root: Path) -> dict[str, tuple[str, int, str]]:
    entries = {}
    for path in sorted(root.rglob("*")):
        relative = path.relative_to(root).as_posix()
        if path.is_dir():
            entries[relative] = ("directory", path.stat().st_mtime_ns, "")
        else:
            digest = hashlib.sha256(path.read_bytes()).hexdigest()
            entries[relative] = ("file", path.stat().st_mtime_ns, digest)
    return entries


@contextlib.contextmanager
def core_sections_of(tree: Path):
    """Point the core assembler at a fixture tree's global/core-sections."""
    original = BUILD_MODULE.CORE_SECTIONS_DIR
    BUILD_MODULE.CORE_SECTIONS_DIR = tree / "global" / "core-sections"
    try:
        yield
    finally:
        BUILD_MODULE.CORE_SECTIONS_DIR = original


def write_rule(tree: Path, name: str, frontmatter: str = "", body: str = "rule text"):
    path = tree / "global" / "rules-situational" / name
    path.parent.mkdir(parents=True, exist_ok=True)
    head = f"---\n{frontmatter.rstrip()}\n---\n" if frontmatter else ""
    path.write_text(f"{head}\n{body}\n", encoding="utf-8")
    return path


def write_section(tree: Path, name: str, frontmatter: str, body: str = ""):
    path = tree / "global" / "core-sections" / name
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(f"---\n{frontmatter.rstrip()}\n---\n{body}", encoding="utf-8")
    return path


class CoreIncludeTests(unittest.TestCase):
    """`include:` — one rule text, carried always-on inside the Claude core.

    An always-on rule has no store of its own any more: its text lives in
    global/rules-situational/ like every other rule, and a core section names
    it. The section is frontmatter only, so the text is never forked.
    """

    def test_a_section_may_include_a_rule_text_instead_of_carrying_a_body(self):
        with tempfile.TemporaryDirectory(prefix="hive-core-include-") as tmp:
            tree = Path(tmp)
            write_rule(tree, "security-floor.md", body="## Security\n\n- Gate text.")
            write_section(
                tree, "rule-security-floor.md",
                "order: 240\ntargets: [claude]\n"
                "include: rules-situational/security-floor.md\n",
            )
            write_section(tree, "header.md", "order: 10\ntargets: [claude, agents]\n",
                          "\n## Header\n")

            with core_sections_of(tree):
                claude = BUILD_MODULE._assemble_core("claude")
                agents = BUILD_MODULE._assemble_core("agents")

            self.assertIn("## Security", claude)
            self.assertIn("- Gate text.", claude)
            # The included file's own frontmatter is stripped, like a reference.
            self.assertNotIn("include:", claude)
            # `targets: [claude]`: the condensed core never receives the full text.
            self.assertNotIn("- Gate text.", agents)

    def test_an_include_naming_a_file_that_does_not_exist_fails_the_build(self):
        with tempfile.TemporaryDirectory(prefix="hive-core-include-") as tmp:
            tree = Path(tmp)
            # A real store with a real rule in it: without one, the missing-
            # store guard fires first and this stops testing what it names.
            write_rule(tree, "gate.md")
            write_section(
                tree, "rule-ghost.md",
                "order: 200\ntargets: [claude]\ninclude: rules-situational/ghost.md\n",
            )
            with core_sections_of(tree):
                with self.assertRaises(SystemExit) as raised:
                    BUILD_MODULE._assemble_core("claude")
            message = str(raised.exception)
            self.assertIn("rule-ghost.md", message)
            self.assertIn("rules-situational/ghost.md", message)

    def test_an_included_file_may_not_also_be_hook_delivered(self):
        # Always-on and delivered-on-a-touch are exclusive by construction: a
        # file carrying both would reach a packed agent twice and a hook would
        # push what the core already holds.
        for key, value in (("globs", '  - "**/*.ts"'), ("commands", "  - git commit")):
            with self.subTest(key=key):
                with tempfile.TemporaryDirectory(prefix="hive-core-include-") as tmp:
                    tree = Path(tmp)
                    write_rule(tree, "doubled.md", frontmatter=f"{key}:\n{value}")
                    write_section(
                        tree, "rule-doubled.md",
                        "order: 200\ntargets: [claude]\n"
                        "include: rules-situational/doubled.md\n",
                    )
                    with core_sections_of(tree):
                        with self.assertRaises(SystemExit) as raised:
                            BUILD_MODULE._assemble_core("claude")
                    message = str(raised.exception)
                    self.assertIn("doubled.md", message)
                    self.assertIn(f"{key}:", message)

    def test_a_section_may_not_carry_both_a_body_and_an_include(self):
        with tempfile.TemporaryDirectory(prefix="hive-core-include-") as tmp:
            tree = Path(tmp)
            write_rule(tree, "gate.md")
            write_section(
                tree, "rule-gate.md",
                "order: 200\ntargets: [claude]\ninclude: rules-situational/gate.md\n",
                "\n## A body nobody would ever see\n",
            )
            with core_sections_of(tree):
                with self.assertRaises(SystemExit) as raised:
                    BUILD_MODULE._assemble_core("claude")
            self.assertIn("rule-gate.md", str(raised.exception))

    def test_an_include_may_not_escape_the_rule_store(self):
        # `include:` reaches the filesystem and its body is inlined into BOTH
        # generated cores — tracked files in a public repo — and from there
        # into five harnesses. `global/../CLAUDE.local.md` is the owner's
        # gitignored business context; an absolute path discards the left side
        # of the join entirely. Same class as the `packs:` escape, which
        # convert-agents.py still guards.
        escapes = {
            "parent traversal": "../CLAUDE.local.md",
            "absolute path": "/etc/hosts",
            "traversal through the store": "rules-situational/../../CLAUDE.local.md",
            "a sibling folder": "core-sections/README.md",
            "self-inclusion": "../AGENTS.md",
        }
        for label, include in escapes.items():
            with self.subTest(escape=label):
                with tempfile.TemporaryDirectory(prefix="hive-core-escape-") as tmp:
                    tree = Path(tmp)
                    (tree / "CLAUDE.local.md").write_text(
                        "PRIVATE BUSINESS CONTEXT\n", encoding="utf-8")
                    (tree / "AGENTS.md").write_text("hand-written\n", encoding="utf-8")
                    write_rule(tree, "gate.md")
                    write_section(
                        tree, "rule-gate.md",
                        f"order: 200\ntargets: [claude]\ninclude: {include}\n",
                    )
                    with core_sections_of(tree):
                        with self.assertRaises(SystemExit) as raised:
                            BUILD_MODULE._assemble_core("claude")
                    message = str(raised.exception)
                    self.assertIn("rule-gate.md", message)
                    self.assertNotIn("PRIVATE BUSINESS CONTEXT", message)

    def test_an_include_may_not_follow_a_symlink_out_of_the_store(self):
        with tempfile.TemporaryDirectory(prefix="hive-core-link-") as tmp:
            tree = Path(tmp)
            outside = tree / "CLAUDE.local.md"
            outside.write_text("PRIVATE BUSINESS CONTEXT\n", encoding="utf-8")
            write_rule(tree, "gate.md")
            (tree / "global" / "rules-situational" / "escapee.md").symlink_to(outside)
            write_section(
                tree, "rule-escapee.md",
                "order: 200\ntargets: [claude]\n"
                "include: rules-situational/escapee.md\n",
            )
            with core_sections_of(tree):
                with self.assertRaises(SystemExit) as raised:
                    BUILD_MODULE._assemble_core("claude")
            self.assertNotIn("PRIVATE BUSINESS CONTEXT", str(raised.exception))

    def test_a_symlinked_rule_store_makes_containment_vacuous_and_is_refused(self):
        # The containment check compares against the RESOLVED store, so if the
        # store directory is itself a symlink the check resolves to wherever it
        # points and then happily confirms the target is "inside" it. Every
        # escape closed one layer up reopens here.
        with tempfile.TemporaryDirectory(prefix="hive-store-link-") as tmp:
            tree = Path(tmp)
            elsewhere = tree / "private-notes"
            elsewhere.mkdir()
            (elsewhere / "zz-leak.md").write_text(
                "PRIVATE BUSINESS CONTEXT\n", encoding="utf-8")
            (tree / "global").mkdir()
            (tree / "global" / "rules-situational").symlink_to(
                elsewhere, target_is_directory=True)
            write_section(
                tree, "rule-leak.md",
                "order: 200\ntargets: [claude]\n"
                "include: rules-situational/zz-leak.md\n",
            )
            with core_sections_of(tree):
                with self.assertRaises(SystemExit) as raised:
                    BUILD_MODULE._assemble_core("claude")
            message = str(raised.exception)
            self.assertNotIn("PRIVATE BUSINESS CONTEXT", message)
            self.assertIn("rules-situational", message)

    def test_an_include_spelling_that_dodges_the_manifest_key_is_refused(self):
        # The second half of the same defect: the key was built from the
        # AUTHORED string, so a `..`-spelled include produced a key matching no
        # manifest source — `always_on` stayed False for a rule the core
        # carries in full, and the hook kept holding on it.
        with tempfile.TemporaryDirectory(prefix="hive-core-spelling-") as tmp:
            tree = Path(tmp)
            write_rule(tree, "gate.md")
            write_section(
                tree, "rule-gate.md",
                "order: 200\ntargets: [claude]\n"
                "include: rules-situational/../rules-situational/gate.md\n",
            )
            with core_sections_of(tree):
                with self.assertRaises(SystemExit) as raised:
                    BUILD_MODULE._assemble_core("claude")
            self.assertIn("rule-gate.md", str(raised.exception))

    def test_an_included_rule_is_always_on_in_the_manifest(self):
        # The normalized key has to MATCH the manifest source, or the whole
        # point of the include (never delivered twice) silently fails.
        with tempfile.TemporaryDirectory(prefix="hive-core-key-") as tmp:
            tree = Path(tmp)
            write_rule(tree, "gate.md")
            write_section(
                tree, "rule-gate.md",
                "order: 200\ntargets: [claude]\ninclude: rules-situational/gate.md\n",
            )
            entries = {e["name"]: e
                       for e in BUILD_MODULE.build_rule_manifest(tree)["rules"]}
            self.assertTrue(entries["gate"]["always_on"])

    def test_the_same_rule_text_may_not_be_included_twice(self):
        with tempfile.TemporaryDirectory(prefix="hive-core-include-") as tmp:
            tree = Path(tmp)
            write_rule(tree, "gate.md")
            for order, name in ((200, "rule-gate.md"), (210, "rule-gate-again.md")):
                write_section(
                    tree, name,
                    f"order: {order}\ntargets: [claude]\n"
                    "include: rules-situational/gate.md\n",
                )
            with core_sections_of(tree):
                with self.assertRaises(SystemExit) as raised:
                    BUILD_MODULE._assemble_core("claude")
            message = str(raised.exception)
            self.assertIn("gate.md", message)
            self.assertIn("rule-gate-again.md", message)


class RuleReachabilityTests(unittest.TestCase):
    """Every rule text reaches a model through a channel the build can name."""

    def build(self, tree: Path):
        return {entry["name"]: entry
                for entry in BUILD_MODULE.build_rule_manifest(tree)["rules"]}

    def test_always_on_means_included_by_a_core_section(self):
        with tempfile.TemporaryDirectory(prefix="hive-alwayson-") as tmp:
            tree = Path(tmp)
            # Fixture rules are NAMED after injected ones: a trigger with no
            # router reference is refused outright (see
            # test_a_hook_trigger_with_no_reference_path_fails_the_build), so a
            # made-up name would fail this test for an unrelated reason.
            write_rule(tree, "gate.md")
            write_rule(tree, "typescript-standards.md",
                       frontmatter='globs:\n  - "**/*.ts"')
            write_section(
                tree, "rule-gate.md",
                "order: 200\ntargets: [claude]\ninclude: rules-situational/gate.md\n",
            )

            rules = self.build(tree)

            self.assertTrue(rules["gate"]["always_on"])
            self.assertEqual(rules["gate"]["globs"], [])
            # The store is not the criterion any more — inclusion is.
            self.assertFalse(rules["typescript-standards"]["always_on"])

    def test_the_manifest_emits_the_commands_a_rule_declares(self):
        with tempfile.TemporaryDirectory(prefix="hive-commands-") as tmp:
            tree = Path(tmp)
            # Fixture rules are NAMED after injected ones: a trigger with no
            # router reference is refused outright (see
            # test_a_hook_trigger_with_no_reference_path_fails_the_build), so a
            # made-up name would fail this test for an unrelated reason.
            write_rule(tree, "git-mechanics.md",
                       frontmatter="commands:\n  - git commit\n  - git push")
            write_rule(tree, "browser-automation.md",
                       frontmatter="commands: agent-browser")
            write_rule(tree, "typescript-standards.md",
                       frontmatter='globs:\n  - "**/*.ts"')

            rules = self.build(tree)

            self.assertEqual(rules["git-mechanics"]["commands"],
                             ["git commit", "git push"])
            self.assertEqual(rules["browser-automation"]["commands"],
                             ["agent-browser"])
            # Absent key -> an empty list, never a missing key: the hook reads it.
            self.assertEqual(rules["typescript-standards"]["commands"], [])

    def test_command_prefixes_reach_the_manifest_verbatim(self):
        # A trailing ` +` is a MATCHING RULE the hook reads ("at least one
        # further non-option argument"): `npm install +` holds
        # `npm install lodash` and lets the lockfile install through. Strip it
        # here and the rule silently starts holding every `npm install`.
        with tempfile.TemporaryDirectory(prefix="hive-commands-") as tmp:
            tree = Path(tmp)
            # Fixture rules are NAMED after injected ones: a trigger with no
            # router reference is refused outright (see
            # test_a_hook_trigger_with_no_reference_path_fails_the_build), so a
            # made-up name would fail this test for an unrelated reason.
            write_rule(tree, "git-mechanics.md",
                       frontmatter='commands:\n  - "npm install +"\n  - "git branch +"\n'
                                   '  - "git commit"')
            write_rule(tree, "context7.md", frontmatter='commands: "npm i +"')

            rules = self.build(tree)

            self.assertEqual(rules["git-mechanics"]["commands"],
                             ["npm install +", "git branch +", "git commit"])
            self.assertEqual(rules["context7"]["commands"], ["npm i +"])

    def test_a_malformed_command_entry_fails_the_build(self):
        # The hook drops a whole rule whose `commands` is not a list of
        # non-empty strings — undeliverable, silently. Refuse to emit one.
        cases = {
            "empty key": "commands:\n",
            "empty entry": 'commands:\n  - "git commit"\n  - ""',
            "nested mapping": "commands:\n  - prefix: git commit",
            "number": "commands:\n  - 42",
            "dotted version": 'commands:\n  - "1.2.3"',
            "option first": 'commands:\n  - "--force"',
            "glued plus": 'commands:\n  - "npm install+"',
            # The marker is only meaningful in trailing position: anywhere else
            # the hook compiles it as a literal token and demands a real `+` in
            # argv, so the rule fires on nothing.
            "plus in the middle": 'commands:\n  - "npm + install"',
            # `<program> +` alone is indistinguishable from a typo for the
            # fuller prefix, and the marker exists to separate a verb's bare
            # form from its argument form.
            "plus with no verb": 'commands:\n  - "install +"',
            # `prefix_matches` compares basename(argv[0]) case-sensitively.
            "uppercase program": 'commands:\n  - "Git commit"',
        }
        for label, frontmatter in cases.items():
            with self.subTest(case=label):
                with tempfile.TemporaryDirectory(prefix="hive-commands-") as tmp:
                    tree = Path(tmp)
                    # An INJECTED name, so the inert-trigger check cannot be
                    # what fires: this test is about the command validator, and
                    # a made-up name would make every case pass vacuously.
                    write_rule(tree, "git-mechanics.md", frontmatter=frontmatter)
                    with self.assertRaises(SystemExit) as raised:
                        BUILD_MODULE.build_rule_manifest(tree)
                    message = str(raised.exception)
                    self.assertIn("git-mechanics.md", message)
                    self.assertIn("command prefix", message)

    def test_a_prefix_the_hook_can_never_match_fails_the_build(self):
        # The validator exists to catch a trigger that fires on nobody, and
        # these two families are exactly that. The hook matches argv tokens
        # AFTER unwrapping a leading sudo/env/bash -c, so a prefix that opens
        # with a wrapper is compared against a command that no longer has it;
        # and it tokenizes with shlex (punctuation_chars, posix quotes), so an
        # authored token carrying `;`, `>`, `(`, `#` or a quote can never equal
        # any argv token.
        cases = {
            "sudo wrapper": "sudo git commit",
            "env wrapper": "env pnpm add",
            "time wrapper": "time make build",
            "exec wrapper": "exec docker run",
            "shell wrapper": "bash -c",
            "trailing semicolon": "terraform apply;",
            "redirection": "docker run>log",
            "parentheses": "make (all)",
            "comment token": "gh pr create #x",
            "quoted argument": "git commit 'x'",
        }
        for label, prefix in cases.items():
            with self.subTest(case=label):
                with tempfile.TemporaryDirectory(prefix="hive-commands-") as tmp:
                    tree = Path(tmp)
                    write_rule(tree, "git-mechanics.md",
                               frontmatter=f'commands:\n  - "{prefix}"')
                    with self.assertRaises(SystemExit) as raised:
                        BUILD_MODULE.build_rule_manifest(tree)
                    message = str(raised.exception)
                    self.assertIn("git-mechanics.md", message)
                    self.assertIn("command prefix", message)

    def test_the_wrapper_list_matches_the_hook_that_strips_them(self):
        # Mirrored, so the build never imports a hook — and pinned, so the two
        # copies cannot drift into a validator that rejects what the hook
        # accepts (or worse, accepts what it strips).
        BUILD_MODULE.check_command_wrapper_parity()

        # And it can actually fail: a check nobody has seen go red is not
        # protection. Drop one name from the mirror and it must say so.
        original = BUILD_MODULE.COMMAND_IGNORED_PREFIXES
        BUILD_MODULE.COMMAND_IGNORED_PREFIXES = original - {"sudo"}
        try:
            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.check_command_wrapper_parity()
        finally:
            BUILD_MODULE.COMMAND_IGNORED_PREFIXES = original
        self.assertIn("sudo", str(raised.exception))

    def test_extra_spacing_inside_a_prefix_is_accepted(self):
        # The hook tokenizes with .split(), so `git  commit` matches exactly as
        # `git commit` does — rejecting it would be stricter than the consumer.
        with tempfile.TemporaryDirectory(prefix="hive-commands-") as tmp:
            tree = Path(tmp)
            write_rule(tree, "git-mechanics.md", frontmatter='commands:\n  - "git  commit"')
            self.assertEqual(self.build(tree)["git-mechanics"]["commands"],
                             ["git  commit"])

    def test_a_hook_trigger_with_no_reference_path_fails_the_build(self):
        # The hook gates only what it can tell the agent to READ: a rule with
        # globs or commands and no router injection has a trigger that fires on
        # nobody, and nothing downstream says so. A pack that carries the text
        # does not redeem it — then the trigger is the lie, and it goes.
        with tempfile.TemporaryDirectory(prefix="hive-inert-") as tmp:
            tree = Path(tmp)
            write_rule(tree, "inert.md", frontmatter="commands: agent-browser")
            with self.assertRaises(SystemExit) as raised:
                self.build(tree)
            message = str(raised.exception)
            self.assertIn("inert.md", message)
            self.assertIn("SKILL_REFERENCE_INJECTIONS", message)

    def test_a_pack_does_not_excuse_a_trigger_with_no_reference(self):
        # The carve-out the check deliberately does NOT have: a pack reaches
        # the five specialized agents, while a glob or a command fires for
        # everyone else — so "it is carried as a pack" leaves the trigger
        # exactly as inert as before. `test-gate` is the real instance.
        with tempfile.TemporaryDirectory(prefix="hive-inert-") as tmp:
            tree = Path(tmp)
            write_rule(tree, "packed-and-inert.md", frontmatter='globs:\n  - "**/*.ts"')
            original = BUILD_MODULE._agent_packs
            BUILD_MODULE._agent_packs = lambda root: {"an-agent": ["packed-and-inert"]}
            try:
                with self.assertRaises(SystemExit) as raised:
                    self.build(tree)
            finally:
                BUILD_MODULE._agent_packs = original
            self.assertIn("packed-and-inert.md", str(raised.exception))

    def test_a_rule_no_channel_can_reach_fails_the_build(self):
        with tempfile.TemporaryDirectory(prefix="hive-unreachable-") as tmp:
            tree = Path(tmp)
            write_rule(tree, "unreachable.md")

            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.build_rule_manifest(tree)

            message = str(raised.exception)
            self.assertIn("unreachable.md", message)
            # The fix it offers has to be a fix. Advertising `globs:` on its
            # own would send the author straight into the inert-trigger error
            # one check later.
            self.assertNotIn("`globs:` or `commands:` (rule-delivery hook),", message)
            self.assertIn("SKILL_REFERENCE_INJECTIONS", message)

    def test_a_trigger_alone_is_not_a_delivery_channel(self):
        # The sharp edge of the new semantics. A trigger says WHEN to hold; the
        # reference is WHAT the hold tells the agent to read, and the hook drops
        # a rule whose reference is missing. So `globs:`/`commands:` on their
        # own deliver nothing — they are half a channel, and the build says so
        # rather than emitting a rule that fires on nobody.
        for channel, frontmatter in (("globs", 'globs:\n  - "**/*.ts"'),
                                     ("commands", "commands: git commit")):
            with self.subTest(channel=channel):
                with tempfile.TemporaryDirectory(prefix="hive-reach-") as tmp:
                    tree = Path(tmp)
                    write_rule(tree, "unserved.md", frontmatter=frontmatter)
                    with self.assertRaises(SystemExit) as raised:
                        self.build(tree)
                    self.assertIn("unserved.md", str(raised.exception))

    def test_a_trigger_plus_a_reference_is_a_delivery_channel(self):
        # The same two triggers on a rule the injection map serves.
        for channel, (name, frontmatter) in {
            "globs": ("typescript-standards.md", 'globs:\n  - "**/*.ts"'),
            "commands": ("git-mechanics.md", "commands: git commit"),
        }.items():
            with self.subTest(channel=channel):
                with tempfile.TemporaryDirectory(prefix="hive-reach-") as tmp:
                    tree = Path(tmp)
                    write_rule(tree, name, frontmatter=frontmatter)
                    entry = self.build(tree)[Path(name).stem]
                    self.assertTrue(entry[channel])
                    self.assertIsNotNone(entry["references"])

    def test_include_injection_and_pack_each_reach_a_rule_on_their_own(self):
        # The three channels that DO stand alone: none of them needs the hook,
        # so none of them needs a reference for the hook to name.

        # A core section that includes it.
        with self.subTest(channel="include"):
            with tempfile.TemporaryDirectory(prefix="hive-reach-") as tmp:
                tree = Path(tmp)
                write_rule(tree, "reached.md")
                write_section(
                    tree, "rule-reached.md",
                    "order: 200\ntargets: [claude]\n"
                    "include: rules-situational/reached.md\n",
                )
                self.assertTrue(self.build(tree)["reached"]["always_on"])

        # A router skill that carries it as a reference.
        with self.subTest(channel="router reference"):
            with tempfile.TemporaryDirectory(prefix="hive-reach-") as tmp:
                tree = Path(tmp)
                write_rule(tree, "agent-routing.md")
                self.assertIsNotNone(self.build(tree)["agent-routing"]["references"])

        # An agent that carries it inlined as a pack.
        with self.subTest(channel="pack"):
            with tempfile.TemporaryDirectory(prefix="hive-reach-") as tmp:
                tree = Path(tmp)
                write_rule(tree, "packed-only.md")
                original = BUILD_MODULE._agent_packs
                BUILD_MODULE._agent_packs = lambda root: {"some-agent": ["packed-only"]}
                try:
                    self.assertIn("packed-only", self.build(tree))
                finally:
                    BUILD_MODULE._agent_packs = original


class RoutingTableParityTests(unittest.TestCase):
    """The routing table is stated twice: canon in the rule, condensed in the core.

    `global/rules-situational/agent-routing.md` decides WHO takes a task;
    `global/core-sections/work-style-delegation.md` restates it for the always-on
    core Codex, opencode and Pi read. Nothing tied the two together, so a row
    could disagree — and did: three NOT-to cells contradicted the canon and two
    agents had no core row at all. These tests pin the tie.
    """

    AGENTS = (
        "angular-developer", "backend-developer", "cloud-architect",
        "database-specialist", "performance-engineer", "react-developer",
        "review-code", "review-ux", "sdd-design", "sdd-explore",
        "sdd-spec-reviewer", "sdd-spec-writer", "sdd-verify", "solution-architect",
        "state-fetcher", "test-engineer", "ts-backend-developer",
        "workspace-custodian",
    )

    def fixture(self, tree: Path, canon_rows, core_rows, agents=None,
                canon_start=None, core_anchor=None):
        """Write a canon file, a core file and an agents tree; return the paths."""
        canon = tree / "agent-routing.md"
        head = BUILD_MODULE.ROUTING_CANON_START if canon_start is None else canon_start
        canon.write_text(
            f"{head}\n\n"
            "| Signal in task | Route to | NOT to |\n|---|---|---|\n"
            + "".join(f"| {s} | {r} | {n} |\n" for s, r, n in canon_rows)
            + f"\n{BUILD_MODULE.ROUTING_CANON_END}\nPost-table prose.\n",
            encoding="utf-8",
        )
        core = tree / "AGENTS.md"
        anchor = BUILD_MODULE.ROUTING_CORE_ANCHOR if core_anchor is None else core_anchor
        core.write_text(
            "- **Delegation gates are hard.** ~20 calls.\n\n"
            f"{anchor}\n|---|---|---|\n"
            + "".join(f"| {s} | {r} | {n} |\n" for s, r, n in core_rows)
            + "\n- A bullet after the table.\n",
            encoding="utf-8",
        )
        agents_dir = tree / "agents"
        role = agents_dir / "development"
        role.mkdir(parents=True, exist_ok=True)
        for name in (self.AGENTS if agents is None else agents):
            (role / f"{name}.md").write_text(f"---\nname: {name}\n---\n", encoding="utf-8")
        return canon, core, agents_dir

    def run_check(self, tree, canon_rows, core_rows, **kwargs):
        canon, core, agents_dir = self.fixture(tree, canon_rows, core_rows, **kwargs)
        with contextlib.redirect_stdout(io.StringIO()) as out:
            BUILD_MODULE.check_routing_table_parity(
                canon=canon, core=core, agents_dir=agents_dir)
        return out.getvalue()

    def expect_failure(self, tree, canon_rows, core_rows, **kwargs):
        canon, core, agents_dir = self.fixture(tree, canon_rows, core_rows, **kwargs)
        with self.assertRaises(SystemExit) as raised:
            with contextlib.redirect_stdout(io.StringIO()):
                BUILD_MODULE.check_routing_table_parity(
                    canon=canon, core=core, agents_dir=agents_dir)
        return str(raised.exception)

    def test_the_repo_tables_agree(self):
        # Against the real files, with no fixture in sight: the check has to be
        # satisfiable by the repo it guards, or it is a permanent red light.
        with contextlib.redirect_stdout(io.StringIO()) as out:
            BUILD_MODULE.check_routing_table_parity()
        self.assertIn("routing table", out.getvalue())

    def test_a_core_not_to_the_canon_does_not_support_fails_naming_the_row(self):
        # The real drift: the core's sdd-verify row said `review-code` where the
        # canon says test-engineer. Same agent routed to, a different agent
        # warned off — a contradiction no reader of one file alone could see.
        with tempfile.TemporaryDirectory(prefix="hive-routing-") as tmp:
            message = self.expect_failure(
                Path(tmp),
                canon_rows=[("functional verification of a running app",
                             "sdd-verify", "review-ux, test-engineer")],
                core_rows=[("Functional verification in a real browser",
                            "sdd-verify", "review-ux, review-code")],
            )
        self.assertIn("sdd-verify", message)
        self.assertIn("review-code", message)
        # The row, so the reader can find it without diffing two tables.
        self.assertIn("Functional verification in a real browser", message)
        # And the canon cell it was checked against.
        self.assertIn("test-engineer", message)

    def test_a_core_row_routing_to_an_agent_the_canon_never_routes_to_fails(self):
        # The extra row's NOT-to is free text on purpose: soundness cannot fire
        # on it, so only the coverage half can fail this — a row naming an agent
        # would let a one-directional coverage check pass the test anyway.
        with tempfile.TemporaryDirectory(prefix="hive-routing-") as tmp:
            message = self.expect_failure(
                Path(tmp),
                canon_rows=[("review a diff", "review-code", "the implementing agent")],
                core_rows=[("review a diff", "review-code", "the implementing agent"),
                           ("cloud topology", "cloud-architect", "the main thread")],
            )
        self.assertIn("cloud-architect", message)
        self.assertIn("only in core", message)

    def test_a_canon_row_routing_to_an_agent_the_core_omits_fails(self):
        # The other direction, and the second real drift: solution-architect and
        # the Next.js server-code row existed in the canon and in no core row,
        # so Codex/opencode/Pi routed those tasks by vibes.
        with tempfile.TemporaryDirectory(prefix="hive-routing-") as tmp:
            message = self.expect_failure(
                Path(tmp),
                canon_rows=[("review a diff", "review-code", "the implementing agent"),
                            ("architecture decision before contracts",
                             "solution-architect", "sdd-design")],
                core_rows=[("review a diff", "review-code", "the implementing agent")],
            )
        self.assertIn("solution-architect", message)
        self.assertIn("only in canon", message)

    def test_mode_suffixes_and_backticks_do_not_split_an_agent_into_two(self):
        # sdd-spec-writer (`docs`) and sdd-spec-writer (`spec`) are one agent in
        # two modes, and the core folds them into one row. Comparing the cells
        # verbatim would report drift on a file that is correct.
        with tempfile.TemporaryDirectory(prefix="hive-routing-") as tmp:
            out = self.run_check(
                Path(tmp),
                canon_rows=[("README, ADR, API docs", "sdd-spec-writer (`docs`)",
                             "the implementing agent"),
                            ("draft an épica", "sdd-spec-writer (`spec`)",
                             "the main thread, sdd-spec-reviewer"),
                            ("raw client requirements",
                             "sdd-spec-reviewer (intake mode, `requirements-rubric.md`)",
                             "the main thread")],
                core_rows=[("README, ADR, API docs; draft an épica",
                            "sdd-spec-writer (`docs` / `spec`)",
                            "the implementing agent, the main thread"),
                           ("Raw client requirements",
                            "sdd-spec-reviewer (intake mode)", "the main thread")],
            )
        self.assertIn("routing table", out)

    def test_a_not_to_agent_supported_by_a_sibling_canon_row_is_accepted(self):
        # backend-developer is routed to by two canon rows with different NOT-to
        # cells, and the core keeps both. The union is the comparison.
        with tempfile.TemporaryDirectory(prefix="hive-routing-") as tmp:
            out = self.run_check(
                Path(tmp),
                canon_rows=[("backend on a non-Node stack", "backend-developer",
                             "ts-backend-developer, database-specialist"),
                            ("server-only Kotlin", "backend-developer",
                             "kotlin-multiplatform-developer")],
                core_rows=[("Backend on any non-Node stack", "backend-developer",
                            "ts-backend-developer, kotlin-multiplatform-developer")],
                agents=self.AGENTS + ("kotlin-multiplatform-developer",),
            )
        self.assertIn("routing table", out)

    def test_the_backend_agents_alias_expands_to_both_agents(self):
        # The core writes "the backend agents" where the canon lists both by
        # name. Treating the phrase as free text would let a real contradiction
        # through; treating it as one name would flag a correct file.
        with tempfile.TemporaryDirectory(prefix="hive-routing-") as tmp:
            out = self.run_check(
                Path(tmp),
                canon_rows=[("schema design, migration", "database-specialist",
                             "ts-backend-developer, backend-developer")],
                core_rows=[("Schema design, migration", "database-specialist",
                            "the backend agents")],
            )
        self.assertIn("routing table", out)

        with tempfile.TemporaryDirectory(prefix="hive-routing-") as tmp:
            message = self.expect_failure(
                Path(tmp),
                canon_rows=[("schema design, migration", "database-specialist",
                             "ts-backend-developer")],
                core_rows=[("Schema design, migration", "database-specialist",
                            "the backend agents")],
            )
        # It expands, so the half the canon does not support is named.
        self.assertIn("backend-developer", message)

    def test_free_text_not_to_entries_are_not_treated_as_agent_names(self):
        # "the main thread", "the implementing agent", "/memory-sync" are not
        # agents; only file stems under global/agents/ are.
        with tempfile.TemporaryDirectory(prefix="hive-routing-") as tmp:
            out = self.run_check(
                Path(tmp),
                canon_rows=[("low-reasoning external state", "state-fetcher",
                             "workspace-custodian (files/ledger), /memory-sync "
                             "(decides staleness), the main thread")],
                core_rows=[("Low-reasoning external state", "state-fetcher",
                            "workspace-custodian, the main thread beyond one "
                            "quick call, the implementing agent")],
            )
        self.assertIn("routing table", out)

    def test_a_substring_of_a_longer_agent_name_is_not_a_match(self):
        # backend-developer is a suffix of ts-backend-developer: a naive
        # substring scan reads a canon that only ever warned off
        # `ts-backend-developer` as support for the core's `backend-developer`.
        with tempfile.TemporaryDirectory(prefix="hive-routing-") as tmp:
            message = self.expect_failure(
                Path(tmp),
                canon_rows=[("schema design", "database-specialist",
                             "ts-backend-developer")],
                core_rows=[("Schema design", "database-specialist",
                            "backend-developer")],
            )
        self.assertIn("backend-developer", message)

    def test_missing_anchors_fail_closed(self):
        rows = [("review a diff", "review-code", "the implementing agent")]
        with self.subTest(anchor="canon section"):
            with tempfile.TemporaryDirectory(prefix="hive-routing-") as tmp:
                message = self.expect_failure(
                    Path(tmp), rows, rows, canon_start="### Renamed Section")
            self.assertIn(BUILD_MODULE.ROUTING_CANON_START, message)
        with self.subTest(anchor="core table header"):
            with tempfile.TemporaryDirectory(prefix="hive-routing-") as tmp:
                message = self.expect_failure(
                    Path(tmp), rows, rows,
                    core_anchor="| Signal | Agent | Never |")
            self.assertIn("ROUTING_CORE_ANCHOR", message)

    def test_a_table_that_matched_zero_rows_is_not_a_pass(self):
        with tempfile.TemporaryDirectory(prefix="hive-routing-") as tmp:
            message = self.expect_failure(
                Path(tmp),
                canon_rows=[],
                core_rows=[("review a diff", "review-code", "the implementing agent")],
            )
        self.assertIn("zero rows", message)


class GeneratedTreeParityTests(unittest.TestCase):
    def test_skill_generation_excludes_python_runtime_cache(self):
        with tempfile.TemporaryDirectory(prefix="hive-skill-cache-") as tmp:
            root = Path(tmp)
            source = root / "source" / "flow-core"
            script = source / "scripts" / "plan.py"
            script.parent.mkdir(parents=True)
            script.write_text("print('validator')\n", encoding="utf-8")
            cache = script.parent / "__pycache__"
            cache.mkdir()
            (cache / "plan.cpython-314.pyc").write_bytes(b"machine-specific cache")
            (source / "SKILL.md").write_text("# Shared flow library\n", encoding="utf-8")
            output = root / "generated"
            subprocess.run(
                [sys.executable, str(ROOT / "harness/build/convert-skills.py"),
                 str(root / "source"), str(output)],
                check=True, capture_output=True, text=True,
            )
            self.assertEqual((output / "flow-core/scripts/plan.py").read_bytes(), script.read_bytes())
            self.assertFalse((output / "flow-core/scripts/__pycache__").exists())

    def create_generated_fixture(self, parent: Path) -> Path:
        fixture = parent / "actual"
        BUILD_MODULE.generate_generated_trees(fixture)
        return fixture

    def test_generated_tree_inventory_covers_every_harness_surface(self):
        self.assertEqual(
            BUILD_MODULE.GENERATED_TREE_RELATIVE_PATHS,
            (
                Path("harness/agents-skills"),
                Path("harness/codex/agents"),
                Path("harness/opencode/agents"),
                Path("harness/grok/agents"),
                Path("harness/pi/agents"),
                # The whole directory, not just agents/: a stray file dropped
                # directly under harness/claude/ is generated-tree drift too.
                Path("harness/claude"),
            ),
        )
        self.assertEqual(
            BUILD_MODULE.GENERATED_FILE_RELATIVE_PATHS,
            (Path("harness/rule-manifest.json"),),
        )
        # The agents tree deploys verbatim into ~/.claude/agents/, so it holds
        # agent definitions only — a README there would land beside them.
        self.assertFalse(
            (ROOT / "harness/claude/agents/README.md").exists(),
            "a README inside the Claude agents tree would be deployed as an agent",
        )

    def test_reports_a_stray_file_directly_under_the_claude_root(self):
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            stray = Path("harness/claude/leftover.md")
            (fixture / stray).write_text("stray\n", encoding="utf-8")

            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.check_generated_tree_parity(actual_root=fixture)

            self.assertIn(f"extra: {stray.as_posix()}", str(raised.exception))

    def test_reports_a_hand_edited_generated_claude_agent(self):
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            edited = Path("harness/claude/agents/development/backend-developer.md")
            target = fixture / edited
            target.write_bytes(target.read_bytes() + b"\n- hand edit\n")

            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.check_generated_tree_parity(actual_root=fixture)

            self.assertIn(f"stale: {edited.as_posix()}", str(raised.exception))

    def test_clean_check_is_read_only(self):
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            before = snapshot(fixture)

            BUILD_MODULE.check_generated_tree_parity(actual_root=fixture)

            self.assertEqual(snapshot(fixture), before)

    def test_reports_stale_agent_skill_reference_and_rule(self):
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            stale_paths = (
                Path("harness/pi/agents/review-code.md"),
                Path(
                    "harness/agents-skills/language-rules/references/"
                    "python-standards.md"
                ),
            )
            for relative in stale_paths:
                target = fixture / relative
                target.write_bytes(target.read_bytes() + b"\nstale\n")

            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.check_generated_tree_parity(actual_root=fixture)

            message = str(raised.exception)
            for relative in stale_paths:
                self.assertIn(f"stale: {relative.as_posix()}", message)

    def test_rule_manifest_describes_every_rule_text_with_its_delivery_paths(self):
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            manifest = json.loads(
                (fixture / "harness/rule-manifest.json").read_text(encoding="utf-8")
            )
            rules = {entry["name"]: entry for entry in manifest["rules"]}

            # A README is not a rule text; nothing downstream should deliver it.
            self.assertNotIn("README", rules)
            # One entry per rule text, from the ONE store. Derived rather than
            # hardcoded on purpose: with a single flat store the count carries
            # no information a literal would add, while the invariant it pins
            # — every text present, README excluded, nothing from elsewhere —
            # survives every rule that lands or leaves.
            store = sorted(p.stem for p in (ROOT / "global/rules-situational").glob("*.md")
                           if p.name != "README.md")
            self.assertEqual(sorted(rules), store)
            # The executor gates reach an agent only as its required pack:
            # no glob, no command, no core section — the hook never pushes them.
            core_gates = rules["agent-core-gates"]
            self.assertEqual(core_gates["globs"], [])
            self.assertEqual(core_gates["commands"], [])
            self.assertFalse(core_gates["always_on"])
            self.assertIn("agent-core-gates",
                          {p for packs in manifest["agents"].values() for p in packs})
            # The glob-scoped set the hook delivers by touched file.
            glob_scoped = sorted(n for n, e in rules.items() if e["globs"])
            # 19 before the always-on store was dissolved; `development-
            # principles`, `security` (the mechanics half of the floor) and
            # `test-gate` gained a glob scope when their gate halves moved into
            # the core, and `config-authoring` when the authoring bullets left
            # it. A literal, because a rule silently losing its scope is
            # invisible otherwise — it just stops being delivered.
            self.assertEqual(len(glob_scoped), 23, glob_scoped)

            # The config/rule authoring surface: every glob is an act a third
            # party can point to in the transcript — a write to a file of that
            # kind. Pinned whole because nothing else delivers this rule: a
            # glob dropped here silently returns the policy to "never read".
            config_authoring = rules["config-authoring"]
            self.assertEqual(
                config_authoring["globs"],
                ["**/CLAUDE.md", "**/CLAUDE.local.md", "**/AGENTS.md",
                 "**/.claude/{agents,commands,rules}/**", "**/skills/**/*.md",
                 "**/rules-situational/**", "**/core-sections/**",
                 "**/global/agents/**", "**/opencode/commands/**"],
            )
            self.assertEqual(config_authoring["commands"], [])
            self.assertFalse(config_authoring["always_on"])
            self.assertEqual(
                config_authoring["references"],
                {
                    "claude": "~/.claude/skills/workspace-conventions/references/"
                              "config-authoring.md",
                    "agents": "~/.agents/skills/workspace-conventions/references/"
                              "config-authoring.md",
                },
            )

            typescript = rules["typescript-standards"]
            self.assertEqual(
                typescript["source"],
                "global/rules-situational/typescript-standards.md",
            )
            self.assertEqual(typescript["globs"], ["**/*.{ts,tsx,mts,cts,js,jsx,mjs,cjs}"])
            self.assertFalse(typescript["always_on"])
            self.assertFalse(typescript["readers"])
            self.assertEqual(
                typescript["references"],
                {
                    "claude": "~/.claude/skills/language-rules/references/"
                              "typescript-standards.md",
                    "agents": "~/.agents/skills/language-rules/references/"
                              "typescript-standards.md",
                },
            )

            # `*.service.ts` is Angular's and NestJS's alike: each rule names
            # the pack that proves the file belongs to the other framework, so
            # the hook never holds a packed agent on its rival's rule.
            self.assertEqual(rules["angular-patterns"]["exclusive_with"],
                             ["nestjs-patterns"])
            self.assertEqual(rules["nestjs-patterns"]["exclusive_with"],
                             ["angular-patterns"])
            self.assertEqual(typescript["exclusive_with"], [])

            session_capture = rules["session-capture"]
            self.assertEqual(
                session_capture["globs"],
                ["**/_support/**", "**/*-specs/**", "**/_support/sessions/**",
                 "**/*-specs/sessions/**"],
            )
            # Reviewers need these at read time even when they carry no pack.
            self.assertTrue(session_capture["readers"])
            self.assertTrue(rules["project-structure"]["readers"])
            self.assertFalse(rules["support-artifacts"]["readers"])

            # A read-only agent receives `readers` rules and nothing else, so
            # a command trigger on a rule a reviewer can fire is delivered only
            # if the rule is flagged: `review-ux` drives agent-browser and is
            # in read_only_agents, and browser-automation has no glob that
            # would announce it any other way.
            browser = rules["browser-automation"]
            self.assertTrue(browser["readers"], "read-only agents drive agent-browser")
            self.assertEqual(browser["globs"], [])
            self.assertTrue(browser["commands"])
            self.assertIn("review-ux", manifest["read_only_agents"])
            # Flagging it is useless unless the hook can name a file to read.
            self.assertIsNotNone(browser["references"])

            # Always-on is inclusion by a core section, and the two cores are
            # the only place that decides it — nothing about the file's
            # location does.
            included = BUILD_MODULE._core_includes(ROOT / "global/core-sections")
            self.assertTrue(included, "no core section includes a rule text")
            self.assertEqual(
                sorted(name for name, entry in rules.items() if entry["always_on"]),
                sorted(Path(source).stem for source in included),
            )
            for name, entry in rules.items():
                if entry["always_on"]:
                    # A rule the core already carries is never hook-delivered.
                    self.assertEqual(entry["globs"], [], name)
                    self.assertEqual(entry["commands"], [], name)

            agent_routing = rules["agent-routing"]
            self.assertFalse(agent_routing["always_on"])
            self.assertEqual(agent_routing["globs"], [])
            self.assertEqual(
                agent_routing["references"]["agents"],
                "~/.agents/skills/task-routing/references/agent-routing.md",
            )
            # The floor is core-included, so every harness reading the core
            # already holds it — injecting it would deliver it twice. Its
            # SITUATIONAL half is the one the router names and the hook gates.
            self.assertIsNone(rules["security-floor"]["references"])
            self.assertIsNotNone(rules["security"]["references"])

            # Every hook trigger resolves to a file the hook can name. A glob
            # or a command without a reference path holds nobody — the gate
            # has nothing to ask for — and nothing downstream reports it.
            inert = sorted(name for name, entry in rules.items()
                           if (entry["globs"] or entry["commands"])
                           and not entry["references"])
            self.assertEqual(inert, [], inert)

    def test_an_exclusive_with_naming_an_unknown_rule_fails_the_build(self):
        # The key is a cross-reference between rule files; a typo would silently
        # gate nothing forever, so the build refuses it rather than emit it.
        with tempfile.TemporaryDirectory(prefix="hive-build-exclusive-") as tmp:
            source = Path(tmp)
            store = source / "global/rules-situational"
            store.mkdir(parents=True)
            (store / "nestjs-patterns.md").write_text(
                '---\nglobs:\n  - "**/*.service.ts"\n---\n', encoding="utf-8")
            (store / "angular-patterns.md").write_text(
                '---\nglobs:\n  - "**/*.service.ts"\nexclusive-with: nestjs-pattern\n'
                "---\n", encoding="utf-8")

            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.build_rule_manifest(source)

            message = str(raised.exception)
            self.assertIn("angular-patterns.md", message)
            self.assertIn("nestjs-pattern", message)

            # Spelled right, the same pair builds.
            (store / "angular-patterns.md").write_text(
                '---\nglobs:\n  - "**/*.service.ts"\nexclusive-with: nestjs-patterns\n'
                "---\n", encoding="utf-8")
            manifest = BUILD_MODULE.build_rule_manifest(source)
            entries = {entry["name"]: entry for entry in manifest["rules"]}
            self.assertEqual(entries["angular-patterns"]["exclusive_with"],
                             ["nestjs-patterns"])

    def test_a_rule_declaring_paths_fails_the_build(self):
        # `paths:` was Claude Code's native path-scoping key and the mechanism is
        # retired: rules are delivered by the hook, off `globs:`/`commands:`. A
        # file still carrying `paths:` would look scoped while being delivered
        # by nothing, so the build refuses it.
        with tempfile.TemporaryDirectory(prefix="hive-build-paths-") as tmp:
            source = Path(tmp)
            (source / "global/rules-situational").mkdir(parents=True)
            scoped = source / "global/rules-situational/typescript-standards.md"
            scoped.write_text('---\nglobs:\n  - "**/*.ts"\n---\n', encoding="utf-8")

            relative = "global/rules-situational/react-nextjs.md"
            offender = source / relative
            offender.write_text('---\npaths: "**/*.tsx"\n---\n', encoding="utf-8")
            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.build_rule_manifest(source)
            message = str(raised.exception)
            self.assertIn(relative, message)
            self.assertIn("paths:", message)
            offender.unlink()

            # Without it the same tree builds.
            manifest = BUILD_MODULE.build_rule_manifest(source)
            entries = {entry["name"]: entry for entry in manifest["rules"]}
            self.assertEqual(entries["typescript-standards"]["globs"], ["**/*.ts"])
            self.assertFalse(entries["typescript-standards"]["always_on"])

    def test_the_manifest_maps_each_agent_to_the_packs_it_carries(self):
        # The rule-delivery hook must skip what an agent already holds inlined,
        # and `packs:` is stripped from every generated output — the manifest is
        # the only place that survives the build.
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            manifest = json.loads(
                (fixture / "harness/rule-manifest.json").read_text(encoding="utf-8")
            )
            self.assertIn("agents", manifest)
            # The specialized development agents, and only they, carry packs;
            # the generic backend agent stays hook-fed.
            self.assertEqual(
                sorted(manifest["agents"]),
                [
                    "angular-developer",
                    "database-specialist",
                    "kotlin-multiplatform-developer",
                    "react-developer",
                    "ts-backend-developer",
                ],
            )
            for agent, packs in manifest["agents"].items():
                self.assertEqual(packs[0], "agent-core-gates", agent)
            # Built from the source agents, so it cannot silently go missing.
            self.assertEqual(
                BUILD_MODULE.build_rule_manifest()["agents"], manifest["agents"]
            )

    def test_the_manifest_lists_the_agents_that_cannot_write(self):
        # The rule-delivery hook gives a read-only agent only the rules a
        # READER needs. "Read-only" is not a label anyone maintains: it is the
        # converter's own write-capability verdict, so an agent that gains or
        # loses Write/Edit moves lists without anyone remembering to.
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            manifest = json.loads(
                (fixture / "harness/rule-manifest.json").read_text(encoding="utf-8")
            )
            read_only = manifest["read_only_agents"]
            self.assertEqual(read_only, sorted(read_only))
            self.assertIn("review-code", read_only)
            self.assertIn("sdd-explore", read_only)
            # sdd-verify denies Edit but keeps Write for its report: still a writer.
            self.assertNotIn("sdd-verify", read_only)
            self.assertNotIn("backend-developer", read_only)

            converter = BUILD_MODULE._load_agent_converter()
            expected = sorted(
                converter.parse_agent(path)["name"]
                for path in (ROOT / "global/agents").rglob("*.md")
                if not converter.can_write(converter.parse_agent(path))
            )
            self.assertEqual(read_only, expected)

    def test_inline_globs_keep_their_brace_groups_intact(self):
        with tempfile.TemporaryDirectory(prefix="hive-globs-") as tmp:
            source = Path(tmp)
            rules = source / "global/rules-situational"
            rules.mkdir(parents=True)
            # Named after injected rules: a trigger with no router reference
            # is refused, so a made-up name would fail for an unrelated reason.
            (rules / "typescript-standards.md").write_text(
                '---\nglobs:\n  - "**/*.{ts,tsx}"\n  - "**/*.vue"\n---\n\ntext\n',
                encoding="utf-8",
            )
            (rules / "react-nextjs.md").write_text(
                '---\nglobs: "**/*.{ts,tsx,mts}"\n---\n\ntext\n', encoding="utf-8"
            )
            rules_by_name = {
                entry["name"]: entry
                for entry in BUILD_MODULE.build_rule_manifest(source)["rules"]
            }
            self.assertEqual(
                rules_by_name["typescript-standards"]["globs"],
                ["**/*.{ts,tsx}", "**/*.vue"],
            )
            # A comma inside braces is part of ONE glob, not a separator.
            self.assertEqual(rules_by_name["react-nextjs"]["globs"],
                             ["**/*.{ts,tsx,mts}"])

    def test_reports_a_stale_rule_manifest(self):
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            target = fixture / "harness/rule-manifest.json"
            target.write_text(
                target.read_text(encoding="utf-8").replace('"readers"', '"reader"'),
                encoding="utf-8",
            )

            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.check_generated_tree_parity(actual_root=fixture)

            self.assertIn("stale: harness/rule-manifest.json", str(raised.exception))

    def test_reports_missing_and_extra_entries_together(self):
        with tempfile.TemporaryDirectory(prefix="hive-build-parity-") as tmp:
            fixture = self.create_generated_fixture(Path(tmp))
            missing = Path("harness/codex/agents/review-code.toml")
            extra = Path("harness/grok/agents/orphan.md")
            (fixture / missing).unlink()
            (fixture / extra).write_text("orphan\n", encoding="utf-8")

            with self.assertRaises(SystemExit) as raised:
                BUILD_MODULE.check_generated_tree_parity(actual_root=fixture)

            message = str(raised.exception)
            self.assertIn(f"missing: {missing.as_posix()}", message)
            self.assertIn(f"extra: {extra.as_posix()}", message)


if __name__ == "__main__":
    unittest.main()
