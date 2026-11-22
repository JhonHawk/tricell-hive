## ADDED Requirements

### Requirement: A native app's screens are verified on a device or emulator

The UI verification of a native or cross-platform mobile app SHALL drive the device or emulator the brief assigns by its serial instead of a browser, SHALL capture its screen and publish those captures with the mobile viewport, and SHALL report the platforms it did not exercise. The accessibility audit of a native app SHALL be reported as not run.

#### Scenario: An Android screen change

- **GIVEN** a change to a screen of an Android app and a connected device assigned by serial
- **WHEN** the change is verified
- **THEN** the captures come from that device and are published with the mobile viewport, and the report names iOS as not exercised when no simulator ran
