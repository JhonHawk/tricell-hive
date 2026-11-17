import datetime as dt
import importlib.util
import json
import pathlib
import unittest

SCRIPT = pathlib.Path(__file__).resolve().parents[2] / "content/skills/flow-build/scripts/check_package.py"


def load_module():
    spec = importlib.util.spec_from_file_location("check_package", SCRIPT)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {SCRIPT}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class CheckPackageTests(unittest.TestCase):
    def setUp(self):
        self.mod = load_module()
        self.now = dt.datetime(2026, 11, 16, 12, 0, 0, tzinfo=dt.timezone.utc)

    def test_npm_package_older_than_7_days_without_advisories_passes(self):
        def fake_fetch(url, method="GET", body=None):
            if "registry.npmjs.org/typescript" in url:
                return {
                    "dist-tags": {"latest": "5.6.3"},
                    "time": {"5.6.3": "2026-10-01T00:00:00.000Z"},
                    "versions": {"5.6.3": {"_npmUser": {"name": "tsbot"}}},
                }
            if "api.osv.dev" in url:
                return {"vulns": []}
            raise AssertionError(f"unexpected url {url}")

        res = self.mod.evaluate_package("npm", "typescript", min_days=7, now=self.now, fetcher=fake_fetch)
        self.assertTrue(res["ok"])
        self.assertEqual(res["version"], "5.6.3")
        self.assertEqual(res["publisher"], "tsbot")
        self.assertGreaterEqual(res["age_days"], 46.0)
        self.assertEqual(res["advisories"], [])

    def test_npm_package_published_less_than_7_days_ago_fails(self):
        def fake_fetch(url, method="GET", body=None):
            if "registry.npmjs.org/fresh-pkg" in url:
                return {
                    "dist-tags": {"latest": "1.0.1"},
                    "time": {"1.0.1": "2026-11-14T12:00:00.000Z"},
                    "versions": {"1.0.1": {"_npmUser": {"name": "alice"}}},
                }
            if "api.osv.dev" in url:
                return {"vulns": []}
            raise AssertionError(f"unexpected url {url}")

        res = self.mod.evaluate_package("npm", "fresh-pkg@1.0.1", min_days=7, now=self.now, fetcher=fake_fetch)
        self.assertFalse(res["ok"])
        self.assertAlmostEqual(res["age_days"], 2.0, places=2)
        self.assertTrue(any("7 days" in r for r in res["reasons"]))

    def test_pypi_package_with_active_advisory_fails(self):
        def fake_fetch(url, method="GET", body=None):
            if "pypi.org/pypi/badpkg/json" in url:
                return {
                    "info": {"version": "2.0.0", "author": "maintainer"},
                    "releases": {"2.0.0": [{"upload_time_iso_8601": "2026-01-01T00:00:00Z"}]},
                }
            if "api.osv.dev" in url:
                return {"vulns": [{"id": "GHSA-1234-5678-90ab", "summary": "RCE in badpkg"}]}
            raise AssertionError(f"unexpected url {url}")

        res = self.mod.evaluate_package("pypi", "badpkg@2.0.0", min_days=7, now=self.now, fetcher=fake_fetch)
        self.assertFalse(res["ok"])
        self.assertEqual(len(res["advisories"]), 1)
        self.assertEqual(res["advisories"][0]["id"], "GHSA-1234-5678-90ab")

    def test_go_and_cargo_ecosystems_resolve_publish_time(self):
        def fake_fetch(url, method="GET", body=None):
            if "proxy.golang.org" in url and url.endswith("/@latest"):
                return {"Version": "v1.2.0", "Time": "2026-09-01T00:00:00Z"}
            if "crates.io/api/v1/crates/serde" in url:
                return {
                    "crate": {"max_stable_version": "1.0.210"},
                    "versions": [
                        {
                            "num": "1.0.210",
                            "created_at": "2026-09-10T00:00:00Z",
                            "published_by": {"login": "dtolnay"},
                        }
                    ],
                }
            if "api.osv.dev" in url:
                return {"vulns": []}
            raise AssertionError(f"unexpected url {url}")

        go_res = self.mod.evaluate_package("go", "golang.org/x/sync", min_days=7, now=self.now, fetcher=fake_fetch)
        self.assertTrue(go_res["ok"])
        self.assertEqual(go_res["version"], "v1.2.0")

        cargo_res = self.mod.evaluate_package("cargo", "serde", min_days=7, now=self.now, fetcher=fake_fetch)
        self.assertTrue(cargo_res["ok"])
        self.assertEqual(cargo_res["version"], "1.0.210")
        self.assertEqual(cargo_res["publisher"], "dtolnay")


if __name__ == "__main__":
    unittest.main()
