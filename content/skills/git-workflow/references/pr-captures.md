# Pull request captures

Read at [git-workflow](../SKILL.md) step 6 when the verification walk produced captures, and again before editing the body of a pull request that holds them. The images go to the pull request, never to a commit, so they neither fill the repository nor stay behind in the work's folder once the merged body holds them.

## Where they may go

Check three facts before writing the body: `gh pr create --help` lists `--attach`; the remote's host, from `gh repo view --json url`, is github.com or a GitHub Enterprise Server host; and the repository's visibility, from `gh repo view --json visibility`.

- **A private or internal repository on github.com:** attach the captures. Only people with access to the repository can view an image uploaded there, so the delivery answer that opens the pull request covers it.
- **A public repository, or any repository on GitHub Enterprise Server:** an uploaded image is visible to anyone with its URL, and removing the reference does not delete it; only GitHub Support can. Attach only captures that show seeded data and test accounts, with no internal URL and no real email address, and only after the user authorized the public upload for this work. Ask that once per work, as a decision of the delivery round or of the plan's effects round, naming the repository, how many images and what they show.
- **Otherwise,** when `--attach` is missing, the host is another forge, the authorization is missing, or you cannot tell that a capture shows only test data: keep the body in its text form, add one line that names the captures as kept locally, and upload nothing.

## The table

Put one table per changed screen in the body, after the acceptance criteria. Each row is one region at one width, theme and state, as the before and after captures pair them in [UI verification](../../flow-verify/references/ui-verification.md#before-captures):

```markdown
| Region · width · theme · state | Before | After |
| --- | --- | --- |
| Filters · 1280 · light · empty | ![Filters before](images/before/filters-1280-light-empty.png) | ![Filters after](images/filters-1280-light-empty.png) |
| Export dialog · 440 · light · open | New screen | ![Export dialog after](images/export-440-light-open.png) |
```

When a side is missing, its cell says why: `New screen`, or `No before capture` with the reason, such as an edit made before the capture. When the before capture used other data than the verification, say so under the table.

## Upload and edit

- Write the body file in the work's `workspace/YYYY-MM-DD-<work>/` folder and reference each image by its path relative to that folder, since `gh` resolves a body file's paths from its own directory. Pass each referenced image with `--attach`, up to 50 per command: `gh pr create --body-file <file> --attach <path> …`. `gh` uploads each one and rewrites its reference to the uploaded URL.
- Never stage an image. `git diff --stat origin/<base>...HEAD` lists none before the push.
- When some uploads fail, `gh` still creates the pull request and exits non-zero. Then save the published body with `gh pr view <n> --json body --jq .body` into a file in the same folder, keep the local references of the images that failed in it, and run `gh pr edit <n> --body-file <file> --attach <path> …` with only those images.
- Every later edit of the body, such as the review result step 6 adds or a corrected capture, starts from the published body saved the same way, never from the first body file, whose local references would replace the uploaded URLs. To replace a row after a correction, put the new capture's local path in that row and pass it with `--attach`.

## Read back

After each create or edit, read the body with `gh pr view <n> --json body` and check that every image reference is a GitHub URL, such as `https://github.com/user-attachments/assets/…`, and that no local path remains. Record in the build's checkpoint which captures have a read-back URL: at the close, only those count as replaced by the pull request, as [artifact layout](../../flow-close/references/artifact-layout.md#retain-evidence) sets.
