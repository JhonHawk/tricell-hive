#!/usr/bin/env python3
"""Initialize or structurally check a project verification index; never run it."""

import argparse
import json
from pathlib import Path
import re
import sys


CAPABILITIES = ("launch", "doctor", "drive", "evidence", "cleanup")
STATUSES = {"unverified", "verified", "unsupported", "not_applicable"}


def require(condition, message):
    if not condition:
        raise ValueError(message)


def fields(value, expected, label):
    require(isinstance(value, dict) and set(value) == set(expected),
            f"{label}: expected fields {', '.join(expected)}")


def text(value, label):
    require(isinstance(value, str) and bool(value.strip()), f"{label}: expected nonempty text")


def identifier(value, label):
    text(value, label)
    require(re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9_.-]*", value), f"{label}: invalid ID")


def entries(value, label):
    require(isinstance(value, list), f"{label}: expected array")
    return value


def strings(value, label):
    for item in entries(value, label):
        text(item, label)
    require(len(value) == len(set(value)), f"{label}: duplicate entry")


def confined(repo, value, label, *, cli=False, kind=None):
    text(value, label)
    path = Path(value)
    require(".." not in path.parts, f"{label}: parent traversal is forbidden")
    require(cli or not path.is_absolute(), f"{label}: use a repo-relative path")
    resolved = (repo / path).resolve()
    require(resolved.is_relative_to(repo), f"{label}: path escapes repository")
    if kind == "file":
        require(resolved.is_file(), f"{label}: referenced file does not exist")
    elif kind == "directory":
        require(resolved.is_dir(), f"{label}: directory does not exist")
    return resolved


def unique_id(item, seen, label):
    identifier(item, label)
    require(item not in seen, f"{label}: duplicate ID {item}")
    seen.add(item)


def validate(data, repo):
    fields(data, ("schema_version", "targets"), "contract")
    require(type(data["schema_version"]) is int and data["schema_version"] == 1,
            "schema_version: only version 1 is supported")
    targets = entries(data["targets"], "targets")
    require(targets, "targets: at least one target is required")
    target_ids = set()
    for target in targets:
        fields(target, ("id", "cwd", "capabilities", "features"), "target")
        unique_id(target["id"], target_ids, "target.id")
        confined(repo, target["cwd"], "target.cwd", kind="directory")
        capabilities = target["capabilities"]
        fields(capabilities, CAPABILITIES, "capabilities")
        reference_ids = set()
        references_by_capability = {}
        for name in CAPABILITIES:
            capability = capabilities[name]
            label = f"{target['id']}.{name}"
            fields(capability, ("status", "reason", "references", "requirements",
                                "effects", "success", "evidence"), label)
            status = capability["status"]
            require(isinstance(status, str) and status in STATUSES, f"{label}: unknown status")
            if capability["reason"] is not None:
                text(capability["reason"], f"{label}.reason")
            if status in ("unsupported", "not_applicable"):
                text(capability["reason"], f"{label}.reason")
            strings(capability["requirements"], f"{label}.requirements")
            strings(capability["effects"], f"{label}.effects")
            if capability["success"] is not None:
                text(capability["success"], f"{label}.success")
            references_by_capability[name] = set()
            for reference in entries(capability["references"], f"{label}.references"):
                fields(reference, ("id", "path", "command"), f"{label}.reference")
                unique_id(reference["id"], reference_ids, f"{label}.reference.id")
                references_by_capability[name].add(reference["id"])
                confined(repo, reference["path"], f"{label}.reference.path", kind="file")
                if reference["command"] is not None:
                    text(reference["command"], f"{label}.reference.command")
            evidence = capability["evidence"]
            if evidence is not None:
                fields(evidence, ("revision", "result", "path"), f"{label}.evidence")
                text(evidence["revision"], f"{label}.evidence.revision")
                require(evidence["result"] in ("pass", "failed", "blocked", "inconclusive"),
                        f"{label}.evidence.result: unknown result")
                confined(repo, evidence["path"], f"{label}.evidence.path", kind="file")
            if status == "verified":
                require(capability["references"], f"{label}: verified requires references")
                text(capability["success"], f"{label}.success")
                require(evidence is not None and evidence["result"] == "pass",
                        f"{label}: verified requires passing evidence with revision and path")
        feature_ids = set()
        for feature in entries(target["features"], "features"):
            fields(feature, ("id", "sources", "drive", "evidence"), "feature")
            unique_id(feature["id"], feature_ids, "feature.id")
            strings(feature["sources"], "feature.sources")
            require(feature["sources"], "feature.sources: at least one source is required")
            for source in feature["sources"]:
                confined(repo, source, "feature.source", kind="file")
            for name in ("drive", "evidence"):
                refs = feature[name]
                strings(refs, f"feature.{name}")
                require(set(refs) <= references_by_capability[name],
                        f"feature.{name}: unknown reference for this target and capability")
                if capabilities[name]["status"] == "verified":
                    require(refs, f"feature.{name}: verified capability requires a reference")


def no_duplicates(pairs):
    value = {}
    for key, item in pairs:
        require(key not in value, f"JSON: duplicate key {key}")
        value[key] = item
    return value


def load(path):
    return json.loads(path.read_text(encoding="utf-8"), object_pairs_hook=no_duplicates,
                      parse_constant=lambda _: (_ for _ in ()).throw(ValueError("JSON: invalid constant")))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("action", choices=("init", "check"))
    parser.add_argument("--repo", required=True, help="absolute repository root")
    parser.add_argument("--path", required=True, help="contract path, absolute or repo-relative")
    args = parser.parse_args()
    try:
        repo_arg = Path(args.repo)
        require(repo_arg.is_absolute(), "repo: use an absolute repository root")
        repo = repo_arg.resolve(strict=True)
        require(repo.is_dir(), "repo: expected directory")
        path = confined(repo, args.path, "contract path", cli=True)
        if args.action == "init":
            requested = repo / args.path
            require(not requested.exists() and not requested.is_symlink(),
                    "contract path: already exists; init never overwrites")
            template = Path(__file__).resolve().parents[1] / "assets/project-verification.json"
            data = load(template)
            validate(data, repo)
            path.parent.mkdir(parents=True, exist_ok=True)
            # Exclusive creation preserves an existing contract, including concurrent init.
            with path.open("x", encoding="utf-8") as stream:
                stream.write(json.dumps(data, indent=2) + "\n")
            print("Initialized unverified project contract; no commands executed.")
        else:
            validate(load(path), repo)
            print("Valid project contract (structure only); no capabilities executed or proven.")
        return 0
    except (OSError, ValueError, RecursionError) as error:
        # Do not echo input content, commands or arbitrary filesystem error strings.
        message = str(error) if isinstance(error, ValueError) and not isinstance(error, json.JSONDecodeError) else type(error).__name__
        print(f"Contract rejected: {message}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
