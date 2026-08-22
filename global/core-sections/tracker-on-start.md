---
order_claude: 141
order_agents: 53
targets: [claude, agents]
join: tight
---

- **The tracker moves when development starts — not when the ticket is read, not at close.** Where the project declares a tracker, the write happens at the crossing from analysis to execution: the first code change for that ticket. Reading, analyzing, estimating or planning it never moves it — those are evaluations, and a ticket marked in-progress for work that may not happen misreports state as badly as one left in `Todo`. **The observable signal is your own first local record of having started** — a plan going to `building`, the branch created, the first edit: whatever declares the work underway locally owes the same write to the tracker, in that moment. One write, no confirmation, outside the close batch; it records that work is underway, it changes nothing. Same the moment a ticket turns out blocked. A local artifact saying `building` while the board says `Todo` means one of the two is lying, and it is not the board people read. Batching, closes, comments and evidence stay in `memory-routing.md > Tracker sync` (router: `memory-policy`).
