# Device tool

Read from [UI verification](../ui-verification.md) when a native or cross-platform mobile app is verified, captured or reviewed.

- The main thread picks the device: an emulator or simulator already running, from `adb devices -l` or `xcrun simctl list devices booted`, and otherwise asks the user which connected device to use. It names that serial, or the simulator's UDID, in the brief of each child that installs, runs or captures the app. A child uses only the device its brief names, and without one runs nothing on a device and reports the check as not run. Never address a device by a fixed name such as `emulator-5554` or `booted`.
- Android: pass `-s <serial>` to every `adb` command: `install -r <debug apk>`, `shell am start -n <package>/<activity>`, `exec-out screencap -p > <file>.png` for a capture, and `reverse tcp:<port> tcp:<port>` so the app reaches a server on this machine. iOS: use `xcrun simctl install <udid> <app>`, `launch <udid> <bundle id>` and `io <udid> screenshot <file>.png` when full Xcode is installed; otherwise report iOS as not exercised. Use `mobile-mcp` or Maestro for taps and flows when they are installed.
- On a physical device, never uninstall an app, clear its data or install over a build signed with another key without asking the user, since each can delete data that cannot be restored. Restore every setting you change, such as dark mode or font size, before you finish.
- Publish captures with `--viewport mobile`, and name with the evidence the device model, its Android or iOS version and the platforms not exercised.
