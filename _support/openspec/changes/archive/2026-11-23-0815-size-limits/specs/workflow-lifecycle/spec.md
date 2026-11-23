## MODIFIED Requirements

### Requirement: The two reporting homes have bounded size and reading cost

`report-endings` and `close-report` SHALL stay within the 24,000-byte whole-read limit every rule leaf has, with no limit of their own, and no text SHALL be cut to fit. The reading-footprint test for the plan and build route SHALL accept the growth that comes from reading those two leaves whole at each stage's end, as a second allowance of 9,230 bytes beside the earlier 182 bytes. The global guidance SHALL stay at or under 30,000 bytes.

#### Scenario: The leaves fit their limits

- **GIVEN** the candidate
- **WHEN** the content tests run
- **THEN** `report-endings` and `close-report` are each at most 24,000 bytes, and the global guidance at most 30,000 bytes
