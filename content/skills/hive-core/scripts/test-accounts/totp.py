#!/usr/bin/env python3
"""Print the current one-time code (RFC 6238) of a test account's second factor.

Usage: totp.py <key-file | ->

The key file holds the base32 key as an authenticator app shows it (spaces,
lowercase and missing padding are accepted) or an otpauth:// URI. The key is
read from the file or stdin, never from an argument, and only the code is
printed.
"""

import base64
import hashlib
import hmac
import struct
import sys
import time
import urllib.parse

ALGORITHMS = {"SHA1": hashlib.sha1, "SHA256": hashlib.sha256, "SHA512": hashlib.sha512}
INVALID = "the input is not a base32 key or an otpauth:// URI with a secret"


def parse(text):
    text = text.strip()
    digits, period, algorithm = 6, 30, "SHA1"
    if text.lower().startswith("otpauth://"):
        query = urllib.parse.parse_qs(urllib.parse.urlsplit(text).query)
        text = query.get("secret", [""])[0]
        digits = int(query.get("digits", ["6"])[0])
        period = int(query.get("period", ["30"])[0])
        algorithm = query.get("algorithm", ["SHA1"])[0].upper()
    key = "".join(text.split()).upper().rstrip("=")
    if not key or algorithm not in ALGORITHMS or not 6 <= digits <= 10 or period <= 0:
        raise ValueError(INVALID)
    key_bytes = base64.b32decode(key + "=" * (-len(key) % 8))
    return key_bytes, digits, period, ALGORITHMS[algorithm]


def code(key_bytes, at, digits, period, digest):
    mac = hmac.new(key_bytes, struct.pack(">Q", int(at) // period), digest).digest()
    offset = mac[-1] & 0x0F
    value = struct.unpack(">I", mac[offset : offset + 4])[0] & 0x7FFFFFFF
    return str(value % 10**digits).zfill(digits)


def main(argv):
    if len(argv) != 1:
        print("usage: totp.py <key-file | ->", file=sys.stderr)
        return 2
    try:
        if argv[0] == "-":
            text = sys.stdin.read()
        else:
            with open(argv[0], encoding="utf-8") as handle:
                text = handle.read()
        key_bytes, digits, period, digest = parse(text)
    except OSError as error:
        print(f"cannot read the key file: {error.strerror}", file=sys.stderr)
        return 2
    except ValueError:
        print(INVALID, file=sys.stderr)
        return 2
    print(code(key_bytes, time.time(), digits, period, digest))
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
