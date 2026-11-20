# Run a check with its output saved

Read from [checks](../SKILL.md) before you run a check whose output you save to a file or filter other than through a plain pipe, such as a test suite, a typecheck or a local CI job.

## Run it

```sh
python3 -I <hive-rule-checks-dir>/scripts/run_check.py [--tail N | --grep REGEX] [--max-bytes N] -- <command> [args…]
```

- Put the command and its arguments after `--`. The script runs them without a shell. For a command that needs one, such as `cd` or `&&`, pass `bash -c '<script>'` as the command.
- The first line it prints is `log: <path>`, a new `hive-check-*.log` in the host's temporary location. The log holds the whole output, stdout and stderr together. Read an unclear result from it.
- When the command ends, the script closes the log and prints its last 40 lines (`--tail`) or up to 40 lines that match `--grep`, at most 16 KiB.
- Its exit status is the check's own. 127 means the command could not start, and 128+n means signal n ended it. 125 means the output passed `--max-bytes` (64 MiB by default) and the script stopped the command's process group; a line on stderr then names the log and the cap.
- One check per call. Run independent checks as separate calls, side by side when the host allows it.

## Never write your own wrapper

Never redirect a block's output to a file and read that file in the same command, in a `trap` or inside the block, as in `trap 'tail "$log"' EXIT; { checks; } >"$log" 2>&1`. When `set -e` stops the block, the trap still writes into the redirected file and copies the log into itself until the disk fills.

## Without python3

When `python3` cannot run, run the check with its output to a new file in the host's temporary location in one command, and read the file's end in a later command. Keep the check's status from the first command.
