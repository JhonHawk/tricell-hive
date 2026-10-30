---
name: flow-close
description: Close a work item and its session hygiene. Use when a work item reaches its completion point, is abandoned or is paused; when the user ends the session; or when the user asks for hygiene or cleanup, such as "haz higiene" or "clean up". Closes the change record, cleans merged branches, task temporaries and processes, and reports the outcome.
---

# Close a work item and the session

1. When the work is integrated into the base branch, move its ticket to the state the project uses for that.
2. Move an integrated change folder to `<Specs>/changes/archive/YYYY-MM-DD-<change-id>/`, and deliver that move as the delivery says.
3. When the work folder holds a `walkthrough.json`, mark it closed first with `python3 <flow-walkthrough dir>/scripts/render_walkthrough.py <work folder> --session-status closed --confirm-viewer`; when it exits 2, keep that folder. Clean up: delete the merged work branches and their clean worktrees, stop the processes this work started, and remove its reproducible temporary files. Keep anything with uncommitted work or uncertain ownership.
4. Report what was integrated and where, the ticket's state, and anything still pending with who owns it.
5. Write the outcome to persistent memory.
6. Ask the user whether to continue with other work or end the session.
