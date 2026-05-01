# Tailor CV

Tailor the CV for the target variant without inventing facts.

Rules:
- Read `cv.yaml`, `AGENTS.md`, and the target `variants/*.yaml` first.
- Preserve employers, titles, dates, degrees, links, and locations.
- Do not add metrics unless they already exist in source data.
- Prefer reordering, pruning, and clarifying over adding new claims.
- Keep uncertain changes labeled.
- After edits, run lint, normalize, render, and report changes.
