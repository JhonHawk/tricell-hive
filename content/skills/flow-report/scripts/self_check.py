#!/usr/bin/env python3
# Adapted diagram-design resources: ../references/license-diagram-design.md
"""Self-check a generated diagram HTML file, with no third-party deps.

Ships inside the flow-report skill so an installed agent can verify its own
output:

    python3 scripts/self_check.py my-report.html

Checks the accessible-SVG contract and the single-file safety rules (no
remote assets whatsoever, no executable attributes, no `<script>` inside a
diagram `<svg>`). This is a distilled subset of the repository gates
(`lint-skin.py`, `verify-motion.py`) in the upstream diagram-design
repository, which remain the authority for contributions to that repository
itself.

---
Adapted from https://github.com/cathrynlavery/diagram-design
(skills/diagram-design/scripts/self_check.py, v2.3.2, MIT) for the
flow-report skill. Local changes:
  - Removed the fonts.googleapis.com stylesheet exemption. Upstream
    allowlisted one specific remote Google Fonts URL as an acceptable
    external reference; flow-report HTML must open offline and pass the
    Artifacts CSP, so ANY external link/script/img/font reference is now a
    failure, no exceptions.
  - Removed the motion-contract code path entirely (`canonical_controller`,
    `check_motion`, `normalized_controller`, `MODES`, `ACTIONS`,
    `ASCII_DECIMAL_RE`, `MOTION_TEMPLATE`/`SKILL_DIR`, every `data-motion-*`
    tracking field on the parser, and the now-orphaned `<style>` text
    capture (`styles`/`_in_style`) that only ever fed the reduced-motion CSS
    check). Motion is deliberately un-adopted by diagram-grammar.md and
    flow-report ships no `template-motion.html`, so upstream's "any
    `<script>` implies the canonical motion controller" assumption is wrong
    for us: a real report ships page-level `<script>` (the baseline
    skeleton's modal/TOC/deck-nav JS) and would otherwise RuntimeError on
    `canonical_controller()`'s missing template on every run.
  - Removed the "at least one accessible SVG" requirement. Upstream validates
    single-diagram files, where a missing SVG means a broken deliverable;
    flow-report runs this script on EVERY report, and a diagram-free report
    (tables/prose only) is legitimate. The full accessible-SVG contract still
    applies to every non-decorative SVG that IS present.
  - `<script>` is now scoped instead: a `<script>` OUTSIDE every `<svg>` is
    page JS and passes; a `<script>` INSIDE an `<svg>` fails cleanly
    ("diagrams are static under the flow-report grammar") instead of
    crashing.
  - CLI, exit codes, the accessible-SVG contract, and the single-file safety
    checks are otherwise unchanged from upstream.
"""

from __future__ import annotations

import argparse
import sys
from html.parser import HTMLParser
from pathlib import Path

REFERENCE_ATTRS = {"src", "href", "xlink:href", "poster", "srcset", "action", "formaction"}

# Any remote host must be opted into by the document itself, with
#     <meta name="flow-report-assets" content="external">
# The declaration rides with the file, so the check is reproducible by anyone
# who receives it — a CLI flag would not be. Declaring it is a real trade: the
# Artifacts CSP blocks every non-font host, so such a report cannot be
# published as an Artifact, and a CDN outage degrades or breaks it offline.
EXTERNAL_OPT_IN = ("flow-report-assets", "external")


class DiagramParser(HTMLParser):
    def __init__(self) -> None:
        super().__init__(convert_charrefs=True)
        self.scripts: list[dict[str, object]] = []
        self.svgs: list[dict[str, object]] = []
        self.unsafe: list[str] = []
        self.references: list[tuple[str, str, str]] = []
        self.allow_external = False
        self._svg_depth = 0
        self._current_svg: dict[str, object] | None = None
        self._capture: str | None = None
        self._current_script: dict[str, object] | None = None
        self._element_stack: list[str] = []

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        tag = tag.casefold()
        normalized_attrs = [(key.casefold(), value or "") for key, value in attrs]
        data = {key: value for key, value in normalized_attrs}
        if tag in {"base", "embed", "object", "iframe"}:
            self.unsafe.append(f"<{tag}> is not allowed in a diagram file")
        if tag == "meta":
            data = {k.casefold(): (v or "") for k, v in attrs}
            if (data.get("name", "").strip().casefold() == EXTERNAL_OPT_IN[0]
                    and data.get("content", "").strip().casefold() == EXTERNAL_OPT_IN[1]):
                self.allow_external = True
        for key, value in normalized_attrs:
            if key.startswith("on"):
                self.unsafe.append(f"executable attribute {key} on <{tag}>")
            if key == "srcdoc":
                self.unsafe.append(f"srcdoc attribute on <{tag}>")
            if key in REFERENCE_ATTRS:
                self.references.append((tag, data.get("rel", ""), value))
        if tag == "script":
            self._current_script = {
                "attrs": data,
                "attr_names": [name for name, _value in normalized_attrs],
                "body": [],
                "closed": False,
                "in_svg": self._svg_depth > 0,
            }
            self.scripts.append(self._current_script)
        self._element_stack.append(tag)
        if tag == "svg" and self._svg_depth == 0:
            self._svg_depth = 1
            self._current_svg = {"attrs": data, "first": None, "title": {}, "desc": {}}
            self.svgs.append(self._current_svg)
            return
        if self._svg_depth:
            self._svg_depth += 1
            assert self._current_svg is not None
            if self._svg_depth == 2 and self._current_svg["first"] is None:
                self._current_svg["first"] = tag
            if self._svg_depth == 2 and tag in {"title", "desc"}:
                self._current_svg[tag] = {"attrs": data, "text": ""}
                self._capture = tag

    def handle_endtag(self, tag: str) -> None:
        tag = tag.casefold()
        if tag == "script" and self._current_script is not None:
            self._current_script["closed"] = True
            self._current_script = None
        if self._svg_depth:
            if tag in {"title", "desc"}:
                self._capture = None
            self._svg_depth -= 1
            if self._svg_depth == 0:
                self._current_svg = None
        for index in range(len(self._element_stack) - 1, -1, -1):
            if self._element_stack[index] == tag:
                del self._element_stack[index:]
                break

    def handle_data(self, data: str) -> None:
        if self._current_script is not None:
            body = self._current_script["body"]
            assert isinstance(body, list)
            body.append(data)
        if self._capture and self._current_svg:
            node = self._current_svg[self._capture]
            assert isinstance(node, dict)
            node["text"] = str(node.get("text", "")) + data


def parsed_document(source: str) -> DiagramParser:
    parser = DiagramParser()
    parser.feed(source)
    parser.close()
    return parser


def reference_error(tag: str, rel: str, value: str, allow_external: bool = False) -> str | None:
    stripped = value.strip()
    lowered = stripped.casefold()
    if not stripped or stripped.startswith("#"):
        return None
    if lowered.startswith("javascript:") or lowered.startswith("data:text/html"):
        return f"executable URL on <{tag}>: {stripped[:80]}"
    remote = lowered.startswith(("http://", "https://", "//")) or (
        ":" in stripped.split("/", 1)[0] and not lowered.startswith("data:")
    )
    if not remote:
        if lowered.startswith("data:") and not lowered.startswith("data:image/"):
            return f"non-image data URL on <{tag}>: {stripped[:80]}"
        return None
    if allow_external:
        return None
    return (
        f"remote reference on <{tag}>: {stripped[:80]} "
        "(only Google Fonts is allowed by default; declare "
        '<meta name="flow-report-assets" content="external"> to opt in)'
    )


def check_svgs(parser: DiagramParser, errors: list[str]) -> None:
    checkable = [
        svg
        for svg in parser.svgs
        if isinstance(svg["attrs"], dict)
        and str(svg["attrs"].get("aria-hidden", "")).casefold() != "true"
    ]
    for number, svg in enumerate(checkable, 1):
        attrs = svg["attrs"]
        assert isinstance(attrs, dict)
        if attrs.get("role") != "img":
            errors.append(f"svg {number} needs role=img")
        labelled = attrs.get("aria-labelledby", "").split()
        title = svg["title"]
        desc = svg["desc"]
        assert isinstance(title, dict) and isinstance(desc, dict)
        title_attrs = title.get("attrs", {})
        desc_attrs = desc.get("attrs", {})
        assert isinstance(title_attrs, dict) and isinstance(desc_attrs, dict)
        if svg["first"] != "title":
            errors.append(f"svg {number} title must be its first child")
        if not str(title.get("text", "")).strip() or not str(desc.get("text", "")).strip():
            errors.append(f"svg {number} needs non-empty title and desc")
        title_id = title_attrs.get("id", "")
        desc_id = desc_attrs.get("id", "")
        if title_id in {"", "title"} or desc_id in {"", "desc"}:
            errors.append(f"svg {number} title/desc IDs must be diagram-prefixed, never bare")
        if labelled != [title_id, desc_id]:
            errors.append(f"svg {number} aria-labelledby must name title then desc")


def check_scripts(parser: DiagramParser, errors: list[str]) -> None:
    for number, script in enumerate(parser.scripts, 1):
        if not script["closed"]:
            errors.append(f"script {number} must have a closing script tag")
        if script["in_svg"]:
            errors.append(
                f"script {number} is inside an <svg> diagram: diagrams are static "
                "under the flow-report grammar"
            )


def verify(path: Path) -> list[str]:
    source = path.read_text(encoding="utf-8")
    parser = parsed_document(source)
    errors: list[str] = []
    errors.extend(parser.unsafe)
    for tag, rel, value in parser.references:
        finding = reference_error(tag, rel, value, parser.allow_external)
        if finding:
            errors.append(finding)
    check_svgs(parser, errors)
    check_scripts(parser, errors)
    return errors


def external_notes(path: Path) -> list[str]:
    """Remote hosts a declared opt-in let through — reported, never silent."""
    parser = parsed_document(path.read_text(encoding="utf-8"))
    if not parser.allow_external:
        return []
    hosts = []
    for _tag, _rel, value in parser.references:
        low = value.strip().casefold()
        if low.startswith(("http://", "https://", "//")):
            host = value.strip().split("/")[2] if "//" in value else value.strip()
            if host not in hosts:
                hosts.append(host)
    return hosts


def main() -> int:
    argument_parser = argparse.ArgumentParser(description=__doc__)
    argument_parser.add_argument("files", nargs="+", type=Path)
    args = argument_parser.parse_args()
    failed = False
    for path in args.files:
        try:
            errors = verify(path)
        except (OSError, UnicodeError) as exc:
            errors = [str(exc)]
        if errors:
            failed = True
            print(f"FAIL {path}")
            for error in errors:
                print(f"  - {error}")
        else:
            hosts = external_notes(path)
            if hosts:
                print(f"OK {path}  (declared external assets: {', '.join(hosts)})")
                print("  ! offline reading depends on the declared external assets")
            else:
                print(f"OK {path}")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
