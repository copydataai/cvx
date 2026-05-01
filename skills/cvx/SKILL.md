---
name: cvx
description: Agent workflow for maintaining, tailoring, verifying, rendering, and diffing structured CV repositories without inventing facts.
---

# cvx Agent Workflow

Use this skill when working in a `cvx` repository or when the user asks to tailor, compress, verify, review, render, or compare a CV.

## Mission

Help the human maintain a professional CV from structured source data while preserving factual accuracy.

The CLI is a small deterministic kernel. The agent owns judgment; the CLI owns validation, normalization, rendering, schemas, and reports.

## Hard Rules

- Do not invent facts.
- Do not invent metrics.
- Do not alter employers, titles, dates, degrees, links, locations, awards, publications, or certifications unless the human provides the correction.
- Treat `cv.yaml` as source of truth.
- Treat `output/**` as generated.
- Prefer editing structured files over generated artifacts.
- If uncertain, mark the claim with `[Unverified]`, `[Missing]`, or `[Suggestion]`.

## Source Files

Read these before content changes:

1. `AGENTS.md`
2. `cv.yaml`
3. The target `variants/*.yaml`, if relevant
4. The relevant `prompts/*.md`

## Standard Workflow

For tailoring:

```bash
cvx lint --variant variants/<target>.yaml
cvx normalize --variant variants/<target>.yaml --output output/current.json
cvx render --variant variants/<target>.yaml
```

Or run the bundled workflow:

```bash
cvx check --variant variants/<target>.yaml
```

For comparing snapshots:

```bash
cvx diff --from output/previous.json --to output/current.json
```

For editor/schema support:

```bash
cvx schema
```

For task prompt text:

```bash
cvx prompt tailor
```

## Editing Policy

Allowed changes:

- grammar
- bullet clarity
- ordering
- concision
- section emphasis
- variant selection rules
- template presentation
- validation/reporting code

Bullet provenance may be added when the source is known:

```yaml
bullets:
  - text: Built scheduling workflows for operators.
    source: human
    verified: true
```

Forbidden changes without human-provided facts:

- new employers
- new titles
- new dates
- new degrees
- new metrics
- new funding/revenue claims
- new awards/publications/certifications
- new technologies not already supported by source data

## Reporting Format

```txt
Changed:
- ...

Did not change:
- Factual claims not supported by source data.
- Generated output as source of truth.

Validation:
- cvx lint: pass/fail
- cvx normalize: pass/fail
- cvx render: pass/fail
- cvx diff: pass/fail, if used

Notes:
- ...
```
