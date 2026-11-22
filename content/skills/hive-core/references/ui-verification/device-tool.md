# Device tool

Read from [UI verification](../ui-verification.md) when a native or cross-platform mobile app is verified, captured or reviewed.

- Use the device or emulator your brief assigns by its serial, preferring an emulator when one exists; never pick one by a fixed name such as `emulator-5554`. Without an assigned device, run nothing on one and report the check as not run.
- Android: pass `-s <serial>` to every `adb` command: `install -r <debug apk>`, `shell am start -n <package>/<activity>`, `exec-out screencap -p > <file>.png` for a capture, and `reverse tcp:<port> tcp:<port>` so the app reaches a server on this machine. iOS: use `xcrun simctl` on a booted simulator (`install`, `launch`, `io booted screenshot <file>.png`) when full Xcode is installed; otherwise report iOS as not exercised. Use `mobile-mcp` or Maestro for taps and flows when they are installed.
- On a physical device, never uninstall an app, clear its data or install over a build signed with another key without asking the user, since each can delete data that cannot be restored. Restore every setting you change, such as dark mode or font size, before you finish.
- Publish captures with `--viewport mobile`, and name with the evidence the device model, its Android or iOS version and the platforms not exercised.
