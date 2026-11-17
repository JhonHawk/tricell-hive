## MODIFIED Requirements

### Requirement: Kernel references validated before installation
The manager SHALL validate the global core's packaged instruction dependencies, including Markdown heading anchors (`#fragment`) on relative, in-file and `skill:` links across the global core, roles and skills, before preparing or applying installation writes, including saved plans with internally recomputed hashes. It SHALL retain the existing exclusions for fenced examples and external URLs.

#### Scenario: Missing global core dependency
- **WHEN** the global core links a skill resource absent from the release payload
- **THEN** preparing and applying the installation reject the payload before writing native destinations.

#### Scenario: Broken heading anchor in an instruction link
- **WHEN** the global core, a role or a skill links a Markdown file with a `#fragment` that matches no ATX heading slug in the target file
- **THEN** preparing and applying the installation reject the payload before writing native destinations.
