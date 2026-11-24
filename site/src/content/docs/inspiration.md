---
title: Inspiration
description: Public projects that inspire Hive's approach to deliberate work with coding agents.
---

Hive draws inspiration from these public projects:

## OpenSpec

[OpenSpec](https://github.com/Fission-AI/OpenSpec) provides an approach to
spec-driven development for AI coding assistants.

## gentle-ai

[gentle-ai](https://github.com/Gentleman-Programming/gentle-ai) configures coding
agents with curated skills, memory, and development workflows.

## Superpowers

[Superpowers](https://github.com/obra/superpowers) brings a library of agent skills
and a software development methodology.

## Skills For Real Engineers

[Skills For Real Engineers](https://github.com/mattpocock/skills) offers agent
skills for engineering work organized around an issue tracker.

## Impeccable

[Impeccable](https://github.com/pbakaus/impeccable) by Paul Bakaus (Apache-2.0)
inspired the methodology of Hive's UI criteria skill: screen modes, product and
design records, accessibility minimums and refinement passes. No file or passage
is redistributed.

## improve

[improve](https://github.com/shadcn/improve) by shadcn (MIT) inspired the
method of Hive mode's code audit: categories audited in parallel, findings
checked at the code they cite, ranking by impact and effort, and a record of
rejected findings. No file or passage is redistributed.

## pstack

[pstack](https://github.com/cursor/plugins/tree/main/pstack) by Lauren Tan (MIT)
inspired these parts of Hive. No file or passage is redistributed; the text is
written in Hive's own words.

- The code principles: type system discipline, boundary discipline, model the
  domain, make operations idempotent, migrate callers then delete legacy APIs,
  minimize reader load, separate before serializing shared state, and explain
  the number. Each comes from pstack's principle skill of the same subject.
- The TypeScript rules, from its `typescript-best-practices` skill.
- The prose rules: pattern cutting and the layered technical-writing standard.
- The blast-radius check: the one fact the change is safe because of, proven by
  running real code.
- The comment pass: remove narrating comments, test each comment kept, and offer
  a check for a constraint.
- The adversarial review: adversaries that must prove each finding with a run,
  and a rerun before a finding is kept.
- The plan's delivery units, with `unit`, `live` and `perf` checks ticked from
  evidence, and its pause field, from its `multi-phase-plan` and `pause-safely`
  playbooks.
- The decision log of unattended delegation, from its `show-me-your-work` skill.

## DESIGN.md

The [DESIGN.md format](https://github.com/google-labs-code/design.md) from Google
Labs (Apache-2.0) is the format of the design record that the UI criteria skill
keeps for a project.

## Site foundations

The documentation site uses [Starlight](https://starlight.astro.build/) and the
[Ion theme](https://github.com/louisescher/starlight-ion-theme), with Hive's own
red accents, logos, and content.
