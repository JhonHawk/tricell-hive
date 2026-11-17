#!/usr/bin/env python3
"""Verify that a package version is at least N days old and has no active OSV advisories.

Usage:
  python3 check_package.py --ecosystem <npm|pypi|go|cargo> <pkg>[@<version>] [--min-days 7] [--expected-publisher <name>]

Exit codes:
  0: package version is >= min-days old, matches expected publisher (if given), and has 0 advisories
  1: package version is younger than min-days, has active advisories, or mismatches publisher
  2: invalid arguments or registry/OSV lookup error
"""

import argparse
import datetime as dt
import json
import sys
import urllib.error
import urllib.parse
import urllib.request

OSV_ECOSYSTEMS = {
    "npm": "npm",
    "pypi": "PyPI",
    "go": "Go",
    "cargo": "crates.io",
}


def split_spec(ecosystem, spec):
    spec = spec.strip()
    if not spec:
        raise ValueError("empty package specification")
    if ecosystem == "npm" and spec.startswith("@"):
        idx = spec.find("@", 1)
        if idx != -1:
            return spec[:idx], spec[idx + 1 :]
        return spec, ""
    if "==" in spec:
        name, ver = spec.split("==", 1)
        return name.strip(), ver.strip()
    if "@" in spec:
        name, ver = spec.rsplit("@", 1)
        return name.strip(), ver.strip()
    return spec, ""


def default_fetcher(url, method="GET", body=None):
    data = json.dumps(body).encode("utf-8") if body is not None else None
    headers = {
        "Accept": "application/json",
        "User-Agent": "tricell-hive-check-package/1.0",
    }
    if data is not None:
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    with urllib.request.urlopen(req, timeout=15) as resp:
        return json.loads(resp.read().decode("utf-8"))


def parse_iso_utc(raw):
    if not raw or not isinstance(raw, str):
        raise ValueError(f"invalid timestamp: {raw!r}")
    text = raw.strip()
    if text.endswith("Z"):
        text = text[:-1] + "+00:00"
    parsed = dt.datetime.fromisoformat(text)
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=dt.timezone.utc)
    return parsed.astimezone(dt.timezone.utc)


def escape_go_module(mod):
    out = []
    for ch in mod:
        if "A" <= ch <= "Z":
            out.append("!" + ch.lower())
        else:
            out.append(ch)
    return "".join(out)


def lookup_registry(ecosystem, name, version, fetcher):
    if ecosystem == "npm":
        encoded = urllib.parse.quote(name, safe="@/")
        doc = fetcher(f"https://registry.npmjs.org/{encoded}")
        resolved = version or doc.get("dist-tags", {}).get("latest", "")
        if not resolved:
            raise ValueError(f"could not resolve latest version for npm:{name}")
        times = doc.get("time", {})
        if resolved not in times:
            raise ValueError(f"version {resolved!r} not found in npm time map for {name}")
        ver_meta = doc.get("versions", {}).get(resolved, {})
        npm_user = ver_meta.get("_npmUser") or {}
        publisher = npm_user.get("name") or npm_user.get("email") or ""
        return resolved, times[resolved], publisher

    if ecosystem == "pypi":
        encoded = urllib.parse.quote(name, safe="")
        doc = fetcher(f"https://pypi.org/pypi/{encoded}/json")
        info = doc.get("info", {})
        resolved = version or info.get("version", "")
        if not resolved:
            raise ValueError(f"could not resolve latest version for pypi:{name}")
        files = doc.get("releases", {}).get(resolved) or []
        if not files:
            raise ValueError(f"no release files found for pypi:{name}@{resolved}")
        timestamps = [
            f.get("upload_time_iso_8601") or f.get("upload_time")
            for f in files
            if f.get("upload_time_iso_8601") or f.get("upload_time")
        ]
        if not timestamps:
            raise ValueError(f"no upload timestamp for pypi:{name}@{resolved}")
        publisher = info.get("author") or info.get("maintainer") or ""
        return resolved, min(timestamps), publisher

    if ecosystem == "go":
        esc = escape_go_module(name)
        if version:
            ver = version if version.startswith("v") else "v" + version
            doc = fetcher(f"https://proxy.golang.org/{esc}/@v/{urllib.parse.quote(ver, safe='')}.info")
        else:
            doc = fetcher(f"https://proxy.golang.org/{esc}/@latest")
        resolved = doc.get("Version", "")
        published = doc.get("Time", "")
        if not resolved or not published:
            raise ValueError(f"missing Version/Time from Go proxy for {name}")
        publisher = name.split("/")[0] if "/" in name else name
        return resolved, published, publisher

    if ecosystem == "cargo":
        encoded = urllib.parse.quote(name, safe="")
        doc = fetcher(f"https://crates.io/api/v1/crates/{encoded}")
        crate = doc.get("crate", {})
        resolved = version or crate.get("max_stable_version") or crate.get("max_version") or ""
        for v in doc.get("versions", []):
            if v.get("num") == resolved:
                pub = (v.get("published_by") or {}).get("login") or ""
                return resolved, v.get("created_at", ""), pub
        raise ValueError(f"version {resolved!r} not found in crates.io metadata for {name}")

    raise ValueError(f"unsupported ecosystem: {ecosystem}")


def lookup_osv(ecosystem, name, version, fetcher):
    osv_eco = OSV_ECOSYSTEMS[ecosystem]
    payload = {"package": {"name": name, "ecosystem": osv_eco}, "version": version}
    doc = fetcher("https://api.osv.dev/v1/query", method="POST", body=payload)
    vulns = []
    for item in doc.get("vulns") or []:
        vulns.append(
            {
                "id": item.get("id", ""),
                "summary": item.get("summary") or item.get("details", "")[:160],
            }
        )
    return vulns


def evaluate_package(ecosystem, spec, min_days=7, expected_publisher=None, now=None, fetcher=None):
    if ecosystem not in OSV_ECOSYSTEMS:
        raise ValueError(f"unsupported ecosystem {ecosystem!r}")
    fetcher = fetcher or default_fetcher
    now = now or dt.datetime.now(dt.timezone.utc)
    name, req_version = split_spec(ecosystem, spec)
    version, published_raw, publisher = lookup_registry(ecosystem, name, req_version, fetcher)
    published_dt = parse_iso_utc(published_raw)
    age_days = (now - published_dt).total_seconds() / 86400.0
    advisories = lookup_osv(ecosystem, name, version, fetcher)

    reasons = []
    if age_days < float(min_days):
        reasons.append(f"published {age_days:.2f} days ago (< {min_days} days minimum)")
    if expected_publisher and publisher.lower() != expected_publisher.strip().lower():
        reasons.append(f"publisher {publisher!r} does not match expected {expected_publisher!r}")
    if advisories:
        ids = ", ".join(a["id"] for a in advisories if a.get("id")) or f"{len(advisories)} advisory(s)"
        reasons.append(f"flagged by advisory: {ids}")

    return {
        "ecosystem": ecosystem,
        "package": name,
        "version": version,
        "published_at": published_dt.isoformat(),
        "age_days": round(age_days, 2),
        "min_days": min_days,
        "publisher": publisher,
        "advisories": advisories,
        "ok": len(reasons) == 0,
        "reasons": reasons,
    }


def main(argv=None):
    parser = argparse.ArgumentParser(description="Check package publish age (>= 7 days) and OSV advisories.")
    parser.add_argument("--ecosystem", required=True, choices=sorted(OSV_ECOSYSTEMS), help="Package registry ecosystem")
    parser.add_argument("--min-days", type=float, default=7.0, help="Minimum publish age in days (default: 7)")
    parser.add_argument("--expected-publisher", default=None, help="Optional expected publisher handle/name")
    parser.add_argument("spec", help="Package name and optional version, e.g. pkg or pkg@1.2.3")
    args = parser.parse_args(argv)

    try:
        result = evaluate_package(
            args.ecosystem,
            args.spec,
            min_days=args.min_days,
            expected_publisher=args.expected_publisher,
        )
    except (ValueError, urllib.error.URLError, TimeoutError, json.JSONDecodeError) as exc:
        print(json.dumps({"ok": False, "error": str(exc)}), file=sys.stderr)
        return 2

    print(json.dumps(result, indent=2))
    return 0 if result["ok"] else 1


if __name__ == "__main__":
    sys.exit(main())
