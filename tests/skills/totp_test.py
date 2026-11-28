import base64
import contextlib
import importlib.util
import io
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest import mock
import urllib.parse


ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "content/skills/hive-core/scripts/test-accounts/totp.py"
sys.dont_write_bytecode = True  # keep __pycache__ out of the distributed skill folder

# RFC 6238 appendix B seeds and their 8-digit codes.
SHA1_SEED = b"12345678901234567890"
SHA256_SEED = b"12345678901234567890123456789012"
SHA512_SEED = b"1234567890123456789012345678901234567890123456789012345678901234"
RFC_VECTORS = [
    (59, "SHA1", SHA1_SEED, "94287082"),
    (59, "SHA256", SHA256_SEED, "46119246"),
    (59, "SHA512", SHA512_SEED, "90693936"),
    (1111111109, "SHA1", SHA1_SEED, "07081804"),
    (1111111109, "SHA256", SHA256_SEED, "68084774"),
    (1111111109, "SHA512", SHA512_SEED, "25091201"),
    (20000000000, "SHA1", SHA1_SEED, "65353130"),
]


def b32(seed):
    return base64.b32encode(seed).decode()


def load_script():
    spec = importlib.util.spec_from_file_location("totp_under_test", SCRIPT)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class TotpTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.module = load_script()

    def tearDown(self):
        self.tmp.cleanup()

    def key_file(self, text):
        path = Path(self.tmp.name) / "key"
        path.write_text(text, encoding="utf-8")
        return str(path)

    def run_at(self, at, key_text):
        out, err = io.StringIO(), io.StringIO()
        with mock.patch("time.time", return_value=at), contextlib.redirect_stdout(out), contextlib.redirect_stderr(err):
            status = self.module.main([self.key_file(key_text)])
        return status, out.getvalue(), err.getvalue()

    def test_rfc_6238_vectors_through_an_otpauth_uri(self):
        for at, algorithm, seed, code in RFC_VECTORS:
            with self.subTest(at=at, algorithm=algorithm):
                query = urllib.parse.urlencode({"secret": b32(seed), "digits": 8, "algorithm": algorithm, "period": 30})
                uri = f"otpauth://totp/Example:ana?{query}"
                status, out, err = self.run_at(at, uri)
                self.assertEqual(status, 0, err)
                self.assertEqual(out, code + "\n")

    def test_bare_key_gives_six_digits_with_sha1(self):
        status, out, _ = self.run_at(59, b32(SHA1_SEED))
        self.assertEqual((status, out), (0, "287082\n"))

    def test_key_as_an_authenticator_shows_it(self):
        groups = [b32(SHA1_SEED).lower()[i : i + 4] for i in range(0, 32, 4)]
        for grouped in [" ".join(groups) + "\n", "-".join(groups)]:
            with self.subTest(grouped=grouped):
                status, out, _ = self.run_at(59, grouped)
                self.assertEqual((status, out), (0, "287082\n"))

    def test_key_without_its_padding(self):
        seed = b"1234567890123"
        padded = b32(seed)
        self.assertIn("=", padded)
        with_padding = self.run_at(59, padded)
        without_padding = self.run_at(59, padded.rstrip("="))
        self.assertEqual(with_padding[0], 0, with_padding[2])
        self.assertEqual(without_padding, with_padding)

    def test_text_that_is_not_a_key_fails_without_echoing_it(self):
        hotp = "otpauth://hotp/x?" + urllib.parse.urlencode({"secret": b32(SHA1_SEED), "counter": 1})
        for text in ["pbpaste | python3 totp.py", "", "otpauth://totp/x?issuer=Example", "ABCDEFGH1", hotp]:
            with self.subTest(text=text):
                status, out, err = self.run_at(59, text)
                self.assertEqual(status, 2)
                self.assertEqual(out, "")
                if text.strip():
                    self.assertNotIn(text.strip(), err)

    def test_command_reads_the_key_file_or_stdin(self):
        path = self.key_file(b32(SHA1_SEED))
        for args, stdin in [([path], ""), (["-"], b32(SHA1_SEED))]:
            with self.subTest(args=args):
                result = subprocess.run(
                    [sys.executable, "-I", str(SCRIPT), *args], input=stdin, capture_output=True, text=True, timeout=60
                )
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertRegex(result.stdout, r"^\d{6}\n$")


if __name__ == "__main__":
    unittest.main()
