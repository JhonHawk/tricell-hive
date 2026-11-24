# Device tool

Read from [UI verification](../ui-verification.md) when a native or cross-platform mobile app is verified, captured or reviewed.

- The main thread picks the device: an emulator or simulator already running, from `adb devices -l` or `xcrun simctl list devices booted`; otherwise, when `adb devices -l` lists a physical Android device, it asks the user whether to use it; with neither, it reports the device check as not run. A physical iPhone is reported as not exercised, since `simctl` drives only simulators. It names the serial or the simulator's UDID, read from that listing, in the brief of each child that installs, runs or captures the app; a child uses only the device its brief names, and without one runs nothing on a device and reports the check as not run. Use no name that was not read from the listing, such as a hard-coded `emulator-5554` or `booted`.
- Pass `-s <serial>` to every `adb` command and `xcrun simctl` the UDID, on the same device any `mobile-mcp` or Maestro flow uses; with `adb`, `reverse tcp:<port> tcp:<port>` lets the app reach a server on this machine. Report iOS as not exercised when full Xcode is not installed.
- On a physical device, never uninstall an app, clear its data or install over a build signed with another key without asking the user, since each can delete data that cannot be restored. Restore every setting you change before you finish.
- Publish captures with `--viewport mobile`, and name with the evidence the device model, its Android or iOS version and the platforms not exercised.
