# Run a check with its output saved

Read from [checks](../checks.md) before you run a check whose output you save to a file or filter other than through a plain pipe, such as a test suite, a typecheck or a local CI job.

```sh
python3 -I <hive-core-dir>/scripts/checks/run_check.py [--tail N | --grep REGEX] [--max-bytes N] -- <command> [args…]
```

- The script runs the command after `--` without a shell; for one that needs a shell, such as `cd` or `&&`, pass `bash -c '<script>'`. One check per call; run independent checks as separate calls, side by side when the host allows it.
- Its first line is `log: <path>`, a new `hive-check-*.log` in the host's temporary location holding the whole output, stdout and stderr together. Read an unclear result from it. At the end it prints the last 40 lines (`--tail`) or up to 40 lines that match `--grep`, at most 16 KiB.
- Its exit status is the check's own, except 127 (the command could not start), 128+n (signal n ended it) and 125 (the output passed `--max-bytes`, 64 MiB by default, and the script stopped the command's process group, naming the log and the cap on stderr).

## Never write your own wrapper

Never redirect a block's output to a file and read that file in the same command, in a `trap` or inside the block, as in `trap 'tail "$log"' EXIT; { checks; } >"$log" 2>&1`. When `set -e` stops the block, the trap still writes into the redirected file and copies the log into itself until the disk fills.

## Without python3

Run the check with its output to a new file in the host's temporary location in one command, read the file's end in a later command, and keep the check's status from the first.
