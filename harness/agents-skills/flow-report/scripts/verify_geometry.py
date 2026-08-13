#!/usr/bin/env python3
# Adapted diagram-design resources: ../references/license-diagram-design.md
"""Verify no diagram label mask is clipped by a node painted after it.

diagram-grammar.md §6 keeps an arrow label 6-10px clear of its connector, and
§5 fixes the paint order as background -> zones -> arrows -> labels -> nodes.
Nothing keeps a label mask off a *node*, so a label whose mask lands partly
inside a node rectangle that is painted later gets covered by the node fill:
the text renders as a fragment sitting on the node border.

Paint order is what makes this a defect rather than a stylistic choice:

* A mask overlapping a zone container is fine - zones are painted before labels,
  so the label stays on top. Zone eyebrows rely on this.
* A mask overlapping a node declared *later* in the document is clipped by that
  node. That is the failure this check reports.

Shape heuristics follow the shipped templates:

* A node is a `<rect>` at least 60x40 - large enough for a title and sublabel.
* A label mask is a `<rect>` 20-120 wide and 8-14 tall - the masking plate that
  diagram-grammar.md prescribes for arrow labels and zone eyebrows.
* A mask fully contained in a node is a badge chip (`EXT`, `EDGE`, `ORIG`) and
  is legal.

Usage:
    python3 scripts/verify_geometry.py my-report.html
    python3 scripts/verify_geometry.py --all

---
Adapted from https://github.com/cathrynlavery/diagram-design
(scripts/verify-geometry.py, v2.3.2, MIT) for the flow-report skill
(tricell-hive). Local changes:
  - Node/mask overlap checks are now scoped per `<svg>...</svg>` block
    instead of running across the whole document. A flow-report HTML file
    can embed several diagrams plus surrounding report markup (tables,
    prose, nav) in one file, where upstream always ships exactly one
    diagram per HTML file — without scoping, a label mask in one diagram
    could spuriously "overlap" a node in a different diagram that merely
    shares nearby coordinates in its own, unrelated viewBox.
  - `--all` still walks a shipped `assets/*.html` directory for parity with
    upstream's CLI shape, but flow-report ships no bundled example assets;
    pass explicit file paths instead.
  - Everything else (shape heuristics, paint-order logic, CLI exit code) is
    unchanged from upstream.
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
ASSET_DIR = ROOT / "skills/diagram-design/assets"

RECT_RE = re.compile(
    r"<rect\b[^>]*?"
    r'\bx="(?P<x>-?[\d.]+)"\s+'
    r'y="(?P<y>-?[\d.]+)"\s+'
    r'width="(?P<w>[\d.]+)"\s+'
    r'height="(?P<h>[\d.]+)"',
    re.IGNORECASE,
)

SVG_BLOCK_RE = re.compile(r"<svg\b.*?</svg>", re.IGNORECASE | re.DOTALL)

NODE_MIN_W = 60.0
NODE_MIN_H = 40.0
MASK_MIN_W = 20.0
MASK_MAX_W = 120.0
MASK_MIN_H = 8.0
MASK_MAX_H = 14.0
EPSILON = 0.5


class Rect:
    __slots__ = ("x", "y", "w", "h", "line", "offset")

    def __init__(self, x, y, w, h, line, offset) -> None:
        self.x, self.y, self.w, self.h = x, y, w, h
        self.line, self.offset = line, offset

    @property
    def right(self) -> float:
        return self.x + self.w

    @property
    def bottom(self) -> float:
        return self.y + self.h

    def __repr__(self) -> str:
        return f"({self.x:g},{self.y:g} {self.w:g}x{self.h:g})"


def parse_rects(source: str) -> list[Rect]:
    rects: list[Rect] = []
    for match in RECT_RE.finditer(source):
        rects.append(
            Rect(
                float(match.group("x")),
                float(match.group("y")),
                float(match.group("w")),
                float(match.group("h")),
                source.count("\n", 0, match.start()) + 1,
                match.start(),
            )
        )
    return rects


def svg_spans(source: str) -> list[tuple[int, int]]:
    """Character (start, end) spans of every `<svg>...</svg>` block.

    Used to scope overlap checks to a single diagram at a time - a
    flow-report HTML file can embed multiple `<svg>` diagrams side by side
    with report markup around them.
    """
    return [(match.start(), match.end()) for match in SVG_BLOCK_RE.finditer(source)]


def overlap(a: Rect, b: Rect) -> tuple[float, float]:
    return (
        min(a.right, b.right) - max(a.x, b.x),
        min(a.bottom, b.bottom) - max(a.y, b.y),
    )


def contained(inner: Rect, outer: Rect) -> bool:
    return (
        inner.x >= outer.x - EPSILON
        and inner.y >= outer.y - EPSILON
        and inner.right <= outer.right + EPSILON
        and inner.bottom <= outer.bottom + EPSILON
    )


def check(path: Path) -> list[str]:
    source = path.read_text(encoding="utf-8")
    rects = parse_rects(source)

    findings: list[str] = []
    for start, end in svg_spans(source):
        scoped = [r for r in rects if start <= r.offset < end]
        nodes = [r for r in scoped if r.w >= NODE_MIN_W and r.h >= NODE_MIN_H]
        masks = [
            r
            for r in scoped
            if MASK_MIN_W <= r.w <= MASK_MAX_W and MASK_MIN_H <= r.h <= MASK_MAX_H
        ]

        for mask in masks:
            for node in nodes:
                if node.offset <= mask.offset:
                    continue  # painted before the label; the label stays on top
                dx, dy = overlap(mask, node)
                if dx <= 1.0 or dy <= 1.0 or contained(mask, node):
                    continue
                findings.append(
                    f"{path.name}:{mask.line}: label mask {mask} is clipped by node "
                    f"{node} declared later at line {node.line} (overlap {dx:g}x{dy:g}px)"
                    f" - move the label onto a free segment of its connector"
                )
                break
    return findings


def targets(args: argparse.Namespace) -> list[Path]:
    if args.all:
        return sorted(ASSET_DIR.glob("*.html"))
    return [Path(p) for p in args.files]


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("files", nargs="*", help="HTML diagrams to check")
    parser.add_argument("--all", action="store_true", help="check every shipped asset")
    args = parser.parse_args()

    paths = targets(args)
    if not paths:
        parser.error("pass one or more files, or --all")

    findings: list[str] = []
    for path in paths:
        if not path.exists():
            findings.append(f"{path}: file not found")
            continue
        findings.extend(check(path))

    for finding in findings:
        print(finding)
    print(f"Summary: {len(paths)} file(s) checked, {len(findings)} finding(s).")
    return 1 if findings else 0


if __name__ == "__main__":
    sys.exit(main())
