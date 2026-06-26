# Requirements intake rubric

Scored 1-5 per dimension by `requirement-analyst`. Plain markdown so any harness can
consume it. This rubric judges a RAW client document — it tolerates informality but not
silence: a dimension the document never touches scores 1, and that absence is the finding.

| # | Dimension | What a 5 looks like |
|---|---|---|
| 1 | Business problem | The pain and its cost are explicit — not just the requested feature |
| 2 | Actors & roles | Every user type named, with what each can do and see |
| 3 | Core flows | The main operations described end-to-end, even informally |
| 4 | Explicit scope | What's included is enumerable; a delivery could be checked against it |
| 5 | Implicit scope surfaced | Adjacent needs (admin, reports, notifications, exports) stated in or out |
| 6 | Business rules | Limits, validations, lifecycle states, visibility rules written down |
| 7 | Data & integrations | What data exists today, where it lives, what systems must connect |
| 8 | Volumes & growth | Order of magnitude of users/records/traffic — enough to size decisions |
| 9 | Constraints | Deadlines, budget signals, compliance, devices, languages |
| 10 | Success definition | The client said what "done and working" means to them |

## Question discipline

- Dimensions 1-6 scoring ≤2 normally produce `blocking` questions.
- Dimensions 7-10 scoring ≤2 normally produce `nice-to-know` questions with a proposed
  default — intake should not stall on them.
