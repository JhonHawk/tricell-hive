#!/usr/bin/env python3
"""WCAG contrast checker for starlight-docs-site audits.

Computes the WCAG 2.1 contrast ratio between two CSS colors (any syntax
coloraide parses: oklch(), hex, rgb(), named). Converts through sRGB, so OKLCH
theme values are handled directly.

Usage (no system install — coloraide is fetched by uv per-run):
    uv run --with coloraide python scripts/contrast-check.py <foreground> <background>

Example:
    uv run --with coloraide python scripts/contrast-check.py \\
        "oklch(70% 0.16 260)" "oklch(12% 0 0)"

Exit code is 0 when the pair meets AA for normal text (>= 4.5:1), else 1 — so
the audit can branch on it. The ratio and per-level verdicts are always printed.
"""

import sys

# WCAG 2.1 thresholds (contrast ratio).
AA_NORMAL = 4.5
AA_LARGE = 3.0
AAA_NORMAL = 7.0
AAA_LARGE = 4.5


def main(argv: list[str]) -> int:
    if len(argv) != 3:
        print(f"usage: {argv[0]} <foreground> <background>", file=sys.stderr)
        print('example: ... "oklch(70% 0.16 260)" "oklch(12% 0 0)"', file=sys.stderr)
        return 2

    from coloraide import Color

    fg_in, bg_in = argv[1], argv[2]
    try:
        fg = Color(fg_in)
        bg = Color(bg_in)
    except Exception as exc:  # noqa: BLE001 - surface any parse error to the caller
        print(f"error: could not parse a color ({exc})", file=sys.stderr)
        return 2

    # coloraide's contrast() defaults to the WCAG 2.1 algorithm.
    ratio = fg.contrast(bg)

    def verdict(level: float) -> str:
        return "PASS" if ratio >= level else "FAIL"

    print(f"foreground: {fg_in}  ->  {fg.convert('srgb').to_string(hex=True)}")
    print(f"background: {bg_in}  ->  {bg.convert('srgb').to_string(hex=True)}")
    print(f"contrast ratio: {ratio:.2f}:1")
    print(f"  AA  normal (>= {AA_NORMAL}): {verdict(AA_NORMAL)}")
    print(f"  AA  large  (>= {AA_LARGE}): {verdict(AA_LARGE)}")
    print(f"  AAA normal (>= {AAA_NORMAL}): {verdict(AAA_NORMAL)}")
    print(f"  AAA large  (>= {AAA_LARGE}): {verdict(AAA_LARGE)}")

    return 0 if ratio >= AA_NORMAL else 1


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
